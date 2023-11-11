package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func GetTotalStats(root string) error {
	cnt := 0
	ts := AnalyzerStats{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != ".json" {
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

		bytes, err := io.ReadAll(readFile)
		if err != nil {
			return err
		}

		var stats AnalyzerStats
		if err := json.Unmarshal(bytes, &stats); err != nil {
			return err
		}

		ts.Add(stats)
		cnt++

		return nil
	})

	fmt.Printf("total files: [%v], skipped files: [%v], accuracy: [%v], args cover: [%v]", ts.TotalFiles, ts.SkippedFiles, ts.Accuracy/RoundedFloat(cnt), ts.ArgsCover/RoundedFloat(cnt))

	return err
}
