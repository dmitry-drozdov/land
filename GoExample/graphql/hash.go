package main

type Hash interface {
	Hash() uint64
}

var getName = func(d Def) string {
	return d.Name
}

func (i Input) Hash() uint64 {
	return 3*hash(i.Name) ^ 7*sliceHash(i.Defs, getName)
}

func (t Type) Hash() uint64 {
	return 3*hash(t.Name) ^ 7*sliceHash(t.Defs, getName)
}

func (f Func) Hash() uint64 {
	return hash(f.Parent) ^ 3*hash(f.Name) ^ 7*sliceHash(f.Args, getName) ^ 31*hash(f.Return)
}
