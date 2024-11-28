package dd

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var _ sdklog.Processor = &ddProcessor{}

type ddProcessor struct {
}

func (p *ddProcessor) Shutdown(_ context.Context) error {
	return nil
}

func (p *ddProcessor) ForceFlush(_ context.Context) error {
	return nil
}

func (p *ddProcessor) OnEmit(ctx context.Context, record *sdklog.Record) error {
	if span, ok := ddtracer.SpanFromContext(ctx); ok {
		record.AddAttributes(
			log.String("dd.span_id", strconv.FormatUint(span.Context().SpanID(), 10)),
			log.String("dd.trace_id", strconv.FormatUint(span.Context().TraceID(), 10)),
		)
	}
	return nil
}

func NewDatadogProcessor() sdklog.Processor {
	return &ddProcessor{}
}
