package main

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

type FMp func(map[struct{}]func() int) map[interface{}]func(int) struct{ f func(bool) }

type MpSimple map[struct{ f bool }]bool

func ReturnMap() *map[int][]*string {
	return nil
}

type Mp map[struct{}]func() int

var i int

func ReturnPointer() *****int {
	panic(i)
}

func ReturnSlict() [][][]int {
	return [][][]int{}
}

type IWeatherService interface {
	Forecast() int
}

type Pointer *struct{}

func PointerF() Pointer {
	return nil
}

type DoublePointer func(**struct {
	Some *int
}) func() struct{}

func DoublePointerF() {
	panic(0)
}

type Slice []func()

var some int

var other func(struct{}) map[struct{}]func() (interface{}, error)

type ArraySlice *[10][]func() interface{}

type WeatherService struct{}

func (ws *WeatherService) Forecast() int {
	rand.Seed(time.Now().Unix())
	value := rand.Intn(31)
	sign := rand.Intn(2)
	if sign == 1 {
		value = -value
	}
	return value
}

type MockWeatherService struct {
	Val int
}

func (ws *MockWeatherService) Forecast() int {
	return ws.Val
}

func (ws *MockWeatherService) SetVal(x int) {
	ws.Val = x
}

type Weather struct {
	service IWeatherService
}

func (w Weather) Forecast() string {
	deg := w.service.Forecast()
	switch {
	case deg < 10:
		return "холодно"
	case deg >= 10 && deg < 15:
		return "прохладно"
	case deg >= 15 && deg < 20:
		return "идеально"
	case deg >= 20:
		return "жарко"
	}
	return "инопланетно"
}

type testCase struct {
	deg  int
	want string
}

var tests []testCase = []testCase{
	{-10, "холодно"},
	{0, "холодно"},
	{5, "холодно"},
	{10, "прохладно"},
	{15, "идеально"},
	{20, "жарко"},
}

var c = struct{ F struct{} }{}

var a = 1

var f = func() {}

var m0 = func(i int) bool { return false }

var m1 = new(MockWeatherService)

var m2 = &MockWeatherService{}

var m3 = math.Round(10)

func F(i int) bool { return false }

var m8 = func(i int) bool {
	var f1 = func(i int) bool { return false }
	_ = f1
	return false
}

func G(i int) {
	m9()
	m10()
}

var m9 = func() {
	var f1 = func() {}
	func() { f1() }()
}

var m10 = func() bool {
	var f1 = func() {}
	func() { f1() }()
	return m8(0)
}

var m11 = func() func() bool {
	var f1 = func() {}
	func() { f1() }()
	return nil
}

var m12 = func() bool {
	var f1 = func() {}
	func() { f1() }()
	var f2 = func(i int) bool { return false }
	m11()
	return f2(0)
}

var m13 = func() func() struct{} {
	var f1 = func() {}
	func() {
		f1()
		m12()
	}()
	return nil
}

var m14 = func() func() func(i int) func(int) {
	f2()
	return nil
}

var (
	m4 = 0
	m5 = math.Round(11)
	m6 = struct{}{}
	m7 = func(x int) {
		x++
	}
	f2 = func() func() func(i int) func(int) { return nil }
)

func SuppressWarnings() interface{} {
	_ = m0
	_ = m1
	_ = m2
	_ = m6
	_ = m8
	m7(a + int(m3) + int(m4) + int(m5))
	f()
	m13()
	m14()
	_ = some
	_ = other
	_ = c
	return struct{}{}
}

func ReturnInterface() interface{ Action(int) struct{} } {
	return nil
}

func ReturnInterface2() **[]*[10][]interface{ Action(int) struct{} } {
	return nil
}

func ReturnStruct() struct{} {
	return struct{}{}
}

func ReturnStruct2() []struct{ test bool } {
	var f1 = func(i int) bool { return false }
	_ = f1
	return []struct{ test bool }{{false}}
}

func ReturnStruct3() struct{ test bool } {
	return struct{ test bool }{false}
}

func ReturnSeveral(n int, s struct{ p bool }, i interface{}) (struct{}, interface{}) {
	return struct{}{}, nil
}

func ReturnSeveral3(n int, s struct{ p bool }, i interface{ Action() struct{} }) (*[]int, [][10]*bool, func(int, bool) []int) {
	return &[]int{}, nil, nil
}

func WithMap(map[string]func()) (map[string]struct{}, error) {
	return nil, nil
}

func TestForecast(t *testing.T) {
	service := &MockWeatherService{}
	weather := Weather{service}
	for _, test := range tests {
		name := fmt.Sprintf("%v", test.deg)
		if i, ok := weather.service.(interface{ SetVal(x int) }); ok {
			i.SetVal(test.deg)
		}
		t.Run(name, func(t *testing.T) {
			got := weather.Forecast()
			if got != test.want {
				t.Errorf("%s: got %s, want %s", name, got, test.want)
			}
		})
	}
}

func main() {
	SuppressWarnings()
	fmt.Println("main func OK")
}
