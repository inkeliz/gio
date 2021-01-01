package main

import (
	"flag"
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"os"
	"strings"
)

func main() {
	flag.Parse()
	pkg := flag.Arg(0)
	if pkg == "" {
		panic("specify a package and the module")
	}

	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
		Env:  append(os.Environ(), "GOOS=js", "GOARCH=wasm"),
		Dir:  pkg,
	}

	pkgs, err := packages.Load(cfg, pkg)
	if err != nil {
		panic(err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	var Files []*ast.File

	for _, file := range pkgs[0].Syntax {
		for _, imp := range file.Imports {
			if strings.Trim(imp.Path.Value, `"`) == "syscall/js" {
				Files = append(Files, file)
			}
		}
	}

	for _, file := range Files {
		fmt.Println(file)
	}

	//fmt.Println(pkgs[0].Syntax[1].Imports[2].Path.Value)
}
