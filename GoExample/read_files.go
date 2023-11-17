package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ReadResults(root string) (map[string]map[string]*FuncStat, error) {
	res := make(map[string]map[string]*FuncStat, 10000)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
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

		pathBk := path
		pathBk = strings.ReplaceAll(pathBk, root, "")
		pathBk = strings.ReplaceAll(pathBk, `\`, "")

		_, ok := res[pathBk]
		if !ok {
			res[pathBk] = make(map[string]*FuncStat, 10)
		}

		for fileScanner.Scan() {
			line := fileScanner.Text()
			ln := &FuncStat{}
			if err := json.Unmarshal([]byte(line), ln); err != nil {
				return err
			}

			res[pathBk][ln.Name] = ln
		}

		return nil
	})

	if len(res) == 0 {
		return nil, fmt.Errorf("empty output")
	}
	return res, err
}
