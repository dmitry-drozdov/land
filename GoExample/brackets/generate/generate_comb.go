package generate

import (
	"fmt"
	"strings"
)

var (
	template = "F { F } F { F { F } F } F"
	symbol   = "f(x);"
)

func GenerateCombinations() (*GenerateRes, error) {
	countF := strings.Count(template, "F")
	totalCombinations := 1 << countF

	res := NewGenerateRes(totalCombinations * 2)

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
		if err := res.Add(result, code); err != nil {
			return nil, err
		}

		for strings.Contains(result, "{ }") {
			result = strings.Replace(result, "{ }", "", -1)
			result = strings.Replace(result, "  ", " ", -1)
			result = strings.TrimSpace(result)
			code += "9" // means "removed { }"
			if err := res.Add(result, code); err != nil {
				return nil, err
			}
		}
	}

	// for i, r := range res {
	// 	res[i] = strings.ReplaceAll(r, " ", "\n")
	// }

	return res, nil
}
