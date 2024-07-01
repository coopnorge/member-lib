package coordinator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ExampleWorkflow is a simple implementation of the Workflow interface
type ExampleWorkflow struct {
	Name          string
	StatusPattern []WorkflowStatus // Determines the status pattern of Execute calls
	callCount     int              // Tracks the number of Execute calls
}

func (ew *ExampleWorkflow) Execute(config *WorkflowConfig) (WorkflowStatus, error) {
	// Simulate work.
	time.Sleep(time.Millisecond * 10)
	if ew.callCount < len(ew.StatusPattern) {
		status := ew.StatusPattern[ew.callCount]
		ew.callCount++
		if status == WorkflowCompleted {
			return status, nil
		}
		return status, fmt.Errorf("predetermined failure occurred")
	}
	ew.callCount++
	return WorkflowFailed, fmt.Errorf("predetermined failure occurred")
}

func (ew *ExampleWorkflow) OnStart(config *WorkflowConfig) {

	fmt.Printf("Starting workflow: %s\n", ew.Name)
}

func (ew *ExampleWorkflow) OnEnd(config *WorkflowConfig, status WorkflowStatus) {
	fmt.Printf("Ending workflow: %s (Status: %s, Retries: %d)\n", ew.Name, status, config.RetryCount)
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

func TestWorkflowExecutor(t *testing.T) {
	tests := []struct {
		name           string
		retry          bool
		retryCount     uint8
		retryDelay     time.Duration
		statusPattern  []WorkflowStatus
		expectedStatus WorkflowStatus
	}{
		{
			name:           "Single attempt, fails",
			retry:          false,
			retryCount:     0,
			retryDelay:     time.Millisecond,
			statusPattern:  []WorkflowStatus{WorkflowFailed},
			expectedStatus: WorkflowFailed,
		},
		{
			name:           "Multiple attempts, succeeds on last try",
			retry:          true,
			retryCount:     2,
			retryDelay:     time.Millisecond,
			statusPattern:  []WorkflowStatus{WorkflowFailed, WorkflowFailed, WorkflowCompleted},
			expectedStatus: WorkflowCompleted,
		},
		{
			name:           "Multiple attempts, all fail",
			retry:          true,
			retryCount:     2,
			retryDelay:     time.Millisecond,
			statusPattern:  []WorkflowStatus{WorkflowFailed, WorkflowFailed, WorkflowFailed},
			expectedStatus: WorkflowFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &ExampleWorkflow{
				Name:          tt.name,
				StatusPattern: tt.statusPattern,
			}
			config := &WorkflowConfig{
				Retry:      tt.retry,
				RetryCount: tt.retryCount,
				RetryDelay: tt.retryDelay,
				Timeout:    nil, // Not used in this test
			}
			status, err := config.Execute(context.Background(), workflow)
			assert.Equal(t, tt.expectedStatus, status, "Unexpected workflow status")
			if tt.expectedStatus == WorkflowCompleted {
				assert.NoError(t, err, "Expected no error, but got: %v", err)
			} else {
				assert.Error(t, err, "Expected an error, but got none")
			}
			assert.Equal(t, len(tt.statusPattern), workflow.callCount, "Unexpected number of Execute calls")
		})
	}
}
