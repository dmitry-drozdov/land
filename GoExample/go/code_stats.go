package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"utils/code"

	"golang.org/x/sync/errgroup"
)

type CodeStats struct {
	CodeLinesCnt uint
	FilesCnt     uint
}

func codeStats(root string) (*CodeStats, error) {
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)

	res := &CodeStats{}
	m := sync.Mutex{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" ||
			strings.HasPrefix(path, root+`\results`) ||
			strings.Contains(path, `\vendor\`) {
			return nil
		}

		pathBk := path

		g.Go(func() error {
			file, err := os.Open(pathBk)
			if err != nil {
				return err
			}

			content, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			err = file.Close()
			if err != nil {
				return err
			}

			cnt := code.GetLOC(string(content))

			m.Lock()
			defer m.Unlock()

			res.CodeLinesCnt += cnt
			res.FilesCnt++

			return nil
		})

		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return res, nil

}
