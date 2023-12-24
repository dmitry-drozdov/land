package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/introspection"
)

var mp = map[string]string{
	"apollographql": `E:\phd\test_repos_graphql\git\apollographql`,
	"dgraph-io":     `E:\phd\test_repos_graphql\git\dgraph-io`,
	"wasmerio":      `E:\phd\test_repos_graphql\git\wasmerio`,
	"qmd":           `e:\phd\test_repos_graphql\qmd\`,
	"mts":           `E:\phd\test_repos_graphql\mts\`,
}

var includeDepricated = &struct{ IncludeDeprecated bool }{IncludeDeprecated: true}

func main() {
	res := make(map[string]Result, len(mp))

	for name, path := range mp {
		r, err := parse(path)
		if err != nil {
			log.Fatalf("[%s]: [%+v]", name, err)
		}
		if r != nil {
			res[name] = *r
		}
	}

	// b, _ := json.Marshal(res["qmd"])
	// fmt.Println(string(b))
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

	schema, err := graphql.ParseSchema(content.String(), nil)
	if err != nil {
		return nil, err
	}

	insp := schema.Inspect()

	types := make([]Type, 0, len(insp.Types())/3+1)
	inputs := make([]Input, 0, len(insp.Types())/3+1)
	funcs := make([]Func, 0, len(insp.Types())/3+1)

	for _, f := range insp.Types() {
		if f.InputFields() != nil {
			defs := make([]Def, 0, len(*f.InputFields()))
			for _, iff := range *f.InputFields() {
				defs = append(defs, Def{Name: iff.Name(), Type: getType(iff.Type())})
			}
			inputs = append(inputs, Input{Name: *f.Name(), Defs: defs})
		}

		if f.Fields(includeDepricated) != nil {
			tp := Type{
				Name: *f.Name(),
				Defs: make([]Def, 0, len(*f.Fields(includeDepricated))),
			}
			fn := make([]Func, 0, len(*f.Fields(includeDepricated)))
			for _, ff := range *f.Fields(includeDepricated) {
				if ignore(getType(ff.Type())) || ignore(ff.Name()) {
					continue
				}
				if len(ff.Args()) > 0 { // => it's function
					args := make([]Def, 0, len(ff.Args()))
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
						Name:   ff.Name(),
						Args:   args,
						Return: getType(ff.Type()),
					})
				} else { // => it's variable
					tp.Defs = append(tp.Defs, Def{
						Name: ff.Name(),
						Type: getType(ff.Type()),
					})
				}
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
	return strings.Contains(name, "_")
}
