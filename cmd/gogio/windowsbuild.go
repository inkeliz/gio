package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

func buildWindows(tmpDir string, bi *buildInfo) error {
	if bi == nil {
		panic("invalid build info")
	}

	builder := &windowsBuilder{TempDir: tmpDir, BuildInfo: bi}
	builder.DestDir = *destPath
	if builder.DestDir == "" {
		builder.DestDir = bi.pkgPath
	}

	if err := builder.setVersion(bi); err != nil {
		return err
	}
	if err := builder.setName(bi); err != nil {
		return err
	}
	if err := builder.setWindowsVersion(bi); err != nil {
		return err
	}

	if ok := builder.isResourceCompilerAvailable(); !ok {
		fmt.Println("WARNING: Unable to find `windrest` or `rc.exe` in your current %PATH%. \n" +
			"Your file will be compiled, but some content (such as icons) may be missing.")

		// Skip process of generate icons, include resource and manifest.
		goto build
	}

	if err := builder.createIcon(); err != nil {
		return err
	}

	if err := builder.createManifest(); err != nil {
		return fmt.Errorf("can't create manifest: %s", err)
	}
	if err := builder.createResource(); err != nil {
		return fmt.Errorf("can't create resource: %s", err)
	}

	if err := builder.buildResource(); err != nil {
		return fmt.Errorf("can't build the syso: %s", err)
	}

build:
	if err := builder.buildProgram(); err != nil {
		return fmt.Errorf("can't build the go program: %s", err)
	}

	return nil
}

type (
	windowsResources struct {
		IconPath     string
		ManifestPath string
		Version      string
		Name         string
		CompanyName  string
	}
	windowsManifest struct {
		Version        string
		WindowsVersion int
		Name           string
	}
	windowsFiles struct {
		Resources     windowsResources
		ResourcesPath string
		Manifest      windowsManifest
	}
)

func (f *windowsFiles) setVersion(b *buildInfo) error {
	if b.version > math.MaxUint16 {
		return fmt.Errorf("max version (%d) reached", b.version)
	}

	v := strconv.FormatInt(int64(b.version), 10)
	f.Resources.Version = v + ",0,0,0"
	f.Manifest.Version = v + ".0.0.0"
	return nil
}

func (f *windowsFiles) setName(b *buildInfo) error {
	name := b.name
	if *destPath != "" {
		name = filepath.Base(*destPath)
		name = strings.Split(name, ".")[0]
	}
	if name == "" {
		return fmt.Errorf("invalid empty name")
	}
	f.Resources.Name = name
	f.Manifest.Name = name
	return nil
}

func (f *windowsFiles) setWindowsVersion(b *buildInfo) error {
	sdk := b.minsdk
	if sdk == defaultSDK {
		sdk = 0
	}
	if sdk > 10 {
		return fmt.Errorf("invalid minsdk (%d) it's higher than Windows 10", sdk)
	}
	f.Manifest.WindowsVersion = sdk
	return nil
}

type windowsBuilder struct {
	TempDir   string
	DestDir   string
	BuildInfo *buildInfo
	windowsFiles
}

func (b *windowsBuilder) createIcon() error {
	originalIconPath := *iconPath
	if originalIconPath == "" {
		originalIconPath = filepath.Join(b.BuildInfo.pkgDir, "appicon.png")
	}

	if _, err := os.Stat(originalIconPath); err != nil {
		return nil
	}

	originalIconFile, err := os.Open(originalIconPath)
	if err != nil {
		return fmt.Errorf("can't read the icon located at %s", originalIconPath)
	}
	defer originalIconFile.Close()

	originalIconImage, err := png.Decode(originalIconFile)
	if err != nil {
		return fmt.Errorf("can't decode the PNG file (%s)", originalIconPath)
	}

	b.Resources.IconPath = filepath.Join(b.TempDir, "appicon.ico")
	exeIcon, err := os.Create(b.Resources.IconPath)
	if err != nil {
		return fmt.Errorf("impossibe to create icon file at %s", b.Resources.IconPath)
	}
	defer exeIcon.Close()

	if err := convertPNGtoICO(exeIcon, originalIconImage); err != nil {
		return err
	}

	return nil
}

func convertPNGtoICO(w io.Writer, img image.Image) error {
	// The file must be in .ICO format.

	size := uint8(255)
	if img.Bounds().Dy() <= math.MaxUint8 || img.Bounds().Dx() <= math.MaxUint8 {
		// We need to scale anyway to remove any useless content from the PNG file
		// otherwise it might break the offset in ICONDIRENTRY. (:
		size = uint8(math.Min(float64(img.Bounds().Dy()), float64(img.Bounds().Dx())))
	}

	scaledImage := resizeIcon(iconVariant{size: int(size), fill: false}, img)

	scaledIcon := bytes.NewBuffer(nil)
	if err := png.Encode(scaledIcon, scaledImage); err != nil {
		return fmt.Errorf("can't encode image: %s", err)
	}

	var errs [5]error
	// ICONDIR structure
	errs[0] = binary.Write(w, binary.LittleEndian, [3]uint16{0, 1, 1})
	// ICONDIRENTRY 0-3 structure
	errs[1] = binary.Write(w, binary.LittleEndian, [4]uint8{size, size, 0, 0})
	// ICONDIRENTRY 4-6 structure
	errs[2] = binary.Write(w, binary.LittleEndian, [2]uint16{1, 32})
	// ICONDIRENTRY 8-12 structure
	errs[3] = binary.Write(w, binary.LittleEndian, [2]uint32{uint32(scaledIcon.Len()), (2 * 3) + (4 * 1) + (2 * 2) + (4 * 2)})
	// Copy scaled image
	_, errs[4] = io.Copy(w, scaledIcon)

	for _, err := range errs {
		if err != nil {
			return fmt.Errorf("can't write content/header of the ICO: %s", err)
		}
	}

	return nil
}

