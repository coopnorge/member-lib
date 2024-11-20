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

func (ew *ExampleWorkflow) Execute(config *WorkflowRunner) (WorkflowStatus, error) {
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

func (ew *ExampleWorkflow) OnStart(config *WorkflowRunner) {
	fmt.Printf("Starting workflow: %s\n", ew.Name)
}

func (ew *ExampleWorkflow) OnEnd(config *WorkflowRunner, status WorkflowStatus) {
	fmt.Printf("Workflow ended with status: %s\n", status)
}

func Example_implementation() {
	// Create a new ExampleWorkflow
	workflow := &ExampleWorkflow{Name: "my_new_workflow"}

	// Create a WorkflowRunner
	runner := &WorkflowRunner{
		Retry:      true,
		RetryCount: 3,
		RetryDelay: 10 * time.Millisecond,
		Timeout:    &[]time.Duration{30 * time.Millisecond}[0],
	}

	// Create a context
	ctx := context.Background()

	// Trigger the workflow.
	status, _ := runner.Trigger(ctx, workflow)

	fmt.Printf("Final status: %s\n", status)
	// Output:
	// Starting workflow: my_new_workflow
	// Workflow ended with status: NotStarted
	// Final status: TimedOut
}

func TestWorkflowExecutor(t *testing.T) {
	tests := []struct {
		name           string
		retry          bool
		retryCount     uint8
		retryDelay     time.Duration
		statusPattern  []WorkflowStatus
		expectedStatus WorkflowStatus
		expectedError  bool
		expectedCalls  int
	}{
		{
			name:           "Single attempt, fails",
			retry:          false,
			retryCount:     0,
			retryDelay:     time.Millisecond,
			statusPattern:  []WorkflowStatus{WorkflowFailed},
			expectedStatus: WorkflowFailed,
			expectedError:  true,
			expectedCalls:  1,
		},
		{
			name:           "Multiple attempts, succeeds on last try",
			retry:          true,
			retryCount:     2,
			retryDelay:     time.Millisecond,
			statusPattern:  []WorkflowStatus{WorkflowFailed, WorkflowFailed, WorkflowCompleted},
			expectedStatus: WorkflowCompleted,
			expectedError:  false,
			expectedCalls:  3,
		},
		{
			name:           "Multiple attempts, all fail",
			retry:          true,
			retryCount:     2,
			retryDelay:     time.Millisecond,
			statusPattern:  []WorkflowStatus{WorkflowFailed, WorkflowFailed, WorkflowFailed},
			expectedStatus: WorkflowFailed,
			expectedError:  true,
			expectedCalls:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &ExampleWorkflow{
				Name:          tt.name,
				StatusPattern: tt.statusPattern,
			}
			config := &WorkflowRunner{
				Retry:      tt.retry,
				RetryCount: tt.retryCount,
				RetryDelay: tt.retryDelay,
				Timeout:    nil, // Not used in this test
			}
			status, err := config.Trigger(context.Background(), workflow)
			assert.Equal(t, tt.expectedStatus, status, "Unexpected workflow status")
			if tt.expectedError {
				assert.Error(t, err, "Expected an error, but got none")
			} else {
				assert.NoError(t, err, "Expected no error, but got: %v", err)
			}
			assert.Equal(t, tt.expectedCalls, workflow.callCount, "Unexpected number of Execute calls")
		})
	}
}
