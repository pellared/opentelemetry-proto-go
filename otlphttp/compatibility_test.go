// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otlphttp_test

import (
	"bytes"
	"testing"

	collectlogs "go.opentelemetry.io/proto/otlphttp/collector/logs/v1"
	collectmetrics "go.opentelemetry.io/proto/otlphttp/collector/metrics/v1"
	collecttrace "go.opentelemetry.io/proto/otlphttp/collector/trace/v1"
	common "go.opentelemetry.io/proto/otlphttp/common/v1"
	logs "go.opentelemetry.io/proto/otlphttp/logs/v1"
	metrics "go.opentelemetry.io/proto/otlphttp/metrics/v1"
	resource "go.opentelemetry.io/proto/otlphttp/resource/v1"
	trace "go.opentelemetry.io/proto/otlphttp/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestDescriptorNamespace(t *testing.T) {
	tests := []struct {
		message proto.Message
		name    string
		path    string
	}{
		{&common.AnyValue{}, "opentelemetry.proto.otlphttp.common.v1.AnyValue", "opentelemetry/proto/otlphttp/common/v1/common.proto"},
		{&resource.Resource{}, "opentelemetry.proto.otlphttp.resource.v1.Resource", "opentelemetry/proto/otlphttp/resource/v1/resource.proto"},
		{&trace.ResourceSpans{}, "opentelemetry.proto.otlphttp.trace.v1.ResourceSpans", "opentelemetry/proto/otlphttp/trace/v1/trace.proto"},
		{&metrics.ResourceMetrics{}, "opentelemetry.proto.otlphttp.metrics.v1.ResourceMetrics", "opentelemetry/proto/otlphttp/metrics/v1/metrics.proto"},
		{&logs.ResourceLogs{}, "opentelemetry.proto.otlphttp.logs.v1.ResourceLogs", "opentelemetry/proto/otlphttp/logs/v1/logs.proto"},
		{&collecttrace.ExportTraceServiceRequest{}, "opentelemetry.proto.otlphttp.collector.trace.v1.ExportTraceServiceRequest", "opentelemetry/proto/otlphttp/collector/trace/v1/trace_service.proto"},
		{&collectmetrics.ExportMetricsServiceRequest{}, "opentelemetry.proto.otlphttp.collector.metrics.v1.ExportMetricsServiceRequest", "opentelemetry/proto/otlphttp/collector/metrics/v1/metrics_service.proto"},
		{&collectlogs.ExportLogsServiceRequest{}, "opentelemetry.proto.otlphttp.collector.logs.v1.ExportLogsServiceRequest", "opentelemetry/proto/otlphttp/collector/logs/v1/logs_service.proto"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			desc := test.message.ProtoReflect().Descriptor()
			if got := string(desc.FullName()); got != test.name {
				t.Errorf("full name = %q, want %q", got, test.name)
			}
			if got := desc.ParentFile().Path(); got != test.path {
				t.Errorf("descriptor path = %q, want %q", got, test.path)
			}
		})
	}
}

func TestCollectorRequestWireFormat(t *testing.T) {
	const schemaURL = "schema"
	// Field 1 is the repeated resource message in every OTLP export request.
	// Field 3 in each resource message is schema_url. This is the canonical OTLP
	// protobuf encoding for one resource message containing only that value.
	wantBinary := []byte{0x0a, 0x08, 0x1a, 0x06, 's', 'c', 'h', 'e', 'm', 'a'}

	tests := []struct {
		name     string
		message  proto.Message
		wantJSON string
	}{
		{
			"traces",
			&collecttrace.ExportTraceServiceRequest{ResourceSpans: []*trace.ResourceSpans{{SchemaUrl: schemaURL}}},
			`{"resourceSpans":[{"schemaUrl":"schema"}]}`,
		},
		{
			"metrics",
			&collectmetrics.ExportMetricsServiceRequest{ResourceMetrics: []*metrics.ResourceMetrics{{SchemaUrl: schemaURL}}},
			`{"resourceMetrics":[{"schemaUrl":"schema"}]}`,
		},
		{
			"logs",
			&collectlogs.ExportLogsServiceRequest{ResourceLogs: []*logs.ResourceLogs{{SchemaUrl: schemaURL}}},
			`{"resourceLogs":[{"schemaUrl":"schema"}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotBinary, err := proto.Marshal(test.message)
			if err != nil {
				t.Fatalf("marshal binary protobuf: %v", err)
			}
			if !bytes.Equal(gotBinary, wantBinary) {
				t.Errorf("binary protobuf = %x, want %x", gotBinary, wantBinary)
			}

			gotJSON, err := protojson.Marshal(test.message)
			if err != nil {
				t.Fatalf("marshal JSON protobuf: %v", err)
			}
			if string(gotJSON) != test.wantJSON {
				t.Errorf("JSON protobuf = %s, want %s", gotJSON, test.wantJSON)
			}
		})
	}
}