func (b *windowsBuilder) createManifest() error {
	// The manifest have some information about the executable itself,
	// such as the supported Windows and Execution Level/Permissions.
	b.Resources.ManifestPath = filepath.Join(b.TempDir, "manifest_windows.xml")
	manifest, err := os.Create(b.Resources.ManifestPath)
	if err != nil {
		return err
	}
	defer manifest.Close()

	return b.Manifest.encode(manifest)
}

func (b *windowsBuilder) createResource() error {
	// The resource includes the icon and manifest previously created
	// it also defines the version and some other information about the
	// program and the developer.
	b.ResourcesPath = filepath.Join(b.TempDir, "main_windows.rc")
	resources, err := os.Create(b.ResourcesPath)
	if err != nil {
		return err
	}
	defer resources.Close()

	return b.Resources.encode(resources)
}

func (b *windowsBuilder) isResourceCompilerAvailable() bool {
	// The windres is required to compile the resources for now.
	// It's included in MSYS/MINGW, and it's also available on Linux.
	// @TODO remove windres depedency.
	_, err := exec.LookPath("windres")
	if err != nil {
		return false
	}
	return true
}

func (b *windowsBuilder) buildResource() error {
	_, err := runCmd(exec.Command("windres", b.ResourcesPath, filepath.Join(filepath.Dir(b.DestDir), "main_windows.syso")))
	return err
}

func (b *windowsBuilder) buildProgram() error {
	cmd := exec.Command(
		"go",
		"build",
		"-ldflags= -s -w -H=windowsgui "+b.BuildInfo.ldflags,
		"-tags="+b.BuildInfo.tags,
		"-o", b.DestDir,
		b.BuildInfo.pkgPath,
	)
	cmd.Env = append(
		os.Environ(),
		"GOOS=windows",
	)
	_, err := runCmd(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (f *windowsManifest) encode(w io.Writer) error {
	t := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly manifestVersion="1.0" xmlns="urn:schemas-microsoft-com:asm.v1" xmlns:asmv3="urn:schemas-microsoft-com:asm.v3">
    <assemblyIdentity type="win32" name="{{.Name}}" version="{{.Version}}" processorArchitecture="x86"/>
    <description>{{.Name}}</description>
    <compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1">
        <application>
            {{if (le .WindowsVersion 10)}}<supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/>
{{end	}}
            {{if (le .WindowsVersion 9)}}<supportedOS Id="{1f676c76-80e1-4239-95bb-83d0f6d0da78}"/>
{{end}}
            {{if (le .WindowsVersion 8)}}<supportedOS Id="{4a2f28e3-53b9-4441-ba9c-d69d4a4a6e38}"/>
{{end}}
            {{if (le .WindowsVersion 7)}}<supportedOS Id="{35138b9a-5d96-4fbd-8e2d-a2440225f93a}"/>
{{end}}
            {{if (le .WindowsVersion 6)}}<supportedOS Id="{e2011457-1546-43c5-a5fe-008deee3d3f0}"/>
{{end}}
        </application>
    </compatibility>
    <trustInfo xmlns="urn:schemas-microsoft-com:asm.v3">
        <security>
            <requestedPrivileges>
                <requestedExecutionLevel level="asInvoker" uiAccess="false" />   
            </requestedPrivileges>
        </security>
    </trustInfo>
	<asmv3:application>
		<asmv3:windowsSettings>
			<dpiAware xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">true</dpiAware>
		</asmv3:windowsSettings>
	</asmv3:application>
</assembly>`
	template, err := template.New("manifest").Parse(t)
	if err != nil {
		return err
	}

	return template.Execute(w, f)
}

func (f *windowsResources) encode(w io.Writer) error {
	const t = `{{if .IconPath}}#define IDI_ICON1 1
IDI_ICON1 ICON "{{escapePath .IconPath}}"{{end}}

#define IDI_MANIFEST 1
IDI_MANIFEST 24 "{{escapePath .ManifestPath}}"

#define IDI_VERSION 1
IDI_VERSION VERSIONINFO
FILEVERSION     {{.Version}}
PRODUCTVERSION  {{.Version}}
FILEFLAGSMASK   0X3FL
FILEFLAGS       0x0L
FILEOS          0X40004L
FILETYPE        0X1L
FILESUBTYPE     0x0L
BEGIN
    BLOCK "StringFileInfo"
    BEGIN
        BLOCK "04000400"
        BEGIN
            VALUE "ProductVersion", "{{.Version}}"
            VALUE "FileVersion", "{{.Version}}"
            VALUE "FileDescription", "{{.Name}}"
            VALUE "ProductName", "{{.Name}}"
{{if .CompanyName}}
		    VALUE "CompanyName", "{{.CompanyName}}"
{{end}}
        END
    END
    BLOCK "VarFileInfo"
    BEGIN
            VALUE "Translation", 0x0400, 0x0400
    END
END`
	template, err := template.New("rc").Funcs(template.FuncMap{"escapePath": func(s string) string {
		return strings.Replace(s, `\`, `\\`, -1)
	}}).Parse(t)
	if err != nil {
		return err
	}

	return template.Execute(w, f)
}
