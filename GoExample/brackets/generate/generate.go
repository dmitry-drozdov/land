package generate

import (
	"brackets/node"
	"strings"
	"sync"
)

type GenerateRes struct {
	CodeToText map[string]string
	CodeToNode map[string]*node.Node
	dups       map[string]struct{}
	mx         sync.Mutex
}

func NewGenerateRes(len int) *GenerateRes {
	return &GenerateRes{
		CodeToText: make(map[string]string, len),
		CodeToNode: make(map[string]*node.Node, len),
		dups:       make(map[string]struct{}, len),
		mx:         sync.Mutex{},
	}
}

func (g *GenerateRes) Add(s string, code string) error {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "<", "{")
	s = strings.ReplaceAll(s, ">", "}")
	s = strings.ReplaceAll(s, "_", "\n")
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

func (g1 *GenerateRes) Merge(g2 *GenerateRes) {
	if g2 == nil || g1 == nil {
		return
	}
	g1.mx.Lock()
	defer g1.mx.Unlock()
	for k, v := range g2.CodeToText {
		if _, ok := g1.dups[k]; ok {
			continue
		}

		g1.CodeToText[k] = v
		g1.CodeToNode[k] = g2.CodeToNode[k]

		g1.dups[k] = struct{}{}
	}
}
