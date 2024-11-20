package dd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/iancoleman/strcase"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var _ sdkmetric.Exporter = *(*ddMetricExporter)(nil)

type ddMetricExporter struct {
	client *statsd.Client
}

// NewDatadogMetricExporter creates a metric.Exporter that forwards the instrument
// values to the Datadog statsd.
func NewDatadogMetricExporter(sc *statsd.Client) (sdkmetric.Exporter, error) {
	if sc == nil {
		return nil, errors.New("statsd.Client is nil")
	}
	return &ddMetricExporter{
		client: sc,
	}, nil
}

// Temporality returns metricdata.Temporality that work with the statsd client.
func (d ddMetricExporter) Temporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
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
	default:
		// Should not happen, but failing here is unnecessary.
		return metricdata.CumulativeTemporality
	}
}

// Aggregation returns metric.Aggregation that work with the statsd client.
func (d ddMetricExporter) Aggregation(ik sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	switch ik {
	case sdkmetric.InstrumentKindCounter, sdkmetric.InstrumentKindUpDownCounter, sdkmetric.InstrumentKindObservableCounter,
		sdkmetric.InstrumentKindObservableUpDownCounter:
		return sdkmetric.AggregationSum{}
	case sdkmetric.InstrumentKindObservableGauge, sdkmetric.InstrumentKindGauge:
		return sdkmetric.AggregationLastValue{}
	case sdkmetric.InstrumentKindHistogram:
		return sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
			NoMinMax:   false,
		}
	default:
		return sdkmetric.AggregationDefault{}
	}
}

func (d ddMetricExporter) Export(_ context.Context, metrics *metricdata.ResourceMetrics) (err error) {
	res := normalizeAttributes(metrics.Resource.Attributes())
	for _, scopeMetrics := range metrics.ScopeMetrics {
		// TODO(hadrienk): Add scope scopeMetrics.Scope
		for _, m := range scopeMetrics.Metrics {
			name := m.Name
			switch dp := m.Data.(type) {
			case metricdata.Gauge[int64]:
				err = d.exportGaugeInt(dp.DataPoints, name, res)
			case metricdata.Gauge[float64]:
				err = d.exportGaugeFloat(dp.DataPoints, name, res)
			case metricdata.Sum[int64]:
				err = d.exportSumInt(dp.DataPoints, name, res)
			case metricdata.Sum[float64]:
				err = d.exportSumFloat(dp.DataPoints, name, res)
			case metricdata.Histogram[int64]:
				err = d.exportHistogramInt(dp.DataPoints, name, res)
			case metricdata.Histogram[float64]:
				err = d.exportHistogramFloat(dp.DataPoints, name, res)
			default:
				err = errors.Join(err, errors.New("unknown metric type"))
			}
		}
	}
	return err
}

func (d ddMetricExporter) exportHistogramFloat(dp []metricdata.HistogramDataPoint[float64], name string, res []attribute.KeyValue) (err error) {
	for _, point := range dp {
		err = errors.Join(err, d.histogramFloat64(name, point, res))
	}
	return err
}

func (d ddMetricExporter) exportHistogramInt(dp []metricdata.HistogramDataPoint[int64], name string, res []attribute.KeyValue) (err error) {
	for _, point := range dp {
		err = errors.Join(err, d.histogramInt64(name, point, res))
	}
	return err
}

func (d ddMetricExporter) exportSumFloat(dp []metricdata.DataPoint[float64], name string, res []attribute.KeyValue) (err error) {
	for _, point := range dp {
		err = errors.Join(err, d.countFloat64(name, point, res))
	}
	return err
}

func (d ddMetricExporter) exportSumInt(dp []metricdata.DataPoint[int64], name string, res []attribute.KeyValue) (err error) {
	for _, point := range dp {
		err = errors.Join(err, d.countInt64(name, point, res))
	}
	return err
}

func (d ddMetricExporter) exportGaugeFloat(dp []metricdata.DataPoint[float64], name string, res []attribute.KeyValue) (err error) {
	for _, point := range dp {
		err = errors.Join(err, d.gaugeFloat64(name, point, res))
	}
	return err
}

func (d ddMetricExporter) exportGaugeInt(dp []metricdata.DataPoint[int64], name string, res []attribute.KeyValue) (err error) {
	for _, point := range dp {
		err = errors.Join(err, d.gaugeInt64(name, point, res))
	}
	return err
}

func (d ddMetricExporter) countInt64(name string, p metricdata.DataPoint[int64], res []attribute.KeyValue) error {
	return d.client.Count(name, p.Value, toTags(res, p.Attributes.ToSlice()), 1.0)
}

func (d ddMetricExporter) countFloat64(name string, p metricdata.DataPoint[float64], res []attribute.KeyValue) error {
	return d.client.Count(name, int64(p.Value), toTags(res, p.Attributes.ToSlice()), 1.0)
}

func (d ddMetricExporter) gaugeFloat64(name string, p metricdata.DataPoint[float64], res []attribute.KeyValue) error {
	return d.client.Gauge(name, p.Value, toTags(res, p.Attributes.ToSlice()), 1.0)
}

