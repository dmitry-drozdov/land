package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"utils/filter"

	"golang.org/x/sync/errgroup"
)

func (a *goAST) ParseFilesBodies(root string) error {
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)
	//l := sync.Mutex{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" {
			return nil
		}

		pathBk := path

		g.Go(func() error {
			err := a.ParseFileBodies(pathBk)
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

	return nil
}

func (a *goAST) ParseFileBodies(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return err
	}

	nested := filter.NewNestedFuncs()

	pathOut := strings.Replace(path, `\test_repos\`, `\test_repos_body\`, 1)

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

		return true
	})
	if err != nil {
		return err
	}

	return nil
}
