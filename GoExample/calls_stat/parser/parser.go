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
)

type Parser struct {
	*ast_type.NameConverter
	Counter    uint64
	FilesCache map[string]struct{}
	Dups       uint64
}

func NewParser(fc map[string]struct{}) *Parser {
	return &Parser{
		ast_type.NewNameConverter(),
		0,
		fc,
		0,
	}
}

func (p *Parser) ParseFiles(root string) (map[string]int, error) {
	res := concurrency.NewSaveMapNumeric[string, int](20000)

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

func (p *Parser) ParseFile(path string, pathOut string, res *concurrency.SaveMapNumeric[string, int]) error {
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

		if x.Recv == nil || len(x.Recv.List) == 0 {
			return true // не ресивер
		}

		start := fset.Position(x.Body.Pos())
		end := fset.Position(x.Body.End())
		nodeText := string(src[start.Offset:end.Offset])
		if len(nodeText) < 3 {
			return true // функция с пустым телом
		}

		if p.Dub(nodeText) {
			return true
		}

		// проход по МЕТОДУ в поиске ОБЫЧНЫХ ВЫЗОВОВ
		allCnt := p.innerInspectPureCalls(x.Body)

		key := fmt.Sprint(f.Name.Name, ".", p.HumanType(x.Recv.List[0].Type))

		res.Add(key, allCnt)
		res.Add(key+".cnt", 1)

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
			case *ast.IndexExpr:
				cnt += 1 + p.innerInspectPureCalls(y.Index)
				return false //array[GetIndex()]()
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
