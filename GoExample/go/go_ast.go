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
	"time"

	"golang.org/x/sync/errgroup"
)

func ParseFiles(root string) (map[string]map[string]*FuncStat, map[string]map[string]*StructStat, int, error) {
	resFun := make(map[string]map[string]*FuncStat, 5000)
	resStruct := make(map[string]map[string]*StructStat, 5000)
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)
	l := sync.Mutex{}

	duplicates := 0

	timeSpent := int64(0)
	maxMemory := uint64(0)

	runtime.GC()
	files := make([]*ast.File, 0, 1000)
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// alreadyDone := make(map[uint64]struct{}, 10000)
	// m := sync.Mutex{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" {
			return nil
		}

		pathBk := path

		// uncomment when need to check duplicates
		// g.Go(func() error {
		// 	file, err := os.Open(pathBk)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	defer file.Close()

		// 	content, err := io.ReadAll(file)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	key := HashFile(content)

		// 	m.Lock()
		// 	defer m.Unlock()
		// 	_, ok := alreadyDone[key]
		// 	if ok {
		// 		duplicates++
		// 		return nil
		// 	}
		// 	alreadyDone[key] = struct{}{}

		// 	return nil
		// })

		g.Go(func() error {
			t0 := time.Now()
			f, dataFunc, dataStruct, err := ParseFile(pathBk)
			if err != nil {
				return err
			}

			files = append(files, f) // do something with f to not be GCollected

			pathBk = strings.ReplaceAll(pathBk, root, "")
			pathBk = strings.ReplaceAll(pathBk, `\`, "")
			l.Lock()
			resFun[pathBk] = dataFunc
			resStruct[pathBk] = dataStruct
			timeSpent += int64(time.Since(t0))
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

	runtime.ReadMemStats(&m2)
	if diff := m2.TotalAlloc - m1.TotalAlloc; diff > maxMemory {
		maxMemory = diff // no need mutex run in 1 goroutine
	}

	fmt.Printf("%.2f sec.\n", float64(timeSpent)/(float64(time.Second)))
	fmt.Printf("%d Mb \n", maxMemory/(1024*1024))
	fmt.Printf("%d ast objs\n", cnt(files))

	return resFun, resStruct, duplicates, nil
}

func ParseFile(path string) (*ast.File, map[string]*FuncStat, map[string]*StructStat, error) {
	resFun := make(map[string]*FuncStat, 100)
	resStruct := make(map[string]*StructStat, 100)

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, nil, nil, err
	}

	type pair struct{ start, end token.Pos }
	funcPos := make([]pair, 0, 100) // to filter type declared inside functions

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
				resStruct[x.Name.Name] = &StructStat{Name: x.Name.Name, Types: make([]string, 0, 5)}
				for _, f := range y.Fields.List {
					for i := 0; i < len(f.Names); i++ { // a, b, c int => append 3 int
						resStruct[x.Name.Name].Types = append(resStruct[x.Name.Name].Types, HumanType(f.Type))
					}
					if len(f.Names) == 0 { // embedded type
						resStruct[x.Name.Name].Types = append(resStruct[x.Name.Name].Types, HumanType(f.Type))
					}
				}
			}

		case *ast.FuncDecl:
			ret := byte(0)
			if x.Type.Results != nil {
				ret = byte(x.Type.Results.NumFields())
			}

			ptr := &FuncStat{
				Name:    x.Name.Name,
				ArgsCnt: byte(x.Type.Params.NumFields()),
				Return:  ret,
				Args:    make([]string, 0, 3),
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
				// only types (=0) or mixed only types + name with type (=1)
				if len(y.Names) != 1 {
					ptr.RequirePostProcess = true
				}
			}
			if x.Body == nil {
				ptr.NoBody = true
			}

			funcPos = append(funcPos, pair{x.Pos(), x.End()})

		case *ast.FuncLit:
			funcPos = append(funcPos, pair{x.Pos(), x.End()})
		}
		return true
	})

	return f, resFun, resStruct, nil
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

func cnt(files []*ast.File) int {
	res := 0
	for _, f := range files {
		res += len(f.Comments)
		res += len(f.Decls)
		res += len(f.Imports)
		res += len(f.Unresolved)
	}
	return res
}
