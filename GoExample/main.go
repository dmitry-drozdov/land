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

	source := `e:\phd\my\tidb\`

	fmt.Println("parsing files with go ast...")
	full, err := ParseFiles(source)
	if err != nil {
		panic(err)
	}
	fmt.Println("parsing files with go ast DONE")

	notAllArgs := 0
	mismatch := 0
	match := 0

	for kf, vf := range full {
		kl, ok := light[kf]
		if !ok {
			continue
		}

		for k, v := range vf {
			funcs, ok := kl[k]
			if !ok {
				mismatch++
				continue
			}

			if len(v.Args) != len(funcs.Args) || len(funcs.Args) != funcs.ArgsCnt {
				notAllArgs++
			}

			if !v.EqualTo(funcs) {
				// fmt.Println()
				// fmt.Println(kf, v, funcs)
				mismatch++
				continue
			}

			match++
		}
	}

	total := mismatch + match
	skipped := len(full) - len(light)
	fmt.Printf("source: [%v] skipped: [%v (%.1f%%)] fail: [%v] ok: [%v] accuracy: [%.1f%%] argsCover: [%.1f%%]",
		source, skipped, ratio(skipped, len(full)), mismatch, match, ratio(match, total), ratio(total-notAllArgs, total))
}

func ratio(part, total int) float64 {
	return float64(part) / float64(total) * 100
}
