package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"utils/code"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/introspection"
)

var includeDepricated = &struct{ IncludeDeprecated bool }{IncludeDeprecated: true}

func parseFolders() (map[string]Result, error) {
	res := make(map[string]Result, len(source))

	for name, path := range source {
		r, err := parse(path)
		if err != nil {
			return nil, err
		}
		if r != nil {
			res[name] = *r
		}
	}

	return res, nil
}

func parse(folder string) (*Result, error) {
	var content strings.Builder
	content.Grow(1 << 20)

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".graphql" {
			return nil
		}
		if err != nil {
			return err
		}

		readFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer readFile.Close()

		bytes, err := io.ReadAll(readFile)
		if err != nil {
			return err
		}

		content.Write(bytes)
		content.WriteByte('\n')

		return err
	})
	if err != nil {
		return nil, err
	}

	schemaStr := content.String()
	schema, err := graphql.ParseSchema(schemaStr, nil)
	if err != nil {
		fErr := os.WriteFile("schema.graphql", []byte(schemaStr), 0)
		if fErr != nil {
			fmt.Println(fErr)
		}
		return nil, err
	}

	funcWithoutArgs := func(fname string) bool {
		return strings.Contains(schemaStr, fname+"():")
	}

	insp := schema.Inspect()

	types := make([]Type, 0, len(insp.Types())/3+1)
	inputs := make([]Input, 0, len(insp.Types())/3+1)
	funcs := make([]Func, 0, len(insp.Types())/3+1)

	for _, f := range insp.Types() {
		if ignore(*f.Name()) {
			continue
		}

		if f.InputFields() != nil { // it's an input only
			defs := make([]Def, 0, len(*f.InputFields()))
			for _, iff := range *f.InputFields() {
				defs = append(defs, Def{Name: iff.Name(), Type: getType(iff.Type())})
			}
			inputs = append(inputs, Input{Name: *f.Name(), Defs: defs})
		}

		if f.Fields(includeDepricated) != nil { // type or func
			tp := Type{
				Name: *f.Name(),
				Defs: make([]Def, 0, len(*f.Fields(includeDepricated))),
			}
			fn := make([]Func, 0, len(*f.Fields(includeDepricated)))
			for _, ff := range *f.Fields(includeDepricated) {
				if ignore(getType(ff.Type())) || ignore(ff.Name()) {
					continue
				}

				cntArgs := len(ff.Args())
				// => it's function
				if cntArgs > 0 || cntArgs == 0 && funcWithoutArgs(ff.Name()) {
					args := make([]Def, 0, cntArgs)
					for _, arg := range ff.Args() {
						if ignore(arg.Name()) {
							panic("unreached")
						}
						args = append(args, Def{
							Name: arg.Name(),
							Type: getType(arg.Type()),
						})
					}
					fn = append(fn, Func{
						Parent: tp.Name,
						Name:   ff.Name(),
						Args:   args,
						Return: getType(ff.Type()),
					})
					continue
				}

				// => it's variable
				tp.Defs = append(tp.Defs, Def{
					Name: ff.Name(),
					Type: getType(ff.Type()),
				})
			}
			if len(tp.Defs) > 0 {
				types = append(types, tp)
			}
			funcs = append(funcs, fn...)
		}
	}

	return &Result{
		Inputs: inputs,
		Types:  types,
		Funcs:  funcs,
		LOC:    code.GetLOC(schemaStr),
	}, nil
}

func getType(t *introspection.Type) string {
	if t.Name() != nil {
		return *t.Name()
	}
	switch t.Kind() {
	case "LIST":
		return "[" + getType(t.OfType()) + "]"
	case "NON_NULL":
		return getType(t.OfType()) + "!"
	default:
		panic("unexpected type")
	}
}

func ignore(name string) bool {
	return strings.HasPrefix(name, "_")
}
