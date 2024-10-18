package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type N Node
type Ns []*Node
type S Shift

func Test_MergeTrees(t *testing.T) {
	tests := []struct {
		n1, n2, res *N
	}{
		{ // equals
			n1:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 10)}}},
			n2:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(11, 15)}}},
			res: &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 10)}, {Shft: Shft(11, 15)}}},
		},
		{ // equals
			n1:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 10)}, {Shft: Shft(11, 15)}}},
			n2:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(11, 15)}, {Shft: Shft(16, 20)}}},
			res: &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 10)}, {Shft: Shft(11, 15)}, {Shft: Shft(16, 20)}}},
		},
		{ // equals + merge
			n1:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 10)}, {Shft: Shft(11, 15), Chldren: Ns{{Shft: Shft(11, 13)}}}}},
			n2:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(11, 15), Chldren: Ns{{Shft: Shft(14, 15)}}}, {Shft: Shft(16, 20)}}},
			res: &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 10)}, {Shft: Shft(11, 15), Chldren: Ns{{Shft: Shft(11, 13)}, {Shft: Shft(14, 15)}}}, {Shft: Shft(16, 20)}}},
		},
		{ // nil check
			n1:  &N{Shft: Shft(2, 20)},
			n2:  nil,
			res: &N{Shft: Shft(2, 20)},
		},
		{ // n2 внутри n1
			n1:  &N{Shft: Shft(2, 20)},
			n2:  &N{Shft: Shft(5, 10)},
			res: &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 10)}}},
		},
		{ // n2 внутри n1
			n1:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 10)}}},
			n2:  &N{Shft: Shft(15, 19)},
			res: &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 10)}, {Shft: Shft(15, 19)}}},
		},
		{ // n2 внутри n1
			n1:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(15, 19)}}},
			n2:  &N{Shft: Shft(5, 10)},
			res: &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 10)}, {Shft: Shft(15, 19)}}},
		},
		{ // n2 внутри n1
			n1:  &N{Shft: Shft(5, 20), Chldren: Ns{{Shft: Shft(15, 20)}}},
			n2:  &N{Shft: Shft(5, 10)},
			res: &N{Shft: Shft(5, 20), Chldren: Ns{{Shft: Shft(5, 10)}, {Shft: Shft(15, 20)}}},
		},
		{ // n2 внутри n1
			n1:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(3, 7)}, {Shft: Shft(15, 19)}}},
			n2:  &N{Shft: Shft(8, 12)},
			res: &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(3, 7)}, {Shft: Shft(8, 12)}, {Shft: Shft(15, 19)}}},
		},
		{ // n2 внутри n1 матрешкой
			n1:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 15)}}},
			n2:  &N{Shft: Shft(6, 12)},
			res: &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(5, 15), Chldren: Ns{{Shft: Shft(6, 12)}}}}},
		},
		{ // n2 внутри n1 матрешкой
			n1:  &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(2, 4)}, {Shft: Shft(5, 15)}, {Shft: Shft(15, 19)}}},
			n2:  &N{Shft: Shft(6, 12)},
			res: &N{Shft: Shft(2, 20), Chldren: Ns{{Shft: Shft(2, 4)}, {Shft: Shft(5, 15), Chldren: Ns{{Shft: Shft(6, 12)}}}, {Shft: Shft(15, 19)}}},
		},
		{ // не пересекаются
			n1:  &N{Shft: Shft(0, 50), Chldren: Ns{{Shft: Shft(10, 20)}}},
			n2:  &N{Shft: Shft(60, 90), Chldren: Ns{{Shft: Shft(65, 75)}}},
			res: &N{Shft: Shft(0, 90), Chldren: Ns{{Shft: Shft(0, 50), Chldren: Ns{{Shft: Shft(10, 20)}}}, {Shft: Shft(60, 90), Chldren: Ns{{Shft: Shft(65, 75)}}}}},
		},
		{ // перекрестное
			n1: &N{Shft: Shft(0, 80), Chldren: Ns{
				{Shft: Shft(10, 30)},
				{Shft: Shft(50, 69)},
			}},
			n2: &N{Shft: Shft(0, 80), Chldren: Ns{
				{Shft: Shft(20, 29), Chldren: Ns{{Shft: Shft(23, 28)}}},
				{Shft: Shft(40, 70), Chldren: Ns{{Shft: Shft(60, 65)}}},
			}},
			res: &N{Shft: Shft(0, 80), Chldren: Ns{
				{Shft: Shft(10, 30), Chldren: Ns{{Shft: Shft(20, 29), Chldren: Ns{{Shft: Shft(23, 28)}}}}},
				{Shft: Shft(40, 70), Chldren: Ns{{Shft: Shft(50, 69), Chldren: Ns{{Shft: Shft(60, 65)}}}}},
			}},
		},
	}

	for _, tt := range tests {
		res := MergeTrees((*Node)(tt.n1), (*Node)(tt.n2))
		if !assert.EqualValues(t, tt.res.Chldren, res.Chldren) {
			res.Print()
			continue
		}

		// коммутативность
		res = MergeTrees((*Node)(tt.n2), (*Node)(tt.n1))
		assert.EqualValues(t, tt.res, res)
	}
}
