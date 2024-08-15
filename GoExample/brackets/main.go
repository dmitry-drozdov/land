package main

import (
	"brackets/generate"
	"brackets/node"
	"brackets/provider"
	"fmt"
	"log"

	"net/http"
	_ "net/http/pprof"
)

var (
	templateControl = "G_{_G_}_G_{_G_{_G_}_G_}_G"
	template        = "F_{_F_}_F_{_F_{_F_}_F_}_F"
	symbol          = "f(x);"
)

func main() {
	go func() {
		log.Println("Starting pprof server on :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	Controls()
}

func Controls() {
	root := `e:\phd\test_repos_controls\`
	orig, err := generate.GenerateControlCombinations(templateControl)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\ngenerated %v files and %v nodes\n", len(orig.CodeToText), len(orig.CodeToNode))

	err = Dump(root, orig.CodeToText)
	if err != nil {
		panic(err)
	}
	fmt.Printf("dump to %v\n", root)

	resFolder := root + `results\`
	landNodes, err := provider.ReadFolder(resFolder)
	if err != nil {
		panic(err)
	}
	fmt.Printf("read %v land nodes\n", len(landNodes))

	_ = node.CompareMaps(landNodes, orig.CodeToNode)
}

func Brackets() {
	root := `e:\phd\test_repos_brackets\`
	orig, err := generate.GenerateCombinations(template, symbol)
	if err != nil {
		panic(err)
	}
	fmt.Printf("generated %v files and %v nodes\n", len(orig.CodeToText), len(orig.CodeToNode))

	err = Dump(root, orig.CodeToText)
	if err != nil {
		panic(err)
	}
	fmt.Printf("dump to %v\n", root)

	resFolder := root + `results\`
	landNodes, err := provider.ReadFolder(resFolder)
	if err != nil {
		panic(err)
	}
	fmt.Printf("read %v land nodes\n", len(landNodes))

	fmt.Println(node.CompareMaps(landNodes, orig.CodeToNode))
}
