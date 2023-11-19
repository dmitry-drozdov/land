package main

import (
	"fmt"
	"strings"
)

type RoundedFloat float64

func (r RoundedFloat) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf("%.2f", r)
	str = strings.TrimRight(str, "0")
	str = strings.TrimRight(str, ".")
	return []byte(str), nil
}
