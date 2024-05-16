package contextext

import (
	"context"
	"sync"
	"time"
)

type (
	// ContextExtended enhances the standard context with additional features.
	ContextExtended[T any] struct {
		mu       sync.RWMutex
		values   map[string]T
		ctx      context.Context
		cancel   context.CancelFunc
		deadline time.Time
	}
)

// NewContextExtended constructor for ContextExtended with a given original context.
func NewContextExtended[T any](base context.Context) *ContextExtended[T] {
	ctx, cancel := context.WithCancel(base)
	return &ContextExtended[T]{
		values: make(map[string]T),
		ctx:    ctx,
		cancel: cancel,
	}
}

// ExtendTimout extends the context's timeout.
func (cp *ContextExtended[T]) ExtendTimout(d time.Duration) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	newDeadline := time.Now().Add(d)
	if cp.deadline.IsZero() || newDeadline.After(cp.deadline) {
		cp.ctx, cp.cancel = context.WithDeadline(cp.ctx, newDeadline)
		cp.deadline = newDeadline
	}
}

// AddValue safely adds a value to the context.
func (cp *ContextExtended[T]) AddValue(key string, value T) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.values[key] = value
}

// GetValue safely retrieves a value from the context.
func (cp *ContextExtended[T]) GetValue(key string) (T, bool) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	value, ok := cp.values[key]
	return value, ok
}

// RemoveValue safely removes a value from the context.
func (cp *ContextExtended[T]) RemoveValue(key string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	delete(cp.values, key)
}

// Deadline returns the time when work done on behalf of this context should be canceled.
func (cp *ContextExtended[T]) Deadline() (deadline time.Time, ok bool) {
	return cp.ctx.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this context should be canceled.
func (cp *ContextExtended[T]) Done() <-chan struct{} {
	return cp.ctx.Done()
}

// Err returns a non-nil error value after Done is closed.
func (cp *ContextExtended[T]) Err() error {
	return cp.ctx.Err()
}

// Value returns the value associated with this context for key, or nil if no value is associated with key.
func (cp *ContextExtended[T]) Value(key interface{}) interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	if keyStr, ok := key.(string); ok {
		if value, exists := cp.values[keyStr]; exists {
			return value
		}
	}
	return cp.ctx.Value(key)
}

// Cancel cancels the context.
func (cp *ContextExtended[T]) Cancel() {
	cp.cancel()
}
