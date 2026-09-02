// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package slimotlpcompat

import (
	"bytes"
	"strings"
	"testing"

	otlpcollectlogs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	otlpcollectmetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	otlpcollectprofiles "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"
	otlpcollecttrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	otlpcommon "go.opentelemetry.io/proto/otlp/common/v1"
	otlplogs "go.opentelemetry.io/proto/otlp/logs/v1"
	otlpmetrics "go.opentelemetry.io/proto/otlp/metrics/v1"
	otlpprocesscontext "go.opentelemetry.io/proto/otlp/processcontext/v1development"
	otlpprofiles "go.opentelemetry.io/proto/otlp/profiles/v1development"
	otlpresource "go.opentelemetry.io/proto/otlp/resource/v1"
	otlptrace "go.opentelemetry.io/proto/otlp/trace/v1"
	slimcollectlogs "go.opentelemetry.io/proto/slim/otlp/collector/logs/v1"
	slimcollectmetrics "go.opentelemetry.io/proto/slim/otlp/collector/metrics/v1"
	slimcollectprofiles "go.opentelemetry.io/proto/slim/otlp/collector/profiles/v1development"
	slimcollecttrace "go.opentelemetry.io/proto/slim/otlp/collector/trace/v1"
	slimcommon "go.opentelemetry.io/proto/slim/otlp/common/v1"
	slimlogs "go.opentelemetry.io/proto/slim/otlp/logs/v1"
	slimmetrics "go.opentelemetry.io/proto/slim/otlp/metrics/v1"
	slimprocesscontext "go.opentelemetry.io/proto/slim/otlp/processcontext/v1development"
	slimprofiles "go.opentelemetry.io/proto/slim/otlp/profiles/v1development"
	slimresource "go.opentelemetry.io/proto/slim/otlp/resource/v1"
	slimtrace "go.opentelemetry.io/proto/slim/otlp/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	canonicalPathPrefix = "opentelemetry/proto/"
	slimPathPrefix      = "opentelemetry/proto/slim/"
	canonicalNamePrefix = "opentelemetry.proto."
	slimNamePrefix      = "opentelemetry.proto.slim."
)

type filePair struct {
	canonical protoreflect.FileDescriptor
	slim      protoreflect.FileDescriptor
}

var allOTLPFiles = []filePair{
	{otlpcommon.File_opentelemetry_proto_common_v1_common_proto, slimcommon.File_opentelemetry_proto_common_v1_common_proto},
	{otlpresource.File_opentelemetry_proto_resource_v1_resource_proto, slimresource.File_opentelemetry_proto_resource_v1_resource_proto},
	{otlptrace.File_opentelemetry_proto_trace_v1_trace_proto, slimtrace.File_opentelemetry_proto_trace_v1_trace_proto},
	{otlpmetrics.File_opentelemetry_proto_metrics_v1_metrics_proto, slimmetrics.File_opentelemetry_proto_metrics_v1_metrics_proto},
	{otlplogs.File_opentelemetry_proto_logs_v1_logs_proto, slimlogs.File_opentelemetry_proto_logs_v1_logs_proto},
	{otlpcollecttrace.File_opentelemetry_proto_collector_trace_v1_trace_service_proto, slimcollecttrace.File_opentelemetry_proto_collector_trace_v1_trace_service_proto},
	{otlpcollectmetrics.File_opentelemetry_proto_collector_metrics_v1_metrics_service_proto, slimcollectmetrics.File_opentelemetry_proto_collector_metrics_v1_metrics_service_proto},
	{otlpcollectlogs.File_opentelemetry_proto_collector_logs_v1_logs_service_proto, slimcollectlogs.File_opentelemetry_proto_collector_logs_v1_logs_service_proto},
	{otlpprocesscontext.File_opentelemetry_proto_processcontext_v1development_process_context_proto, slimprocesscontext.File_opentelemetry_proto_processcontext_v1development_process_context_proto},
	{otlpprofiles.File_opentelemetry_proto_profiles_v1development_profiles_proto, slimprofiles.File_opentelemetry_proto_profiles_v1development_profiles_proto},
	{otlpcollectprofiles.File_opentelemetry_proto_collector_profiles_v1development_profiles_service_proto, slimcollectprofiles.File_opentelemetry_proto_collector_profiles_v1development_profiles_service_proto},
}

