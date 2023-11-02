package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "test.go", nil, 0)
	if err != nil {
		panic(err)
	}

	// Inspect the AST and print all identifiers and literals.
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		// case *ast.FuncDecl:
		// 	fmt.Println(x.Name.Name, x.Type.Params.List, x.Type.Results)
		case *ast.TypeSpec:
			typeSpec := n.(*ast.TypeSpec)
			switch typeSpec.Type.(type) {
			case *ast.StructType:
				//fmt.Println(x)
			case *ast.InterfaceType:
				//fmt.Println(x)
			default:
				fmt.Println(x)
			}
		}
		return true
	})

}
