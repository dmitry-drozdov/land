package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

func ParseFiles(root string) (map[string]map[string]*FuncStat, map[string]map[string]*StructStat, int, error) {
	resFun := make(map[string]map[string]*FuncStat, 10000)
	resStruct := make(map[string]map[string]*StructStat, 10000)
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)
	l := sync.Mutex{}

	alreadyDone := make(map[uint64]struct{}, 10000)
	duplicates := 0
	m := sync.Mutex{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" {
			return nil
		}

		pathBk := path

		g.Go(func() error {
			file, err := os.Open(pathBk)
			if err != nil {
				return err
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			key := HashFile(content)

			m.Lock()
			defer m.Unlock()
			_, ok := alreadyDone[key]
			if ok {
				duplicates++
				return nil
			}
			alreadyDone[key] = struct{}{}

			return nil
		})

		g.Go(func() error {
			dataFunc, dataStruct, err := ParseFile(pathBk)
			if err != nil {
				return err
			}

			pathBk = strings.ReplaceAll(pathBk, root, "")
			pathBk = strings.ReplaceAll(pathBk, `\`, "")
			l.Lock()
			resFun[pathBk] = dataFunc
			resStruct[pathBk] = dataStruct
			l.Unlock()
			return nil
		})

		return nil
	})
	g.Wait()
	if err != nil {
		return nil, nil, 0, err
	}

	if len(resFun) == 0 {
		return nil, nil, 0, fmt.Errorf("empty output")
	}

	return resFun, resStruct, duplicates, nil
}

func ParseFile(path string) (map[string]*FuncStat, map[string]*StructStat, error) {
	resFun := make(map[string]*FuncStat, 10000)
	resStruct := make(map[string]*StructStat, 10000)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, nil, err
	}

	type pair struct{ start, end token.Pos }
	funcPos := []pair{} // to filter type declared inside functions

	checkInside := func(start, end token.Pos) bool {
		for _, pos := range funcPos {
			if start > pos.start && end < pos.end {
				return true
			}
		}
		return false
	}

	ast.Inspect(f, func(n ast.Node) bool {
		if n != nil && checkInside(n.Pos(), n.End()) {
			return true
		}

		switch x := n.(type) {
		case *ast.TypeSpec:
			switch y := x.Type.(type) {
			case *ast.StructType:
				resStruct[x.Name.Name] = &StructStat{Name: x.Name.Name, Types: []string{}}
				for _, f := range y.Fields.List {
					resStruct[x.Name.Name].Types = append(resStruct[x.Name.Name].Types, HumanType(f.Type))
				}
			}

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
			resFun[x.Name.Name] = ptr

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

			funcPos = append(funcPos, pair{x.Pos(), x.End()})

		case *ast.FuncLit:
			funcPos = append(funcPos, pair{x.Pos(), x.End()})
		}
		return true
	})

	return resFun, resStruct, nil
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

func HashFile(bytes []byte) uint64 {
	h := fnv.New64a()
	h.Write([]byte(bytes))
	return h.Sum64()
}
