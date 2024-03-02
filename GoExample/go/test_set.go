package main

import (
	"os"
	"path/filepath"
	"strings"
)

func getFiles(root string) ([]string, error) {
	res := make([]string, 0, 3000)
	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || strings.HasPrefix(path, root+`\results`) || filepath.Ext(info.Name()) != ".go" {
			return nil
		}
		res = append(res, path)
		return nil
	})
	return res, err
}
