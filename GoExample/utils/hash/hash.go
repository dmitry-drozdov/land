package hash

import "hash/fnv"

func HashSlice[T any](sl []T, f func(s T) string) uint64 {
	res := uint64(0)
	for _, s := range sl {
		res += HashString(f(s))
	}
	return res
}

func HashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}
