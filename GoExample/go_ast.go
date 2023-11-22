package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

func ParseFiles(root string) (map[string]map[string]*FuncStat, error) {
	res := make(map[string]map[string]*FuncStat, 10000)
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 5)
	l := sync.Mutex{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" {
			return nil
		}

		pathBk := path

		g.Go(func() error {
			readFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer readFile.Close()

			data, err := ParseFile(pathBk)
			if err != nil {
				return err
			}

			pathBk = strings.ReplaceAll(pathBk, root, "")
			pathBk = strings.ReplaceAll(pathBk, `\`, "")
			l.Lock()
			res[pathBk] = data
			l.Unlock()
			return nil
		})

		return nil
	})
	g.Wait()
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("empty output")
	}

	return res, nil
}

func ParseFile(path string) (map[string]*FuncStat, error) {
	res := make(map[string]*FuncStat, 10000)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			ret := 0
			if x.Type.Results != nil {
				ret = x.Type.Results.NumFields()
			}

			ptr := &FuncStat{
				Name:    x.Name.Name,
				ArgsCnt: x.Type.Params.NumFields(),
				Return:  ret,
			}
			res[x.Name.Name] = ptr

			add := func(a string) {
				ptr.Args = append(ptr.Args, a)
			}

			for _, y := range x.Type.Params.List {
				// can be several args with 1 type: n int, j, k, l float
				for i := 0; i < len(y.Names); i++ {
					add(HumanType(y.Type))
				}
				// at least 1 name always presented
				if len(y.Names) == 0 {
					add(HumanType(y.Type))
				}
			}
		}
		return true
	})

	return res, nil
}

func HumanType(tp ast.Expr) string {
	switch t := tp.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%v.%v\n", t.X, t.Sel.Name)
	case *ast.StarExpr:
		return HumanType(t.X)
	case *ast.ArrayType:
		return HumanType(t.Elt)
	case *ast.IndexExpr:
		return HumanType(t.X)
	case *ast.IndexListExpr:
		return HumanType(t.X)
	case *ast.ParenExpr:
		return HumanType(t.X)
	case *ast.FuncType:
		return "anon_func_title"
	case *ast.ChanType:
		return "chan"
	case *ast.MapType:
		return "map"
	case *ast.StructType:
		return "anon_struct"
	case *ast.InterfaceType:
		return "anon_interface"
	case *ast.Ellipsis:
		return HumanType(t.Elt)
	}
	return fmt.Sprintf("%T", tp)
}
