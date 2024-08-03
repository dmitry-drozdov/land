package ast_type

import (
	"fmt"
	"go/ast"
	"sync"
)

type TypeStats map[string]int

type NameConverter struct {
	stats  TypeStats
	reqCnt int
	mx     sync.Mutex
}

func NewNameConverter() *NameConverter {
	return &NameConverter{
		stats: make(TypeStats, 10),
		mx:    sync.Mutex{},
	}
}

func (a *NameConverter) Stats() TypeStats {
	return a.stats
}
func (a *NameConverter) ReqCnt() int {
	return a.reqCnt
}

func (a *NameConverter) HumanType(tp ast.Expr) (res string) {
	defer func() { a.mx.Lock(); a.stats[res]++; a.mx.Unlock() }()

	a.reqCnt++
	switch t := tp.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%v.%v\n", t.X, t.Sel.Name)
	case *ast.StarExpr:
		return a.HumanType(t.X)
	case *ast.ArrayType:
		return a.HumanType(t.Elt)
	case *ast.IndexExpr:
		return a.HumanType(t.X)
	case *ast.IndexListExpr:
		return a.HumanType(t.X)
	case *ast.ParenExpr:
		return a.HumanType(t.X)
	case *ast.FuncType:
		return "anon_func_title"
	case *ast.ChanType:
		return "chan"
	case *ast.MapType:
		return "map"
	case *ast.StructType:
		return "anon_struct"
	case *ast.InterfaceType:
		return "anon_interface"
	case *ast.Ellipsis:
		return a.HumanType(t.Elt)
	}
	return fmt.Sprintf("%T", tp)
}
