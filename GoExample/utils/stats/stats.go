package stats

import "fmt"

type Stats struct {
	Good int
	Bad  int
}

func (s *Stats) Ok() {
	s.Good++
}

func (s *Stats) Fail() {
	s.Bad++
}

func (s *Stats) Add(other *Stats) {
	if other != nil {
		s.Good += other.Good
		s.Bad += other.Bad
	}
}

func (s Stats) total() int {
	return s.Good + s.Bad
}

func (s Stats) String() string {
	return fmt.Sprintf("bad [%d (%.1f%%)] good [%d (%.1f%%)] total [%d]",
		s.Bad, ratio(s.Bad, s.total()),
		s.Good, ratio(s.Good, s.total()),
		s.total(),
	)
}

func ratio(x, y int) float64 {
	return float64(x) / float64(y) * 100.0
}
