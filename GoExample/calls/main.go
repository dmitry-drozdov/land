package main

import (
	"calls/parser"
	"calls/provider"
	"errors"
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
	"test",
	"backend",
	"azure-service-operator",
	"kubernetes",
	"go-redis",
	"docker-ce",
	"tidb",
	"moby",
}

var stats = struct {
	ok    int
	total int
}{}

func main() {
	for _, f := range folders {
		if err := doWork(f); err != nil {
			fmt.Printf("[%v] <ERROR>: [%v]\n", f, err)
			//panic(-1)
		}
	}

	fmt.Printf("TOTAL ratio: %.3f\n", float64(stats.ok)/float64(stats.total)*100)
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

	err = compareMaps(orig, land)
	if err != nil {
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
	var errs []error
	if len(orig) != len(land) {
		errs = append(errs, fmt.Errorf("len mismatch %v %v", len(orig), len(land)))
	}

	okCnt := 0
	for origK, origV := range orig {
		landV, ok := land[origK]
		if !ok {
			errs = append(errs, fmt.Errorf("key not found %v", origK))
			continue
		}
		if landV != origV {
			errs = append(errs, fmt.Errorf("val mismatch [%v] [%v] [%v]", landV, origV, origK))
			continue
		}
		okCnt++
	}

	fmt.Printf("ratio: %.2f\n", float64(okCnt)/float64(len(orig))*100)
	stats.ok += okCnt
	stats.total += len(orig)

	return errors.Join(errs...)
}
