package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"utils/ast_type"
	"utils/concurrency"
	"utils/filter"
	"utils/hash"
)

type Parser struct {
	*ast_type.NameConverter
	Balancer *concurrency.Balancer
}

func NewParser(balancer *concurrency.Balancer) *Parser {
	return &Parser{ast_type.NewNameConverter(), balancer}
}

func (p *Parser) ParseFiles(root string) (map[string]int, error) {
	res := concurrency.NewSaveMap[string, int](20000)

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
	if err != nil {
		return nil, err
	}

	return res.Unsafe(), nil
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

	ai := 0
	autoInc := func() int { ai++; return ai }

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
		nodeText := string(src()[start.Offset:end.Offset])
		if len(nodeText) < 3 {
			return true // функция с пустым телом
		}

		var suffix string
		if x.Recv != nil && len(x.Recv.List) > 0 {
			suffix = fmt.Sprint("_", hash.HashStrings(p.HumanType(x.Recv.List[0].Type), x.Name.Name), ".go")
		} else {
			suffix = fmt.Sprint("_", hash.HashString(x.Name.Name), "_", autoInc(), ".go")
		}

		pathOut := pathOut[:len(pathOut)-3] + suffix

		//проход по МЕТОДУ в поиске АНОНИМНЫХ ФУНКЦИЙ
		//allCnt := p.innerInspectAnonCalls(x.Body)

		// проход по МЕТОДУ в поиске ОБЫЧНЫХ ВЫЗОВОВ
		allCnt := p.innerInspectPureCalls(x.Body)

		if allCnt == 0 && !p.Balancer.CanSubAction() {
			return true
		}

		// if suffix == "_2271423108252276541.go" {
		// 	ast.Print(fset, x.Body)
		// }

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

		p.Balancer.MainAction(allCnt)

		key := strings.Split(strings.TrimSuffix(pathOut, ".go"), "\\")
		res.Set(key[len(key)-1], allCnt)

		return true
	})

	return nil
}

// func (p *Parser) innerInspectAnonCalls(root ast.Node) int {
// 	cnt := 0
// 	ast.Inspect(root, func(n ast.Node) bool {
// 		switch x := n.(type) {
// 		case *ast.CallExpr:
// 			fn, ok := x.Fun.(*ast.FuncLit)
// 			if ok {
// 				cnt++
// 				if fn.Type != nil && fn.Type.Results != nil {
// 					cnt += p.innerInspectAnonCalls(fn.Type.Results)
// 				}
// 			}
// 			return true // continue
// 		case *ast.FuncLit:
// 			return false // interrupt, не анализируем тела вложенных функций (это не вызов, а переменная)
// 		default:
// 			return true // continue
// 		}
// 	})
// 	return cnt
// }

func (p *Parser) innerInspectPureCalls(root ast.Node) int {
	cnt := 0
	ast.Inspect(root, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			_, ok := x.Fun.(*ast.FuncLit)
			if ok {
				return true // тело внутри анонимной функции тоже просматриваем для удобства тестирования
			}

			_, ok = x.Fun.(*ast.CallExpr)
			if ok {
				cnt++
				return false // interrupt, кейс f()()()
			}

			_, ok = x.Fun.(*ast.ParenExpr)
			if ok {
				return false // interrupt, кейс (*int)(&c)
			}

			sel, ok := x.Fun.(*ast.SelectorExpr)
			if ok {
				cnt += 1 + p.innerInspectPureCalls(sel.X)
				return false // interrupt, кейс a.f(x).g(y)
			}

			cnt++
			return false // interrupt, внутренние вызовы нам не интересны
		case *ast.FuncLit:
			return true // continue, анализируем тела вложенных функций (внутри мб вызов)
		default:
			return true // continue
		}
	})
	return cnt
}
