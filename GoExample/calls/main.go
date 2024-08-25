package main

import (
	"calls/parser"
	"calls/provider"
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
	fmt.Printf("===== %s START =====\n", sname)

	source := fmt.Sprintf(`e:\phd\test_repos\%s\`, sname)
	orig, err := parser.NewParser().ParseFiles(source)
	if err != nil {
		return err
	}

	resFolder := fmt.Sprintf(`e:\phd\test_repos_calls\results\%s\`, sname)
	land, err := provider.ReadFolder(resFolder)
	if err != nil {
		return err
	}

	fmt.Printf("===== %s END [%v] [%v]=====\n", sname, sum(orig), sum(land))

	if err := compareMaps(orig, land); err != nil {
		return err
	}

	return nil
}

func sum(mp map[string]int) int {
	res := 0
	for _, v := range mp {
		res += v
	}
	return res
}

func compareMaps(orig, land map[string]int) error {
	if len(orig) != len(land) {
		fmt.Printf("len mismatch %v %v\n", len(orig), len(land))
	}

	okCnt := 0
	for origK, origV := range orig {
		landV, ok := land[origK]
		if !ok {
			fmt.Printf("key not found %v\n", origK)
			continue
		}
		if landV != origV {
			fmt.Printf("val mismatch %v %v\n", landV, origV)
			continue
		}
		okCnt++
	}

	fmt.Printf("ratio: %.2f\n", float64(okCnt)/float64(len(orig))*100)

	return nil
}
