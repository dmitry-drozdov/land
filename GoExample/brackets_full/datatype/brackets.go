package datatype

import (
	"errors"
	"fmt"
)

type Brackets struct {
	Depth    int
	Children []*Brackets
}

func (b *Brackets) ControlNumber() (int, int) {
	return b.maxDepth(), b.count()
}

func (b *Brackets) maxDepth() int {
	if b == nil {
		return 0
	}
	mx := 0
	for _, c := range b.Children {
		mx = max(mx, c.maxDepth())
	}
	return mx
}

func (b *Brackets) count() int {
	if b == nil {
		return 0
	}
	cnt := 0
	for _, c := range b.Children {
		cnt += c.count()
	}
	return cnt
}

func (b1 *Brackets) EqualTo(b2 *Brackets) error {
	if (b1 == nil) != (b2 == nil) {
		return errors.New("xor failed")
	}
	if b1.Depth != b2.Depth {
		return fmt.Errorf("depth failed: [land=%v] [go=%v]", b1.Depth, b2.Depth)
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
