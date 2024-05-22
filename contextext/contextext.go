package contextext

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type (
	// ContextExtended enhances the standard context with additional features.
	ContextExtended[T any] struct {
		values   sync.Map
		ctx      context.Context
		cancel   context.CancelFunc
		deadline time.Time
	}
)

// NewContextExtended constructor for ContextExtended with a given original context.
func NewContextExtended[T any](base context.Context) *ContextExtended[T] {
	ctx, cancel := context.WithCancel(base)
	ctxExt := &ContextExtended[T]{cancel: cancel}
	ctxExt.ctx = storeContextExtended(ctx, ctxExt)

	return ctxExt
}

// ExtendTimout extends the context's timeout.
func (ce *ContextExtended[T]) ExtendTimout(d time.Duration) {
	newDeadline := time.Now().Add(d)
	if ce.deadline.IsZero() || newDeadline.After(ce.deadline) {
		ce.ctx, ce.cancel = context.WithDeadline(ce.ctx, newDeadline)
		ce.deadline = newDeadline
	}
}

// AddValue safely adds a value to the context.
func (ce *ContextExtended[T]) AddValue(key any, value T) {
	ce.values.Store(key, value)
}

// GetValue safely retrieves a value from the context.
func (ce *ContextExtended[T]) GetValue(key any) (T, bool) {
	val, ok := ce.values.Load(key)
	if ok {
		return val.(T), ok
	}

	var zero T
	return zero, false
}

// RemoveValue safely removes a value from the context.
func (ce *ContextExtended[T]) RemoveValue(key any) {
	ce.values.Delete(key)
}

// Deadline returns the time when work done on behalf of this context should be canceled.
func (ce *ContextExtended[T]) Deadline() (deadline time.Time, ok bool) {
	return ce.ctx.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this context should be canceled.
func (ce *ContextExtended[T]) Done() <-chan struct{} {
	return ce.ctx.Done()
}

// Err returns a non-nil error value after Done is closed.
func (ce *ContextExtended[T]) Err() error {
	return ce.ctx.Err()
}

// Value returns the value associated with this context for key, or nil if no value is associated with key.
func (ce *ContextExtended[T]) Value(key any) any {
	// Try access collected values from map in case if direct access to ContextExtended[T].
	v, exist := ce.GetValue(key)
	if exist {
		return v
	}

	// Try access to ContextExtended[T] if vanilla context.Context was called.
	if v := ce.ctx.Value(key); v != nil {
		return v
	}

	return nil
}

// Cancel cancels the context.
func (ce *ContextExtended[T]) Cancel() {
	ce.cancel()
}

// SafelyExtractExtendedContextFromInterface by casting interface to extended generic context.
// T type of values that are stored in context.
func SafelyExtractExtendedContextFromInterface[T any](ctx context.Context) (*ContextExtended[T], error) {
	key := getContextExtendedKey[T]()
	val := ctx.Value(key)
	if val != nil {
		if ce, ok := val.(*ContextExtended[T]); ok {
			return ce, nil
		}
	}

	return nil, fmt.Errorf("given context (%v) not a type of %v", reflect.TypeOf(ctx), reflect.TypeOf(ContextExtended[T]{}))
}

// getContextExtendedKey generates a unique key for storing ContextExtended in the context.
func getContextExtendedKey[T any]() string {
	return reflect.TypeOf((*ContextExtended[T])(nil)).Elem().String()
}

// StoreContextExtended stores the ContextExtended in the context.
func storeContextExtended[T any](ctx context.Context, ce *ContextExtended[T]) context.Context {
	return context.WithValue(ctx, getContextExtendedKey[T](), ce)
}
