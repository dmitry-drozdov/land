package main

import (
	"fmt"
	"strings"
)

// F { F } F { F { F } F } F

var (
	template = "F { F } F { F { F } F } F"
	symbol   = "f(x);"
)

func generateCombinations() map[string]string { // file name => code text
	countF := strings.Count(template, "F")
	totalCombinations := 1 << countF

	mp := make(map[string]string, totalCombinations*2)
	dups := make(map[string]struct{}, totalCombinations*2)

	add := func(s string, code string) {
		if s != "" {
			_, ok := dups[s]
			if !ok {
				mp[code] = s
			}
			dups[s] = struct{}{}
		}
	}

	for i := 0; i < totalCombinations; i++ {
		bitMask := fmt.Sprintf("%07b", i)
		result := template

		for j := 0; j < countF; j++ {
			if bitMask[j] == '0' {
				result = strings.Replace(result, "F", "", 1)
			} else {
				result = strings.Replace(result, "F", symbol, 1)
			}
		}

		result = strings.Replace(result, "  ", " ", -1)
		result = strings.TrimSpace(result)

		code := bitMask
		add(result, code)

		for strings.Contains(result, "{ }") {
			result = strings.Replace(result, "{ }", "", -1)
			result = strings.Replace(result, "  ", " ", -1)
			result = strings.TrimSpace(result)
			code += "9" // means remove { }
			add(result, code)
		}
	}

	// for i, r := range res {
	// 	res[i] = strings.ReplaceAll(r, " ", "\n")
	// }

	return mp
}
