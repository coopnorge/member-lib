// Package process is a subpackage of coordinator, providing simple way to create and manage processes such as servers, metrics, and profiling in a clean and controlled manner.
//
// This package allows you to focus on the logic for starting and stopping processes
// within goroutines, while the coordinator handles the application lifecycle.
//
// To use this package, implement the Process interface, which includes methods to
// trigger events like OnStart and OnStop, ensuring integration and control
// over your application's processes via coordinator.ServiceCoordinator.
package process

import "context"

// Severity is an enumerated type representing the importance of a background task to be executed.
type Severity byte

const (
	// TaskSeverityMajor is a constant representing processes critical for execution.
	TaskSeverityMajor Severity = iota
	// TaskSeverityMinor is a constant representing processes classified as non-critical or optional for execution.
	TaskSeverityMinor
)

// Process interface represents a task that requires execution within a microservice.
// This could include a variety of tasks such as servers, events, parsers, aggregators, etc.
type Process interface {
	// GetName method returns the name of the task.
	GetName() string
	// GetSeverity method returns the severity/importance of the task.
	GetSeverity() Severity
	// OnStart method is an event to be called when the main loop starts.
	OnStart(ctx context.Context) error
	// OnStop method is an event to be called when the main loop stops.
	OnStop(ctx context.Context) error
}

// IsCriticalToStop checks if the task is essential for execution.
func IsCriticalToStop(t Process) bool {
	return t.GetSeverity() == TaskSeverityMajor
}
