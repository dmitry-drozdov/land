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

	Source         string
	SkippedFiles   int
	SkippedPerCent RoundedFloat
	Fail           int
	Ok             int
	Accuracy       RoundedFloat
	ArgsCover      RoundedFloat
}

func (a *AnalyzerStats) init() {
	total := a.mismatch + a.match
	skipped := a.lnFull - a.lnLight

	a.SkippedFiles = skipped
	a.SkippedPerCent = ratio(skipped, a.lnFull)
	a.Fail = a.mismatch
	a.Ok = a.match
	a.Accuracy = ratio(a.match, total)
	a.ArgsCover = ratio(total-a.notAllArgs, total)
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

func ratio(part, total int) RoundedFloat {
	return RoundedFloat(float64(part) / float64(total) * 100)
}
