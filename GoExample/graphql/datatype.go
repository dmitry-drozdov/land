package main

import (
	"fmt"
	"reflect"
	"sort"
)

type Result struct {
	Inputs []Input
	Types  []Type
	Funcs  []Func
}

var getName = func(d Def) string {
	return d.Name
}

func (r *Result) Sort() {
	for _, input := range r.Inputs {
		sortDefs(input.Defs)
	}
	sort.Slice(r.Inputs, func(i, j int) bool {
		if r.Inputs[i].Name == r.Inputs[j].Name {
			return sliceHash(r.Inputs[i].Defs, getName) > sliceHash(r.Inputs[j].Defs, getName)
		}
		return r.Inputs[i].Name > r.Inputs[j].Name
	})
	r.Inputs = removeDuplicates(r.Inputs)

	for _, tp := range r.Types {
		sortDefs(tp.Defs)
	}
	sort.Slice(r.Types, func(i, j int) bool {
		if r.Types[i].Name == r.Types[j].Name {
			return sliceHash(r.Types[i].Defs, getName) > sliceHash(r.Types[j].Defs, getName)
		}
		return r.Types[i].Name > r.Types[j].Name
	})
	r.Types = removeDuplicates(r.Types)

	for _, fun := range r.Funcs {
		sortDefs(fun.Args)
	}
	sort.Slice(r.Funcs, func(i, j int) bool {
		if r.Funcs[i].Name == r.Funcs[j].Name {
			sh1 := sliceHash(r.Funcs[i].Args, getName)
			sh2 := sliceHash(r.Funcs[j].Args, getName)
			if sh1 == sh2 {
				return r.Funcs[i].Return > r.Funcs[j].Return
			}
			return sh1 > sh2
		}
		return r.Funcs[i].Name > r.Funcs[j].Name
	})
	r.Funcs = removeDuplicates(r.Funcs)
}

func sortDefs(s []Def) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Name > s[j].Name
	})
}

func removeDuplicates[T Hash](sl []T) []T {
	result := make([]T, 0, len(sl))
	last := *new(T)
	for _, v := range sl {
		if v.Hash() == last.Hash() {
			continue
		}
		result = append(result, v)
		last = v
	}
	return result
}

func (r *Result) Empty() bool {
	return r == nil || (len(r.Inputs) == 0 && len(r.Types) == 0 && len(r.Funcs) == 0)
}

func (r *Result) EqualTo(o *Result) error {
	r.Sort()
	o.Sort()

	if len(r.Funcs) != len(o.Funcs) {
		return fmt.Errorf("len func mismatch [%+v] [%+v]", len(r.Funcs), len(o.Funcs))
	}
	if len(r.Types) != len(o.Types) {
		return fmt.Errorf("len type mismatch [%+v] [%+v]", len(r.Types), len(o.Types))
	}
	if len(r.Inputs) != len(o.Inputs) {
		return fmt.Errorf("len inputs mismatch [%+v] [%+v]", len(r.Inputs), len(o.Inputs))
	}

	for i := range r.Funcs {
		if !reflect.DeepEqual(r.Funcs[i], o.Funcs[i]) {
			return fmt.Errorf("func [%+v] != [%+v]", r.Funcs[i], o.Funcs[i])
		}
	}
	for i := range r.Types {
		if !reflect.DeepEqual(r.Types[i], o.Types[i]) {
			return fmt.Errorf("type [%+v] != [%+v]", r.Types[i], o.Types[i])
		}
	}
	for i := range r.Inputs {
		if !reflect.DeepEqual(r.Inputs[i], o.Inputs[i]) {
			return fmt.Errorf("input [%+v] != [%+v]", r.Inputs[i], o.Inputs[i])
		}
	}
	return nil
}

type Input struct {
	Name string
	Defs []Def
}

type Type struct {
	Name string
	Defs []Def
}

type Func struct {
	Name   string
	Args   []Def
	Return string
}

type Def struct {
	Name string
	Type string
}
