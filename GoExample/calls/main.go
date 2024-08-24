package main

import (
	"calls/parser"
	"fmt"
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
	//"test",
	"backend",
	"azure-service-operator",
	"kubernetes",
	"go-redis",
	"docker-ce",
	"tidb",
	"moby",
}

func main() {
	for _, f := range folders {
		if err := doWork(f); err != nil {
			fmt.Printf("[%v] <ERROR>: [%v]\n", f, err)
		}
	}
}

func doWork(sname string) error {
	fmt.Printf("\n===== %s START =====\n", sname)
	defer fmt.Printf("===== %s END =====\n", sname)

	source := fmt.Sprintf(`e:\phd\test_repos\%s\`, sname)
	fmt.Println("parsing files with go ast...")
	res, err := parser.NewParser().ParseFiles(source)
	if err != nil {
		return err
	}

	fmt.Printf("parsing files with go ast DONE: %v\n", sum(res))

	return nil
}

func sum(mp map[string]int) int {
	res := 0
	for _, v := range mp {
		res += v
	}
	return res
}
