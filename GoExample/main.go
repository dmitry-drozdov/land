package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

type FuncStat struct {
	Name   string
	Args   int
	Return int
}

func main() {

	light, err := ReadResults(`e:\phd\my\results\`)
	if err != nil {
		panic(err)
	}

	fmt.Println("read results done")

	source := `e:\phd\my\docker-ce\`

	full, err := ParseFiles(source)
	if err != nil {
		panic(err)
	}

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
			if v != funcs {
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

func ParseFiles(root string) (map[string]map[string]FuncStat, error) {
	res := make(map[string]map[string]FuncStat, 10000)
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

func ParseFile(path string) (map[string]FuncStat, error) {
	res := make(map[string]FuncStat, 10000)

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

			res[x.Name.Name] = FuncStat{
				Name:   x.Name.Name,
				Args:   x.Type.Params.NumFields(),
				Return: ret,
			}
		}
		return true
	})

	return res, nil
}

func ReadResults(root string) (map[string]map[string]FuncStat, error) {
	res := make(map[string]map[string]FuncStat, 10000)
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
			res[info.Name()] = make(map[string]FuncStat, 10)
		}

		for fileScanner.Scan() {
			line := fileScanner.Text()
			words := strings.Split(line, " ")
			if len(words) != 3 {
				return fmt.Errorf("incorrect length")
			}

			nArgs, err := strconv.Atoi(words[1])
			if err != nil {
				return err
			}

			nReturn, err := strconv.Atoi(words[2])
			if err != nil {
				return err
			}

			res[info.Name()][words[0]] = FuncStat{
				Name:   words[0],
				Args:   nArgs,
				Return: nReturn,
			}
		}

		return nil
	})
	return res, err
}
