package main

import (
	"fmt"
	"strings"
	"utils/slice"

	"github.com/mohae/shuffle"
	cp "github.com/otiai10/copy"
)

type GrammarType string

const (
	GrammarTypeHighLevel GrammarType = "GrammarTypeHighLevel"
	GrammarTypeMarkup    GrammarType = "GrammarTypeMarkup"
)

var currentMode = GrammarTypeMarkup

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
	// err := makeTestSet(40)
	// if err != nil {
	// 	panic(err)
	// }
	// return

	// err := GenerateLargeFile(`e:\phd\test_repos_light`, `e:\phd\large.go`)
	// if err != nil {
	// 	panic(err)
	// }
	// return

	// err := GenerateLargeFileStandard(`e:\phd\large\large`, "go")
	// if err != nil {
	// 	panic(err)
	// }
	// return

	// err = GenerateLargeFileStandardSharp(`e:\phd\large\large`, "cs")
	// if err != nil {
	// 	panic(err)
	// }
	// return

	// t0 := time.Now()
	// fset := token.NewFileSet()
	// _, err = parser.ParseFile(fset, `e:\phd\large.go`, nil, 0)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(time.Since(t0))
	// return

	// cnt, err := deleteDups(`e:\phd\test_repos`)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("deleted %v duplicates\n", cnt)

	for _, f := range folders {
		if err := doWork(f, currentMode); err != nil {
			fmt.Printf("[%v] <ERROR>: [%v]\n", f, err)
		}
	}

	var goFiles, goLines, goVendor, goLinesVendor uint
	for _, f := range folders {
		st, err := getCodeStats(f)
		if err != nil {
			fmt.Printf("[%v] <ERROR>: [%v]\n", f, err)
			continue
		}
		fmt.Println(f, st["go"], st["graphql"])
		goFiles += st["go"].FilesCnt
		goLines += st["go"].CodeLinesCnt
		goVendor += st["go"].FilesVendorCnt
		goLinesVendor += st["go"].CodeLinesVendorCnt
	}
	fmt.Println(
		"projects", len(folders),
		"go files", goFiles,
		"go vendor", (goVendor*100)/goFiles,
		"go lines", goLines,
		"go lines vendor", (goLinesVendor*100)/goLines,
	)

	err := GetTotalStats("results")
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

func doWork(sname string, gt GrammarType) error {
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
	ast := NewGoAST()
	fullFunc, fullStruct, duplicates, err := ast.ParseFiles(source)
	if err != nil {
		return err
	}
	fmt.Println("parsing files with go ast DONE")

	total := 0
	mp := map[string]int{}
	for k, v := range ast.stats {
		total += v
		switch k {
		case "map":
			mp[k] += v
		case "chan":
			mp[k] += v
		default:
			if strings.Contains(k, "anon") {
				mp[k] += v
			}
		}
	}
	fmt.Printf("anon_func [%d] anon_inter [%d] anon_struct [%d] map [%d] chan [%d] total [%d]\n",
		mp["anon_func_title"], mp["anon_interface"], mp["anon_struct"], mp["map"], mp["chan"], total)

	a := &AnalyzerFuncStats{
		Source:         sname,
		Duplicates:     duplicates,
		lnFull:         len(fullFunc),
		lnLight:        len(lightFunc),
		postProcessReq: 0,
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

	argsDepth := &DepthStats[int]{}
	retsDepth := &DepthStats[int]{}
	totalDepth := &DepthStats[int]{}
	avgDepth := &DepthStats[float64]{}
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

			if gt == GrammarTypeMarkup {
				if len(v.Args) != len(funcs.Args) || byte(len(funcs.Args)) != funcs.ArgsCnt {
					a.notAllArgs++
				}
			}

			sum := 0
			for _, argDepth := range funcs.ArgsDepth {
				sum += argsDepth.Process(argDepth, funcs.Name)
			}
			for _, retDepth := range funcs.ReturnsDepth {
				sum += retsDepth.Process(retDepth, funcs.Name)
			}
			totalDepth.Process(sum, funcs.Name)
			if funcs.ArgsCnt+funcs.Return > 0 {
				avgDepth.Process(float64(sum)/float64(funcs.ArgsCnt+funcs.Return), funcs.Name)
			} else {
				avgDepth.Process(0, funcs.Name)
			}

			if !v.EqualTo(funcs, gt) {
				countMismatch()
				continue
			}

			a.match++
			if v.RequirePostProcess {
				a.postProcessReq++
			}
			if v.NoBody {
				a.noBody++
			}
		}
	}

	fmt.Printf("argsDepth: " + argsDepth.String())
	fmt.Printf("retsDepth: " + retsDepth.String())
	fmt.Printf("totalDepth: " + totalDepth.String())
	fmt.Printf("avgDepth: " + avgDepth.String())

	return a.Dump()
}

var trimSpace = func(s string) string { return strings.TrimSpace(s) }
