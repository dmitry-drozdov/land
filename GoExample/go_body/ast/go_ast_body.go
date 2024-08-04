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
	"utils/ast_type"
	"utils/concurrency"
	"utils/filter"
	"utils/hash"
	"utils/inspect"

	"golang.org/x/exp/rand"
	"golang.org/x/sync/errgroup"
)

const (
	maxFiles = 300
)

type Parser struct {
	*ast_type.NameConverter
}

func NewParser() *Parser {
	return &Parser{ast_type.NewNameConverter()}
}

func (p *Parser) ParseFilesBodies(root string) (map[string]*inspect.Node, error) {
	paths := make([]string, 0, 1000)
	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	rand.Seed(2024)
	rand.Shuffle(len(paths), func(i, j int) {
		paths[i], paths[j] = paths[j], paths[i]
	})

	ln := len(paths)
	take := min(ln, maxFiles)
	paths = paths[:take]
	fmt.Printf("\t take %.2f%% of data\n", float64(take)/float64(ln)*100)

	nodes := concurrency.NewSaveMap[string, *inspect.Node](1000)

	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)

	for _, path := range paths {
		pathBk := path
		g.Go(func() error {
			return p.ParseFileBodies(pathBk, strings.Replace(pathBk, `\test_repos\`, `\test_repos_body\`, 1), nodes)
		})
	}

	if gErr := g.Wait(); gErr != nil {
		return nil, gErr
	}

	return nodes.Unsafe(), nil
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
		pathOut := pathOut[:len(pathOut)-3] + suffix

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
		nodes.Set(strings.TrimSuffix(pathOut, ".go"), node)

		return true
	})
	if err != nil {
		return
	}

	return
}
