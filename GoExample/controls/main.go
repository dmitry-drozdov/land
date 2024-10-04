package main

import (
	"context"
	"controls/datatype"
	"controls/parser"
	"controls/provider"
	"errors"
	"fmt"
	"time"
	"utils/concurrency"

	"github.com/fatih/color"
	"gitlab.services.mts.ru/lp/backend/libs/tracer"
	"go.opentelemetry.io/otel/trace"
)

const (
	RATIO = 1
)

var folders = []string{
	"Lp\\address-service",
	"Lp\\bitrix-adapter",
	"Lp\\channel-profile",
	"Lp\\delivery-offering",
	"Lp\\delivery-ordering",
	"Lp\\efin-courier",
	"Lp\\logportal-adapter",
	"Lp\\polygons",
	"Lp\\protovar-adapter",
	"Lp\\rtk-assembling-adapter",
	"Lp\\rtk-pickup",
	"Lp\\rtk-stock",
	"Lp\\rtk-stores-loader",
	"Lp\\stock-managment",
	"Lp\\warehouses",
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
	ctx := context.Background()
	// cancel := tracer.NewTracer(ctx,
	// 	tracer.WithInsecure(true),
	// 	tracer.WithServiceName("phd"),
	// 	tracer.WithEndpoint("localhost:4317"),
	// )
	// defer func() {
	// 	if err := cancel(); err != nil {
	// 		panic(err)
	// 	}
	// }()
	tracer.ReplaceGlobals(&tracer.Tracer{T: trace.NewNoopTracerProvider().Tracer("phd")})

	ctx, end := tracer.Start(ctx, "main")
	defer end(nil)

	color.New(color.FgRed, color.Bold).Printf("START %v\n", time.Now().Format(time.DateTime))

	b := concurrency.NewBalancer(RATIO) // на каждые RATIO файлов с вызовами 1 файл без вызовов
	fc := make(map[string]struct{}, 1_900_000)
	for _, f := range folders {
		if err := doWork(ctx, f, b, fc); err != nil {
			color.New(color.FgBlack, color.Bold).Printf("[%v] <ERROR>: [%v]\n", f, err)
		}
	}

	totalFuncs := stats.hasCalls + stats.hasNoCalls
	color.Green(
		"TOTAL has calls: %v (%.2f%%), has no calls: %v (%.2f%%)\n",
		stats.hasCalls, ratio(stats.hasCalls, totalFuncs),
		stats.hasNoCalls, ratio(stats.hasNoCalls, totalFuncs),
	)
	color.Green("TOTAL func call: %v, bodies: %v\n", b.CntMain(), b.CntSub())
	color.Green("TOTAL ratio: %.5f [bad=%v]\n", ratio(stats.ok, stats.total), stats.total-stats.ok)
}

func doWork(ctx context.Context, sname string, balancer *concurrency.Balancer, fc map[string]struct{}) error {
	ctx, end := tracer.Start(ctx, "doWork "+sname)
	defer end(nil)

	color.Cyan("===== %s START =====\n", sname)

	source := fmt.Sprintf(`e:\phd\test_repos\%s\`, sname)
	p := parser.NewParser(balancer, fc)
	orig, err := p.ParseFiles(ctx, source)
	if err != nil {
		return err
	}

	resFolder := fmt.Sprintf(`e:\phd\test_repos_calls\results\%s\`, sname)
	land, err := provider.ReadFolder(ctx, resFolder)
	if err != nil {
		return err
	}
	color.Cyan("===== %s END [%v] [%v] [dups %v]=====\n", sname, -1, -1, p.Dups)

	err = compareMaps(ctx, orig, land)
	if err != nil {
		return err
	}

	return nil
}

func compareMaps(ctx context.Context, orig, land map[string]*datatype.Control) error {
	_, end := tracer.Start(ctx, "compareMaps")
	defer end(nil)

	var errs []error
	if len(orig) != len(land) {
		errs = append(errs, fmt.Errorf("len mismatch %v %v", len(orig), len(land)))
	}

	okCnt := 0
	for origK, origV := range orig {
		landV, ok := land[origK]
		if !ok {
			errs = append(errs, fmt.Errorf("%v: key not found", origK))
			continue
		}
		if err := landV.EqualTo(origV); err != nil {
			errs = append(errs, fmt.Errorf("%v: %w", origK, err))
			continue
		}
		okCnt++
	}

	fmt.Printf("ratio: %.2f\n", float64(okCnt)/float64(len(orig))*100)
	stats.ok += okCnt
	stats.total += len(orig)

	return errors.Join(errs...)
}

func ratio(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
