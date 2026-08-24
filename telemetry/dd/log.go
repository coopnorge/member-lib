//nolint:revive // TODO: add documentation
package dd

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
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

//nolint:gocritic // hugeParam: method defined by external interface
func (p *ddProcessor) Enabled(ctx context.Context, _ sdklog.EnabledParameters) bool { return true }

func (p *ddProcessor) OnEmit(ctx context.Context, record *sdklog.Record) error {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		osCtx := ctxWrapper{span.SpanContext()}
		record.AddAttributes(
			attribute.String("dd.span_id", strconv.FormatUint(osCtx.SpanID(), 10)),
			attribute.String("dd.trace_id", strconv.FormatUint(osCtx.TraceID(), 10)),
		)
	}
	return nil
}

//nolint:revive // TODO: add documentation
func NewDatadogProcessor() sdklog.Processor {
	return &ddProcessor{}
}
