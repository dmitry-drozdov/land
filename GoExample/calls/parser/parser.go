package parser

import (
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
	p.Queue.Wait()
	if err != nil {
		return nil, err
	}

	return res.Unsafe(), nil
}

func (p *Parser) ParseFile(path string, pathOut string, res *concurrency.SaveMap[string, int]) error {
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

		// проход по МЕТОДУ в поиске ОБЫЧНЫХ ВЫЗОВОВ
		allCnt := p.innerInspectPureCalls(x.Body)

		// if allCnt == 0 && !p.Balancer.CanSubAction() {
		// 	return true
		// }

		// if suffix == "_10367583230383768386_1923.go" {
		// 	ast.Print(fset, x.Body)
		// 	panic(0)
		// }

		p.Balancer.MainAction(allCnt)
		res.Set(fname, allCnt)

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
			switch y := x.Fun.(type) {
			case *ast.FuncLit:
				return true // тело внутри анонимной функции тоже просматриваем для удобства тестирования
			case *ast.CallExpr:
				cnt++
				return false // interrupt, кейс f()()()
			case *ast.ParenExpr:
				for _, arg := range x.Args {
					cnt += p.innerInspectPureCalls(arg)
				}
				cnt += p.innerInspectPureCalls(y.X)
				return false // interrupt, кейс *(*uint64)(unsafe.Pointer(&c.elemBuf[0]))
			case *ast.SelectorExpr:
				if excluded[y.Sel.Name] {
					return true
				}
				cnt += 1 + p.innerInspectPureCalls(y.X)
				return false // interrupt, кейс a.f(x).g(y)
			case *ast.MapType, *ast.InterfaceType:
				return false // interrupt, кейс map[int]string(oldMap) и interface{}(oldMap)
			case *ast.Ident:
				if excluded[y.Name] {
					return true // внешний вызов нам не подошел - продолжаем внутри
				}
			case *ast.ArrayType:
				if ident, ok := y.Elt.(*ast.Ident); ok && excluded[ident.Name] {
					return true // внешний вызов нам не подошел - продолжаем внутри
				}
				if _, ok := y.Elt.(*ast.InterfaceType); ok {
					return false // внутрь интерфейса не лезем, там нет вызовов, и []interface{}(smth) - это каст, а не вызов
				}
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

var excluded = map[string]bool{
	"bool":       true,
	"string":     true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,
	"byte":       true,
	"rune":       true,
	"float32":    true,
	"float64":    true,
	"complex64":  true,
	"complex128": true,
	"close":      true,
	"len":        true,
	"cap":        true,
	"copy":       true,
	"delete":     true,
	"complex":    true,
	"real":       true,
	"imag":       true,
	"new":        true,
	"make":       true,
	"append":     true,
	"panic":      true,
	"recover":    true,
	"print":      true,
	"println":    true,
}
