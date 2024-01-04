package main

import (
	"sort"
	"strings"
)

type FuncStat struct {
	Name    string
	Args    []string
	ArgsCnt int
	Return  int
}

func (f *FuncStat) EqualTo(g *FuncStat) bool {
	if f == nil && g == nil {
		return true
	}
	if f == nil || g == nil {
		return false
	}
	if f.Name != g.Name || f.ArgsCnt != g.ArgsCnt || f.Return != g.Return {
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