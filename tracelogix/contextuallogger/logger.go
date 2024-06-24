// Package contextuallogger provides a service for collecting and managing contextual information
// to correlate related logs throughout the lifecycle of a request or operation. This package is
// designed to be integrated as a service within your application, enabling you to log events
// with contextual data seamlessly during context passing.
//
// The primary goal is to enhance log tracing and debugging by maintaining context-specific data,
// making it easier to understand the flow and state of an application at various points.
package contextuallogger

import (
	"context"
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
	LogAdapterDebug LogAdapterLevel = iota - 1
	LogAdapterInfo
	LogAdapterWarn
	LogAdapterError
	LogAdapterFatal
)

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
