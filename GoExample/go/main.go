package main

import (
	"fmt"
	"strings"
	"utils/slice"

	"github.com/mohae/shuffle"
	cp "github.com/otiai10/copy"
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
	err := makeTestSet(50)
	if err != nil {
		panic(err)
	}
	return

	cnt, err := deleteDups(`e:\phd\test_repos`)
	if err != nil {
		panic(err)
	}
	fmt.Printf("deleted %v duplicates\n", cnt)

	for _, f := range folders {
		if err := doWork(f); err != nil {
			fmt.Printf("[%v] <ERROR>: [%v]\n", f, err)
		}
	}

	for _, f := range folders {
		st, err := getCodeStats(f)
		if err != nil {
			fmt.Printf("[%v] <ERROR>: [%v]\n", f, err)
			continue
		}
		fmt.Println(f, st["go"], st["graphql"])
	}

	err = GetTotalStats("results")
	if err != nil {
		panic(err)
	}
}

func getCodeStats(sname string) (map[string]*CodeStats, error) {
	return codeStats(fmt.Sprintf(`e:\phd\test_repos\%s\`, sname))
}

func makeTestSet(percent int) error {
	if percent <= 0 || percent > 100 {
		return fmt.Errorf("incorrect percent")
	}
	allFiles := make([]string, 0, 40000)
	for _, sname := range folders {
		files, err := getFiles(fmt.Sprintf(`e:\phd\test_repos\%s\`, sname))
		if err != nil {
			return err
		}
		allFiles = append(allFiles, files...)
	}

	ln := len(allFiles)
	if err := shuffle.String(allFiles); err != nil {
		return err
	}

	allFiles = allFiles[:(ln * percent / 100)]

	d := make(map[string]int, len(folders))
	for _, f := range allFiles {
		for _, folder := range folders {
			if strings.HasPrefix(f, fmt.Sprintf(`e:\phd\test_repos\%s\`, folder)) {
				d[folder]++
			}
		}
	}

	fmt.Println(d)

	for _, f := range allFiles {
		err := cp.Copy(f, strings.Replace(f, `e:\phd\test_repos\`, `e:\phd\test_repos_light\`, 1))
		if err != nil {
			return err
		}
	}

	return nil
}

func doWork(sname string) error {
	fmt.Printf("\n===== %s START =====\n", sname)
	defer fmt.Printf("===== %s END =====\n", sname)

	fmt.Println("reading results...")
	lightFunc, lightStruct, err := ReadResults(fmt.Sprintf(`e:\phd\test_repos\results\%s`, sname))
	if err != nil {
		return err
	}
	fmt.Println(len(lightStruct))
	fmt.Println("reading results DONE")

	source := fmt.Sprintf(`e:\phd\test_repos\%s\`, sname)
	fmt.Println("parsing files with go ast...")
	fullFunc, fullStruct, duplicates, err := ParseFiles(source)
	if err != nil {
		return err
	}
	fmt.Println("parsing files with go ast DONE")

	a := &AnalyzerFuncStats{
		Source:     sname,
		Duplicates: duplicates,
		lnFull:     len(fullFunc),
		lnLight:    len(lightFunc),
	}

	s := &a.StructStats

	for kf, vf := range fullStruct {
		lk, ok := lightStruct[kf]
		if !ok {
			fmt.Printf("info for %s not found", kf)
			continue
		}

		for k, v := range vf {
			sl, ok := lk[k]
			if !ok {
				s.FailNotFound++
				for _, t := range v.Types {
					if t == "anon_func_title" {
						s.FailNotFoundHasFunc++
						break
					}
				}
				continue
			}
			if !slice.Compare(v.Types, sl.Types, trimSpace) {
				s.FailIncorrectTypes++
				continue
			}
			s.Ok++
		}
	}

	for kf, vf := range fullFunc {
		kl, ok := lightFunc[kf]
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

			if len(v.Args) != len(funcs.Args) || byte(len(funcs.Args)) != funcs.ArgsCnt {
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

var trimSpace = func(s string) string { return strings.TrimSpace(s) }
