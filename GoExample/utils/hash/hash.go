package hash

import (
	"hash/fnv"
	"regexp"
	"unsafe"
)

func HashSlice[T any](sl []T, f func(s T) string) uint64 {
	res := uint64(0)
	for _, s := range sl {
		res += HashString(f(s))
	}
	return res
}

func HashString(s string) uint64 {
	h := fnv.New64a()
	data := unsafe.StringData(s)
	b := unsafe.Slice((*byte)(unsafe.Pointer(data)), len(s))
	h.Write(b)
	return h.Sum64()
}

func HashStrings(ss ...string) uint64 {
	h := fnv.New64a()
	for _, s := range ss {
		data := unsafe.StringData(s)
		b := unsafe.Slice((*byte)(unsafe.Pointer(data)), len(s))
		h.Write(b)
	}
	return h.Sum64()
}

func HashFile(bytes []byte) uint64 {
	str := string(bytes)

	re := regexp.MustCompile(`[\s]`) // to unify files formatting
	str = re.ReplaceAllString(str, "")

	h := fnv.New64a()
	h.Write([]byte(str))
	return h.Sum64()
}
