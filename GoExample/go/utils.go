package main

import (
	"sort"
	"strings"

	"golang.org/x/exp/constraints"
)

var trimSpace = func(s string) string { return strings.TrimSpace(s) }

func compareSlice[T constraints.Ordered](a []T, b []T, f func(T) T) bool {
	if len(a) != len(b) {
		return false
	}
	if (a == nil) != (b == nil) { // nil != []T{}
		return false
	}

	sort.Slice(a, func(i, j int) bool {
		return a[i] < a[j]
	})
	sort.Slice(b, func(i, j int) bool {
		return b[i] < b[j]
	})

	for i := range a {
		if f(a[i]) != f(b[i]) {
			return false
		}
	}

	return true
}
