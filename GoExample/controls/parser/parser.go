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

		start := fset.Position(x.Body.Pos())
		end := fset.Position(x.Body.End())

		// более быстрый вариант nodeText := string(src[start.Offset+1:end.Offset-1])
		nodeText := unsafe.String(&src[start.Offset+1], end.Offset-start.Offset-2)
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
		controls := &datatype.Control{Type: "root"}
		p.innerInspectControls(x.Body, controls)

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
