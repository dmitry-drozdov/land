package main

import (
	"errors"
	"fmt"
	"gobody/ast"
	"gobody/provider"
	"strings"
	"utils/inspect"
	"utils/stats"
)

var folders = []string{
	"sourcegraph",
	"delivery-offering",
	"boost",
	"chainlink",
	"modules",
	"go-ethereum",
	"grafana",
	"gvisor",
	"backend",
	"azure-service-operator",
	"kubernetes",
	"go-redis",
	"docker-ce",
	"tidb",
	"moby",
}

func main() {
	err := generateAndParseBodies()
	if err != nil {
		panic(err)
	}
}

func generateAndParseBodies() error {
	p := ast.NewParser()
	st := &stats.Stats{}
	for _, f := range folders {
		source := fmt.Sprintf(`e:\phd\test_repos\%s\`, f)
		fmt.Printf("process files bodies with go ast [%s]...\n", f)
		nodes, err := p.ParseFilesBodies(source)
		if err != nil {
			return err
		}
		fmt.Printf("\t got [%d] nodes\n", len(nodes))

		source = fmt.Sprintf(`e:\phd\test_repos_body\results\%s\`, f)
		nodesFromLand, err := provider.ReadFolder(source)
		if err != nil {
			return err
		}
		fmt.Printf("\t got [%d] nodes from LanD\n", len(nodesFromLand))

		st2, _ := compareMaps(nodes, nodesFromLand)
		st.Add(st2)
	}
	fmt.Println("TOTAL STATS: ", st)
	return nil
}

func compareMaps(m1, m2 map[string]*inspect.Node) (*stats.Stats, error) {
	if len(m1) != len(m2) {
		return nil, fmt.Errorf("maps len mismatch")
	}
	st := &stats.Stats{}
	errs := make([]error, 100)
	for k1, v1 := range m1 {
		k2 := strings.Replace(k1, `e:\phd\test_repos_body`, `e:\phd\test_repos_body\results`, 1)
		v2, ok2 := m2[k2]
		if !ok2 {
			return nil, fmt.Errorf("file not found: %v", k1)
		}

		err := v1.EqualTo(v2)
		if err != nil {
			// b, _ := json.Marshal(v1)
			// fmt.Println(err, k1, string(b))
			// fmt.Println()
			errs = append(errs, err)
			st.Fail()
			continue
		}
		st.Ok()
	}
	fmt.Println(st)
	return st, errors.Join(errs...)
}
