package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type AnalyzerStructStats struct {
	Ok                  int
	FailNotFound        int
	FailNotFoundHasFunc int // subtype of FailNotFound
	FailIncorrectTypes  int
	Ratio               RoundedFloat
}

func (a *AnalyzerStructStats) Total() int {
	return a.FailIncorrectTypes + a.FailNotFound + a.Ok
}

type AnalyzerFuncStats struct {
	notAllArgs int
	mismatch   int
	match      int
	lnFull     int
	lnLight    int
	cntVendor  int

	Duplicates         int
	Source             string
	TotalFiles         int
	SkippedFiles       int
	SkippedPerCent     RoundedFloat
	Fail               int
	Ok                 int
	Accuracy           RoundedFloat
	ArgsCover          RoundedFloat
	VendorFuncs        int
	VendorFuncsPerCent RoundedFloat

	StructStats AnalyzerStructStats
}

func (a *AnalyzerFuncStats) init() {
	total := a.mismatch + a.match
	skipped := a.lnFull - a.lnLight

	a.TotalFiles = a.lnFull
	a.SkippedFiles = skipped
	a.SkippedPerCent = ratio(skipped, a.lnFull)
	a.Fail = a.mismatch
	a.Ok = a.match
	a.Accuracy = ratio(a.match, total)
	a.ArgsCover = ratio(total-a.notAllArgs, total)

	a.VendorFuncs = a.cntVendor
	a.VendorFuncsPerCent = ratio(a.cntVendor, a.mismatch)

	a.StructStats.Ratio = ratio(a.StructStats.Ok, a.StructStats.Total())
}

func (a *AnalyzerFuncStats) marshal() ([]byte, error) {
	return json.MarshalIndent(a, "", " ")
}

func (a *AnalyzerFuncStats) String() string {
	bytes, err := a.marshal()
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

func (a *AnalyzerFuncStats) Dump() error {
	a.init()
	bytes, err := a.marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(fmt.Sprintf("results/%s.json", a.Source), bytes, 0644)
}

func (a *AnalyzerStructStats) Add(b AnalyzerStructStats) {
	a.Ok += b.Ok
	a.FailIncorrectTypes += b.FailIncorrectTypes
	a.FailNotFound += b.FailNotFound
	a.FailNotFoundHasFunc += b.FailNotFoundHasFunc
}

func (a *AnalyzerFuncStats) Add(b AnalyzerFuncStats) {
	a.Source = ""
	a.TotalFiles += b.TotalFiles
	a.SkippedFiles += b.SkippedFiles
	a.SkippedPerCent += b.SkippedPerCent
	a.Fail += b.Fail
	a.Ok += b.Ok
	a.Accuracy += b.Accuracy
	a.ArgsCover += b.ArgsCover
	a.VendorFuncs += b.VendorFuncs
	a.VendorFuncsPerCent += b.VendorFuncsPerCent
	a.Duplicates += b.Duplicates
	a.StructStats.Add(b.StructStats)
}

func ratio(part, total int) RoundedFloat {
	if total == 0 {
		return 0
	}
	return RoundedFloat(float64(part) / float64(total) * 100)
}
