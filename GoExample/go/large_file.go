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

func GenerateLargeFileStandard(out string) error {
	var sb strings.Builder
	sb.Grow(1e9)

	sb.WriteString("package test\n")
	sb.WriteString(`import "os"`)

	for i := 0; i < 750000; i++ {
		sb.WriteString(fmt.Sprintf(`
			type MyType%v struct {}
			func (t *MyType%v) DoAction%v(a *int, b struct{}, c interface{}, d []int, m map[<-chan int]*int) (float64, error) {
				os.Getenv("")
				return 0, nil
			}
		`, i, i, i))
	}

	return os.WriteFile(out, []byte(sb.String()), fs.ModePerm)
}
