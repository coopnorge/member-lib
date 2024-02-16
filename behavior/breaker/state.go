package breaker

import "fmt"

// State is a type that represents a state of CircuitBreaker.
type State byte

// These constants related to State of CircuitBreaker.
const (
	// StateClosed invoke Action and if the call to the operation is unsuccessful - increments of failures.
	StateClosed State = iota
	// StateHalfOpen a limited number of successful operations are allowed to pass through and invoke Action.
	StateHalfOpen
	// StateOpen requires to fail immediately and an ErrCircuitOpen is returned.
	StateOpen
)

// String implements stringer interface.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "Closed"
	case StateHalfOpen:
		return "Half-Open"
	case StateOpen:
		return "Open"
	default:
		return fmt.Sprintf("unknown state: %d", s)
	}
}

// IsState current matching to the needed.
func (s State) IsState(needed State) bool {
	return s == needed
}
