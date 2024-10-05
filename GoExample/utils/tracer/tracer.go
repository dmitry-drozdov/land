package tracer

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	T trace.Tracer
}

func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return FromContext(ctx).startSpan(ctx, spanName, opts...)
}

func (s *Tracer) startSpan(
	ctx context.Context,
	spanName string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	return s.T.Start(ctx, spanName, opts...)
}

func Start(
	ctx context.Context,
	spanName string,
	opts ...trace.SpanStartOption,
) (context.Context, func(error)) {
	ctx, span := StartSpan(ctx, spanName, opts...)
	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}
