// Package tracelogix provides a service for collecting and managing contextual information
// to correlate related logs throughout the lifecycle of a request or operation. This package is
// designed to be integrated as a service within your application, enabling you to log events
// with contextual data seamlessly during context passing.
//
// The primary goal is to enhance log tracing and debugging by maintaining context-specific data,
// making it easier to understand the flow and state of an application at various points.
package tracelogix

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
		// SetContextMetadata allows preallocate request metadata for TraceLog.
		// Useful to expand log information across application level.
		SetContextMetadata(reqCtx context.Context, reqLogMetadataFields MetadataFields) context.Context
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
	LogAdapterFatal LogAdapterLevel = iota
	LogAdapterError
	LogAdapterWarn
	LogAdapterInfo
	LogAdapterDebug
)

// LoggerMetadataContextKey that must be utilized for metadata sharing inside context.
const LoggerMetadataContextKey loggerContextKey = "traceLogMetadataContextKey"

// NewTraceLog constructor.
func NewTraceLog(adapter LogAdapter) *TraceLog {
	return &TraceLog{logAdapter: adapter}
}

// SetContextMetadata allows preallocate request metadata for TraceLog.
// Useful to expand log information across application level.
func (tl *TraceLog) SetContextMetadata(srcCtx context.Context, newMetadata MetadataFields) context.Context {
	existingMetadata, ok := srcCtx.Value(LoggerMetadataContextKey).(MetadataFields)
	if !ok {
		return context.WithValue(srcCtx, LoggerMetadataContextKey, newMetadata)
	}

	mergedMeta := maps.Clone(newMetadata)
	maps.Copy(mergedMeta, existingMetadata)

	for k, v := range newMetadata {
		if existingMetadata[k] != v {
			mergedMeta[k] = v
		}
	}

	return context.WithValue(srcCtx, LoggerMetadataContextKey, mergedMeta)
}

// Log information with intercepted metadata from request.
func (tl *TraceLog) Log(srcCtx context.Context, severityLvl LogAdapterLevel, logMessage string, metadata *MetadataFields) {
	if tl.logAdapter == nil {
		return
	}

	metadataToLog := make(MetadataFields)
	if metadata != nil {
		metadataToLog = *metadata
	}

	if existingMetadata, ok := srcCtx.Value(LoggerMetadataContextKey).(MetadataFields); ok {
		maps.Copy(metadataToLog, existingMetadata)
	}

	tl.logAdapter.LogAdaptersWriter(srcCtx, severityLvl, logMessage, &metadataToLog)
}
