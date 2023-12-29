package main

import (
	"fmt"
	"log"
)

var mp = map[string]string{
	"apollographql": `E:\phd\test_repos_graphql\git\apollographql`,
	"dgraph-io":     `E:\phd\test_repos_graphql\git\dgraph-io`,
	"wasmerio":      `E:\phd\test_repos_graphql\git\wasmerio`,
	"qmd":           `E:\phd\test_repos_graphql\qmd\`,
	"mts":           `E:\phd\test_repos_graphql\mts\`,
}

func main() {
	res, err := parseFolders()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
}
