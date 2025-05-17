package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OsSignalAnalyzer = &analysis.Analyzer{
	Name: "ossignalcheck",
	Doc:  "checks for os.Exit calls from main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	function := func(f *ast.FuncDecl) {
		ast.Inspect(f, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.SelectorExpr:
				if i, ok := x.X.(*ast.Ident); ok && i.Name == "os" && x.Sel.Name == "Exit" {
					pass.Reportf(x.Pos(), "direct os.Exit call in main function of package main")
				}
			}
			return true
		})
	}
	if pass.Pkg.Name() == "main" {
		for _, pkg := range pass.Pkg.Imports() {
			if pkg.Name() == "os" {
				for _, file := range pass.Files {
					ast.Inspect(file, func(node ast.Node) bool {
						if x, ok := node.(*ast.FuncDecl); ok && x.Name.Name == "main" {
							function(x)
						}
						return true
					})
				}
				break
			}
		}
	}
	return nil, nil
}
