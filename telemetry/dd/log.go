//nolint:revive // TODO: add documentation
package dd

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
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
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		osCtx := ctxWrapper{span.SpanContext()}
		record.AddAttributes(
			log.String("dd.span_id", strconv.FormatUint(osCtx.SpanID(), 10)),
			log.String("dd.trace_id", strconv.FormatUint(osCtx.TraceID(), 10)),
		)
	}
	return nil
}

//nolint:revive // TODO: add documentation
func NewDatadogProcessor() sdklog.Processor {
	return &ddProcessor{}
}
