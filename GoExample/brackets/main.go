package main

import "fmt"

func main() {
	comb := generateCombinations()
	for _, c := range comb {
		fmt.Println(c)
	}
}
