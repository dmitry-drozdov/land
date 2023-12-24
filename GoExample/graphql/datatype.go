package main

type Result struct {
	Inputs []Input
	Types  []Type
	Funcs  []Func
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
