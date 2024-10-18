package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MergeTrees(t *testing.T) {
	tests := []struct {
		n1, n2, res *Node
	}{
		{ // equals
			n1:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}}},
			n2:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{11, 15}}}},
			res: &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}, {Shft: Shift{11, 15}}}},
		},
		{ // equals
			n1:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}, {Shft: Shift{11, 15}}}},
			n2:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{11, 15}}, {Shft: Shift{16, 20}}}},
			res: &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}, {Shft: Shift{11, 15}}, {Shft: Shift{16, 20}}}},
		},
		{ // equals + merge
			n1:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}, {Shft: Shift{11, 15}, Chldren: []*Node{{Shft: Shift{11, 13}}}}}},
			n2:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{11, 15}, Chldren: []*Node{{Shft: Shift{14, 15}}}}, {Shft: Shift{16, 20}}}},
			res: &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}, {Shft: Shift{11, 15}, Chldren: []*Node{{Shft: Shift{11, 13}}, {Shft: Shift{14, 15}}}}, {Shft: Shift{16, 20}}}},
		},
		{ // nil check
			n1:  &Node{Shft: Shift{2, 20}},
			n2:  nil,
			res: &Node{Shft: Shift{2, 20}},
		},
		{ // n2 внутри n1
			n1:  &Node{Shft: Shift{2, 20}},
			n2:  &Node{Shft: Shift{5, 10}},
			res: &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}}},
		},
		{ // n2 внутри n1
			n1:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}}},
			n2:  &Node{Shft: Shift{15, 19}},
			res: &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}, {Shft: Shift{15, 19}}}},
		},
		{ // n2 внутри n1
			n1:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{15, 19}}}},
			n2:  &Node{Shft: Shift{5, 10}},
			res: &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}, {Shft: Shift{15, 19}}}},
		},
		{ // n2 внутри n1
			n1:  &Node{Shft: Shift{5, 20}, Chldren: []*Node{{Shft: Shift{15, 20}}}},
			n2:  &Node{Shft: Shift{5, 10}},
			res: &Node{Shft: Shift{5, 20}, Chldren: []*Node{{Shft: Shift{5, 10}}, {Shft: Shift{15, 20}}}},
		},
		{ // n2 внутри n1
			n1:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{3, 7}}, {Shft: Shift{15, 19}}}},
			n2:  &Node{Shft: Shift{8, 12}},
			res: &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{3, 7}}, {Shft: Shift{8, 12}}, {Shft: Shift{15, 19}}}},
		},
		{ // n2 внутри n1 матрешкой
			n1:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 15}}}},
			n2:  &Node{Shft: Shift{6, 12}},
			res: &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{5, 15}, Chldren: []*Node{{Shft: Shift{6, 12}}}}}},
		},
		{ // n2 внутри n1 матрешкой
			n1:  &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{2, 4}}, {Shft: Shift{5, 15}}, {Shft: Shift{15, 19}}}},
			n2:  &Node{Shft: Shift{6, 12}},
			res: &Node{Shft: Shift{2, 20}, Chldren: []*Node{{Shft: Shift{2, 4}}, {Shft: Shift{5, 15}, Chldren: []*Node{{Shft: Shift{6, 12}}}}, {Shft: Shift{15, 19}}}},
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
