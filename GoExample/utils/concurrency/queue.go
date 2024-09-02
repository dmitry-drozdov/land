package concurrency

import (
	"sync"
)

type F func() error

type Queue struct {
	ch chan F
	wg sync.WaitGroup
}

func NewQueue() *Queue {
	q := &Queue{
		ch: make(chan F, 1024),
		wg: sync.WaitGroup{},
	}
	for range 4 {
		q.wg.Add(1)
		go func() {
			defer q.wg.Done()
			q.run()
		}()
	}
	return q
}

func (q *Queue) Add(f F) {
	q.ch <- f
}

func (q *Queue) Wait() {
	close(q.ch)
	q.wg.Wait()
}

func (q *Queue) run() {
	for f := range q.ch {
		err := f()
		if err != nil {
			panic(err)
		}
	}
}
