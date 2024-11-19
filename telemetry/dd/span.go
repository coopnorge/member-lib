package dd

import (
	"encoding/binary"
	"go.opentelemetry.io/otel/sdk/trace"
	trace2 "go.opentelemetry.io/otel/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
)

func getParent(span trace.ReadOnlySpan) ddtrace.SpanContextW3C {
	if parent := span.Parent(); parent.IsValid() {
		return ctxWrapper{otelCtx: parent}
	}
	// Passing a "fake" child relationship so that the trace id generation
	// doesn't kick in. If not, the datadog lib
	return noParent{ctxWrapper{span.SpanContext()}}
}

var _ ddtrace.SpanContextW3C = &ctxWrapper{}

// ctxWrapper converts open telemetry span context to ddtrace.SpanContextW3C
type ctxWrapper struct {
	otelCtx trace2.SpanContext
}

func (p ctxWrapper) TraceID128() string {
	return p.otelCtx.TraceID().String()
}

func (p ctxWrapper) TraceID128Bytes() [16]byte {
	return p.otelCtx.TraceID()
}

func (p ctxWrapper) SpanID() uint64 {
	id := p.otelCtx.SpanID()
	return binary.BigEndian.Uint64(id[:])
}

func (p ctxWrapper) TraceID() uint64 {
	id := p.otelCtx.TraceID()
	return binary.BigEndian.Uint64(id[8:])
}

func (p ctxWrapper) ForeachBaggageItem(handler func(k string, v string) bool) {
	p.otelCtx.TraceState().Walk(handler)
}

var _ ddtrace.SpanContextW3C = &noParent{}

// noParent hides the SpanID from parent span context.
type noParent struct {
	ddtrace.SpanContextW3C
}

func (n noParent) SpanID() uint64 {
	return 0
}

type spanEvent struct {
	Name         string                 `json:"name"`
	TimeUnixNano int64                  `json:"time_unix_nano"`
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
}
