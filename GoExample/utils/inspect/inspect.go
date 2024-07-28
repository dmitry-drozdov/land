package inspect

import (
	"fmt"
	"go/ast"
)

type Node struct {
	Type     string
	Name     string
	Children []*Node
}

func Inspect(x *ast.FuncDecl) *Node {
	res := &Node{
		Type: "func_body",
	}
	inspect(x.Body, res)
	return res
}

func inspect(body ast.Node, node *Node) {
	if body == nil {
		return
	}
	ast.Inspect(body, func(n ast.Node) bool {
		child := &Node{}
		switch x := n.(type) {
		case *ast.CallExpr:
			child.Type = "call"
			child.Name = getExprName(x.Fun)
			for _, arg := range x.Args {
				inspect(arg, child)
			}
		case *ast.IfStmt:
			child.Type = "if"
			inspect(x.Init, child)
			inspect(x.Body, child)
			inspect(x.Cond, child)
			inspect(x.Else, child)
		case *ast.ForStmt:
			child.Type = "for"
			inspect(x.Init, child)
			inspect(x.Body, child)
			inspect(x.Cond, child)
			inspect(x.Post, child)
		case *ast.SwitchStmt:
			child.Type = "switch"
			inspect(x.Init, child)
			inspect(x.Body, child)
			inspect(x.Tag, child)
		case *ast.SelectStmt:
			child.Type = "select"
			inspect(x.Body, child)
		default:
			return true // continue inspection
		}
		node.Children = append(node.Children, child)
		return false // interrupt inspection (will be inspected in recursion)
	})
}

func (n *Node) Depth() int {
	mx := 0
	for _, n := range n.Children {
		mx = max(mx, n.Depth())
	}
	return mx + 1
}

func getExprName(expr ast.Expr) string {
	switch fun := expr.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr: // X.Sel
		return fmt.Sprint(getExprName(fun.X), ".", fun.Sel.Name)
	default:
		return ""
	}
}
