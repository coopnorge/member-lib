package dd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	ljson "github.com/coopnorge/member-lib/telemetry/dd/json"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func Providers(ctx context.Context, res *resource.Resource, traceURL, metricURL string) (tp *sdktrace.TracerProvider, mp *sdkmetric.MeterProvider, lp *sdklog.LoggerProvider, err error) {
	te, me, le, err := Exporters(ctx, res, traceURL, metricURL)
	if err != nil {
		return nil, nil, nil, err
	}

	tp = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(te, sdktrace.WithBatchTimeout(2*time.Second)),
		sdktrace.WithResource(res),
	)

	mp = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(me, sdkmetric.WithInterval(1*time.Second))),
		sdkmetric.WithResource(res),
	)

	lp = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(NewDatadogProcessor()),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(le)),
		sdklog.WithResource(res),
	)

	return tp, mp, lp, nil
}

// Exporters returns exporters that rely on the datadog libraries to report traces, metrics and logs.
// An error will be reported if the resource lacks a service name, service version or a deployment environment.
func Exporters(_ context.Context, res *resource.Resource, traceURL, metricURL string) (te sdktrace.SpanExporter, me sdkmetric.Exporter, le sdklog.Exporter, err error) {
	// Env sanity check.
	if err := checkEnv(); err != nil {
		return nil, nil, nil, err
	}

	svc, ver, env, err := requiredAttr(res)
	if err != nil {
		return nil, nil, nil, err
	}

	var traceURLOption tracer.StartOption
	if strings.HasPrefix(traceURL, "/") {
		traceURLOption = tracer.WithUDS(traceURL)
	} else {
		traceURLOption = tracer.WithAgentAddr(traceURL)
	}

	te = NewDatadogSpanExporter(
		tracer.WithService(svc),
		tracer.WithServiceVersion(ver),
		tracer.WithEnv(env),
		traceURLOption,
	)

	ddClient, err := statsd.New(metricURL, statsd.WithTags([]string{
		fmt.Sprintf("env:%s", env),
		fmt.Sprintf("service:%s", svc),
		fmt.Sprintf("version:%s", ver),
	}))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to start statsdogd client: %w", err)
	}
	me, err = NewDatadogMetricExporter(ddClient)
	if err != nil {
		return nil, nil, nil, err
	}

	le = ljson.NewJsonExporter(os.Stderr)
	return te, me, le, nil
}

func requiredAttr(res *resource.Resource) (service, version, environment string, err error) {
	set := res.Set()
	if val, ok := set.Value(semconv.ServiceNameKey); ok {
		service = val.AsString()
	} else {
		err = errors.Join(err, errors.New("missing 'service.name'"))
	}
	if val, ok := set.Value(semconv.ServiceVersionKey); ok {
		version = val.AsString()
	} else {
		err = errors.Join(err, errors.New("missing 'service.version'"))
	}
	if val, ok := set.Value(semconv.DeploymentEnvironmentKey); ok {
		environment = val.AsString()
	} else {
		err = errors.Join(err, errors.New("missing 'deployment.environment'"))
	}
	if err != nil {
		err = fmt.Errorf("missing attribute(s): %w", err)
	}
	return
}

var envMapping = map[string]string{
	"OTEL_SERVICE_NAME":        "DD_SERVICE",
	"OTEL_LOG_LEVEL":           "DD_LOG_LEVEL",
	"OTEL_PROPAGATORS":         "DD_TRACE_PROPAGATION_STYLE",
	"OTEL_TRACES_SAMPLER":      "DD_TRACE_SAMPLE_RATE",
	"OTEL_TRACES_EXPORTER":     "DD_TRACE_ENABLED",
	"OTEL_METRICS_EXPORTER":    "DD_RUNTIME_METRICS_ENABLED",
	"OTEL_RESOURCE_ATTRIBUTES": "DD_TAGS",
	"OTEL_SDK_DISABLED":        "DD_TRACE_OTEL_ENABLED",
}

func checkEnv() error {
	var errs error
	for otelKey, ddKey := range envMapping {
		_, hasOtel := os.LookupEnv(otelKey)
		_, hasDd := os.LookupEnv(ddKey)
		if hasOtel && hasDd {
			errs = errors.Join(errs, fmt.Errorf("cannot use both '%s' and '%s'", otelKey, ddKey))
		}
	}
	return errs
}
