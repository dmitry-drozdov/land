package generate

import (
	"fmt"
	"strconv"
	"strings"

	"utils/worker"
)

var ( // нельзя удалять { } после if, поэтому помечаем их символом |
	ifControl       = "if true <_F_>"
	ifElseControl   = "if true <_F_> else <_F_>"
	ifElseIfControl = "if true <_F_> else if true <_F_>"
)

func GenerateControlCombinations(template string) (*GenerateRes, error) {
	countG := strings.Count(template, "G")
	subCombinations := 1 << (2 * countG) // 4^countG

	templates := make([]string, 0, subCombinations)

	for i := 0; i < subCombinations; i++ {
		if i > 7 {
			continue // shrink test data
		}
		mask := strconv.FormatInt(int64(i), 4)
		mask = fmt.Sprintf("%07s", mask)

		result := template

		for j := 0; j < countG; j++ {
			switch mask[j] {
			case '0':
				result = strings.Replace(result, "G", "F", 1)
			case '1':
				result = strings.Replace(result, "G", ifControl, 1)
			case '2':
				result = strings.Replace(result, "G", ifElseControl, 1)
			case '3':
				result = strings.Replace(result, "G", ifElseIfControl, 1)
			default:
				return nil, fmt.Errorf("unknown mask symbol")
			}
		}

		templates = append(templates, result)
	}

	res := NewGenerateRes(len(templates))

	err := worker.Iterate(templates, 6, func(t string) error {
		combs, err := GenerateCombinations(t, "f(x);")
		res.Merge(combs)
		return err
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