func (d ddMetricExporter) gaugeInt64(name string, p metricdata.DataPoint[int64], res []attribute.KeyValue) error {
	return d.client.Gauge(name, float64(p.Value), toTags(res, p.Attributes.ToSlice()), 1.0)
}

func normalizeAttributes(attributes []attribute.KeyValue) (normalized []attribute.KeyValue) {
	for _, kv := range attributes {
		normalized = append(normalized, normalizeAttribute(kv))
	}
	return normalized
}

func normalizeAttribute(kv attribute.KeyValue) attribute.KeyValue {
	switch kv.Key {
	case semconv.ServiceNameKey:
		return attribute.String("service", kv.Value.AsString())
	case semconv.ServiceVersionKey:
		return attribute.String("version", kv.Value.AsString())
	case semconv.DeploymentEnvironmentKey:
		return attribute.String("env", kv.Value.AsString())
	}
	return kv
}

func toTags(attributes ...[]attribute.KeyValue) (tags []string) {
	for _, attrs := range attributes {
		for _, attr := range attrs {
			tags = append(tags, toTag(attr))
		}
	}
	return tags
}

func toTag(attr attribute.KeyValue) string {
	tag := fmt.Sprintf(
		"%s:%s",
		strings.ToLower(strcase.ToSnake(string(attr.Key))),
		attr.Value.AsString(),
	)
	return tag
}

// nolint:gosec // datadog only supports int64.
func safeUintToInt(val uint64) int64 {
	return int64(val)
}

func (d ddMetricExporter) histogramFloat64(name string, p metricdata.HistogramDataPoint[float64], tags []attribute.KeyValue) (err error) {
	err = errors.Join(err, d.client.Count(fmt.Sprintf("%s.sum", name), int64(p.Sum), toTags(tags, p.Attributes.ToSlice()), 1.0))
	err = errors.Join(err, d.client.Count(fmt.Sprintf("%s.count", name), safeUintToInt(p.Count), toTags(tags, p.Attributes.ToSlice()), 1.0))
	if v, ok := p.Min.Value(); ok {
		err = errors.Join(err, d.client.Gauge(fmt.Sprintf("%s.min", name), v, toTags(tags, p.Attributes.ToSlice()), 1.0))
	}
	if v, ok := p.Max.Value(); ok {
		err = errors.Join(err, d.client.Gauge(fmt.Sprintf("%s.max", name), v, toTags(tags, p.Attributes.ToSlice()), 1.0))
	}
	for i := 0; i < len(p.Bounds); i += 2 {
		var lower, upper float64
		if i+1 < len(p.Bounds) {
			lower, upper = p.Bounds[i], p.Bounds[i+1]
		} else {
			lower, upper = p.Bounds[i], math.Inf(1)
		}
		bounds := []attribute.KeyValue{
			attribute.String("lower_bound", fmt.Sprintf("%f", lower)),
			attribute.String("upper_bound", fmt.Sprintf("%f", upper)),
		}
		err = errors.Join(err, d.client.Count(fmt.Sprintf("%s.bucket", name), safeUintToInt(p.BucketCounts[i]), toTags(tags, p.Attributes.ToSlice(), bounds), 1.0))
	}
	return err
}

func (d ddMetricExporter) histogramInt64(name string, p metricdata.HistogramDataPoint[int64], tags []attribute.KeyValue) (err error) {
	err = errors.Join(err, d.client.Count(fmt.Sprintf("%s.sum", name), p.Sum, toTags(tags, p.Attributes.ToSlice()), 1.0))
	err = errors.Join(err, d.client.Count(fmt.Sprintf("%s.count", name), safeUintToInt(p.Count), toTags(tags, p.Attributes.ToSlice()), 1.0))
	if v, ok := p.Min.Value(); ok {
		err = errors.Join(err, d.client.Gauge(fmt.Sprintf("%s.min", name), float64(v), toTags(tags, p.Attributes.ToSlice()), 1.0))
	}
	if v, ok := p.Max.Value(); ok {
		err = errors.Join(err, d.client.Gauge(fmt.Sprintf("%s.max", name), float64(v), toTags(tags, p.Attributes.ToSlice()), 1.0))
	}
	for i := 0; i < len(p.Bounds); i += 2 {
		var lower, upper float64
		if i+1 < len(p.Bounds) {
			lower, upper = p.Bounds[i], p.Bounds[i+1]
		} else {
			lower, upper = p.Bounds[i], math.Inf(1)
		}
		bounds := []attribute.KeyValue{
			attribute.String("lower_bound", fmt.Sprintf("%f", lower)),
			attribute.String("upper_bound", fmt.Sprintf("%f", upper)),
		}
		err = errors.Join(err, d.client.Count(fmt.Sprintf("%s.bucket", name), int64(p.BucketCounts[i]), toTags(tags, p.Attributes.ToSlice(), bounds), 1.0))
	}
	return err
}

func (d ddMetricExporter) ForceFlush(_ context.Context) error {
	return d.client.Flush()
}

func (d ddMetricExporter) Shutdown(_ context.Context) error {
	return d.client.Close()
}
