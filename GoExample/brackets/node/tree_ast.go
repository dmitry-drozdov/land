package node

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// func init() {
// 	n, err := ParseAst("{ } f(x); { f(x); } f(x); { f(x); { f(x); } f(x); } f(x);")
// 	if err != nil {
// 		panic(err)
// 	}
// 	b, _ := json.Marshal(n)
// 	fmt.Println(string(b))
// }

func ParseAst(s string) (*Node, error) {
	node := &Node{
		Type:     "root",
		Children: nil,
	}

	s = fmt.Sprintf(`
		package main
		func main() {
			%s
		}
	`, s) // add main func to be able to use go parser

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "", s, 0)
	if err != nil {
		return nil, err
	}

	inspect(f, node)

	if len(node.Children) > 0 { // remove main func
		node.Children = node.Children[0].Children
	}

	node.processAny()
	return node, nil
}

func inspect(root ast.Node, node *Node) {
	if root == nil {
		return
	}
	ast.Inspect(root, func(n ast.Node) bool {
		switch y := n.(type) {
		case *ast.BlockStmt:
			n := &Node{Type: "block"}
			for _, yy := range y.List {
				inspect(yy, n)
			}
			node.Children = append(node.Children, n)
			return false
		case *ast.IfStmt:
			n := &Node{Type: "if"}
			n.Children = append(n.Children, &Node{Type: "any"}) // if cond
			inspect(y.Body, n)
			inspect(y.Else, n)
			node.Children = append(node.Children, n)
			return false
		case *ast.CallExpr:
			n := &Node{Type: "any"}
			node.Children = append(node.Children, n)
		}
		return true
	})
}
