package contextutils

import (
	"context"
	"fmt"
)

// AddKeyValueToCtx adds a key to the context and assigned to it a value
func AddKeyValueToCtx(ctx context.Context, key, value any) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetKeyValueFromCtx retrieves the value of a specific key in the context
func GetKeyValueFromCtx[K any, T any](ctx context.Context, key K) (T, error) {
	v := ctx.Value(key)
	if v == nil {
		return v.(T), fmt.Errorf("value with key '%v' not found", key)
	}

	return v.(T), nil
}

// RemoveKeyFromCtx removes a key and its value from context
func RemoveKeyFromCtx(ctx context.Context, key any) context.Context {
	return context.WithValue(ctx, key, nil)
}
