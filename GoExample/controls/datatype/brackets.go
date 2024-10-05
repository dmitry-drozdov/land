package datatype

import (
	"errors"
	"fmt"
)

type Control struct {
	Type     string
	Depth    int
	Children []*Control
}

func (b *Control) ControlNumber() (int, int) {
	return b.MaxDepth(), b.Count()
}

func (b *Control) MaxChildrenCount() int {
	if b == nil {
		return 0
	}
	mx := len(b.Children)
	for _, c := range b.Children {
		mx = max(mx, c.MaxChildrenCount())
	}
	return mx
}

func (b *Control) MaxDepth() int {
	if b == nil {
		return 0
	}
	mx := 0
	for _, c := range b.Children {
		mx = max(mx, c.MaxDepth())
	}
	return mx
}

func (b *Control) Count() int {
	if b == nil {
		return 0
	}
	cnt := 1
	for _, c := range b.Children {
		cnt += c.Count()
	}
	return cnt
}

func (b1 *Control) EqualTo(b2 *Control) error {
	if (b1 == nil) != (b2 == nil) {
		return errors.New("xor failed")
	}
	if b1.Depth != b2.Depth {
		return fmt.Errorf("depth failed: [land=%v] [go=%v]", b1.Depth, b2.Depth)
	}
	if b1.Type != b2.Type {
		return fmt.Errorf("type failed: [land=%v] [go=%v]", b1.Type, b2.Type)
	}
	if len(b1.Children) != len(b2.Children) {
		return fmt.Errorf("len failed: [land=%v] [go=%v]", len(b1.Children), len(b2.Children))
	}
	for i := range b1.Children {
		err := b1.Children[i].EqualTo(b2.Children[i])
		if err != nil {
			return fmt.Errorf("sub err [pos=%v] [%w]", i, err)
		}
	}
	return nil
}
