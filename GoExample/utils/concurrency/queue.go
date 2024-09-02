package concurrency

type F func() error

type Queue struct {
	ch   chan F
	done chan struct{}
}

func NewQueue() *Queue {
	q := &Queue{
		ch:   make(chan F, 1024),
		done: make(chan struct{}),
	}
	go q.run()
	return q
}

func (q *Queue) Add(f F) {
	q.ch <- f
}

func (q *Queue) Wait() {
	close(q.ch)
	<-q.done
}

func (q *Queue) run() {
	for f := range q.ch {
		err := f()
		if err != nil {
			panic(err)
		}
	}
	close(q.done)
}
