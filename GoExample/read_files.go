package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

func ReadResults(root string) (map[string]map[string]*FuncStat, error) {
	res := make(map[string]map[string]*FuncStat, 10000)
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
			res[info.Name()] = make(map[string]*FuncStat, 10)
		}

		for fileScanner.Scan() {
			line := fileScanner.Text()
			ln := &FuncStat{}
			if err := json.Unmarshal([]byte(line), ln); err != nil {
				return err
			}

			res[info.Name()][ln.Name] = ln
		}

		return nil
	})
	return res, err
}
