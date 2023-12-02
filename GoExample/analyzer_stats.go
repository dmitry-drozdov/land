package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type AnalyzerStats struct {
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
}

func (a *AnalyzerStats) init() {
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
}

func (a *AnalyzerStats) marshal() ([]byte, error) {
	return json.MarshalIndent(a, "", " ")
}

func (a *AnalyzerStats) String() string {
	bytes, err := a.marshal()
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

func (a *AnalyzerStats) Dump() error {
	a.init()
	bytes, err := a.marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fmt.Sprintf("results/%s.json", a.Source), bytes, 0644)
}

func (a *AnalyzerStats) Add(b AnalyzerStats) {
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
}

func ratio(part, total int) RoundedFloat {
	if total == 0 {
		return 0
	}
	return RoundedFloat(float64(part) / float64(total) * 100)
}
