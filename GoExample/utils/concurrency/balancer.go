package concurrency

import (
	"sync"
)

type Balancer struct {
	cntMain int
	cntSub  int
	mx      sync.Mutex
}

func (b *Balancer) MainAction(points int) {
	b.mx.Lock()
	defer b.mx.Unlock()
	b.cntMain += points
}

func (b *Balancer) CanSubAction() bool {
	b.mx.Lock()
	defer b.mx.Unlock()
	if b.cntSub < b.cntMain {
		b.cntSub++
		return true
	}
	return false
}

func (b *Balancer) CntMain() int {
	return b.cntMain
}

func (b *Balancer) CntSub() int {
	return b.cntSub
}
