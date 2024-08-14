package generate

import (
	"strconv"
	"strings"
)

func GenerateCombinations(template string, symbol string) (*GenerateRes, error) {
	countF := strings.Count(template, "F")
	totalCombinations := 1 << countF

	res := NewGenerateRes(totalCombinations)

	for i := 0; i < totalCombinations; i++ {
		bitMask := strconv.FormatInt(int64(i), 2)
		for len(bitMask) < countF {
			bitMask = "0" + bitMask
		}
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

	return res, nil
}
