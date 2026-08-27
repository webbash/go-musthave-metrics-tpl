package main

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ErrOsExitAnalyzer checks for calling os.Exit in main functions
var ErrOsExitAnalyzer = &analysis.Analyzer{
	Name: "osExit",
	Doc:  "check for calling os.Exit in main functions",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	pkgPath := pass.Pkg.Path()
	if pass.Pkg.Name() != "main" || strings.HasSuffix(pkgPath, ".test") {
		return nil, nil
	}

	for _, file := range pass.Files {
		osPkgName := ""

		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil || path != "os" {
				continue
			}

			osPkgName = "os"
			if imp.Name != nil {
				osPkgName = imp.Name.Name
			}
		}

		if osPkgName == "" || osPkgName == "_" {
			continue
		}

		ast.Inspect(file, func(x ast.Node) bool {
			fn, ok := x.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" || fn.Body == nil {
				return true
			}

			ast.Inspect(fn.Body, func(x ast.Node) bool {
				call, ok := x.(*ast.CallExpr)
				if !ok {
					return true
				}

				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "Exit" {
					return true
				}

				pkg, ok := selector.X.(*ast.Ident)
				if !ok || pkg.Name != osPkgName {
					return true
				}

				pass.Reportf(call.Pos(), "os.Exit() call in main function is not allowed")
				return true // to not stop check on the first file
			})

			return true
		})
	}

	return nil, nil
}
