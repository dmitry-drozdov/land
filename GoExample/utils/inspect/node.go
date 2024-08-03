package inspect

import "fmt"

type Node struct {
	Type     string
	Name     string
	Children []*Node
}

func (n1 *Node) EqualTo(n2 *Node) error {
	if n1.Name != n2.Name {
		return fmt.Errorf("name mismatch: %v %v", n1.Name, n2.Name)
	}
	if n1.Type != n2.Type {
		return fmt.Errorf("type mismatch: %v %v ", n1.Type, n2.Type)
	}
	if len(n1.Children) != len(n2.Children) {
		return fmt.Errorf("children len mismatch: %v %v", len(n1.Children), len(n2.Children))
	}
	return nil
}
