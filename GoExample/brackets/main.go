package main

import "fmt"

func main() {
	root := `e:\phd\test_repos_brackets\`
	comb := generateCombinations()
	fmt.Println(len(comb))
	err := Dump(root, comb)
	if err != nil {
		panic(err)
	}
}
