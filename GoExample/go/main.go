package main

import (
	"fmt"
	"strings"
)

var folders = []string{
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

func main() {
	cnt, err := deleteDups(`e:\phd\test_repos`)
	if err != nil {
		panic(err)
	}
	fmt.Printf("deleted %v duplicates", cnt)

	for _, f := range folders {
		if err := doWork(f); err != nil {
			fmt.Printf("[%v] <ERROR>: [%v]\n", f, err)
		}
	}

	err = GetTotalStats("results")
	if err != nil {
		panic(err)
	}
}

func doWork(sname string) error {
	fmt.Printf("\n===== %s START =====\n", sname)
	defer fmt.Printf("===== %s END =====\n", sname)

	fmt.Println("reading results...")
	light, err := ReadResults(fmt.Sprintf(`e:\phd\test_repos\results\%s`, sname))
	if err != nil {
		return err
	}
	fmt.Println("reading results DONE")

	source := fmt.Sprintf(`e:\phd\test_repos\%s\`, sname)
	fmt.Println("parsing files with go ast...")
	full, duplicates, err := ParseFiles(source)
	if err != nil {
		return err
	}
	fmt.Println("parsing files with go ast DONE")

	a := &AnalyzerStats{
		Source:     sname,
		Duplicates: duplicates,
		lnFull:     len(full),
		lnLight:    len(light),
	}

	for kf, vf := range full {
		kl, ok := light[kf]
		if !ok {
			a.mismatch++
			continue
		}

		for k, v := range vf {
			countMismatch := func() {
				fmt.Println()
				fmt.Println(kf, v)
				if strings.Contains(kf, "vendor") {
					a.cntVendor++
				}
				a.mismatch++
			}

			funcs, ok := kl[k]
			if !ok {
				countMismatch()
				continue
			}

			if len(v.Args) != len(funcs.Args) || len(funcs.Args) != funcs.ArgsCnt {
				a.notAllArgs++
			}

			if !v.EqualTo(funcs) {
				countMismatch()
				continue
			}

			a.match++
		}
	}

	return a.Dump()
}