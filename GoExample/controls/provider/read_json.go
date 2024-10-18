package provider

import (
	"context"
	"controls/datatype"
	"io"
	"os"
	"path/filepath"
	"strings"
	"utils/concurrency"

	"utils/tracer"

	jsoniter "github.com/json-iterator/go"
	"golang.org/x/sync/errgroup"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func ReadFolder(ctx context.Context, root string) (map[string]*datatype.Control, error) {
	_, end := tracer.Start(ctx, "ReadFolder")
	defer end(nil)

	res := concurrency.NewSaveMap[string, *datatype.Control](200000)
	g := errgroup.Group{}
	g.SetLimit(8)

	_ = filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info == nil || info.IsDir() || filepath.Ext(info.Name()) != ".json" {
			return nil
		}

		name, pathBk := info.Name(), path

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

			val := &datatype.Control{}
			err = json.Unmarshal(bytes, val)
			if err != nil {
				return err
			}

			res.Set(strings.TrimSuffix(name, ".json"), val)
			return nil
		})

		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return res.Unsafe(), nil
}