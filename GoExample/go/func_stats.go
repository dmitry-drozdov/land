package main

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/constraints"
)

type FuncStat struct {
	Receiver           string
	Name               string
	Args               []string
	ArgsDepth          []int
	ArgsCnt            byte
	Return             byte
	ReturnsDepth       []int
	RequirePostProcess bool
	NoBody             bool
}

type StructStat struct {
	Name  string
	Types []string
}

func (f *FuncStat) EqualTo(g *FuncStat, gt GrammarType) bool {
	if f == nil && g == nil {
		return true
	}
	if f == nil || g == nil {
		return false
	}
	if f.Name != g.Name || f.Receiver != g.Receiver {
		return false
	}
	if gt == GrammarTypeHighLevel {
		return true
	}

	if f.ArgsCnt != g.ArgsCnt || f.Return != g.Return {
		return false
	}
	if len(f.Args) != len(g.Args) {
		return false
	}
	if len(f.Args) == 0 && len(g.Args) == 0 {
		return true
	}
	sortSlice := func(s []string) {
		sort.Slice(s, func(i, j int) bool {
			return s[i] < s[j]
		})
	}
	sortSlice(f.Args)
	sortSlice(g.Args)

	for i := range f.Args {
		if !compareStrings(f.Args[i], g.Args[i]) {
			return false
		}
	}

	return true
}

func compareStrings(s1, s2 string) bool {
	s1 = strings.TrimSpace(s1)
	s1 = strings.ToLower(s1)
	s2 = strings.TrimSpace(s2)
	s2 = strings.ToLower(s2)
	return s1 == s2
}

type DepthStats[T interface {
	constraints.Float | constraints.Integer
}] struct {
	max, total, cnt T
	names           []string
}

func (d *DepthStats[T]) Process(depth T, name string) T {
	if depth > d.max {
		d.max = depth
		d.names = []string{}
	}
	if depth == d.max {
		d.names = append(d.names, name)
	}
	d.cnt++
	d.total += depth
	return depth
}

func (d *DepthStats[T]) String() string {
	var names []string
	if d.max > 2 {
		names = d.names
	}
	return fmt.Sprintf("max[%v] avg[%.3f] [%+v]\n", d.max, float64(d.total)/float64(d.cnt), names)
}