func TestCanonicalAndSlimDescriptorsCoexist(t *testing.T) {
	if got, want := len(allOTLPFiles), 11; got != want {
		t.Fatalf("tested file descriptors = %d, want %d", got, want)
	}

	for _, files := range allOTLPFiles {
		canonicalPath := files.canonical.Path()
		t.Run(canonicalPath, func(t *testing.T) {
			wantSlimPath := strings.Replace(canonicalPath, canonicalPathPrefix, slimPathPrefix, 1)
			if got := files.slim.Path(); got != wantSlimPath {
				t.Errorf("slim descriptor path = %q, want %q", got, wantSlimPath)
			}

			canonicalPackage := string(files.canonical.Package())
			wantSlimPackage := strings.Replace(canonicalPackage, canonicalNamePrefix, slimNamePrefix, 1)
			if got := string(files.slim.Package()); got != wantSlimPackage {
				t.Errorf("slim protobuf package = %q, want %q", got, wantSlimPackage)
			}

			assertRegisteredFile(t, files.canonical)
			assertRegisteredFile(t, files.slim)
			assertSlimFileNamespace(t, files.slim)
		})
	}
}

func TestCanonicalAndSlimMessagesHaveCompatibleEncoding(t *testing.T) {
	const schemaURL = "https://opentelemetry.io/schemas/1.40.0"
	tests := []struct {
		name      string
		canonical proto.Message
		slim      proto.Message
	}{
		{
			"traces",
			&otlpcollecttrace.ExportTraceServiceRequest{ResourceSpans: []*otlptrace.ResourceSpans{{SchemaUrl: schemaURL}}},
			&slimcollecttrace.ExportTraceServiceRequest{ResourceSpans: []*slimtrace.ResourceSpans{{SchemaUrl: schemaURL}}},
		},
		{
			"metrics",
			&otlpcollectmetrics.ExportMetricsServiceRequest{ResourceMetrics: []*otlpmetrics.ResourceMetrics{{SchemaUrl: schemaURL}}},
			&slimcollectmetrics.ExportMetricsServiceRequest{ResourceMetrics: []*slimmetrics.ResourceMetrics{{SchemaUrl: schemaURL}}},
		},
		{
			"logs",
			&otlpcollectlogs.ExportLogsServiceRequest{ResourceLogs: []*otlplogs.ResourceLogs{{SchemaUrl: schemaURL}}},
			&slimcollectlogs.ExportLogsServiceRequest{ResourceLogs: []*slimlogs.ResourceLogs{{SchemaUrl: schemaURL}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			canonicalBinary, err := proto.Marshal(test.canonical)
			if err != nil {
				t.Fatalf("marshal canonical binary protobuf: %v", err)
			}
			slimBinary, err := proto.Marshal(test.slim)
			if err != nil {
				t.Fatalf("marshal slim binary protobuf: %v", err)
			}
			if !bytes.Equal(slimBinary, canonicalBinary) {
				t.Errorf("slim binary protobuf = %x, canonical = %x", slimBinary, canonicalBinary)
			}

			canonicalJSON, err := protojson.Marshal(test.canonical)
			if err != nil {
				t.Fatalf("marshal canonical JSON protobuf: %v", err)
			}
			slimJSON, err := protojson.Marshal(test.slim)
			if err != nil {
				t.Fatalf("marshal slim JSON protobuf: %v", err)
			}
			if !bytes.Equal(slimJSON, canonicalJSON) {
				t.Errorf("slim JSON protobuf = %s, canonical = %s", slimJSON, canonicalJSON)
			}
		})
	}
}

func TestAnyTypeURLsUseDistinctNamespaces(t *testing.T) {
	canonicalAny, err := anypb.New(&otlpcommon.AnyValue{})
	if err != nil {
		t.Fatalf("pack canonical message: %v", err)
	}
	slimAny, err := anypb.New(&slimcommon.AnyValue{})
	if err != nil {
		t.Fatalf("pack slim message: %v", err)
	}

	if got, want := canonicalAny.TypeUrl, "type.googleapis.com/opentelemetry.proto.common.v1.AnyValue"; got != want {
		t.Errorf("canonical type URL = %q, want %q", got, want)
	}
	if got, want := slimAny.TypeUrl, "type.googleapis.com/opentelemetry.proto.slim.common.v1.AnyValue"; got != want {
		t.Errorf("slim type URL = %q, want %q", got, want)
	}

	resolvedCanonical, err := anypb.UnmarshalNew(canonicalAny, proto.UnmarshalOptions{})
	if err != nil {
		t.Fatalf("resolve canonical type URL: %v", err)
	}
	if _, ok := resolvedCanonical.(*otlpcommon.AnyValue); !ok {
		t.Errorf("canonical type URL resolved to %T", resolvedCanonical)
	}
	resolvedSlim, err := anypb.UnmarshalNew(slimAny, proto.UnmarshalOptions{})
	if err != nil {
		t.Fatalf("resolve slim type URL: %v", err)
	}
	if _, ok := resolvedSlim.(*slimcommon.AnyValue); !ok {
		t.Errorf("slim type URL resolved to %T", resolvedSlim)
	}
}

func TestCoreSchemasDoNotEmbedGoogleProtobufAny(t *testing.T) {
	for _, files := range allOTLPFiles {
		visitMessages(files.slim.Messages(), func(message protoreflect.MessageDescriptor) {
			fields := message.Fields()
			for i := 0; i < fields.Len(); i++ {
				field := fields.Get(i)
				if field.Message() != nil && field.Message().FullName() == "google.protobuf.Any" {
					t.Errorf("%s embeds google.protobuf.Any", field.FullName())
				}
			}
		})
	}
}

func assertRegisteredFile(t *testing.T, descriptor protoreflect.FileDescriptor) {
	t.Helper()
	got, err := protoregistry.GlobalFiles.FindFileByPath(descriptor.Path())
	if err != nil {
		t.Errorf("descriptor path %q is not registered: %v", descriptor.Path(), err)
		return
	}
	if got != descriptor {
		t.Errorf("descriptor path %q resolved to a different file", descriptor.Path())
	}
}

func assertSlimFileNamespace(t *testing.T, file protoreflect.FileDescriptor) {
	t.Helper()

	imports := file.Imports()
	for i := 0; i < imports.Len(); i++ {
		path := imports.Get(i).Path()
		if strings.HasPrefix(path, canonicalPathPrefix) && !strings.HasPrefix(path, slimPathPrefix) {
			t.Errorf("%s imports canonical OTLP descriptor %q", file.Path(), path)
		}
	}

	checkDescriptor := func(descriptor protoreflect.Descriptor) {
		t.Helper()
		name := string(descriptor.FullName())
		if strings.HasPrefix(name, canonicalNamePrefix) && !strings.HasPrefix(name, slimNamePrefix) {
			t.Errorf("descriptor %q does not use the slim namespace", name)
		}
		if _, err := protoregistry.GlobalFiles.FindDescriptorByName(descriptor.FullName()); err != nil {
			t.Errorf("descriptor %q is not registered: %v", name, err)
		}
	}

	visitEnums(file.Enums(), checkDescriptor)
	visitMessages(file.Messages(), func(message protoreflect.MessageDescriptor) {
		checkDescriptor(message)
		visitEnums(message.Enums(), checkDescriptor)

		fields := message.Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			checkDescriptor(field)
			assertSlimReference(t, field.Message())
			assertSlimReference(t, field.Enum())
		}

		oneofs := message.Oneofs()
		for i := 0; i < oneofs.Len(); i++ {
			checkDescriptor(oneofs.Get(i))
		}
	})

	extensions := file.Extensions()
	for i := 0; i < extensions.Len(); i++ {
		checkDescriptor(extensions.Get(i))
	}

	services := file.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		checkDescriptor(service)
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)
			checkDescriptor(method)
			assertSlimReference(t, method.Input())
			assertSlimReference(t, method.Output())
		}
	}
}

func assertSlimReference(t *testing.T, descriptor protoreflect.Descriptor) {
	t.Helper()
	if descriptor == nil {
		return
	}
	name := string(descriptor.FullName())
	if strings.HasPrefix(name, canonicalNamePrefix) && !strings.HasPrefix(name, slimNamePrefix) {
		t.Errorf("internal reference %q does not use the slim namespace", name)
	}
}

func visitMessages(messages protoreflect.MessageDescriptors, visit func(protoreflect.MessageDescriptor)) {
	for i := 0; i < messages.Len(); i++ {
		message := messages.Get(i)
		visit(message)
		visitMessages(message.Messages(), visit)
	}
}

func visitEnums(enums protoreflect.EnumDescriptors, visit func(protoreflect.Descriptor)) {
	for i := 0; i < enums.Len(); i++ {
		enum := enums.Get(i)
		visit(enum)
		values := enum.Values()
		for j := 0; j < values.Len(); j++ {
			visit(values.Get(j))
		}
	}
}
