package contextuallogger

import (
	"context"
	"errors"
	"maps"
)

type (
	// LogAdapterLevel verbosity level type.
	LogAdapterLevel int8
	// LogAdapter that handles writing logs.
	LogAdapter interface {
		// LogAdaptersWriter for log events.
		LogAdaptersWriter(ctx context.Context, lvl LogAdapterLevel, msg string, meta *MetadataFields)
	}

	// TraceLogService interface.
	TraceLogService interface {
		// SetRequestMetadataContext allows preallocate request metadata for TraceLog.
		// Useful to expand log information across application level.
		SetRequestMetadataContext(reqCtx context.Context, reqLogMetadataFields MetadataFields) context.Context
		// Log information with intercepted metadata from request.
		Log(sourceCtx context.Context, logSeverity LogAdapterLevel, logMessage string, logMeta *MetadataFields)
	}

	// TraceLog is a structured logger designed for tracing app-level events
	// and activities, in an informative and standardized way.
	//
	// The primary use of TraceLog is
	// to capture and present logs in a manner that aids debugging, performance monitoring.
	TraceLog struct {
		logAdapter LogAdapter
	}
	loggerContextKey string

	// MetadataFields that must be logged with message.
	MetadataFields map[string]any
)

const (
	LogAdapterFatal LogAdapterLevel = iota // 0
	LogAdapterError                        // 1
	LogAdapterWarn                         // 2
	LogAdapterInfo                         // 3
	LogAdapterDebug                        // 4
)

// Map of log level names to LogAdapterLevel values
var logLevelMap = map[string]LogAdapterLevel{
	"fatal": LogAdapterFatal,
	"error": LogAdapterError,
	"warn":  LogAdapterWarn,
	"info":  LogAdapterInfo,
	"debug": LogAdapterDebug,
}

// Function to convert a string log level to a LogAdapterLevel
func getLogAdapterLevel(level string) (LogAdapterLevel, error) {
	if logLevel, exists := logLevelMap[level]; exists {
		return logLevel, nil
	}
	return -1, errors.New("invalid log level")
}

// LoggerMetadataContextKey that must be utilized for metadata sharing inside context.
const LoggerMetadataContextKey loggerContextKey = "traceLogMetadataContextKey"

// NewTraceLog constructor.
func NewTraceLog(adapter LogAdapter) *TraceLog {
	return &TraceLog{logAdapter: adapter}
}

// SetRequestMetadataContext allows preallocate request metadata for TraceLog.
// Useful to expand log information across application level.
func (tl *TraceLog) SetRequestMetadataContext(reqCtx context.Context, reqLogMetadataFields MetadataFields) context.Context {
	metadata, isFoundMetadata := reqCtx.Value(LoggerMetadataContextKey).(MetadataFields)
	if !isFoundMetadata {
		return context.WithValue(reqCtx, LoggerMetadataContextKey, reqLogMetadataFields)
	}

	mergedMeta := maps.Clone(reqLogMetadataFields)
	maps.Copy(mergedMeta, metadata)

	for k, v := range reqLogMetadataFields {
		if metadata[k] != v {
			mergedMeta[k] = v
		}
	}

	return context.WithValue(reqCtx, LoggerMetadataContextKey, mergedMeta)
}

// Log information with intercepted metadata from request.
func (tl *TraceLog) Log(sourceCtx context.Context, logSeverity LogAdapterLevel, logMessage string, logMeta *MetadataFields) {
	if tl.logAdapter == nil {
		return
	}

	logFieldsToLog := make(MetadataFields)
	if logMeta != nil {
		logFieldsToLog = *logMeta
	}

	if metadata, ok := sourceCtx.Value(LoggerMetadataContextKey).(MetadataFields); ok {
		maps.Copy(logFieldsToLog, metadata)
	}

	tl.logAdapter.LogAdaptersWriter(sourceCtx, logSeverity, logMessage, &logFieldsToLog)
}
