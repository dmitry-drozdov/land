package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"utils/hash"

	"golang.org/x/sync/errgroup"
)

func deleteDups(root string) (int, error) {
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)

	alreadyDone := make(map[uint64]struct{}, 10000)
	duplicates := 0
	m := sync.Mutex{}

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

			content, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			err = file.Close()
			if err != nil {
				return err
			}

			key := hash.HashFile(content)

			m.Lock()
			defer m.Unlock()
			_, ok := alreadyDone[key]
			if ok {
				duplicates++
				return os.Remove(pathBk)
			}
			alreadyDone[key] = struct{}{}

			return nil
		})

		return nil
	})
	if err := g.Wait(); err != nil {
		return 0, err
	}
	if err != nil {
		return 0, err
	}

	return duplicates, nil

}
