// Package breaker implements the Circuit Breaker pattern.
//
// This pattern prevents an application from performing operations
// that are likely to fail, allowing it to maintain high performance
// and availability even when some parts of a system are not functioning
// optimally.
//
// The breaker package provides a simple yet flexible way to
// integrate circuit breaking logic into your applications.
package breaker

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/coopnorge/member-lib/pkg/stringconv"
)

var (
	// ErrCircuitOpen is returned when the state of Circuit Breaker is open.
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// Action function that will be executed inside CircuitBreaker.
type Action func() (any, error)

// Configuration of CircuitBreaker
//
// MaxFailuresThreshold Maximum number of failures allowed.
//
// ResetTimeout in seconds, is the period of the open state. After which the state of the CircuitBreaker becomes half-open.
// fields can be used for example as ENV variables in your project like MY_APP_CB_MAX_FAILURES_THRESHOLD:"3".
type Configuration struct {
	MaxFailuresThreshold string `json:"cb_max_failures_threshold,omitempty"`
	ResetTimeout         string `json:"cb_reset_timeout,omitempty"`
}

// CircuitBreakerHandler is the interface for circuit breaker functionality.
type CircuitBreakerHandler interface {
	// Proceed with Action inside CircuitBreakerHandler, simply wraps Action and changing State on result.
	Proceed(action Action) (any, error)
	// GetState returns current state of the CircuitBreakerHandler - Circuit Breaker State.
	GetState() State
	// Reset or reboot CircuitBreakerHandler State to initial of Circuit Breaker.
	Reset()
}

// CircuitBreaker component that is designed  to prevent sending execution that are likely to fail.
type CircuitBreaker struct {
	OnSuccess func() // OnSuccess will be called when Action can be processed, when state is StateClosed or StateHalfOpen
	OnFailure func() // OnFailure will be triggered when CircuitBreaker moved to StateOpen and Action failed to execute and will return error.

	mu           sync.RWMutex
	currentState State

	timeout      time.Duration // Duration when state must be closed
	lastAttempt  time.Time     // Timestamp of the last attempt to execution
	failureCount uint64        // Current count of consecutive failures
	failureLimit uint64        // Number of failures that will switch the state from closed to open
}

// NewCircuitBreaker creates a new CircuitBreaker instance with the specified configuration.
func NewCircuitBreaker(cfg *Configuration) (*CircuitBreaker, error) {
	errTmpl := "failed to parse parameter for %s"

	restTimout, restTimoutErr := stringconv.ToWholeNumber[int32](cfg.ResetTimeout)
	if restTimoutErr != nil {
		return nil, errors.Join(fmt.Errorf(errTmpl, "ResetTimout in CircuitBreaker"), restTimoutErr)
	}

	maxFailuresThreshold, maxFailuresThresholdErr := stringconv.ToWholeNumber[uint64](cfg.MaxFailuresThreshold)
	if maxFailuresThresholdErr != nil {
		return nil, errors.Join(fmt.Errorf(errTmpl, "MaxFailuresThreshold in CircuitBreaker"), maxFailuresThresholdErr)
	}

	cb := &CircuitBreaker{
		currentState: StateClosed,
		timeout:      time.Second * time.Duration(restTimout),
		OnSuccess:    func() {},
		OnFailure:    func() {},
		failureLimit: maxFailuresThreshold,
	}

	return cb, nil
}

// Proceed with Action inside CircuitBreaker.
func (cb *CircuitBreaker) Proceed(action Action) (any, error) {
	switch cb.GetState() {
	default: // StateOpen
		if cb.isTimeout() {
			cb.setState(StateHalfOpen)
			return cb.Proceed(action)
		} else {
			return nil, ErrCircuitOpen
		}
	case StateHalfOpen, StateClosed:
		result, err := action()
		if err != nil {
			cb.recordFailure()
			cb.OnFailure()
			return nil, err
		}

		cb.OnSuccess()
		cb.recordSuccess()

		return result, nil
	}
}

// GetState returns current state of the CircuitBreaker.
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.currentState
}

// Reset or reboot CircuitBreaker state to initial.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	cb.failureCount = 0
	cb.mu.Unlock()

	cb.setState(StateClosed)
}

func (cb *CircuitBreaker) setState(state State) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.currentState = state
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	cb.failureCount++
	cb.lastAttempt = time.Now()
	isExceededFailureCount := cb.failureCount > cb.failureLimit
	cb.mu.Unlock()

	if isExceededFailureCount {
		cb.setState(StateOpen)
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()

	if cb.failureCount > 0 {
		cb.failureCount--
		cb.mu.Unlock()
	} else if cb.failureCount == 0 {
		cb.mu.Unlock()
		cb.setState(StateClosed)
	}
}

func (cb *CircuitBreaker) isTimeout() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return time.Since(cb.lastAttempt) > cb.timeout
}
