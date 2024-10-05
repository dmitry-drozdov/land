package tracer

import (
	"sync"

	"go.opentelemetry.io/otel/trace/noop"
)

var (
	_globalMu sync.RWMutex
	_globalT  = newNoopTracer()
)

func newNoopTracer() *Tracer {
	t := noop.NewTracerProvider().Tracer("unknown")

	return &Tracer{
		T: t,
	}
}

func T() *Tracer {
	_globalMu.RLock()
	t := _globalT
	_globalMu.RUnlock()
	return t
}

func ReplaceGlobals(tracer *Tracer) func() {
	_globalMu.Lock()
	prev := _globalT
	_globalT = tracer
	_globalMu.Unlock()
	return func() { ReplaceGlobals(prev) }
}
