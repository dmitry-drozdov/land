package tracer

import (
	"context"
	"crypto/tls"
	"fmt"

	"gitlab.services.mts.ru/lp/backend/libs/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc/credentials"
)

func NewTracer(ctx context.Context, opts ...Option) func() error {
	options := &tracerOptions{}

	for _, opt := range opts {
		if err := opt(options); err != nil {
			panic(fmt.Errorf("failed to apply option: %w", err))
		}
	}

	attr := []attribute.KeyValue{semconv.ServiceName(options.serviceName)}

	res, err := resource.New(ctx, resource.WithAttributes(attr...))
	if err != nil {
		logger.Fatalf(ctx, "failed to create resource: %w", err)
	}

	grpcOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(options.endpoint),
	}

	if options.insecure {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure())
	}
	if options.tlsConfig != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	}

	exporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(grpcOpts...))
	if err != nil {
		logger.Fatalf(ctx, "failed to create span exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := otel.Tracer(options.serviceName)

	ReplaceGlobals(&Tracer{tracer})

	return func() error {
		return tracerProvider.Shutdown(ctx)
	}
}
