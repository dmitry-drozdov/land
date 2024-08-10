package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func init() {
	n := Parse("{ } f(x); { f(x); } f(x); { f(x); { f(x); } f(x); } f(x);")
	b, _ := json.Marshal(n)
	fmt.Println(string(b))
}

type Node struct {
	Type     string
	Children []*Node
}

func Parse(s string) *Node {
	n := &Node{
		Type:     "root",
		Children: nil,
	}
	tokens := strings.Split(s, " ")

	parse(tokens, 0, n)
	return n
}

func parse(tokens []string, pos int, node *Node) int {
	for i := pos; i < len(tokens); i++ {
		token := tokens[i]
		switch token {
		case "{":
			n := &Node{Type: "block"}
			node.Children = append(node.Children, n)
			i = parse(tokens, i+1, n)
		case "}":
			return i
		default:
			node.Children = append(node.Children, &Node{Type: "Any"})
		}
	}
	return len(tokens) // inf
}
