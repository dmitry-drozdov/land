package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

func ReadResults(root string) (map[string]map[string]*FuncStat, error) {
	res := make(map[string]map[string]*FuncStat, 10000)
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)
	mx := sync.Mutex{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info == nil || info.IsDir() {
			return nil
		}

		pathBk := path

		g.Go(func() error {
			readFile, err := os.Open(pathBk)
			if err != nil {
				return err
			}
			defer readFile.Close()

			fileScanner := bufio.NewScanner(readFile)
			fileScanner.Split(bufio.ScanLines)

			pathBk = strings.ReplaceAll(pathBk, root, "")
			pathBk = strings.ReplaceAll(pathBk, `\`, "")
			pathBk = strings.ReplaceAll(pathBk, ".json", ".go")

			mx.Lock()
			_, ok := res[pathBk]
			if !ok {
				res[pathBk] = make(map[string]*FuncStat, 10)
			}
			mx.Unlock()

			for fileScanner.Scan() {
				line := fileScanner.Text()
				ln := &FuncStat{}
				if err := json.Unmarshal([]byte(line), ln); err != nil {
					return err
				}

				mx.Lock()
				res[pathBk][ln.Name] = ln
				mx.Unlock()
			}
			return nil
		})

		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("empty output")
	}
	return res, err
}
