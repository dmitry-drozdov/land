package main

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

type IWeatherService interface {
	Forecast() int
}

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

var a = 1

var f = func() {}

var m0 = func(i int) bool { return false }

var m1 = new(MockWeatherService)

var m2 = &MockWeatherService{}

var m3 = math.Round(10)

var (
	m4 = 0
	m5 = math.Round(11)
	m6 = struct{}{}
	m7 = func(x int) {
		x++
	}
)

func SuppressWarnings() interface{} {
	_ = m0
	_ = m1
	_ = m2
	_ = m6
	m7(a + int(m3) + int(m4) + int(m5))
	f()
	return struct{}{}
}

func ReturnInterface() interface{ Action(int) struct{} } {
	return nil
}

func ReturnStruct() struct{} {
	return struct{}{}
}

func ReturnStruct2() struct{ test bool } {
	return struct{ test bool }{false}
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
