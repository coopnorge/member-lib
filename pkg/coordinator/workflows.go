package coordinator

import "time"

// WorkflowConfig holds configuration options for a workflow.
type WorkflowConfig struct {
	MaxRetries int
	RetryDelay time.Duration
	Timeout    time.Duration
}

// WorkflowContext holds the current state and configuration of a workflow.
type WorkflowContext struct {
	Config     WorkflowConfig
	RetryCount int
}

// Workflow interface defines the methods a workflow must implement.
type Workflow interface {
	Execute(ctx *WorkflowContext) (bool, error)
	OnStart(ctx *WorkflowContext)
	OnEnd(ctx *WorkflowContext)
}

// WorkflowExecutor handles the execution of workflows.
type WorkflowExecutor struct{}

// Execute runs a workflow with the given configuration.
func (we *WorkflowExecutor) Execute(w Workflow, config WorkflowConfig) (bool, error) {
	ctx := &WorkflowContext{
		Config: config,
	}

	w.OnStart(ctx)
	defer w.OnEnd(ctx)

	for ctx.RetryCount <= ctx.Config.MaxRetries {
		success, err := w.Execute(ctx)
		if success {
			return true, nil
		}

		if ctx.RetryCount == ctx.Config.MaxRetries {
			return false, err
		}

		ctx.RetryCount++
		time.Sleep(ctx.Config.RetryDelay)
	}

	return false, nil
}
