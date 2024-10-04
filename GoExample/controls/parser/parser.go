package parser

import (
	"controls/datatype"
	"fmt"
	"go/ast"
	"go/parser"
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
	Balancer   *concurrency.Balancer
	Counter    uint64
	FilesCache map[string]struct{}
	Dups       uint64
}

func NewParser(balancer *concurrency.Balancer, fc map[string]struct{}) *Parser {
	return &Parser{
		ast_type.NewNameConverter(),
		concurrency.NewQueue(),
		balancer,
		0,
		fc,
		0,
	}
}

func (p *Parser) ParseFiles(root string) (map[string]*datatype.Control, error) {
	res := concurrency.NewSaveMap[string, *datatype.Control](20000)

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

		return p.ParseFile(path, strings.Replace(path, `\test_repos\`, `\test_repos_calls\`, 1), res)
	})
	p.Queue.Wait()
	if err != nil {
		return nil, err
	}

	return res.Unsafe(), nil
}

func (p *Parser) ParseFile(path string, pathOut string, res *concurrency.SaveMap[string, *datatype.Control]) error {
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

		//проход по МЕТОДУ в поиске АНОНИМНЫХ ФУНКЦИЙ
		//allCnt := p.innerInspectAnonCalls(x.Body)

		// проход по МЕТОДУ в поиске if/for/else
		controls := &datatype.Control{Type: "root"}
		p.innerInspectControls(x.Body, controls)

		// if allCnt == 0 && !p.Balancer.CanSubAction() {
		// 	return true
		// }

		// if suffix == "_10367583230383768386_1923.go" {
		// 	ast.Print(fset, x.Body)
		// 	panic(0)
		// }

		p.Balancer.MainAction(1)
		res.Set(fname, controls)

		p.Queue.Add(func() error {
			nodeText = nodeText[1 : len(nodeText)-1]

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

func (p *Parser) innerInspectControls(root ast.Node, control *datatype.Control) {
	ast.Inspect(root, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt:
			child := &datatype.Control{
				Type:     "if",
				Depth:    control.Depth + 1,
				Children: make([]*datatype.Control, 0, 2),
			}
			control.Children = append(control.Children, child)
			if x.Init != nil {
				p.innerInspectControls(x.Init, child)
			}
			if x.Cond != nil {
				p.innerInspectControls(x.Cond, child)
			}
			p.innerInspectControls(x.Body, child)
			if x.Else != nil {
				p.innerInspectControls(x.Else, child)
			}
			return false
		default:
			return true // continue
		}
	})
}

func (p *Parser) AutoInc() uint64 {
	p.Counter++
	return p.Counter
}

var re = regexp.MustCompile(`[\s]`) // to unify files formatting
func (p *Parser) Dub(str string) bool {
	str = re.ReplaceAllString(str, "")
	if _, ok := p.FilesCache[str]; ok {
		p.Dups++
		return true
	}
	p.FilesCache[str] = struct{}{}
	return false
}
