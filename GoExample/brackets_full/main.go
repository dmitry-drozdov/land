package main

import (
	"brackets/datatype"
	"brackets/parser"
	"brackets/provider"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/color"
)

const (
	RATIO = 1
)

var folders = []string{
	"Lp",
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
	ok    int
	total int
}{}

func main() {
	color.New(color.FgRed, color.Bold).Printf("START %v\n", time.Now().Format(time.DateTime))

	fc := make(map[string]struct{}, 1_900_000)
	for _, f := range folders {
		if err := doWork(f, fc); err != nil {
			color.New(color.FgBlack, color.Bold).Printf("[%v] <ERROR>: [%v]\n", f, err)
		}
	}

	color.Green("TOTAL ratio: %.5f [bad=%v]\n", ratio(stats.ok, stats.total), stats.total-stats.ok)
}

func doWork(sname string, fc map[string]struct{}) error {
	color.Cyan("===== %s START =====\n", sname)

	source := fmt.Sprintf(`e:\phd\test_repos\%s\`, sname)
	p := parser.NewParser(fc)
	orig, err := p.ParseFiles(source)
	if err != nil {
		return err
	}

	resFolder := fmt.Sprintf(`e:\phd\test_repos_calls\results\%s\`, sname)
	land, err := provider.ReadFolder(resFolder)
	if err != nil {
		return err
	}

	color.Cyan("===== %s avg max children count  [%v] =====\n", sname, childrenCount(orig))
	color.Cyan("===== %s END [%v] [%v] [dups %v]=====\n", sname, control(orig), control(land), p.Dups)

	err = compareMaps(orig, land)
	if err != nil {
		return err
	}

	return nil
}

func control(mp map[string]*datatype.Brackets) int {
	res := 0
	for _, v := range mp {
		_, cnt := v.ControlNumber()
		res += cnt
	}
	return res
}

func childrenCount(mp map[string]*datatype.Brackets) map[int]int {
	stats := map[int]int{}
	for _, v := range mp {
		cnt := v.MaxChildrenCount()
		stats[cnt]++
	}
	return stats
}

func compareMaps(orig, land map[string]*datatype.Brackets) error {
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
		if err := landV.EqualTo(origV); err != nil {
			errs = append(errs, fmt.Errorf("%v: %w", origK, err))
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
