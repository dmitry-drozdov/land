package code

import (
	"strings"
)

func GetLOC(code string) uint {
	res := uint(0)
	lines := strings.Split(code, "\n")
	goComment := false
	graphqlComment := false
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}

		// GO
		if strings.Contains(l, "/*") {
			goComment = true
			if !strings.HasPrefix(l, "/*") {
				res++ // something before comment (rare case)
			}
			continue
		}
		if strings.Contains(l, "*/") {
			goComment = false
			continue
		}
		if strings.HasPrefix(l, "//") {
			continue
		}
		if goComment {
			continue
		}

		// GRAPHQL
		if strings.Contains(l, `"""`) {
			graphqlComment = !graphqlComment
			continue
		}
		if strings.HasPrefix(l, "#") {
			continue
		}
		if graphqlComment {
			continue
		}

		res++
	}
	return res
}
