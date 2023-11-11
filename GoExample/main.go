package main

import (
	"fmt"
)

func main() {

	fmt.Println("reading results done...")
	light, err := ReadResults(`e:\phd\my\results\`)
	if err != nil {
		panic(err)
	}
	fmt.Println("reading results DONE")

	sname := "docker-ce"
	source := fmt.Sprintf(`e:\phd\my\%s\`, sname)

	fmt.Println("parsing files with go ast...")
	full, err := ParseFiles(source)
	if err != nil {
		panic(err)
	}
	fmt.Println("parsing files with go ast DONE")

	a := &AnalyzerStats{
		Source:  sname,
		lnFull:  len(full),
		lnLight: len(light),
	}

	for kf, vf := range full {
		kl, ok := light[kf]
		if !ok {
			continue
		}

		for k, v := range vf {
			funcs, ok := kl[k]
			if !ok {
				a.mismatch++
				continue
			}

			if len(v.Args) != len(funcs.Args) || len(funcs.Args) != funcs.ArgsCnt {
				a.notAllArgs++
			}

			if !v.EqualTo(funcs) {
				fmt.Println()
				fmt.Println(kf, v, funcs)
				a.mismatch++
				continue
			}

			a.match++
		}
	}

	if err := a.Dump(); err != nil {
		panic(err)
	}

}
