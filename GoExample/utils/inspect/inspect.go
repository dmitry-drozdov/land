package inspect

import (
	"fmt"
	"go/ast"
)

// func init() {
// 	fset := token.NewFileSet()

// 	f, err := parser.ParseFile(fset, "", []byte(`
// 	package main
// 	func main(){
// 		log.G(context.TODO()).Errorf("Could not resolve driver %s while handling driver table event: %v", n.networkType, err)
// 	}
// `), 0)
// 	if err != nil {
// 		panic(err)
// 	}
// 	res := &Node{
// 		Type: "func_body",
// 	}
// 	ast.Inspect(f, func(n ast.Node) bool {
// 		x, ok := n.(*ast.FuncDecl)
// 		if !ok {
// 			return true
// 		}

// 		inspect(x, res)
// 		return false
// 	})

// 	b, _ := json.Marshal(res)
// 	fmt.Println(string(b))

// 	os.Exit(1)
// }

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
			if child.Name == "" {
				return true // not a function call
			}
			inspect(x.Fun, node)
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
		case *ast.RangeStmt:
			child.Type = "for"
			inspect(x.X, child) // range x.X { x.Body }
			inspect(x.Body, child)
		case *ast.TypeSwitchStmt:
			child.Type = "switch"
			inspect(x.Init, child) // switch a := ssi.(gn) { x.Body}
			inspect(x.Body, child)
			inspect(x.Assign, child)
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
