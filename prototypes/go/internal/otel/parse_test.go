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
		{
			name:   "no-op config",
			config: NoOpConfig,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			cleanup := Configure(logger, tc.config)
			require.NoError(t, cleanup())
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
			name:         "otlp exporter insecure",
			exporterType: "otlp/exporter1",
			cfg: map[string]interface{}{
				"insecure": true,
				"endpoint": "localhost:4317",
			},
			expectedExporter: nil,
			expectError:      false,
		},
		{
			name:         "otlp exporter",
			exporterType: "otlp/exporter1",
			cfg: map[string]interface{}{
				"endpoint": "localhost:4317",
			},
			expectedExporter: nil,
			expectError:      false,
		},
		{
			name:         "otlp http exporter",
			exporterType: "otlp/exporter2",
			cfg: map[string]interface{}{
				"endpoint": "localhost:4317",
				"protocol": "http/protobuf",
			},
			expectedExporter: nil,
			expectError:      false,
		},
		{
			name:             "console exporter",
			exporterType:     "console",
			cfg:              map[string]interface{}{},
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
			// TODO: add test to validate exporter
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
