package main

import (
	"calls/parser"
	"fmt"
	"sync"
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
	var wg sync.WaitGroup
	wg.Add(len(folders))
	for _, f := range folders {
		go func() {
			defer wg.Done()
			if err := doWork(f); err != nil {
				fmt.Printf("[%v] <ERROR>: [%v]\n", f, err)
			}
		}()
	}
	wg.Wait()
}

func doWork(sname string) error {
	fmt.Printf("===== %s START =====\n", sname)

	source := fmt.Sprintf(`e:\phd\test_repos\%s\`, sname)
	res, err := parser.NewParser().ParseFiles(source)
	if err != nil {
		return err
	}

	fmt.Printf("===== %s END [%v]=====\n", sname, sum(res))
	return nil
}

func sum(mp map[string]int) int {
	res := 0
	for _, v := range mp {
		res += v
	}
	return res
}
