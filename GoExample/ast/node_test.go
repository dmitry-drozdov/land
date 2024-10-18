package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MergeTrees(t *testing.T) {
	tests := []struct {
		n1, n2, res *Node
	}{
		{ // nil check
			n1:  &Node{Offs: Offset{2, 20}},
			n2:  nil,
			res: &Node{Offs: Offset{2, 20}},
		},
		{ // n2 внутри n1
			n1:  &Node{Offs: Offset{2, 20}},
			n2:  &Node{Offs: Offset{5, 10}},
			res: &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{5, 10}}}},
		},
		{ // n2 внутри n1
			n1:  &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{5, 10}}}},
			n2:  &Node{Offs: Offset{15, 19}},
			res: &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{5, 10}}, {Offs: Offset{15, 19}}}},
		},
		{ // n2 внутри n1
			n1:  &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{15, 19}}}},
			n2:  &Node{Offs: Offset{5, 10}},
			res: &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{5, 10}}, {Offs: Offset{15, 19}}}},
		},
		{ // n2 внутри n1
			n1:  &Node{Offs: Offset{5, 20}, Chldren: []*Node{{Offs: Offset{15, 20}}}},
			n2:  &Node{Offs: Offset{5, 10}},
			res: &Node{Offs: Offset{5, 20}, Chldren: []*Node{{Offs: Offset{5, 10}}, {Offs: Offset{15, 20}}}},
		},
		{ // n2 внутри n1
			n1:  &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{3, 7}}, {Offs: Offset{15, 19}}}},
			n2:  &Node{Offs: Offset{8, 12}},
			res: &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{3, 7}}, {Offs: Offset{8, 12}}, {Offs: Offset{15, 19}}}},
		},
		{ // n2 внутри n1 матрешкой
			n1:  &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{5, 15}}}},
			n2:  &Node{Offs: Offset{6, 12}},
			res: &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{5, 15}, Chldren: []*Node{{Offs: Offset{6, 12}}}}}},
		},
		{ // n2 внутри n1 матрешкой
			n1:  &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{2, 4}}, {Offs: Offset{5, 15}}, {Offs: Offset{15, 19}}}},
			n2:  &Node{Offs: Offset{6, 12}},
			res: &Node{Offs: Offset{2, 20}, Chldren: []*Node{{Offs: Offset{2, 4}}, {Offs: Offset{5, 15}, Chldren: []*Node{{Offs: Offset{6, 12}}}}, {Offs: Offset{15, 19}}}},
		},
	}

	for _, tt := range tests {
		res := MergeTrees(tt.n1, tt.n2)
		assert.EqualValues(t, tt.res, res)

		// коммутативность
		res = MergeTrees(tt.n2, tt.n1)
		assert.EqualValues(t, tt.res, res)
	}
}
