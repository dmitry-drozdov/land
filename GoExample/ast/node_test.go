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
		{ // вклинивание
			n1: &N{Shft: Shft(0, 80), Chldren: Ns{
				{Shft: Shft(50, 69)},
			}},
			n2: &N{Shft: Shft(0, 80), Chldren: Ns{
				{Shft: Shft(40, 70), Chldren: Ns{{Shft: Shft(60, 65)}}},
			}},
			res: &N{Shft: Shft(0, 80), Chldren: Ns{
				{Shft: Shft(40, 70), Chldren: Ns{{Shft: Shft(50, 69), Chldren: Ns{{Shft: Shft(60, 65)}}}}},
			}},
		},
		{
			n1: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(50, 70), Chldren: Ns{
						{Shft: Shft(66, 69)},
					}},
				}},
			}},
			n2: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(40, 80), Chldren: Ns{
					{Shft: Shft(45, 79), Chldren: Ns{
						{Shft: Shft(60, 65)},
					}},
				}},
			}},
			res: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(40, 80), Chldren: Ns{
						{Shft: Shft(45, 79), Chldren: Ns{
							{Shft: Shft(50, 70), Chldren: Ns{
								{Shft: Shft(60, 65)},
								{Shft: Shft(66, 69)},
							}},
						}},
					}},
				}},
			}},
		},
	}

	for i, tt := range tests {
		res := MergeTrees((*Node)(tt.n1), (*Node)(tt.n2))
		if !assert.EqualValues(t, tt.res.Chldren, res.Chldren, i) {
			res.Print()
			continue
		}

		// коммутативность
		res = MergeTrees((*Node)(tt.n2), (*Node)(tt.n1))
		assert.EqualValues(t, tt.res, res, i)
	}
}

