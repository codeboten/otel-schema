package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	conf "github.com/MrAlias/otel-schema/prototype/go/internal/otel"
)

func main() {
	errs := []error{}
	eh := otel.ErrorHandlerFunc(func(e error) { errs = append(errs, e) })
	otel.SetErrorHandler(eh)
	logger := zap.NewExample()
	defer logger.Sync() // flushes buffer, if any
	cfg, err := conf.ParseAndValidateFromConfigFile(logger, "../../config.yaml", "../../json_schema/schema/schema.json")
	if err != nil {
		logger.Error(err.Error())
		return
	}
	cleanup := conf.Configure(logger, cfg)
	defer cleanup()
	tracer := otel.GetTracerProvider().Tracer("ExampleService")
	_, span := tracer.Start(context.Background(), "foo")
	span.End()
	fmt.Printf("%v", errs)
}

// package app

// import (
// 	"context"
// 	"fmt"
// 	"log"

//
// 	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
// 	"go.opentelemetry.io/otel/sdk/resource"
// 	sdktrace "go.opentelemetry.io/otel/sdk/trace"
// 	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
// 	"go.opentelemetry.io/otel/trace"
// )

// var tracer trace.Tracer

// func newExporter(ctx context.Context)  /* (someExporter.Exporter, error) */ {
// 	// Your preferred exporter: console, jaeger, zipkin, OTLP, etc.
// }

// func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
// 	// Ensure default SDK resources and the required service name are set.
// 	r, err := resource.Merge(
// 		resource.Default(),
// 		resource.NewWithAttributes(
// 			semconv.SchemaURL,
// 			semconv.ServiceNameKey.String("ExampleService"),
// 		),
// 	)

// 	if err != nil {
// 		panic(err)
// 	}

// 	return sdktrace.NewTracerProvider(
// 		sdktrace.WithBatcher(exp),
// 		sdktrace.WithResource(r),
// 	)
// }

// func main() {
// 	ctx := context.Background()

// 	exp, err := newExporter(ctx)
// 	if err != nil {
// 		log.Fatalf("failed to initialize exporter: %v", err)
// 	}

// 	// Create a new tracer provider with a batch span processor and the given exporter.
// 	tp := newTraceProvider(exp)

// 	// Handle shutdown properly so nothing leaks.
// 	defer func() { _ = tp.Shutdown(ctx) }()

// 	otel.SetTracerProvider(tp)

// 	// Finally, set the tracer that can be used for this package.
// 	tracer = tp.Tracer("ExampleService")
// }
