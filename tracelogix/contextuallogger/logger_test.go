package contextuallogger

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name string
		lvl  LogAdapterLevel
	}{
		{name: "Debug log", lvl: LogAdapterDebug},
		{name: "Info log", lvl: LogAdapterInfo},
		{name: "Warn log", lvl: LogAdapterWarn},
		{name: "Error log", lvl: LogAdapterError},
		{name: "Fatal log", lvl: LogAdapterFatal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logWriterAdapter := new(stubLoggerAdapter)

			tl := NewTraceLog(logWriterAdapter)

			originalCtx := context.Background()
			meta := MetadataFields{"key1": "value1"}

			tl.Log(originalCtx, tt.lvl, fmt.Sprintf("Test Log for %s", tt.name), &meta)
			assert.True(t, logWriterAdapter.logger.isLogged, fmt.Sprintf("expected that logger is logged message under level %v", tt.lvl))
		})
	}

}

func TestSetRequestMetadataContext(t *testing.T) {
	tl := NewTraceLog(new(stubLoggerAdapter))
	originalCtx := context.Background()

	// Validate initial context extend and it's metadata
	initialFields := MetadataFields{"key1": "value1", "key2": "value2"}
	ctxWithMetadata := tl.SetRequestMetadataContext(originalCtx, initialFields)

	metadata, isMetadataFound := ctxWithMetadata.Value(LoggerMetadataContextKey).(MetadataFields)
	assert.True(t, isMetadataFound)
	assert.True(t, metadata["key1"] == "value1")
	assert.True(t, metadata["key2"] == "value2")

	metaFieldForLog := MetadataFields{"key1": "valueModified", "logAdditionalKey": "logValue3"}
	tl.Log(ctxWithMetadata, LogAdapterInfo, "Test Log", &metaFieldForLog)

	metadata, isMetadataFound = ctxWithMetadata.Value(LoggerMetadataContextKey).(MetadataFields)
	assert.True(t, isMetadataFound)
	assert.True(t, metadata["key1"] == "value1")
	assert.True(t, metadata["key2"] == "value2")

	// Validate if context metadata must be mutated
	extendContextLogFields := MetadataFields{"key1": "mutatedValue1", "key3": "value3"}
	updatedCtxWithMetadata := tl.SetRequestMetadataContext(ctxWithMetadata, extendContextLogFields)

	updMetadata, isUPDMetadataFound := updatedCtxWithMetadata.Value(LoggerMetadataContextKey).(MetadataFields)
	assert.True(t, isUPDMetadataFound)
	assert.True(t, updMetadata["key1"] == "mutatedValue1")
	assert.True(t, updMetadata["key2"] == "value2")
	assert.True(t, updMetadata["key3"] == "value3")
}

type stubLoggerAdapter struct {
	logger stubLogger
}

func (sla *stubLoggerAdapter) LogAdaptersWriter(_ context.Context, lvl LogAdapterLevel, msg string, _ *MetadataFields) {
	switch lvl {
	case LogAdapterFatal:
		sla.logger.Fatal(msg)
	case LogAdapterError:
		sla.logger.Error(msg)
	case LogAdapterWarn:
		sla.logger.Warn(msg)
	case LogAdapterInfo:
		sla.logger.Info(msg)
	default:
		sla.logger.Debug(msg)
	}
}

type stubLogger struct {
	isLogged bool
	t        *testing.T
}

func (sl *stubLogger) Debug(args ...any) {
	sl.isLogged = true
	assert.NotEmpty(sl.t, args)
}

func (sl *stubLogger) Info(args ...any) {
	sl.isLogged = true
	assert.NotEmpty(sl.t, args)
}

func (sl *stubLogger) Warn(args ...any) {
	sl.isLogged = true
	assert.NotEmpty(sl.t, args)
}

func (sl *stubLogger) Error(args ...any) {
	sl.isLogged = true
	assert.NotEmpty(sl.t, args)
}

func (sl *stubLogger) Fatal(args ...any) {
	sl.isLogged = true
	assert.NotEmpty(sl.t, args)
}
