package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"utils/ast_type"
	"utils/concurrency"
	"utils/filter"
	"utils/hash"

	"golang.org/x/sync/errgroup"
)

type Parser struct {
	*ast_type.NameConverter
}

func NewParser() *Parser {
	return &Parser{ast_type.NewNameConverter()}
}

func (p *Parser) ParseFiles(root string) (map[string]int, error) {
	res := concurrency.NewSaveMap[string, int](1000)

	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" {
			return nil
		}

		pathBk := path

		g.Go(func() error {
			return p.ParseFile(pathBk, strings.Replace(pathBk, `\test_repos\`, `\test_repos_calls\`, 1), res)
		})

		return nil
	})
	g.Wait()
	if err != nil {
		return nil, err
	}

	return res.Unsafe(), g.Wait()
}

func (p *Parser) ParseFile(path string, pathOut string, res *concurrency.SaveMap[string, int]) error {
	var cache []byte
	src := func() []byte {
		if cache == nil {
			src, err := os.ReadFile(path)
			if err != nil {
				panic(err)
			}
			cache = src
		}
		return cache
	}

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return err
	}

	nested := filter.NewNestedFuncs()

	// проход по файлу в поисках МЕТОДОВ
	ast.Inspect(f, func(n ast.Node) bool {
		if n != nil && nested.Nested(n.Pos(), n.End()) {
			return true
		}

		switch n.(type) {
		case *ast.FuncDecl, *ast.FuncLit:
			nested.Add(n)
		}

		x, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if x.Recv == nil || len(x.Recv.List) == 0 {
			return true
		}

		suffix := fmt.Sprint("_", hash.HashStrings(p.HumanType(x.Recv.List[0].Type), x.Name.Name), ".go")
		pathOut := pathOut[:len(pathOut)-3] + suffix

		// проход по МЕТОДУ в поиске АНОНИМНЫХ ФУНКЦИЙ

		allCnt := 0
		ast.Inspect(x.Body, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				_, ok := x.Fun.(*ast.FuncLit)
				if !ok {
					return true // continue
				}
				allCnt++
				return false // interrupt
			default:
				return true // continue
			}
		})
		res.Set(strings.TrimSuffix(pathOut, ".go"), allCnt)

		if allCnt == 0 {
			return true
		}

		start := fset.Position(x.Body.Pos())
		end := fset.Position(x.Body.End())
		nodeText := string(src()[start.Offset:end.Offset])
		if len(nodeText) < 3 {
			return true
		}

		nodeText = nodeText[1 : len(nodeText)-2]

		err = os.MkdirAll(filepath.Dir(pathOut), 0755)
		if err != nil {
			return false
		}

		var file *os.File
		file, err = os.OpenFile(pathOut, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return false
		}
		defer file.Close()

		_, err = file.WriteString(nodeText)
		if err != nil {
			return false
		}

		return true
	})

	return nil
}
