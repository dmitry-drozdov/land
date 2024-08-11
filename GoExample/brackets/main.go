package main

import (
	"brackets/generate"
	"brackets/node"
	"brackets/provider"
	"fmt"
)

func main() {
	root := `e:\phd\test_repos_brackets\`
	orig, err := generate.GenerateCombinations()
	if err != nil {
		panic(err)
	}
	fmt.Printf("generated %v files and %v nodes\n", len(orig.CodeToText), len(orig.CodeToNode))

	err = Dump(root, orig.CodeToText)
	if err != nil {
		panic(err)
	}
	fmt.Printf("dump to %v\n", root)

	resFolder := root + `results\`
	landNodes, err := provider.ReadFolder(resFolder)
	if err != nil {
		panic(err)
	}
	fmt.Printf("read %v land nodes\n", len(landNodes))

	fmt.Println(node.CompareMaps(landNodes, orig.CodeToNode))
}
