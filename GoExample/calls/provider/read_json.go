package provider

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func ReadFolder(root string) (map[string]int, error) {
	res := make(map[string]int, 20000)

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info == nil || info.IsDir() || filepath.Ext(info.Name()) != ".json" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		bytes, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		var val int
		err = json.Unmarshal(bytes, &val)
		if err != nil {
			return err
		}

		res[strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))] = val

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
