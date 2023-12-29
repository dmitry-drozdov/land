package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

func parseLandFolders() (map[string]Result, error) {
	res := make(map[string]Result, len(results))

	for name, path := range results {
		r, err := readResults(path)
		if err != nil {
			return nil, err
		}
		if r != nil {
			res[name] = *r
		}
	}

	return res, nil
}

func readResults(root string) (*Result, error) {
	res := &Result{}
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)
	mx := sync.Mutex{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info == nil || info.IsDir() {
			return nil
		}

		pathBk := path

		g.Go(func() (err error) {
			defer func() {
				if err != nil {
					err = fmt.Errorf("path: [%v], err: [%w]", pathBk, err)
				}
			}()

			readFile, err := os.Open(pathBk)
			if err != nil {
				return err
			}
			defer readFile.Close()

			bytes, err := io.ReadAll(readFile)
			if err != nil {
				return err
			}

			fRes := &Result{}
			if err := json.Unmarshal(bytes, fRes); err != nil {
				return err
			}

			mx.Lock()
			res.Funcs = append(res.Funcs, fRes.Funcs...)
			res.Inputs = append(res.Inputs, fRes.Inputs...)
			res.Types = append(res.Types, fRes.Types...)
			mx.Unlock()

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

	if res.Empty() {
		return nil, fmt.Errorf("empty output")
	}

	return res, nil
}
