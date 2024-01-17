package main

import (
	"hash/fnv"
)

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
