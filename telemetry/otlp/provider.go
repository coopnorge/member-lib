package otlp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func Exporters(ctx context.Context) (se sdktrace.SpanExporter, me sdkmetric.Exporter, le sdklog.Exporter, err error) {
	var errs error
	te, err := otlptracegrpc.New(ctx)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to create OTLP trace exporter: %w", err))
	}

	le, err = otlploggrpc.New(ctx)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to create OTLP log exporter: %w", err))
	}

	me, err = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithTemporalitySelector(datadogTemporality))
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to create OTLP metric exporter: %w", err))
	}

	return te, me, le, errs
}

func Providers(ctx context.Context, res *resource.Resource) (tp *sdktrace.TracerProvider, mp *sdkmetric.MeterProvider, lp *sdklog.LoggerProvider, err error) {
	se, me, le, err := Exporters(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	tp = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(se, sdktrace.WithBatchTimeout(2*time.Second)),
		sdktrace.WithResource(res),
	)

	mp = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(me, sdkmetric.WithInterval(1*time.Second))),
		sdkmetric.WithResource(res),
	)

	lp = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(le)),
		sdklog.WithResource(res),
	)

	return tp, mp, lp, nil
}

func datadogTemporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	switch kind {
	case sdkmetric.InstrumentKindCounter,
		sdkmetric.InstrumentKindGauge,
		sdkmetric.InstrumentKindHistogram,
		sdkmetric.InstrumentKindObservableGauge,
		sdkmetric.InstrumentKindObservableCounter:
		return metricdata.DeltaTemporality
	case sdkmetric.InstrumentKindUpDownCounter,
		sdkmetric.InstrumentKindObservableUpDownCounter:
		return metricdata.CumulativeTemporality
	}
	panic("unknown instrument kind")
}
