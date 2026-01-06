package dd

import (
	"context"
	"encoding/json"
	"errors"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var _ sdktrace.SpanExporter = &ddTraceExporter{}

type ddTraceExporter struct {
}

//nolint:revive // TODO: add documentation
func NewDatadogSpanExporter(opts ...tracer.StartOption) sdktrace.SpanExporter {
	tracer.Start(opts...)
	return &ddTraceExporter{}
}

func (e *ddTraceExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		startOpts := []ddtrace.StartSpanOption{
			tracer.StartTime(span.StartTime()),
			tracer.ResourceName(span.Name()),
			tracer.WithSpanID(ctxWrapper{span.SpanContext()}.SpanID()),
		}
		startOpts = append(startOpts, tracer.ChildOf(getParent(span)))

		ddSpan := tracer.StartSpan(
			span.InstrumentationScope().Name,
			startOpts...,
		)

		ddSpan.SetTag(ext.SpanKind, span.SpanKind().String())

		// Set resource as Datadog tags
		for _, attr := range span.Resource().Attributes() {
			ddSpan.SetTag(string(attr.Key), attr.Value.AsString())
		}

		// Set attributes as Datadog tags
		for _, attr := range span.Attributes() {
			ddSpan.SetTag(string(attr.Key), attr.Value.AsString())
		}

		finishOpts := []ddtrace.FinishOption{
			tracer.FinishTime(span.EndTime()),
		}

		if b, ok := marshalEvents(span); ok {
			ddSpan.SetTag("events", b)
		}

		// Check if there's only one error in the span events, in this case we can
		// add the error itself.
		if status := span.Status(); status.Code == codes.Error {
			eevents := make([]sdktrace.Event, 0)
			for _, event := range span.Events() {
				if event.Name == semconv.ExceptionEventName {
					eevents = append(eevents, event)
				}
			}
			if len(eevents) == 1 {
				for _, attr := range eevents[0].Attributes {
					switch attr.Key {
					case semconv.ExceptionTypeKey:
						ddSpan.SetTag(ext.ErrorType, attr.Value.AsString())
					case semconv.ExceptionMessageKey:

						ddSpan.SetTag(ext.ErrorDetails, attr.Value.AsString())
					case semconv.ExceptionStacktraceKey:
						ddSpan.SetTag(ext.ErrorStack, attr.Value.AsString())
					}
				}
			}
			ddSpan.SetTag(ext.ErrorMsg, status.Description)
			finishOpts = append(finishOpts,
				tracer.WithError(
					errors.New(status.Description),
				),
				tracer.NoDebugStack(),
			)
		}
		ddSpan.Finish(finishOpts...)
	}
	return nil
}

func marshalEvents(span sdktrace.ReadOnlySpan) (string, bool) {
	if len(span.Events()) == 0 {
		return "", false
	}
	events := make([]spanEvent, 0, len(span.Events()))
	for _, event := range span.Events() {
		attrs := make(map[string]interface{})
		for _, a := range event.Attributes {
			attrs[string(a.Key)] = a.Value.AsInterface()
		}
		events = append(events, spanEvent{
			Name:         event.Name,
			TimeUnixNano: event.Time.UnixNano(),
			Attributes:   attrs,
		})
	}
	b, err := json.Marshal(events)
	return string(b), err == nil
}

func (e *ddTraceExporter) Shutdown(_ context.Context) error {
	tracer.Stop()
	return nil
}
