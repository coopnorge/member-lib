package coordinator

import (
	"context"
	"time"
)

type WorkflowStatus int

const (
	WorkflowNotStarted WorkflowStatus = iota
	WorkflowRunning
	WorkflowCompleted
	WorkflowFailed
	WorkflowCancelled
	WorkflowTimedOut
)

func (s WorkflowStatus) String() string {
	switch s {
	case WorkflowNotStarted:
		return "NotStarted"
	case WorkflowRunning:
		return "Running"
	case WorkflowCompleted:
		return "Completed"
	case WorkflowFailed:
		return "Failed"
	case WorkflowCancelled:
		return "Cancelled"
	case WorkflowTimedOut:
		return "TimedOut"
	default:
		return "Unknown"
	}
}

// WorkflowConfig holds configuration options for a workflow.
type WorkflowConfig struct {
	Retry      bool
	RetryCount uint8
	RetryDelay time.Duration
	Timeout    *time.Duration
}

// Workflow interface defines the methods a workflow must implement.
type Workflow interface {
	Execute(config *WorkflowConfig) (WorkflowStatus, error)
	OnStart(config *WorkflowConfig)
	OnEnd(config *WorkflowConfig, status WorkflowStatus)
}

// Execute runs a workflow with the given configuration and context.
func (config *WorkflowConfig) Execute(ctx context.Context, w Workflow) (WorkflowStatus, error) {
	w.OnStart(config)
	defer w.OnEnd(config, WorkflowNotStarted)

	var timeoutCtx context.Context
	var cancel context.CancelFunc
	if config.Timeout != nil {
		timeoutCtx, cancel = context.WithTimeout(ctx, *config.Timeout)
		defer cancel()
	} else {
		timeoutCtx = ctx
	}

	for config.RetryCount < 1 || (config.Retry && config.RetryCount > 0) {
		status, err := w.Execute(config)
		if status == WorkflowCompleted {
			w.OnEnd(config, status)
			return status, nil
		}
		if !config.Retry || config.RetryCount == 0 {
			w.OnEnd(config, status)
			return status, err
		}
		config.RetryCount--
		select {
		case <-timeoutCtx.Done():
			w.OnEnd(config, WorkflowTimedOut)
			return WorkflowTimedOut, timeoutCtx.Err()
		case <-time.After(config.RetryDelay):
		}
	}

	w.OnEnd(config, WorkflowFailed)
	return WorkflowFailed, nil
}
