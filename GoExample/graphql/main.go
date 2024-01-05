package main

import (
	"fmt"
	"log"
)

var source = map[string]string{
	"dagger":      `E:\phd\test_repos_graphql\git\dagger`,
	"sourcegraph": `E:\phd\test_repos_graphql\git\sourcegraph`,
	"apollo":      `E:\phd\test_repos_graphql\git\apollographql`,
	"dgraph":      `E:\phd\test_repos_graphql\git\dgraph-io`,
	"wasmer":      `E:\phd\test_repos_graphql\git\wasmerio`,
	"wiki":        `E:\phd\test_repos_graphql\git\wiki`,
	"qmd":         `E:\phd\test_repos_graphql\qmd\`,
	"mts":         `E:\phd\test_repos_graphql\mts\`,
}

var results = map[string]string{
	"dagger":      `E:\phd\test_repos_graphql\results\git\dagger`,
	"sourcegraph": `E:\phd\test_repos_graphql\results\git\sourcegraph`,
	"apollo":      `E:\phd\test_repos_graphql\results\git\apollographql`,
	"dgraph":      `E:\phd\test_repos_graphql\results\git\dgraph-io`,
	"wasmer":      `E:\phd\test_repos_graphql\results\git\wasmerio`,
	"wiki":        `E:\phd\test_repos_graphql\results\git\wiki`,
	"qmd":         `E:\phd\test_repos_graphql\results\qmd\`,
	"mts":         `E:\phd\test_repos_graphql\results\mts\`,
}

func main() {
	fmt.Println("===============START===============")
	defer fmt.Println("================END================")
	res, err := parseFolders()
	if err != nil {
		log.Fatal(err)
	}

	resLand, err := parseLandFolders()
	if err != nil {
		log.Fatal(err)
	}

	if len(res) != len(resLand) {
		log.Fatal("repos len mismatch")
	}

	for repoName, schema := range res {
		schemaLand, ok := resLand[repoName]
		if !ok {
			log.Fatalf("repo [%v] not found in land schemas", repoName)
		}
		if err := schema.EqualTo(&schemaLand); err != nil {
			fmt.Println(fmt.Errorf("[%s], err: [%w]", repoName, err))
		} else {
			fmt.Println(fmt.Errorf("[%s], LOC: [%d] OK", repoName, schema.LOC))
		}
	}

}
