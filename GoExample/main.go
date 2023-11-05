package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

type FuncStat struct {
	Name    string
	Args    []string
	ArgsCnt int
	Return  int
}

func (f *FuncStat) EqualTo(g *FuncStat) bool {
	if f == nil && g == nil {
		return true
	}
	if f == nil || g == nil {
		return false
	}
	if f.Name != g.Name || f.ArgsCnt != g.ArgsCnt || f.Return != g.Return {
		return false
	}
	if len(f.Args) != len(g.Args) {
		return false
	}
	if len(f.Args) == 0 && len(g.Args) == 0 {
		return true
	}
	sortSlice := func(s []string) {
		sort.Slice(s, func(i, j int) bool {
			return s[i] < s[j]
		})
	}
	sortSlice(f.Args)
	sortSlice(g.Args)

	for i := range f.Args {
		if !compareStrings(f.Args[i], g.Args[i]) {
			return false
		}
	}

	return true
}

func compareStrings(s1, s2 string) bool {
	s1 = strings.TrimSpace(s1)
	s1 = strings.ToLower(s1)
	s2 = strings.TrimSpace(s2)
	s2 = strings.ToLower(s2)
	return s1 == s2
}

func main() {

	fmt.Println("reading results done...")
	light, err := ReadResults(`e:\phd\my\results\`)
	if err != nil {
		panic(err)
	}
	fmt.Println("reading results DONE")

	source := `e:\phd\my\tidb\`

	fmt.Println("parsing files with go ast...")
	full, err := ParseFiles(source)
	if err != nil {
		panic(err)
	}
	fmt.Println("parsing files with go ast DONE")

	mismatch := 0
	match := 0

	for kf, vf := range full {
		kl, ok := light[kf]
		if !ok {
			continue
		}

		for k, v := range vf {
			funcs, ok := kl[k]
			if !ok {
				mismatch++
				continue
			}
			if !v.EqualTo(funcs) {
				fmt.Println()
				fmt.Println(kf, v, funcs)
				mismatch++
				continue
			}

			match++
		}
	}

	total := mismatch + match
	skipped := len(full) - len(light)
	fmt.Printf("source: [%v] skipped: [%v (%.1f%%)] fail: [%v] ok: [%v] accuracy: [%.1f%%]",
		source, skipped, ratio(skipped, len(full)), mismatch, match, ratio(match, total))
}

func ratio(part, total int) float64 {
	return float64(part) / float64(total) * 100
}

func ParseFiles(root string) (map[string]map[string]*FuncStat, error) {
	res := make(map[string]map[string]*FuncStat, 10000)
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 2)
	l := sync.Mutex{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" {
			return nil
		}
		if err != nil {
			return err
		}

		readFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer readFile.Close()

		pathBk := path
		g.Go(func() error {
			data, err := ParseFile(pathBk)
			if err != nil {
				return err
			}

			pathBk = strings.ReplaceAll(pathBk, root, "")
			pathBk = strings.ReplaceAll(pathBk, `\`, "")
			l.Lock()
			res[pathBk] = data
			l.Unlock()
			return nil
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, g.Wait()
}

func ParseFile(path string) (map[string]*FuncStat, error) {
	res := make(map[string]*FuncStat, 10000)

	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	// Inspect the AST and print all identifiers and literals.
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			ret := 0
			if x.Type.Results != nil {
				ret = x.Type.Results.NumFields()
			}

			ptr := &FuncStat{
				Name:    x.Name.Name,
				ArgsCnt: x.Type.Params.NumFields(),
				Return:  ret,
			}
			res[x.Name.Name] = ptr

			add := func(a string) {
				ptr.Args = append(ptr.Args, a)
			}

			for _, y := range x.Type.Params.List {
				add(HumanType(y.Type))
			}
		}
		return true
	})

	return res, nil
}

func HumanType(tp ast.Expr) string {
	switch t := tp.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%v.%v\n", t.X, t.Sel.Name)
	case *ast.StarExpr:
		return HumanType(t.X)
	case *ast.ArrayType:
		return HumanType(t.Elt)
	case *ast.FuncType:
		return "anon_func_title"
	case *ast.ChanType:
		return "chan"
	case *ast.MapType:
		return "map"
	case *ast.StructType:
		return "anon_struct"
	case *ast.InterfaceType:
		return "anon_interface"
	case *ast.Ellipsis:
		return HumanType(t.Elt)
	}
	return fmt.Sprintf("%T", tp)
}

func ReadResults(root string) (map[string]map[string]*FuncStat, error) {
	res := make(map[string]map[string]*FuncStat, 10000)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}

		readFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer readFile.Close()

		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)

		_, ok := res[info.Name()]
		if !ok {
			res[info.Name()] = make(map[string]*FuncStat, 10)
		}

		for fileScanner.Scan() {
			line := fileScanner.Text()
			ln := &FuncStat{}
			if err := json.Unmarshal([]byte(line), ln); err != nil {
				return err
			}

			res[info.Name()][ln.Name] = ln
		}

		return nil
	})
	return res, err
}
