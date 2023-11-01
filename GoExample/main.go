package main

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

// some comment
type FMp func(map[struct{}]func() int) map[interface{}]func(int) struct{ f func(bool) }

type MpSimple map[struct{ f bool }]bool

func ReturnMap() *map[int][]*string {
	return nil // some comment
}

type Mp map[struct{}]func() int

var I int

func ReturnPointer(f func(int, bool) struct{}, i interface{}, f2 func(i int, s struct{}) interface{}) (j *****int) {
	panic(i)
}

func ReturnSlict(int, bool, struct{}, interface{}) (k [][][]int, err error) {
	return [][][]int{}, nil
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

// some comment
// other comment
func (ws *MockWeatherService) SetVal(x int) {
	ws.Val = x
}

type Weather struct {
	service IWeatherService
}

/*
multiline comment /*
*/
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

var c = &struct{ F struct{} }{}

var a = 1

var f = func() {}

var m0 = func(i int) bool { return false }

var m1 = new(MockWeatherService)

/*
/* // multiline comment /*
*/
var m2 *MockWeatherService = &MockWeatherService{}

var m3 float64 = math.Round(10)

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

var m9 = func() chan<- func() {
	var f1 = func() {}
	func() { f1() }()
	return nil
}

var m10 = func() bool {
	var f1 = func() {}
	func() { f1() }()
	return m8(0)
}

func Some3(struct{}) chan<- struct{ f interface{} } {
	return nil
}

var m11 = func() func() chan struct{} {
	var f1 = func() {}
	func() { f1() }()
	return nil
}

func Some() {

}

var m12 = func() bool {
	var f1 = func() {}
	func() { f1() }()
	var f2 = func(i int) bool { return false }
	m11()
	return f2(0)
}

func Some2(struct{}) chan<- struct{ f map[string]struct{} } {
	return nil
}

var m13 func() func() struct{} = func() func() struct{} {
	var f1 = func() {}
	func() {
		f1()
		m12()
	}()
	return nil
}

var m14 func() func() func(i int) func(int) = func() func() func(i int) func(int) {
	f2()
	return nil
}

var (
	m4 [][100]**[][]struct{} = nil
	m5                       = &[]*[10]*[]float64{{{math.Round(math.Sin(float64(len(map[string]struct{}{}))))}}}
	m6                       = map[*struct {
		f chan<- func() chan<- interface{}
	}]chan<- float64{}
	m7 = func(x int) {
		x++
	}
	f2  = func() func() func(i int) func(int) { return nil }
	Ch4 = make(chan<- <-chan struct{ f interface{} })
	S1  = &struct{}{}
)

func SuppressWarnings() interface{} {
	_ = m0
	_ = m1
	_ = m2
	_ = m6
	_ = m8
	m7(a + int(m3) + int(len(m4)) + int(len(*m5)))
	f()
	m13()
	m14()
	_ = some
	_ = other
	_ = c
	return struct{}{}
}

var (
	I1 int
	I2 struct{ f func() }
	I3 map[chan<- int]<-chan struct{ f interface{} }
	I5                           = func(chan<- func()) {}
	I6 *map[struct{}]interface{} = new(map[struct{}]interface{})
)

func ReturnInterface() interface{ Action(int) struct{} } {
	return nil
}

func ReturnInterface2() (**[]*[10][]interface{ Action(int) struct{} }, error, bool) {
	return nil, nil, false
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

/* pure comment

 */
func ReturnSeveral3(n int, s struct{ p bool }, i interface{ Action() struct{} }) (*[]int, [][10]*bool, func(int, bool) []int) {
	return &[]int{}, nil, nil
}

// func func chan<-
func WithMap(map[string]func()) (m map[string]struct{}, e error) {
	return nil, nil
}

// chan struct{}
var ch chan struct{}

// chan map[struct{}]chan float64
var ch3 chan map[struct{}]chan float64

var mp map[chan struct{}]int

var ch2 chan<- func() <-chan int

func Chan(<-chan int) chan<- map[chan<- int]struct{} {
	close(ch)
	delete(mp, make(chan struct{}))
	return nil
}

func Chan2(chan map[chan int]chan string) chan map[chan interface{}]func() {
	close(ch2)
	close(ch3)
	return nil
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
