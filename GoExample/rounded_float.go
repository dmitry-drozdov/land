package main

import (
	"fmt"
)

type RoundedFloat float64

func (r RoundedFloat) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.1f", r)), nil
}
