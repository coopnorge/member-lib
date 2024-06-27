package coordinator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockWorkflow simulates a workflow with deterministic behavior.
type MockWorkflow struct {
	Name          string
	StatusPattern []WorkflowStatus // Determines the status pattern of Execute calls
	callCount     int              // Tracks the number of Execute calls
}

func (mw *MockWorkflow) Execute(config *WorkflowConfig) (WorkflowStatus, error) {
	// Simulate work.
	time.Sleep(time.Millisecond * 10)
	if mw.callCount < len(mw.StatusPattern) {
		status := mw.StatusPattern[mw.callCount]
		mw.callCount++
		if status == WorkflowCompleted {
			return status, nil
		}
		return status, fmt.Errorf("predetermined failure occurred")
	}
	mw.callCount++
	return WorkflowFailed, fmt.Errorf("predetermined failure occurred")
}

func (mw *MockWorkflow) OnStart(config *WorkflowConfig) {
	fmt.Printf("Starting workflow: %s\n", mw.Name)
}

func (mw *MockWorkflow) OnEnd(config *WorkflowConfig, status WorkflowStatus) {
	fmt.Printf("Ending workflow: %s (Status: %s, Retries: %d)\n", mw.Name, status, config.RetryCount)
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
			workflow := &MockWorkflow{
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
