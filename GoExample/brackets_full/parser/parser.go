package parser

import (
	"brackets/datatype"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"utils/ast_type"
	"utils/concurrency"
	"utils/filter"
	"utils/hash"
)

type Parser struct {
	*ast_type.NameConverter
	Queue      *concurrency.Queue
	Counter    uint64
	FilesCache map[string]struct{}
	Dups       uint64
}

func NewParser(fc map[string]struct{}) *Parser {
	return &Parser{
		ast_type.NewNameConverter(),
		concurrency.NewQueue(),
		0,
		fc,
		0,
	}
}

func (p *Parser) ParseFiles(root string) (map[string]*datatype.Brackets, error) {
	res := concurrency.NewSaveMap[string, *datatype.Brackets](20000)

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" { /* ||
			strings.Contains(path, `\mock`) ||
			strings.Contains(path, `\generate`) ||
			strings.Contains(path, `\fake`) ||
			strings.Contains(path, `test\`) ||
			strings.Contains(info.Name(), "mock") ||
			strings.Contains(info.Name(), "generate") ||
			strings.Contains(info.Name(), `fake`) ||
			strings.Contains(info.Name(), "test") {*/
			return nil
		}

		err := p.ParseFile(path, strings.Replace(path, `\test_repos\`, `\test_repos_calls\`, 1), res)
		if err != nil {
			return fmt.Errorf("%v: %w", path, err)
		}
		return nil
	})
	p.Queue.Wait()
	if err != nil {
		return nil, err
	}

	return res.Unsafe(), nil
}

func (p *Parser) ParseFile(path string, pathOut string, res *concurrency.SaveMap[string, *datatype.Brackets]) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "", src, 0)
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

		if x.Body == nil {
			return true // функция без тела
		}

		start := fset.Position(x.Body.Pos())
		end := fset.Position(x.Body.End())
		nodeText := string(src[start.Offset:end.Offset])
		if len(nodeText) < 3 {
			return true // функция с пустым телом
		}

		var suffix uint64
		if x.Recv != nil && len(x.Recv.List) > 0 {
			suffix = hash.HashStrings(p.HumanType(x.Recv.List[0].Type), x.Name.Name)
		} else {
			suffix = hash.HashString(x.Name.Name)
		}

		if p.Dub(nodeText) {
			return true
		}

		pathOut := fmt.Sprint(pathOut[:len(pathOut)-3], "_", suffix, "_", p.AutoInc(), ".go")

		key := strings.Split(strings.TrimSuffix(pathOut, ".go"), "\\")
		fname := key[len(key)-1]

		p.Queue.Add(func() error {
			//brackets := &datatype.Brackets{}
			//p.innerInspectPureCalls(x.Body, brackets)

			nodeText = nodeText[1 : len(nodeText)-1]
			brackets := p.innerInspectPureCallsV2(nodeText)

			// if suffix == "_10367583230383768386_1923.go" {
			// 	ast.Print(fset, x.Body)
			// 	panic(0)
			// }

			res.Set(fname, brackets)

			err = os.MkdirAll(filepath.Dir(pathOut), 0755)
			if err != nil {
				return err
			}

			var file *os.File
			file, err = os.OpenFile(pathOut, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = file.WriteString(nodeText)
			return err
		})

		return true
	})

	return nil
}

func (p *Parser) innerInspectPureCallsV2(nodeText string) *datatype.Brackets {
	fset := token.NewFileSet()
	file := fset.AddFile("example.go", fset.Base(), len(nodeText))

	var s scanner.Scanner
	s.Init(file, []byte(nodeText), nil, scanner.ScanComments)

	root := &datatype.Brackets{
		Depth:    0,
		Children: []*datatype.Brackets{},
	}

	stack := []*datatype.Brackets{root}

	for {
		_, tok, _ := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.LBRACE {
			depth := len(stack)
			bracket := &datatype.Brackets{
				Depth:    depth,
				Children: make([]*datatype.Brackets, 0, 4),
			}

			current := stack[len(stack)-1]
			current.Children = append(current.Children, bracket)

			stack = append(stack, bracket)

		} else if tok == token.RBRACE {
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			} else {
				fmt.Println("Несоответствующая закрывающая скобка на позиции")
			}
		}
	}

	return root
}

func (p *Parser) AutoInc() uint64 {
	p.Counter++
	return p.Counter
}

func (p *Parser) Dub(str string) bool {
	re := regexp.MustCompile(`[\s]`) // to unify files formatting
	str = re.ReplaceAllString(str, "")

	if _, ok := p.FilesCache[str]; ok {
		p.Dups++
		return true
	}
	p.FilesCache[str] = struct{}{}
	return false
}
