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

func generateCombinations() []string {

	countF := strings.Count(template, "F")
	totalCombinations := 1 << countF

	res := make([]string, 0, totalCombinations*2)

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

		if strings.Contains(result, "{ }") {
			res = append(res, result)

			resultWithoutBraces := strings.Replace(result, "{ }", "", -1)
			resultWithoutBraces = strings.Replace(resultWithoutBraces, "  ", " ", -1)
			resultWithoutBraces = strings.TrimSpace(resultWithoutBraces)
			res = append(res, resultWithoutBraces)
		} else {
			res = append(res, result)
		}
	}

	// for i, r := range res {
	// 	res[i] = strings.ReplaceAll(r, " ", "\n")
	// }

	return res
}
