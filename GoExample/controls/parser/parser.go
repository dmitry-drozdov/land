package parser

import (
	"bufio"
	"context"
	"controls/datatype"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unsafe"
	"utils/ast_type"
	"utils/concurrency"
	"utils/hash"

	"utils/tracer"

	"github.com/willf/bloom"
)

type Parser struct {
	*ast_type.NameConverter
	Queue      *concurrency.Queue
	Balancer   byte
	Counter    uint64
	FilesCache *bloom.BloomFilter
	Dups       uint64
}

func NewParser(fc *bloom.BloomFilter) *Parser {
	return &Parser{
		ast_type.NewNameConverter(),
		concurrency.NewQueue(),
		0,
		0,
		fc,
		0,
	}
}

func (p *Parser) ParseFiles(ctx context.Context, root string) (map[string]*datatype.Control, error) {
	ctx, end := tracer.Start(ctx, "ParseFiles")
	defer end(nil)

	res := concurrency.NewSaveMap[string, *datatype.Control](20000)
	pathCache := concurrency.NewSaveMap[string, struct{}](20000)

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

		return p.ParseFile(ctx, path, strings.Replace(path, `\test_repos\`, `\test_repos_calls\`, 1), res, pathCache)
	})
	p.Queue.Wait()
	if err != nil {
		return nil, err
	}

	return res.Unsafe(), nil
}

func (p *Parser) ParseFile(
	ctx context.Context,
	path string,
	pathOut string,
	res *concurrency.SaveMap[string, *datatype.Control],
	pathCache *concurrency.SaveMap[string, struct{}],
) error {
	ctx, end := tracer.Start(ctx, "ParseFile")
	defer end(nil)

	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return err
	}

	once := sync.Once{}

	// проход по файлу в поисках МЕТОДОВ
	ast.Inspect(f, func(n ast.Node) bool {
		x, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if x.Body == nil {
			return true // функция без тела
		}

		start := fset.Position(x.Body.Pos()).Offset + 1 // убираем по краям { и }
		end := fset.Position(x.Body.End()).Offset - 1

		ln := end - start
		if ln < 10 {
			return true // функция с пустым или незначащим телом
		}
		nodeText := unsafe.String(&src[start], ln)

		var suffix uint64
		if x.Recv != nil && len(x.Recv.List) > 0 {
			suffix = hash.HashStrings(p.HumanType(x.Recv.List[0].Type), x.Name.Name)
		} else {
			suffix = hash.HashString(x.Name.Name)
		}

		if p.Dub(nodeText) {
			return true
		}

		once.Do(func() {
			dir := filepath.Dir(pathOut)
			if pathCache.Ok(dir) {
				return
			}
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				panic(err)
			}
			pathCache.Set(dir, struct{}{})
		})

		pathOut := fmt.Sprint(pathOut[:len(pathOut)-3], "_", suffix, "_", p.AutoInc())

		key := strings.Split(pathOut, "\\") // trim .go
		fname := key[len(key)-1]

		// проход по МЕТОДУ в поиске if/for/else
		controls := &datatype.Control{Type: "root", Children: make([]*datatype.Control, 0, 2)}
		p.innerInspectControls(x.Body, controls)

		cnt := len(controls.Children)
		if cnt == 0 { // из тех, кто без операторов, берем только 1/3, все равно ничего интересного
			p.Balancer = (p.Balancer + 1) % 3
			if p.Balancer > 0 { // 1 или 2 - пропуск, 0 - берем
				return true
			}
		}

		res.Set(fname, controls)

		p.Queue.Add(func() error {
			_, end := tracer.Start(ctx, "write to file")
			defer end(nil)

			file, err := os.OpenFile(pathOut+".go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer file.Close()

			writer := bufio.NewWriter(file)

			_, err = writer.WriteString(nodeText)
			if err != nil {
				return err
			}
			return writer.Flush()
		})

		return true
	})

	return nil
}

func (p *Parser) innerInspectControls(root ast.Node, control *datatype.Control) {
	add := func(tp string) *datatype.Control {
		child := &datatype.Control{
			Type:     tp,
			Depth:    control.Depth + 1,
			Children: make([]*datatype.Control, 0, 2),
		}
		control.Children = append(control.Children, child)
		return child
	}

	ast.Inspect(root, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt:
			child := add("if")
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

		case *ast.SwitchStmt:
			child := add("switch")
			if x.Init != nil {
				p.innerInspectControls(x.Init, child)
			}
			if x.Tag != nil {
				p.innerInspectControls(x.Tag, child)
			}
			p.innerInspectControls(x.Body, child)
			return false

		case *ast.TypeSwitchStmt:
			child := add("switch")
			if x.Init != nil {
				p.innerInspectControls(x.Init, child)
			}
			p.innerInspectControls(x.Body, child)
			return false

		case *ast.SelectStmt:
			child := add("select")
			p.innerInspectControls(x.Body, child)
			return false

		case *ast.ForStmt:
			child := add("for")
			if x.Init != nil {
				p.innerInspectControls(x.Init, child)
			}
			if x.Cond != nil {
				p.innerInspectControls(x.Cond, child)
			}
			p.innerInspectControls(x.Body, child)
			return false

		case *ast.RangeStmt:
			child := add("for")
			p.innerInspectControls(x.X, child)
			p.innerInspectControls(x.Body, child)
			return false

		case *ast.CallExpr:
			switch y := x.Fun.(type) {
			case *ast.IndexExpr:
				child := add("call")
				p.innerInspectControls(y.Index, child)
				for _, arg := range x.Args {
					p.innerInspectControls(arg, child)
				}
				return false
			case *ast.FuncLit:
				child := add("anon_func_call")
				p.innerInspectControls(y.Body, child)
				return false
			case *ast.CallExpr:
				_ = add("call")
				return false // interrupt, кейс f()()()
			case *ast.ParenExpr:
				for _, arg := range x.Args {
					p.innerInspectControls(arg, control)
				}
				p.innerInspectControls(y.X, control)
				return false // interrupt, кейс *(*uint64)(unsafe.Pointer(&c.elemBuf[0]))
			case *ast.SelectorExpr:
				if excluded[y.Sel.Name] {
					return true
				}
				child := add("call")
				p.innerInspectControls(y.X, control)
				for _, arg := range x.Args {
					p.innerInspectControls(arg, child)
				}
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

			child := add("call")
			for _, arg := range x.Args {
				p.innerInspectControls(arg, child)
			}
			return false // interrupt, внутренние вызовы нам не интересны

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
	b := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.StringData(str))), len(str))
	if p.FilesCache.TestAndAdd(b) {
		p.Dups++
		return true
	}
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