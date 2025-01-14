package json

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"strconv"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

var _ sdklog.Exporter = &jsonExporter{}

type jsonExporter struct {
	flush   func() error
	encoder json.Encoder
}

func NewJSONExporter(w io.Writer) sdklog.Exporter {
	b := bufio.NewWriter(w)
	return &jsonExporter{
		flush:   b.Flush,
		encoder: *json.NewEncoder(b),
	}
}

// Export implements log.Exporter.
func (j *jsonExporter) Export(_ context.Context, records []sdklog.Record) error {
	for i := 0; i < len(records); i++ {
		err := j.encoder.Encode(convert(&records[i]))
		if err != nil {
			return err
		}
	}
	return j.flush()
}

func convert(record *sdklog.Record) jsonRecord {
	entry := jsonRecord{
		Msg:          record.Body().AsString(),
		Level:        record.Severity().String(),
		Time:         LogTime(record.Timestamp()),
		ObservedTime: LogTime(record.ObservedTimestamp()),
		Scope: jsonScope{
			SchemaURL:  record.InstrumentationScope().SchemaURL,
			Name:       record.InstrumentationScope().Name,
			Version:    record.InstrumentationScope().Version,
			Attributes: attrSetToMap(record.InstrumentationScope().Attributes),
		},
	}

	res := record.Resource()
	if &res != resource.Empty() {
		entry.Resource = &jsonResource{
			SchemaURL: res.SchemaURL(),
		}
		entry.Resource.Attributes = attrSliceToMap(res.Attributes())
	}

	if record.TraceID().IsValid() {
		id := record.TraceID()
		entry.DDTraceID = strconv.FormatUint(binary.BigEndian.Uint64(id[8:]), 10)
	}

	if record.SpanID().IsValid() {
		id := record.SpanID()
		entry.DDSpanID = strconv.FormatUint(binary.BigEndian.Uint64(id[:]), 10)
	}

	if record.AttributesLen() > 0 {
		attrs := make(map[string]value, record.AttributesLen())
		record.WalkAttributes(func(kv log.KeyValue) bool {
			attrs[kv.Key] = newValue(kv.Value)
			return true
		})
		entry.Attributes = attrs
	}
	return entry
}

type jsonScope struct {
	SchemaURL  string         `json:"schema_url"`
	Name       string         `json:"name"`
	Version    string         `json:"version"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type jsonRecord struct {
	Level        string  `json:"level"`
	Msg          string  `json:"msg"`
	Time         LogTime `json:"time,omitempty"`
	ObservedTime LogTime `json:"observed_time,omitempty"`

	Scope    jsonScope     `json:"scope,omitempty"`
	Resource *jsonResource `json:"resource,omitempty"`

	DDTraceID string `json:"dd.trace_id,omitempty"`
	DDSpanID  string `json:"dd.span_id,omitempty"`

	Attributes map[string]value `json:"attributes,omitempty"`
}

type jsonResource struct {
	SchemaURL  string         `json:"schema.url"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// ForceFlush implements log.Exporter.
func (j *jsonExporter) ForceFlush(_ context.Context) error {
	return j.flush()
}

// Shutdown implements log.Exporter.
func (j *jsonExporter) Shutdown(_ context.Context) error {
	j.encoder = *json.NewEncoder(io.Discard)
	return j.flush()
}
