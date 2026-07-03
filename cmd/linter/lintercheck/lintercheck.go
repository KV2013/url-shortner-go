package lintercheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "linter",
	Doc:  "checks for usage of built-in panic and calls to os.Exit/log.Fatal outside main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		isMainPkg := file.Name.Name == "main"
		var currentFunc string

		ast.Inspect(file, func(n ast.Node) bool {
			if fd, ok := n.(*ast.FuncDecl); ok {
				currentFunc = fd.Name.Name
				return true
			}

			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			switch fun := call.Fun.(type) {
			case *ast.Ident:
				if fun.Name == "panic" {
					pass.Reportf(call.Pos(), "use of built-in panic")
				}
			case *ast.SelectorExpr:
				ident, ok := fun.X.(*ast.Ident)
				if !ok {
					return true
				}
				pkgName := ident.Name
				funcName := fun.Sel.Name

				if pkgName == "os" && funcName == "Exit" {
					if !(isMainPkg && currentFunc == "main") {
						pass.Reportf(call.Pos(), "call to os.Exit outside main function")
					}
				}
				if pkgName == "log" && (funcName == "Fatal" || funcName == "Fatalf" || funcName == "Fatalln") {
					if !(isMainPkg && currentFunc == "main") {
						pass.Reportf(call.Pos(), "call to log.%s outside main function", funcName)
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
