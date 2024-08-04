package inspect

import (
	"fmt"
	"sort"
	"utils/hash"
)

type Node struct {
	Type     string
	Name     string
	Children []*Node
}

func (n *Node) Hash() uint64 {
	h := hash.HashStrings(n.Type, n.Name)
	if len(n.Children) == 0 {
		return h
	}
	for _, c := range n.Children {
		h += c.Hash()
	}
	return h
}

func (n1 *Node) EqualTo(n2 *Node) error {
	if n1.Name != n2.Name {
		return fmt.Errorf("name mismatch: %v %v", n1.Name, n2.Name)
	}
	if n1.Type != n2.Type {
		return fmt.Errorf("type mismatch: %v %v ", n1.Type, n2.Type)
	}
	if len(n1.Children) != len(n2.Children) {
		if len(n1.Children) < len(n2.Children) {
			//return nil // LanD can find more children
		}
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
