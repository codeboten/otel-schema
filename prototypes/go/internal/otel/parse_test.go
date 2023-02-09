package otel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestConfigure(t *testing.T) {
	tt := []struct {
		name   string
		config OpenTelemetryConfiguration
	}{
		{
			name: "sdk disabled",
			config: OpenTelemetryConfiguration{
				Sdk: SDK{
					Disabled: true,
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			Configure(logger, tc.config)
		})
	}
}

func TestGetExporter(t *testing.T) {
	tt := []struct {
		name             string
		exporterType     string
		expectedExporter interface{}
		expectError      bool
		cfg              interface{}
	}{
		{
			name:             "invalid exporter type",
			exporterType:     "invalid-type",
			expectedExporter: nil,
			expectError:      true,
		},
		{
			name:         "otlp exporter",
			exporterType: "otlp/exporter1",
			cfg: map[string]interface{}{
				"otlp/exporter1": Otlp{
					Insecure: &[]bool{true}[0],
				},
			},
			expectedExporter: nil,
			expectError:      false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := getExporter(context.Background(), tc.exporterType, tc.cfg)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			// require.Equal(t, tc.expectedExporter, exporter)
		})
	}
}

func TestConfigureTracerProvider(t *testing.T) {
	tt := []struct {
		name string
	}{
		{
			name: "default",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			configureTracerProvider(logger, NoOpConfig.Sdk.TracerProvider)
		})
	}
}
