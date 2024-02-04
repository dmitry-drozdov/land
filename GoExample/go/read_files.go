package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/segmentio/encoding/json"

	"golang.org/x/sync/errgroup"
)

func ReadResults(root string) (map[string]map[string]*FuncStat, map[string]map[string]*StructStat, error) {
	resFun := make(map[string]map[string]*FuncStat, 10000)
	resStruct := make(map[string]map[string]*StructStat, 10000)
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
			if _, ok := resFun[pathBk]; !ok {
				resFun[pathBk] = make(map[string]*FuncStat, 10)
			}
			if _, ok := resStruct[pathBk]; !ok {
				resStruct[pathBk] = make(map[string]*StructStat, 10)
			}
			mx.Unlock()

			for fileScanner.Scan() {
				line := fileScanner.Text()

				mx.Lock()
				if strings.Contains(line, "ArgsCnt") { // func
					ln := &FuncStat{}
					if err := json.Unmarshal([]byte(line), ln); err != nil {
						return err
					}
					resFun[pathBk][ln.Name] = ln
				} else { // struct
					ln := &StructStat{}
					if err := json.Unmarshal([]byte(line), ln); err != nil {
						return err
					}
					resStruct[pathBk][ln.Name] = ln
				}
				mx.Unlock()
			}
			return nil
		})

		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	if len(resFun) == 0 {
		return nil, nil, fmt.Errorf("empty output")
	}
	return resFun, resStruct, err
}
