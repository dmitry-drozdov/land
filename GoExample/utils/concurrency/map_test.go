package concurrency

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func Test_Map(t *testing.T) {
	t.Run("100000 + mod 7", func(t *testing.T) {
		mp := NewMap[string, int](7, func(s string) int64 { return int64(len(s) % 7) })
		cnt := 100000
		for i := range cnt {
			k := fmt.Sprint((i<<10)%(3*i-7), i)
			mp.Add(k, i)
		}

		// mp.Distribution()
		// panic(0)

		assert.Equal(t, (0+(cnt-1))*cnt/2, mp.Sum())
	})

	t.Run("1000000 + mod 5", func(t *testing.T) {
		mp := NewMap[string, int](5, func(s string) int64 { return int64(len(s) % 5) })
		cnt := 1000000
		for i := range cnt {
			k := fmt.Sprint((i<<10)%(3*i-7), i)
			mp.Add(k, i)
		}

		assert.Equal(t, (0+(cnt-1))*cnt/2, mp.Sum())
	})

	t.Run("1000000 + mod 5 + inc", func(t *testing.T) {
		mp := NewMap[string, int](5, func(s string) int64 { return int64(len(s) % 5) })
		cnt := 1000000
		for i := range cnt {
			k := fmt.Sprint((i<<10)%(3*i-7), i)
			mp.Add(k, i)
			mp.Inc(k)
		}

		assert.Equal(t, (0+(cnt-1))*cnt/2+cnt, mp.Sum())
	})

	t.Run("1000000 + one part", func(t *testing.T) {
		mp := NewMap[string, int](1, func(s string) int64 { return 0 })
		cnt := 1000000
		for i := range cnt {
			k := fmt.Sprint((i<<10)%(3*i-7), i)
			mp.Add(k, 2*i)
		}

		assert.Equal(t, (0+(cnt-1)*2)*cnt/2, mp.Sum())
	})

	t.Run("go + 10000000 + mod 9", func(t *testing.T) {
		mp := NewMap[int, int](9, func(s int) int64 { return int64(s % 9) })
		cnt := 10000000
		g := errgroup.Group{}
		g.SetLimit(9)
		t0 := time.Now()
		for i := range cnt {
			g.Go(func() error {
				mp.Add(i, 2*i)
				return nil
			})
		}
		g.Wait()
		dt := time.Since(t0)
		//assert.Equal(t, (0+(cnt-1))*cnt/2+cnt /*Inc*/, mp.Sum())

		mp2 := map[int]int{}
		mx := sync.Mutex{}
		g.SetLimit(9)
		t0 = time.Now()
		for i := range cnt {
			g.Go(func() error {
				mx.Lock()
				mp2[i] = i * 2
				mx.Unlock()
				return nil
			})
		}
		g.Wait()
		dt2 := time.Since(t0)

		assert.Less(t, dt, dt2)
		assert.Greater(t, dt2-dt, time.Millisecond*500)
	})
}
