package main

import (
	"fmt"
	"gobody/ast"
)

var folders = []string{
	"sourcegraph",
	"delivery-offering",
	"boost",
	"chainlink",
	"modules",
	"go-ethereum",
	"grafana",
	"gvisor",
	"backend",
	"azure-service-operator",
	"kubernetes",
	"go-redis",
	"docker-ce",
	"tidb",
	"moby",
}

func main() {
	err := generateAndParseBodies()
	if err != nil {
		panic(err)
	}
}

func generateAndParseBodies() error {
	p := ast.NewParser()
	for _, f := range folders {
		source := fmt.Sprintf(`e:\phd\test_repos\%s\`, f)
		fmt.Printf("parsing files bodies with go ast [%s]...\n", f)
		err := p.ParseFilesBodies(source)
		if err != nil {
			return err
		}
	}
	return nil
}
