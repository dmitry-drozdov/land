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
	LOC    uint
}

func (r *Result) Sort() {
	for _, input := range r.Inputs {
		sortDefs(input.Defs)
	}
	sort.Slice(r.Inputs, func(i, j int) bool {
		return r.Inputs[i].Hash() > r.Inputs[j].Hash()
	})

	for _, tp := range r.Types {
		sortDefs(tp.Defs)
	}
	sort.Slice(r.Types, func(i, j int) bool {
		return r.Types[i].Hash() > r.Types[j].Hash()
	})

	for _, fun := range r.Funcs {
		sortDefs(fun.Args)
	}
	sort.Slice(r.Funcs, func(i, j int) bool {
		return r.Funcs[i].Hash() > r.Funcs[j].Hash()
	})
}

func sortDefs(s []Def) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Name > s[j].Name
	})
}

func (r *Result) CheckDuplicates() error {
	if err := checkDuplicates(r.Inputs); err != nil {
		return fmt.Errorf("inputs: [%w]", err)
	}
	if err := checkDuplicates(r.Funcs); err != nil {
		return fmt.Errorf("funcs: [%w]", err)
	}
	if err := checkDuplicates(r.Types); err != nil {
		return fmt.Errorf("types: [%w]", err)
	}
	return nil
}

func checkDuplicates[T Hash](sl []T) error {
	last := *new(T)
	for _, v := range sl {
		if v.Hash() == last.Hash() {
			return fmt.Errorf("duplicates: [%+v]", v)
		}
		last = v
	}
	return nil
}

func (r *Result) Empty() bool {
	return r == nil || (len(r.Inputs) == 0 && len(r.Types) == 0 && len(r.Funcs) == 0)
}

func (r *Result) EqualTo(o *Result) error {
	r.Sort()
	o.Sort()

	if err := r.CheckDuplicates(); err != nil {
		return err
	}
	if err := o.CheckDuplicates(); err != nil {
		return err
	}

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
	Parent string
	Name   string
	Args   []Def
	Return string
}

type Def struct {
	Name string
	Type string
}
