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
	"utils/ast_type"
	"utils/filter"

	"golang.org/x/sync/errgroup"
)

type goAST struct {
	*ast_type.NameConverter
}

func NewGoAST() *goAST {
	return &goAST{ast_type.NewNameConverter()}
}

func (a *goAST) ParseFiles(root string) (map[string]map[string]*FuncStat, map[string]map[string]*StructStat, int, error) {
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

	totalNodes := 0

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
			nodesAll, f, dataFunc, dataStruct, err := a.ParseFile(pathBk)
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
			totalNodes += nodesAll
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

	runtime.GC()
	runtime.ReadMemStats(&m2)
	if diff := m2.TotalAlloc - m1.TotalAlloc; diff > maxMemory {
		maxMemory = diff // no need mutex run in 1 goroutine
	}

	files[0] = nil // do something with f to not be GCollected

	fmt.Printf("%.2f sec.\n", float64(timeSpent)/(float64(time.Second)))
	fmt.Printf("%d Mb \n", maxMemory/(1024*1024))
	fmt.Printf("%d / %d req/all nodes (%.2f%%)\n", a.ReqCnt(), totalNodes, ratio(a.ReqCnt(), totalNodes))

	return resFun, resStruct, duplicates, nil
}

func (a *goAST) ParseFile(path string) (int, *ast.File, map[string]*FuncStat, map[string]*StructStat, error) {
	resFun := make(map[string]*FuncStat, 100)
	resStruct := make(map[string]*StructStat, 100)

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return 0, nil, nil, nil, err
	}

	nested := filter.NewNestedFuncs()

	allCnt := 0

	ast.Inspect(f, func(n ast.Node) bool {
		allCnt++
		if n != nil && nested.Nested(n.Pos(), n.End()) {
			return true
		}

		switch x := n.(type) {
		case *ast.TypeSpec:
			switch y := x.Type.(type) {
			case *ast.StructType:
				resStruct[x.Name.Name] = &StructStat{Name: x.Name.Name, Types: make([]string, 0, 5)}
				for _, f := range y.Fields.List {
					for i := 0; i < len(f.Names); i++ { // a, b, c int => append 3 int
						resStruct[x.Name.Name].Types = append(resStruct[x.Name.Name].Types, a.HumanType(f.Type))
					}
					if len(f.Names) == 0 { // embedded type
						resStruct[x.Name.Name].Types = append(resStruct[x.Name.Name].Types, a.HumanType(f.Type))
					}
				}
			}

		case *ast.FuncDecl:
			ptr := &FuncStat{
				Name:    x.Name.Name,
				ArgsCnt: byte(x.Type.Params.NumFields()),
				Args:    make([]string, 0, 3),
			}
			if x.Type.Results != nil {
				ptr.Return = byte(x.Type.Results.NumFields())
			}
			if x.Recv != nil && len(x.Recv.List) > 0 {
				ptr.Receiver = a.HumanType(x.Recv.List[0].Type)
			}
			resFun[x.Name.Name] = ptr

			//reqCnt += 1 + int(ret) ???

			add := func(a string) {
				ptr.Args = append(ptr.Args, a)
			}

			for _, y := range x.Type.Params.List {
				// can be several args with 1 type: n int, j, k, l float
				for i := 0; i < len(y.Names); i++ {
					add(a.HumanType(y.Type))
				}
				// at least 1 name always presented
				if len(y.Names) == 0 {
					add(a.HumanType(y.Type))
				}
				// only types (=0) or mixed only types + name with type (=1)
				if len(y.Names) != 1 {
					ptr.RequirePostProcess = true
				}
			}
			if x.Body == nil {
				ptr.NoBody = true
			}

			nested.Add(x)

		case *ast.FuncLit:
			nested.Add(x)
		}
		return true
	})

	return allCnt, f, resFun, resStruct, nil
}
