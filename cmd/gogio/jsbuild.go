// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"fmt"
	"golang.org/x/tools/go/packages"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func buildJS(bi *buildInfo) error {
	out := *destPath
	if out == "" {
		out = bi.name
	}
	if err := os.MkdirAll(out, 0700); err != nil {
		return err
	}
	cmd := exec.Command(
		"go",
		"build",
		"-ldflags="+bi.ldflags,
		"-tags="+bi.tags,
		"-o", filepath.Join(out, "main.wasm"),
		bi.pkgPath,
	)
	cmd.Env = append(
		os.Environ(),
		"GOOS=js",
		"GOARCH=wasm",
	)
	_, err := runCmd(cmd)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(out, "index.html"), []byte(jsIndex), 0600); err != nil {
		return err
	}
	goroot, err := runCmd(exec.Command("go", "env", "GOROOT"))
	if err != nil {
		return err
	}
	wasmJS := filepath.Join(goroot, "misc", "wasm", "wasm_exec.js")
	if _, err := os.Stat(wasmJS); err != nil {
		return fmt.Errorf("failed to find $GOROOT/misc/wasm/wasm_exec.js driver: %v", err)
	}
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps,
		Env:  append(os.Environ(), "GOOS=js", "GOARCH=wasm"),
	}, bi.pkgPath)
	if err != nil {
		return err
	}
	extraJS, err := jsSearchExtraJS(pkgs[0], make(map[string]bool))
	if err != nil {
		return err
	}

	return crateJSFile(filepath.Join(out, "wasm.js"), append([]string{wasmJS}, extraJS...)...)
}

func jsSearchExtraJS(p *packages.Package, visited map[string]bool) (extraJS []string, err error) {
	if len(p.GoFiles) == 0 {
		return nil, nil
	}
	js, err := filepath.Glob(filepath.Join(filepath.Dir(p.GoFiles[0]), "*_js.js"))
	if err != nil {
		return nil, err
	}
	extraJS = append(extraJS, js...)
	for _, imp := range p.Imports {
		if !visited[imp.ID] {
			extra, err := jsSearchExtraJS(imp, visited)
			if err != nil {
				return nil, err
			}
			extraJS = append(extraJS, extra...)
			visited[imp.ID] = true
		}
	}
	return extraJS, nil
}

func crateJSFile(dst string, files ...string) (err error) {
	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := w.Close(); err != nil {
			err = cerr
		}
	}()

	for i := range files {
		r, err := os.Open(files[i])
		if err != nil {
			return err
		}
		if _, err = io.Copy(w, r); err != nil {
			r.Close()
			return err
		}
		if err := r.Close(); err != nil {
			return err
		}

		if i == 0 { // Append after the `wasm_exec.js`
			_, err := io.Copy(w, strings.NewReader(jsNewGo))
			if err != nil {
				return err
			}
		}
	}
	_, err = io.Copy(w, strings.NewReader(jsStreamWasm))
	return err
}

const (
	jsIndex = `<!doctype html>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, user-scalable=no">
		<meta name="mobile-web-app-capable" content="yes">
		<script src="wasm.js"></script>
		<style>
			body,pre { margin:0;padding:0; }
		</style>
	</head>
	<body>
	</body>
</html>`
	jsNewGo = `(() => {
    global.go = new Go();

    // Pick up argv from the argv query argument (if set).
    go.argv = go.argv.concat((new URLSearchParams(location.search).get("argv") ?? "").split(" "));
})();`
	jsStreamWasm = `(() => {
    if (!WebAssembly.instantiateStreaming) { // polyfill
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }
    WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
        go.run(result.instance);
    });
})();`
)
