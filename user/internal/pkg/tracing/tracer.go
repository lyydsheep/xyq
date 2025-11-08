package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TraceInfo 包含追踪信息
type TraceInfo struct {
	TraceID string
	SpanID  string
}

// InitTracer initializes the global tracer
func InitTracer(serviceName string) trace.Tracer {
	return otel.Tracer(serviceName)
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("").Start(ctx, name, opts...)
}

// AddSpanTags adds tags to the current span
func AddSpanTags(ctx context.Context, tags map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	for key, value := range tags {
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
	}
}

func AddSpanEvent(ctx context.Context, eventName string, attributes map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for key, value := range attributes {
		attrs = append(attrs, attribute.String(key, fmt.Sprintf("%v", value)))
	}

	span.AddEvent(eventName, trace.WithAttributes(attrs...))
}

// ExtractTraceInfo 从上下文中提取追踪信息
func ExtractTraceInfo(ctx context.Context) TraceInfo {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsValid() {
		return TraceInfo{
			TraceID: "",
			SpanID:  "",
		}
	}

	sc := span.SpanContext()
	return TraceInfo{
		TraceID: sc.TraceID().String(),
		SpanID:  sc.SpanID().String(),
	}
}
