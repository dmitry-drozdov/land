package main

import (
	"hash/fnv"
	"sort"

	"golang.org/x/exp/constraints"
)

func compareSlice[T constraints.Ordered](a []T, b []T) bool {
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
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func sliceHash[T any](sl []T, f func(s T) string) uint64 {
	res := uint64(0)
	for _, s := range sl {
		res += hash(f(s))
	}
	return res
}

func hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}