func Test_MergeTreesHard(t *testing.T) {
	tests := []struct {
		n1, n2, res *N
	}{
		{ // вклинивание через уровни
			n1: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(50, 70), Chldren: Ns{
						{Shft: Shft(66, 69)},
					}},
				}},
			}},
			n2: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(30, 39)},
				{Shft: Shft(40, 80), Chldren: Ns{
					{Shft: Shft(45, 79), Chldren: Ns{
						{Shft: Shft(60, 65)},
					}},
				}},
			}},
			res: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(30, 39)},
					{Shft: Shft(40, 80), Chldren: Ns{
						{Shft: Shft(45, 79), Chldren: Ns{
							{Shft: Shft(50, 70), Chldren: Ns{
								{Shft: Shft(60, 65)},
								{Shft: Shft(66, 69)},
							}},
						}},
					}},
				}},
			}},
		},
		{ // вклинивание через уровни
			n1: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(50, 70), Chldren: Ns{
						{Shft: Shft(66, 69)},
					}},
				}},
				{Shft: Shft(90, 100)},
			}},
			n2: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(30, 39)},
				{Shft: Shft(40, 80), Chldren: Ns{
					{Shft: Shft(45, 79), Chldren: Ns{
						{Shft: Shft(60, 65)},
					}},
				}},
			}},
			res: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(30, 39)},
					{Shft: Shft(40, 80), Chldren: Ns{
						{Shft: Shft(45, 79), Chldren: Ns{
							{Shft: Shft(50, 70), Chldren: Ns{
								{Shft: Shft(60, 65)},
								{Shft: Shft(66, 69)},
							}},
						}},
					}},
				}},
				{Shft: Shft(90, 100)},
			}},
		},
		{ // вклинивание через уровни
			n1: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(50, 70), Chldren: Ns{
						{Shft: Shft(66, 69)},
					}},
				}},
				{Shft: Shft(90, 120)},
			}},
			n2: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(30, 39)},
				{Shft: Shft(40, 80), Chldren: Ns{
					{Shft: Shft(45, 79), Chldren: Ns{
						{Shft: Shft(60, 65)},
					}},
				}},
				{Shft: Shft(100, 110)},
			}},
			res: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(30, 39)},
					{Shft: Shft(40, 80), Chldren: Ns{
						{Shft: Shft(45, 79), Chldren: Ns{
							{Shft: Shft(50, 70), Chldren: Ns{
								{Shft: Shft(60, 65)},
								{Shft: Shft(66, 69)},
							}},
						}},
					}},
				}},
				{Shft: Shft(90, 120), Chldren: Ns{{Shft: Shft(100, 110)}}},
			}},
		},
		{ // вклинивание через уровни
			n1: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(50, 70), Chldren: Ns{
						{Shft: Shft(66, 69)},
					}},
				}},
				{Shft: Shft(90, 120), Chldren: Ns{{Shft: Shft(90, 110)}}},
			}},
			n2: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(30, 39)},
				{Shft: Shft(40, 80), Chldren: Ns{
					{Shft: Shft(45, 79), Chldren: Ns{
						{Shft: Shft(60, 65)},
					}},
				}},
				{Shft: Shft(100, 109)},
				{Shft: Shft(130, 140)},
			}},
			res: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(30, 39)},
					{Shft: Shft(40, 80), Chldren: Ns{
						{Shft: Shft(45, 79), Chldren: Ns{
							{Shft: Shft(50, 70), Chldren: Ns{
								{Shft: Shft(60, 65)},
								{Shft: Shft(66, 69)},
							}},
						}},
					}},
				}},
				{Shft: Shft(90, 120), Chldren: Ns{
					{Shft: Shft(90, 110), Chldren: Ns{{Shft: Shft(100, 109)}}},
				}},
				{Shft: Shft(130, 140)},
			}},
		},
		{ // вклинивание через уровни
			n1: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(50, 70), Chldren: Ns{
						{Shft: Shft(66, 69)},
					}},
				}},
				{Shft: Shft(100, 110)},
			}},
			n2: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(10, 19), Chldren: Ns{
					{Shft: Shft(11, 17)},
				}},
				{Shft: Shft(30, 39)},
				{Shft: Shft(40, 80), Chldren: Ns{
					{Shft: Shft(45, 79), Chldren: Ns{
						{Shft: Shft(60, 65)},
					}},
				}},
				{Shft: Shft(90, 100)},
				{Shft: Shft(102, 108)},
			}},
			res: &N{Shft: Shft(0, 120), Chldren: Ns{
				{Shft: Shft(10, 19), Chldren: Ns{
					{Shft: Shft(11, 17)},
				}},
				{Shft: Shft(20, 80), Chldren: Ns{
					{Shft: Shft(30, 39)},
					{Shft: Shft(40, 80), Chldren: Ns{
						{Shft: Shft(45, 79), Chldren: Ns{
							{Shft: Shft(50, 70), Chldren: Ns{
								{Shft: Shft(60, 65)},
								{Shft: Shft(66, 69)},
							}},
						}},
					}},
				}},
				{Shft: Shft(90, 100)},
				{Shft: Shft(100, 110), Chldren: Ns{
					{Shft: Shft(102, 108)},
				}},
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

func Test_CorrectedNode(t *testing.T) {
	tests := []struct {
		n   *Node
		res *Node
	}{
		{n: &Node{}, res: &Node{}},
		{n: &Node{Type: "Any"}, res: &Node{Type: "Any"}},
		{n: &Node{Type: "Any", Text: "0123456789", Shft: Shft(0, 9)}, res: &Node{Type: "Any", Text: "0123456789", Shft: Shft(0, 9)}},
		{ // 1 any
			n: &Node{Type: "root", Text: "0123456789A", Shft: Shft(0, 10), Chldren: Ns{
				&Node{Type: "Any", Text: "0123456789A", Shft: Shft(0, 10), Chldren: Ns{
					{Type: "Some", Text: "45678", Shft: Shft(4, 8)},
				}},
			}},
			res: &Node{Type: "root", Text: "0123456789A", Shft: Shft(0, 10), Chldren: Ns{
				&Node{Type: "Any", Text: "0123", Shft: Shft(0, 3)},
				&Node{Type: "Some", Text: "45678", Shft: Shft(4, 8)},
				&Node{Type: "Any", Text: "9A", Shft: Shft(9, 10)},
			}},
		},
		{ // 2 Any
			n: &Node{Type: "root", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
				&Node{Type: "Any", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
					{Type: "Some", Text: "56", Shft: Shft(5, 6)},
					{Type: "Some", Text: "ABCD", Shft: Shft(10, 13)},
				}},
			}},
			res: &Node{Type: "root", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
				{Type: "Any", Text: "01234", Shft: Shft(0, 4)},
				{Type: "Some", Text: "56", Shft: Shft(5, 6)},
				{Type: "Any", Text: "789", Shft: Shft(7, 9)},
				{Type: "Some", Text: "ABCD", Shft: Shft(10, 13)},
				{Type: "Any", Text: "EF", Shft: Shft(14, 15)},
			}},
		},
		{ // Some слева
			n: &Node{Type: "root", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
				&Node{Type: "Any", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
					{Type: "Some", Text: "0123", Shft: Shft(0, 3)},
				}},
			}},
			res: &Node{Type: "root", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
				{Type: "Some", Text: "0123", Shft: Shft(0, 3)},
				{Type: "Any", Text: "456789ABCDEF", Shft: Shft(4, 15)},
			}},
		},
		{ // Some справа
			n: &Node{Type: "root", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
				&Node{Type: "Any", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
					{Type: "Some", Text: "ABCDEF", Shft: Shft(10, 15)},
				}},
			}},
			res: &Node{Type: "root", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
				{Type: "Any", Text: "0123456789", Shft: Shft(0, 9)},
				{Type: "Some", Text: "ABCDEF", Shft: Shft(10, 15)},
			}},
		},
		{ // 2 any
			n: &Node{Type: "root", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
				&Node{Type: "Any", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
					{Type: "Some", Text: "23456789", Shft: Shft(2, 9), Chldren: Ns{
						{Type: "Any", Text: "345678", Shft: Shft(3, 8), Chldren: Ns{
							{Type: "Some", Text: "56", Shft: Shft(5, 6)},
						}},
					}},
				}},
			}},
			res: &Node{Type: "root", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
				&Node{Type: "Any", Text: "01", Shft: Shft(0, 1)},
				&Node{Type: "Some", Text: "23456789", Shft: Shft(2, 9), Chldren: Ns{
					{Type: "Any", Text: "34", Shft: Shft(3, 4)},
					{Type: "Some", Text: "56", Shft: Shft(5, 6)},
					{Type: "Any", Text: "78", Shft: Shft(7, 8)},
				}},
				&Node{Type: "Any", Text: "ABCDEF", Shft: Shft(10, 15)},
			}},
		},
		{ // 2 any
			n: &Node{Type: "root", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
				&Node{Type: "Any", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
					{Type: "Some", Text: "23456789", Shft: Shft(2, 9), Chldren: Ns{
						{Type: "Any", Text: "345678", Shft: Shft(3, 8), Chldren: Ns{
							{Type: "Some", Text: "56", Shft: Shft(5, 6)},
						}},
					}},
				}},
			}},
			res: &Node{Type: "root", Text: "0123456789ABCDEF", Shft: Shft(0, 15), Chldren: Ns{
				&Node{Type: "Any", Text: "01", Shft: Shft(0, 1)},
				&Node{Type: "Some", Text: "23456789", Shft: Shft(2, 9), Chldren: Ns{
					{Type: "Any", Text: "34", Shft: Shft(3, 4)},
					{Type: "Some", Text: "56", Shft: Shft(5, 6)},
					{Type: "Any", Text: "78", Shft: Shft(7, 8)},
				}},
				&Node{Type: "Any", Text: "ABCDEF", Shft: Shft(10, 15)},
			}},
		},
	}

	for _, tt := range tests {
		CorrectedNode(tt.n)
		if !assert.EqualValues(t, tt.res, tt.n) {
			tt.n.Print()
		}
	}
}
