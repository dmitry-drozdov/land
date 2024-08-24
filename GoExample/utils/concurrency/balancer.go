package concurrency

import (
	"sync"
)

type Balancer struct {
	cnt int
	mx  sync.Mutex
}

func (b *Balancer) MainAction() {
	b.mx.Lock()
	defer b.mx.Unlock()
	b.cnt++
}

func (b *Balancer) CanSubAction() bool {
	b.mx.Lock()
	defer b.mx.Unlock()
	if b.cnt > 0 {
		b.cnt--
		return true
	}
	return false
}
