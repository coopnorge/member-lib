package coordinator

import (
	"context"
	"fmt"
	"time"
)

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
	status := WorkflowNotStarted
	defer w.OnEnd(config, status)

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
			return status, nil
		}
		if !config.Retry || config.RetryCount == 0 {
			return status, err
		}
		config.RetryCount--
		select {
		case <-timeoutCtx.Done():
			status = WorkflowTimedOut
			return WorkflowTimedOut, timeoutCtx.Err()
		case <-time.After(config.RetryDelay):
		}
	}

	status = WorkflowFailed
	return status, nil
}

type WorkflowStatus byte

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

// ExampleWorkflow is a simple implementation of the Workflow interface
type ExampleWorkflow struct{}

func (ew *ExampleWorkflow) Execute(config *WorkflowConfig) (WorkflowStatus, error) {
	// Simulate some work
	time.Sleep(50 * time.Millisecond)
	return WorkflowCompleted, nil
}

func (ew *ExampleWorkflow) OnStart(config *WorkflowConfig) {
	fmt.Println("Workflow started")
}

func (ew *ExampleWorkflow) OnEnd(config *WorkflowConfig, status WorkflowStatus) {
	fmt.Printf("Workflow ended with status: %s\n", status)
}

func Example_workflows() {
	// Create a new ExampleWorkflow
	workflow := &ExampleWorkflow{}

	// Create a WorkflowConfig
	config := &WorkflowConfig{
		Retry:      true,
		RetryCount: 3,
		RetryDelay: 10 * time.Millisecond,
		Timeout:    &[]time.Duration{30 * time.Millisecond}[0],
	}

	// Create a context
	ctx := context.Background()

	// Execute the workflow
	status, err := config.Execute(ctx, workflow)

	// Print the result
	fmt.Printf("Final status: %s\n", status)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Workflow started
	// Workflow ended with status: TimedOut
	// Final status: TimedOut
	// Error: context deadline exceeded
}
