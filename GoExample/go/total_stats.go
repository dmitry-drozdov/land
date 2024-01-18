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
	ts := AnalyzerFuncStats{}
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

		var stats AnalyzerFuncStats
		if err := json.Unmarshal(bytes, &stats); err != nil {
			return err
		}

		ts.Add(stats)
		cnt++

		return nil
	})

	fmt.Printf(`total files: [%v], skipped files: [%v], ok methods: [%v], fail methods: [%v], accuracy: [%.3f%%], args cover: [%.2f%%], vendors ratio: [%.2f%%], duplicates: [%v]
struct accuracy: [%.3f%%], missed struct: [%.3f%% (has anon func [%.3f%%])], incorrect struct: [%.3f%%]`,
		ts.TotalFiles,
		ts.SkippedFiles,
		ts.Ok,
		ts.Fail,
		float64(ts.Ok)/float64(ts.Ok+ts.Fail)*100,
		ts.ArgsCover/RoundedFloat(cnt),
		ts.VendorFuncsPerCent/RoundedFloat(cnt),
		ts.Duplicates,
		float64(ts.StructStats.Ok)/float64(ts.StructStats.Total())*100,
		float64(ts.StructStats.FailNotFound)/float64(ts.StructStats.Total())*100,
		float64(ts.StructStats.FailNotFoundHasFunc)/float64(ts.StructStats.FailNotFound)*100,
		float64(ts.StructStats.FailIncorrectTypes)/float64(ts.StructStats.Total())*100,
	)

	return err
}
