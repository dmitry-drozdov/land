package provider

import (
	"brackets/node"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ReadFolder(root string) (map[string]*node.Node, error) {
	nodes := make(map[string]*node.Node, 200)

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

		node := &node.Node{}
		err = json.Unmarshal(bytes, &node)
		if err != nil {
			return err
		}

		nodes[strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))] = node

		return nil
	})
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
