package main

import (
	"calls/parser"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"gonum.org/v1/gonum/stat"
)

const (
	RATIO = 1
)

var folders = []string{
	"Lp",
	"azure-service-operator",
	"kubernetes",
	"docker-ce",
	"sourcegraph",
	"delivery-offering",
	"boost",
	"chainlink",
	"modules",
	"go-ethereum",
	"grafana",
	"gvisor",
	"test",
	"backend",
	"go-redis",
	"tidb",
	"moby",
}

var stats = struct {
	hasCalls   int
	hasNoCalls int
	ok         int
	total      int
}{}

func main() {
	color.New(color.FgRed, color.Bold).Printf("START %v\n", time.Now().Format(time.DateTime))

	fc := make(map[string]struct{}, 1_900_000)
	for _, f := range folders {
		if err := doWork(f, fc); err != nil {
			color.New(color.FgBlack, color.Bold).Printf("[%v] <ERROR>: [%v]\n", f, err)
		}
	}

	printHistogram(resolverValues)
	printHistogram(regularValues)

	totalFuncs := stats.hasCalls + stats.hasNoCalls
	color.Green(
		"TOTAL has calls: %v (%.2f%%), has no calls: %v (%.2f%%)\n",
		stats.hasCalls, ratio(stats.hasCalls, totalFuncs),
		stats.hasNoCalls, ratio(stats.hasNoCalls, totalFuncs),
	)
	color.Green("TOTAL ratio: %.5f [bad=%v]\n", ratio(stats.ok, stats.total), stats.total-stats.ok)
}

var (
	resolverValues = []float64{}
	regularValues  = []float64{}
)

func doWork(sname string, fc map[string]struct{}) error {
	color.Cyan("===== %s START =====\n", sname)

	source := fmt.Sprintf(`e:\phd\test_repos\%s\`, sname)
	p := parser.NewParser(fc)
	orig, err := p.ParseFiles(source)
	if err != nil {
		return err
	}

	for k, v := range orig {
		if strings.HasSuffix(k, ".cnt") {
			continue
		}
		parts := strings.Split(k, ".") // 0 - pkg, 1 - receiver
		cnt, ok := orig[k+".cnt"]
		if !ok {
			panic("missing key")
		}
		val := float64(v) / float64(cnt)
		if strings.Contains(strings.ToLower(parts[1]), "resolver") {
			resolverValues = append(resolverValues, val)
		} else {
			regularValues = append(regularValues, val)
		}
	}

	return nil
}

func printHistogram(data []float64) {
	// Определяем количество бинов (интервалов)
	numBins := 15

	// Создаём слайсы для хранения границ бинов и количества элементов в каждом бине
	bins := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 1000}
	counts := make([]float64, numBins)

	sort.Float64s(data)

	// Заполняем гистограмму
	stat.Histogram(counts, bins, data, nil)

	// Выводим результаты
	fmt.Println("Гистограмма:")
	for i := 0; i < numBins; i++ {
		fmt.Printf("[%.2f - %.2f]: %d\n", bins[i], bins[i+1], int(counts[i]))
	}
}

func minMax(data []float64) (min, max float64) {
	min, max = data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func ratio(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
