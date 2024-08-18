package node

import (
	"fmt"
	"strings"
)

func ParseBracketSequence(s string) (*Node, error) {
	tokens := strings.Split(s, " ")

	stack := make([]*Node, 0, len(tokens))
	root := &Node{
		Type:     "root",
		Children: nil,
	}
	current := root

	for _, token := range tokens {
		switch token {
		case "{":
			newNode := &Node{Type: "block"}
			current.Children = append(current.Children, newNode)
			stack = append(stack, current)
			current = newNode
		case "}":
			if len(stack) == 0 {
				return nil, fmt.Errorf("redundant closing bracket")
			}
			current = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		case "f(x);":
			newNode := &Node{Type: "any"}
			current.Children = append(current.Children, newNode)
		case "", " ":
			// do nothing
		default:
			return nil, fmt.Errorf("unknown token: [%s]", token)
		}
	}

	if len(stack) != 0 {
		return nil, fmt.Errorf("redundant opening bracket")
	}

	return root, nil
}
