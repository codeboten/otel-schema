package otel

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
	"sigs.k8s.io/yaml"
)

type sdk struct {
	Disabled bool           `json:"disabled"`
	Traces   tracerProvider `json:"tracer_provider"`
}

type processor struct {
	Type string         `json:"type"`
	Args map[string]any `json:"args"`
}

type exporter = map[string]any

// type exporter struct {
// 	Args map[string]any
// }

type tracerProvider struct {
	Processors []processor         `json:"span_processors"`
	Exporters  map[string]exporter `json:"exporters"`
}

type Config struct {
	Sdk sdk `json:"sdk"`
}

var NoOpConfig = OpenTelemetryConfiguration{}

func ParseAndValidateFromConfigFile(logger *zap.Logger, filename string, schema string) (OpenTelemetryConfiguration, error) {
	path, err := filepath.Abs(schema)
	if err != nil {
		return NoOpConfig, err
	}
	jsonSchema := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", path))

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return NoOpConfig, err
	}
	buf, err = yaml.YAMLToJSON(buf)
	if err != nil {
		return NoOpConfig, err
	}
	configuration := gojsonschema.NewBytesLoader(buf)

	result, err := gojsonschema.Validate(jsonSchema, configuration)
	if err != nil {
		return NoOpConfig, err
	}

	if result.Valid() {
		logger.Info("The document is valid")
	} else {
		logger.Warn("The document is not valid. see errors:")
		for _, desc := range result.Errors() {
			logger.Warn("", zap.Any("error", desc))
		}
	}
	cfg := OpenTelemetryConfiguration{}
	// TODO: investigate if it would be worth using gojsonschema.Result instead of
	// unmarshaling the data into a custom struct.
	err = json.Unmarshal(buf, &cfg)
	if err != nil {
		logger.Error("Failed to unmarshal configuration")
		return NoOpConfig, err
	}
	return cfg, nil
}

func headersToStringMap(headers Headers) map[string]string {
	result := make(map[string]string, len(headers))
	for key, val := range headers {
		result[key] = fmt.Sprintf("%v", val)
	}
	return result
}

func otlpToExporter(ctx context.Context, cfg Otlp) (*otlptrace.Exporter, error) {
	var secureOpt otlptracegrpc.Option
	if *cfg.Insecure {
		secureOpt = otlptracegrpc.WithInsecure()
	} else {
		secureOpt = otlptracegrpc.WithTLSCredentials(
			credentials.NewClientTLSFromCert(nil, ""),
		)
	}

	if cfg.Protocol == "http/protobuf" {
		return otlptrace.New(
			ctx,
			otlptracehttp.NewClient(
				otlptracehttp.WithEndpoint(cfg.Endpoint),
			),
		)
	}

	return otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			secureOpt,
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithHeaders(headersToStringMap(cfg.Headers)),
		),
	)
}

func getExporter(ctx context.Context, id string, e interface{}) (*otlptrace.Exporter, error) {
	if e == nil {
		return nil, fmt.Errorf("invalid configuration")
	}
	var exporterType, identifer string
	parts := strings.SplitN(id, "/", 2)
	exporterType = parts[0]
	if len(parts) > 1 {
		identifer = parts[1]
	}

	switch exporterType {
	case "otlp":
		return otlpToExporter(ctx, rawToOtlp(e))
	}

	return nil, fmt.Errorf("invalid exporter: %s %s", id, identifer)
}

func configureTracerProvider(logger *zap.Logger, tpCfg *TracerProvider) (func(context.Context) error, error) {
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // TODO: configure this
	}
	for _, p := range tpCfg.SpanProcessors {
		if _, ok := tpCfg.Exporters[p.Args.Exporter]; !ok {
			return nil, fmt.Errorf("exporter %s not found", p.Args.Exporter)
		}
		exporter, _ := getExporter(context.Background(), p.Args.Exporter, tpCfg.Exporters[p.Args.Exporter])

		switch p.Type {
		case "batch":
			opts = append(opts, sdktrace.WithBatcher(exporter))
		default:
			logger.Warn("processor type unsupported", zap.String("type", p.Type))
		}
	}

	provider := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(provider)
	return func(ctx context.Context) error {
		if err := provider.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	}, nil
}

func Configure(logger *zap.Logger, cfg OpenTelemetryConfiguration) func() {
	if cfg.Sdk.Disabled {
		logger.Info("SDK disabled")
		return func() {}
	}

	cleanupTracerProvider, err := configureTracerProvider(logger, cfg.Sdk.TracerProvider)
	if err != nil {
		logger.Error("error configuring tracer provider", zap.Error(err))
		return func() {}
	}

	return func() {
		cleanupTracerProvider(context.Background())
	}
}
