package node

import (
	"encoding/json"
	"fmt"
	"sort"
	"utils/hash"
)

type Node struct {
	Type     string
	Children []*Node
}

func (n *Node) String() string {
	b, _ := json.Marshal(n)
	return string(b)
}

func (n *Node) Hash() uint64 {
	h := hash.HashString(n.Type)
	if len(n.Children) == 0 {
		return h
	}
	for _, c := range n.Children {
		h += c.Hash()
	}
	return h
}

func (n1 *Node) EqualTo(n2 *Node) error {
	if n1.Type != n2.Type {
		return fmt.Errorf("type mismatch: [%v] [%v]", n1.Type, n2.Type)
	}
	if len(n1.Children) != len(n2.Children) {
		return fmt.Errorf("children len mismatch: %v %v", len(n1.Children), len(n2.Children))
	}
	sort.Slice(n1.Children, func(i, j int) bool {
		return n1.Children[i].Hash() < n1.Children[j].Hash()
	})
	sort.Slice(n2.Children, func(i, j int) bool {
		return n2.Children[i].Hash() < n2.Children[j].Hash()
	})

	for i, c1 := range n1.Children {
		if err := c1.EqualTo(n2.Children[i]); err != nil {
			return fmt.Errorf("child err: %w", err)
		}
	}

	return nil
}

func (n *Node) processAny() { //  Any Any => Any; in general Any ... Any => Any;
	if n == nil || len(n.Children) == 0 {
		return
	}
	children := make([]*Node, 0, len(n.Children))
	for i := 0; i < len(n.Children); i++ {
		if i < len(n.Children)-1 && n.Children[i].Type == "any" && n.Children[i+1].Type == "any" {
			continue
		}
		children = append(children, n.Children[i])
		n.Children[i].processAny()
	}
	n.Children = children
}
