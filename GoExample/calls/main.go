package main

import (
	"calls/parser"
	"calls/provider"
	"errors"
	"fmt"
	"time"
	"utils/concurrency"

	"github.com/fatih/color"
)

const (
	RATIO = 1
)

var folders = []string{
	"azure-service-operator",
	"kubernetes",
	"docker-ce",
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
	"go-redis",
	"tidb",
	"moby",
}

var stats = struct {
	hasCalls   int
	hasNoCalls int
	ok         int
	total      int
}{}

func main() {
	color.New(color.FgRed, color.Bold).Printf("START %v\n", time.Now().Format(time.DateTime))

	b := concurrency.NewBalancer(RATIO) // на каждые RATIO файлов с вызовами 1 файл без вызовов
	for _, f := range folders {
		if err := doWork(f, b); err != nil {
			color.New(color.FgBlack, color.Bold).Printf("[%v] <ERROR>: [%v]\n", f, err)
		}
	}

	totalFuncs := stats.hasCalls + stats.hasNoCalls
	color.Green(
		"TOTAL has calls: %v (%.2f%%), has no calls: %v (%.2f%%)\n",
		stats.hasCalls, ratio(stats.hasCalls, totalFuncs),
		stats.hasNoCalls, ratio(stats.hasNoCalls, totalFuncs),
	)
	color.Green("TOTAL func call: %v, bodies: %v\n", b.CntMain(), b.CntSub())
	color.Green("TOTAL ratio: %.3f [bad=%v]\n", ratio(stats.ok, stats.total), stats.total-stats.ok)
}

func doWork(sname string, balancer *concurrency.Balancer) error {
	color.Cyan("===== %s START =====\n", sname)

	source := fmt.Sprintf(`e:\phd\test_repos\%s\`, sname)
	orig, err := parser.NewParser(balancer).ParseFiles(source)
	if err != nil {
		return err
	}

	resFolder := fmt.Sprintf(`e:\phd\test_repos_calls\results\%s\`, sname)
	land, err := provider.ReadFolder(resFolder)
	if err != nil {
		return err
	}
	trackCalls(orig)
	color.Cyan("===== %s END [%v] [%v]=====\n", sname, sum(orig), sum(land))

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

func trackCalls(mp map[string]int) {
	for _, v := range mp {
		if v == 0 {
			stats.hasNoCalls++
		} else {
			stats.hasCalls++
		}
	}
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
		if landV > origV {
			errs = append(errs, fmt.Errorf("val mismatch [land=%v]>[go=%v] [%v]", landV, origV, origK))
			continue
		}
		if landV < origV {
			errs = append(errs, fmt.Errorf("!!!val mismatch [land=%v]<[go=%v] [%v]", landV, origV, origK))
			continue
		}
		okCnt++
	}

	fmt.Printf("ratio: %.2f\n", float64(okCnt)/float64(len(orig))*100)
	stats.ok += okCnt
	stats.total += len(orig)

	return errors.Join(errs...)
}

func ratio(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
