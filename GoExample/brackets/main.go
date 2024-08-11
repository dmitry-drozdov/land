package main

import (
	_ "brackets/node"
	"brackets/generate"
	"fmt"
)

func main() {
	root := `e:\phd\test_repos_brackets\`
	res, err := generate.GenerateCombinations()
	if err != nil {
		panic(err)
	}
	fmt.Println(len(res.CodeToText), len(res.CodeToNode))

	err = Dump(root, res.CodeToText)
	if err != nil {
		panic(err)
	}
}
