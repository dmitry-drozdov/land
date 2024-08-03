package provider

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"utils/concurrency"
	"utils/inspect"

	"golang.org/x/sync/errgroup"
)

func ReadFolder(root string) (map[string]*inspect.Node, error) {
	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 8)

	nodes := concurrency.NewSaveMap[string, *inspect.Node](1000)

	err := filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info == nil || info.IsDir() || filepath.Ext(info.Name()) != ".json" {
			return nil
		}

		pathBk := path

		g.Go(func() error {
			file, err := os.Open(pathBk)
			if err != nil {
				return err
			}
			defer file.Close()

			bytes, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			node := &inspect.Node{}
			err = json.Unmarshal(bytes, &node)
			if err != nil {
				return err
			}

			nodes.Set(strings.TrimSuffix(pathBk, ".json"), node)

			return nil
		})

		return nil
	})
	if gErr := g.Wait(); gErr != nil {
		return nil, gErr
	}
	if err != nil {
		return nil, err
	}

	return nodes.Unsafe(), nil
}
