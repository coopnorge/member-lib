package coordinator

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockWorkflow simulates a workflow with deterministic behavior
type MockWorkflow struct {
	Name           string
	SuccessPattern []bool // Determines the success/failure pattern of Execute calls
	callCount      int    // Tracks the number of Execute calls
}

func (mw *MockWorkflow) Execute(ctx *WorkflowContext) (bool, error) {
	// Simulate work
	time.Sleep(time.Millisecond * 10)

	if mw.callCount < len(mw.SuccessPattern) && mw.SuccessPattern[mw.callCount] {
		mw.callCount++
		return true, nil
	}

	mw.callCount++
	return false, fmt.Errorf("predetermined failure occurred")
}

func (mw *MockWorkflow) OnStart(ctx *WorkflowContext) {
	fmt.Printf("Starting workflow: %s\n", mw.Name)
}

func (mw *MockWorkflow) OnEnd(ctx *WorkflowContext) {
	fmt.Printf("Ending workflow: %s (Retries: %d)\n", mw.Name, ctx.RetryCount)
}

func TestWorkflowExecutor(t *testing.T) {
	tests := []struct {
		name           string
		maxRetries     int
		retryDelay     time.Duration
		successPattern []bool
		expectedResult bool
	}{
		{
			name:           "Single attempt, fails",
			maxRetries:     0,
			retryDelay:     time.Millisecond,
			successPattern: []bool{false},
			expectedResult: false,
		},
		{
			name:           "Multiple attempts, succeeds on last try",
			maxRetries:     2,
			retryDelay:     time.Millisecond,
			successPattern: []bool{false, false, true},
			expectedResult: true,
		},
		{
			name:           "Multiple attempts, all fail",
			maxRetries:     2,
			retryDelay:     time.Millisecond,
			successPattern: []bool{false, false, false},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &MockWorkflow{
				Name:           tt.name,
				SuccessPattern: tt.successPattern,
			}
			executor := &WorkflowExecutor{}

			config := WorkflowConfig{
				MaxRetries: tt.maxRetries,
				RetryDelay: tt.retryDelay,
				Timeout:    time.Second, // Not used in this test
			}

			success, err := executor.Execute(workflow, config)

			if tt.expectedResult {
				assert.True(t, success, "Expected success, but got failure")
				assert.NoError(t, err, "Expected no error, but got: %v", err)
			} else {
				assert.False(t, success, "Expected failure, but got success")
				assert.Error(t, err, "Expected an error, but got none")
			}

			assert.Equal(t, len(tt.successPattern), workflow.callCount, "Unexpected number of Execute calls")
		})
	}
}
