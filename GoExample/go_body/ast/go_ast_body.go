package ast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"utils/filter"
	"utils/inspect"

	"utils/concurrency"

	"utils/ast_type"
	"utils/hash"

	"golang.org/x/sync/errgroup"
)

const (
	maxFuncs = 500 // 25000
)

type Parser struct {
	*ast_type.NameConverter
}

func NewParser() *Parser {
	return &Parser{ast_type.NewNameConverter()}
}

func (p *Parser) ParseFilesBodies(root string) error {
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)

	nodes := concurrency.NewSaveMap[string, *inspect.Node](1000)

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" {
			return nil
		}

		pathBk := path

		g.Go(func() error {
			if nodes.Len() > maxFuncs {
				return nil
			}

			err := p.ParseFileBodies(pathBk, strings.Replace(pathBk, `\test_repos\`, `\test_repos_body\`, 1), nodes)
			if err != nil {
				return err
			}

			return nil
		})

		return nil
	})
	if gErr := g.Wait(); gErr != nil {
		return gErr
	}
	if err != nil {
		return err
	}

	fmt.Println(nodes.Len())

	return nil
}

func (p *Parser) ParseFileBodies(path string, pathOut string, nodes *concurrency.SaveMap[string, *inspect.Node]) (err error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return
	}
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return
	}

	nested := filter.NewNestedFuncs()

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
		pathOut := strings.Replace(pathOut, ".go", suffix, 1)

		start := fset.Position(x.Body.Pos())
		end := fset.Position(x.Body.End())
		nodeText := string(src[start.Offset:end.Offset])
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

		node := inspect.Inspect(x)
		nodes.Set(pathOut, node)

		return true
	})
	if err != nil {
		return
	}

	return
}
