// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otlphttptest

import (
	"bytes"
	"strings"
	"testing"

	otlpcollectlogs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	otlpcollectmetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	otlpcollecttrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	otlplogs "go.opentelemetry.io/proto/otlp/logs/v1"
	otlpmetrics "go.opentelemetry.io/proto/otlp/metrics/v1"
	otlptrace "go.opentelemetry.io/proto/otlp/trace/v1"
	httpcollectlogs "go.opentelemetry.io/proto/otlphttp/collector/logs/v1"
	httpcollectmetrics "go.opentelemetry.io/proto/otlphttp/collector/metrics/v1"
	httpcollecttrace "go.opentelemetry.io/proto/otlphttp/collector/trace/v1"
	httplogs "go.opentelemetry.io/proto/otlphttp/logs/v1"
	httpmetrics "go.opentelemetry.io/proto/otlphttp/metrics/v1"
	httptrace "go.opentelemetry.io/proto/otlphttp/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestCanonicalAndHTTPMessagesCoexist(t *testing.T) {
	const schemaURL = "https://opentelemetry.io/schemas/1.40.0"
	tests := []struct {
		name      string
		canonical proto.Message
		http      proto.Message
	}{
		{
			"traces",
			&otlpcollecttrace.ExportTraceServiceRequest{ResourceSpans: []*otlptrace.ResourceSpans{{SchemaUrl: schemaURL}}},
			&httpcollecttrace.ExportTraceServiceRequest{ResourceSpans: []*httptrace.ResourceSpans{{SchemaUrl: schemaURL}}},
		},
		{
			"metrics",
			&otlpcollectmetrics.ExportMetricsServiceRequest{ResourceMetrics: []*otlpmetrics.ResourceMetrics{{SchemaUrl: schemaURL}}},
			&httpcollectmetrics.ExportMetricsServiceRequest{ResourceMetrics: []*httpmetrics.ResourceMetrics{{SchemaUrl: schemaURL}}},
		},
		{
			"logs",
			&otlpcollectlogs.ExportLogsServiceRequest{ResourceLogs: []*otlplogs.ResourceLogs{{SchemaUrl: schemaURL}}},
			&httpcollectlogs.ExportLogsServiceRequest{ResourceLogs: []*httplogs.ResourceLogs{{SchemaUrl: schemaURL}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			canonicalBinary, err := proto.Marshal(test.canonical)
			if err != nil {
				t.Fatalf("marshal canonical binary protobuf: %v", err)
			}
			httpBinary, err := proto.Marshal(test.http)
			if err != nil {
				t.Fatalf("marshal HTTP binary protobuf: %v", err)
			}
			if !bytes.Equal(httpBinary, canonicalBinary) {
				t.Errorf("HTTP binary protobuf = %x, canonical = %x", httpBinary, canonicalBinary)
			}

			canonicalJSON, err := protojson.Marshal(test.canonical)
			if err != nil {
				t.Fatalf("marshal canonical JSON protobuf: %v", err)
			}
			httpJSON, err := protojson.Marshal(test.http)
			if err != nil {
				t.Fatalf("marshal HTTP JSON protobuf: %v", err)
			}
			if !bytes.Equal(httpJSON, canonicalJSON) {
				t.Errorf("HTTP JSON protobuf = %s, canonical = %s", httpJSON, canonicalJSON)
			}

			canonicalDesc := test.canonical.ProtoReflect().Descriptor()
			httpDesc := test.http.ProtoReflect().Descriptor()
			canonicalName := string(canonicalDesc.FullName())
			httpName := string(httpDesc.FullName())
			wantHTTPName := strings.Replace(canonicalName, "opentelemetry.proto.", "opentelemetry.proto.otlphttp.", 1)
			if httpName != wantHTTPName {
				t.Errorf("HTTP full name = %q, want %q", httpName, wantHTTPName)
			}

			canonicalPath := canonicalDesc.ParentFile().Path()
			httpPath := httpDesc.ParentFile().Path()
			wantHTTPPath := strings.Replace(canonicalPath, "opentelemetry/proto/", "opentelemetry/proto/otlphttp/", 1)
			if httpPath != wantHTTPPath {
				t.Errorf("HTTP descriptor path = %q, want %q", httpPath, wantHTTPPath)
			}

			if _, err := protoregistry.GlobalFiles.FindDescriptorByName(canonicalDesc.FullName()); err != nil {
				t.Errorf("canonical descriptor is not registered: %v", err)
			}
			if _, err := protoregistry.GlobalFiles.FindDescriptorByName(httpDesc.FullName()); err != nil {
				t.Errorf("HTTP descriptor is not registered: %v", err)
			}
		})
	}
}
