package generate

import (
	"brackets/node"
)

type GenerateRes struct {
	CodeToText map[string]string
	CodeToNode map[string]*node.Node
	dups       map[string]struct{}
}

func NewGenerateRes(len int) *GenerateRes {
	return &GenerateRes{
		CodeToText: make(map[string]string, len),
		CodeToNode: make(map[string]*node.Node, len),
		dups:       make(map[string]struct{}, len),
	}
}

func (g *GenerateRes) Add(s string, code string) error {
	if s == "" {
		return nil
	}
	if _, ok := g.dups[s]; ok {
		return nil
	}

	g.CodeToText[code] = s

	n, err := node.ParseAst(s)
	if err != nil {
		return err
	}
	g.CodeToNode[code] = n

	g.dups[s] = struct{}{}

	return nil
}
