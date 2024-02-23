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

func codeStats(root string) (map[string]*CodeStats, error) {
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)

	res := make(map[string]*CodeStats, 2)
	m := sync.Mutex{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() ||
			strings.HasPrefix(path, root+`\results`) ||
			strings.Contains(path, `\vendor\`) {
			return nil
		}

		ext := filepath.Ext(info.Name())
		if len(ext) < 3 {
			return nil
		}
		ext = ext[1:]
		if ext != "go" && ext != "graphql" {
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

			_, ok := res[ext]
			if !ok {
				res[ext] = &CodeStats{
					CodeLinesCnt: cnt,
					FilesCnt:     1,
				}
				return nil
			}

			res[ext].CodeLinesCnt += cnt
			res[ext].FilesCnt++

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
