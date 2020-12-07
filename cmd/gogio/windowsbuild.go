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

	builder := &windowsBuilder{TempDir: tmpDir, BuildInfo: bi}
	builder.DestDir = *destPath
	if builder.DestDir == "" {
		builder.DestDir = bi.pkgPath
	}

	if err := builder.version(bi); err != nil {
		return err
	}
	if err := builder.name(bi); err != nil {
		return err
	}
	if err := builder.osVersion(bi); err != nil {
		return err
	}

	if err := builder.createIcon(); err != nil {
		return err
	}

	if err := builder.createManifest(); err != nil {
		return fmt.Errorf("can't create manifest: %v", err)
	}
	if err := builder.createResource(); err != nil {
		return fmt.Errorf("can't create resource: %v", err)
	}

	if err := builder.buildResource(); err != nil {
		return fmt.Errorf("can't build the syso: %s", err)
	}

	for _, arch := range builder.BuildInfo.archs {
		if err := builder.buildProgram(arch); err != nil {
			return fmt.Errorf("can't build the go program: %s", err)
		}
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
		Arch           string
	}
	windowsFiles struct {
		Resources     windowsResources
		ResourcesPath string
		Manifest      windowsManifest
	}
)

func (f *windowsFiles) version(b *buildInfo) error {
	if max := math.MaxUint16; b.version > max {
		return fmt.Errorf("version (%d) is larger than the maximum (%d)", b.version, max)
	}

	v := strconv.Itoa(b.version)
	f.Resources.Version = "0,0,0," + v
	f.Manifest.Version = "0.0.0." + v
	return nil
}

func (f *windowsFiles) name(b *buildInfo) error {
	name := b.name
	if *destPath != "" {
		if filepath.Ext(*destPath) != ".exe" {
			return fmt.Errorf("invalid filename got suffix (it must end with `.exe`) receives: %s", *destPath)
		}

		name = filepath.Base(*destPath)
		name = name[:strings.LastIndex(name, ".")]
	}
	if name == "" {
		return fmt.Errorf("invalid empty name")
	}
	f.Resources.Name = name
	f.Manifest.Name = name
	return nil
}

func (f *windowsFiles) osVersion(b *buildInfo) error {
	sdk := b.minsdk
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
	if _, err := os.Stat(b.BuildInfo.iconPath); err != nil {
		return nil
	}

	iconFile, err := os.Open(b.BuildInfo.iconPath)
	if err != nil {
		return fmt.Errorf("can't read the icon located at %s: %v", iconPath, err)
	}
	defer iconFile.Close()

	iconImage, err := png.Decode(iconFile)
	if err != nil {
		return fmt.Errorf("can't decode the PNG file (%s): %v", iconPath, err)
	}

	b.Resources.IconPath = filepath.Join(b.TempDir, "appicon.ico")
	exeIcon, err := os.Create(b.Resources.IconPath)
	if err != nil {
		return fmt.Errorf("impossibe to create icon file at %s: %v", b.Resources.IconPath, err)
	}
	defer exeIcon.Close()

	return convertPNGtoICO(exeIcon, iconImage)
}

func convertPNGtoICO(w io.Writer, img image.Image) error {
	// The file must be in .ICO format.
	bounds := img.Bounds()

	size := bounds.Dx()
	if sy := bounds.Dy(); sy < size {
		size = sy
	}
	// Max size of the .ICO is 256
	if size > 256 {
		size = 256
	}

	scaledImage := resizeIcon(iconVariant{size: size, fill: false}, img)

	var scaledIcon bytes.Buffer
	if err := png.Encode(&scaledIcon, scaledImage); err != nil {
		return fmt.Errorf("can't encode image: %v", err)
	}

	// ICONDIR structure
	binary.Write(w, binary.LittleEndian, [3]uint16{0, 1, 1})
	// ICONDIRENTRY 0-3 structure.
	// The width/height is defined from 0 to 255 (uint8). But "0" means 256px.
	binary.Write(w, binary.LittleEndian, [4]uint8{uint8(size % 256), uint8(size % 256), 0, 0})
	// ICONDIRENTRY 4-6 structure
	binary.Write(w, binary.LittleEndian, [2]uint16{1, 32})
	// ICONDIRENTRY 8-12 structure
	binary.Write(w, binary.LittleEndian, [2]uint32{uint32(scaledIcon.Len()), (2 * 3) + (4 * 1) + (2 * 2) + (4 * 2)})
	// Copy scaled image
	io.Copy(w, &scaledIcon)

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
	// @TODO remove windres dependency.
	_, err := exec.LookPath("windres")
	if err != nil {
		return false
	}
	return true
}

func (b *windowsBuilder) buildResource() error {
	cmd := exec.Command(
		"windres",
		b.ResourcesPath,
		filepath.Join(b.BuildInfo.pkgPath, "main_windows.syso"),
	)
	_, err := runCmd(cmd)
	return err
}

func (b *windowsBuilder) buildProgram(arch string) error {
	dest := b.DestDir
	if len(b.BuildInfo.archs) > 1 {
		dest = filepath.Join(filepath.Dir(b.DestDir), b.Resources.Name+"_"+arch+".exe")
	}

	cmd := exec.Command(
		"go",
		"build",
		"-ldflags= -H=windowsgui "+b.BuildInfo.ldflags,
		"-tags="+b.BuildInfo.tags,
		"-o", dest,
		b.BuildInfo.pkgPath,
	)
	cmd.Env = append(
		os.Environ(),
		"GOOS=windows",
		"GOARCH="+arch,
	)
	_, err := runCmd(cmd)
	return err
}

func (f *windowsManifest) encode(w io.Writer) error {
	t := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly manifestVersion="1.0" xmlns="urn:schemas-microsoft-com:asm.v1" xmlns:asmv3="urn:schemas-microsoft-com:asm.v3">
    <assemblyIdentity type="win32" name="{{.Name}}" version="{{.Version}}" />
    <description>{{.Name}}</description>
    <compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1">
        <application>
            {{if (le .WindowsVersion 10)}}<supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/>
{{end}}
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
{{if .CompanyName}}VALUE "CompanyName", "{{.CompanyName}}"{{end}}
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
