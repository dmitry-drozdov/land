package main

import h "utils/hash"

type Hash interface {
	Hash() uint64
	GetName() string
}

var getName = func(d Def) string {
	return d.Name
}

func (i Input) Hash() uint64 {
	return 3*h.HashString(i.Name) ^ 7*h.HashSlice(i.Defs, getName)
}

func (t Type) Hash() uint64 {
	return 3*h.HashString(t.Name) ^ 7*h.HashSlice(t.Defs, getName)
}

func (f Func) Hash() uint64 {
	return h.HashString(f.Parent) ^ 3*h.HashString(f.Name) ^ 7*h.HashSlice(f.Args, getName) ^ 31*h.HashString(f.Return)
}

func (i Input) GetName() string {
	return i.Name
}

func (t Type) GetName() string {
	return t.Name
}

func (f Func) GetName() string {
	return f.Name
}
