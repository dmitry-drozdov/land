package tracer

import (
	"context"
)

type ctxTracer struct{}

// FromContext возвращает трассировщик из контекста
func FromContext(ctx context.Context) *Tracer {
	return getTracer(ctx)
}

func getTracer(ctx context.Context) *Tracer {
	if t, ok := ctx.Value(ctxTracer{}).(*Tracer); ok {
		return t
	}
	return T()
}
