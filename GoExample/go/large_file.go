package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

func GenerateLargeFile(root string, out string) error {
	mx := sync.Mutex{}
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)

	all := make([]byte, 0, 1e9)

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" || strings.HasPrefix(path, root+`\results`) {
			return nil
		}

		pathBk := path

		g.Go(func() error {
			file, err := os.Open(pathBk)
			if err != nil {
				return err
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			mx.Lock()
			all = append(all, content...)
			all = append(all, '\n')
			mx.Unlock()
			return nil
		})

		return nil
	})
	if err != nil {
		return err
	}
	if err := g.Wait(); err != nil {
		return err
	}
	fmt.Println(len(all))

	return os.WriteFile(out, all, fs.ModePerm)
}

func GenerateLargeFileStandard(out string, ext string) error {
	var sb strings.Builder
	sb.Grow(1e9)

	sb.WriteString("package test\n")
	sb.WriteString(`import "os"`)

	n := 100000
	for i := 0; i < n; i++ {
		sb.WriteString(fmt.Sprintf(`
func (t *MyType%v) DoAction%v(*[]struct{}, map[<-chan int]*int) (error) {
	var a = -1 + 2
	return nil
}
		`, i, i))
	}

	return os.WriteFile(fmt.Sprint(out, ".", ext), []byte(sb.String()), fs.ModePerm)
}

func GenerateLargeFileStandardSharp(out string, ext string) error {
	var sb strings.Builder
	sb.Grow(1e9)

	n := 100000
	for i := 0; i < n; i++ {
		sb.WriteString(fmt.Sprintf(`
public IList<Some.Some.Runtime.IToken> GetAllTokens%v(ref string[] name%v, int[] type%v, List<Message> errors = null)
{
	var a = 1 + 2;
	return null;
}
		`, i, i, i))
	}

	return os.WriteFile(fmt.Sprint(out, ".", ext), []byte(sb.String()), fs.ModePerm)
}
