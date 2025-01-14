package json

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/log/logtest"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	trace2 "go.opentelemetry.io/otel/trace"
)

func JsonEqual(t *testing.T, expected, actual string) bool {
	return assertjson.Equal(t, []byte(expected), []byte(actual))
}

func TestJsonExporter_Export(t *testing.T) {
	logger, flush := setupLogger(t, resource.Default())
	t.Run("support simple logs", func(t *testing.T) {
		logger.Info("info")
		JsonEqual(t, `
{
  "level": "INFO",
  "msg": "info",
  "time": "<ignore-diff>",
  "observed_time": "<ignore-diff>",
  "scope": "<ignore-diff>",
  "resource": "<ignore-diff>",
  "attributes": "<ignore-diff>"
}`, flush())
	})

	t.Run("supports attributes", func(t *testing.T) {
		logger.Info("info",
			"string", "value",
			"map", map[string]string{"key": "value"},
			"int", 1234,
			"float", 123.456)
		res := flush()
		JsonEqual(t, `
{
  "level": "INFO",
  "msg": "info",
  "time": "<ignore-diff>",
  "observed_time": "<ignore-diff>",
  "scope": "<ignore-diff>",
  "resource": "<ignore-diff>",
  "attributes": {
    "code.filepath": "/Users/hadrien/Projects/member-lib/telemetry/dd/json/ddjson_test.go",
    "code.function": "func2",
    "code.lineno": "<ignore-diff>",
    "code.namespace": "github.com/coopnorge/member-lib/telemetry/dd/json.TestJsonExporter_Export",
    "float": 123.456,
    "int": 1234,
    "map": {
      "key": "value"
    },
    "string": "value"
  }
}`, res)
	})

	t.Run("supports trace and span id from context", func(t *testing.T) {
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		tracer := otel.Tracer("test-tracer")
		ctx, span := tracer.Start(context.Background(), "span-name")
		logger.DebugContext(ctx, "debug")
		span.End()
		JsonEqual(t, fmt.Sprintf(`
		{
			"dd.span_id": "%d",
			"dd.trace_id": "%d",
			"time": "<ignore-diff>",
  			"observed_time": "<ignore-diff>",
  			"attributes": "<ignore-diff>",
  			"resource": "<ignore-diff>",
  			"scope": "<ignore-diff>",
  			"level": "DEBUG",
			"msg": "debug"
		}`, datadogSpanID(span.SpanContext().SpanID()), datadogTraceID(span.SpanContext().TraceID())), flush())
	})
}

func datadogSpanID(id trace2.SpanID) uint64 {
	return binary.BigEndian.Uint64(id[:])
}

func datadogTraceID(id trace2.TraceID) uint64 {
	return binary.BigEndian.Uint64(id[8:])
}

func setupLogger(t *testing.T, res *resource.Resource) (*slog.Logger, func() string) {
	var buf strings.Builder
	exporter := NewJSONExporter(&buf)
	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)),
	)
	logger := otelslog.NewLogger("test", otelslog.WithLoggerProvider(provider), otelslog.WithSource(true))
	return logger, func() string {
		require.NoError(t, exporter.ForceFlush(context.Background()))
		defer buf.Reset()
		return buf.String()
	}
}

func Test_jsonRecord(t *testing.T) {
	r := jsonRecord{
		Level:        "LEVEL",
		Msg:          "MSD",
		Time:         LogTime(time.UnixMilli(123456789)),
		ObservedTime: LogTime(time.UnixMilli(987654321)),
		DDTraceID:    "",
		DDSpanID:     "",
		Attributes: map[string]value{
			"test": newValue(log.MapValue(log.Int("key", 123)))},
	}

	var buf strings.Builder
	e := json.NewEncoder(&buf)
	e.SetIndent("", "  ")

	err := e.Encode(r)
	require.NoError(t, err)
	t.Log("json result:", buf.String())
}

func Test_convert(t *testing.T) {
	r := logtest.RecordFactory{
		Timestamp:         time.UnixMilli(12340),
		ObservedTimestamp: time.UnixMilli(43210),
		Severity:          log.SeverityError,
		SeverityText:      "Error",
		Body:              log.StringValue("message"),
		Attributes: []log.KeyValue{
			log.Int("int", 1234),
			log.Int64("int64", 4321),
			log.Float64("float64", 3.1415),
			log.Bool("bool", true),
			log.Empty("empty"),
			log.Map("map", log.Int("int", 1234)),
			log.Slice("slice", log.IntValue(1234)),
		},
	}.NewRecord()

	val := convert(&r)
	var buf strings.Builder
	e := json.NewEncoder(&buf)
	e.SetIndent("", "  ")

	err := e.Encode(val)
	require.NoError(t, err)
	JsonEqual(t, `
{
          "level": "ERROR",
          "msg": "message",
          "time": "1970-01-01T00:00:12Z",
          "observed_time": "1970-01-01T00:00:43Z",
          "scope": {
            "schema_url": "",
            "name": "",
            "version": ""
          },
          "resource": {
            "schema.url": ""
          },
          "attributes": {
            "bool": true,
            "empty": null,
            "float64": 3.1415,
            "int": 1234,
            "int64": 4321,
            "map": {
              "int": 1234
            },
            "slice": [
              1234
            ]
          }
        }`, buf.String())
}
