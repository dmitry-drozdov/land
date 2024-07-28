package filter

import (
	"go/ast"
	"go/token"
)

type pair struct{ start, end token.Pos }

type NestedFuncs struct {
	funcPos []pair
}

func NewNestedFuncs() *NestedFuncs {
	return &NestedFuncs{
		funcPos: make([]pair, 0, 100),
	}
}

func (n *NestedFuncs) Nested(start, end token.Pos) bool {
	for _, pos := range n.funcPos {
		if start > pos.start && end < pos.end {
			return true
		}
	}
	return false
}

func (n *NestedFuncs) Add(x ast.Node) {
	n.funcPos = append(n.funcPos, pair{x.Pos(), x.End()})
}
