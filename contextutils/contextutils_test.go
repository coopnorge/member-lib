package contextutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubContextKey struct{}

func TestAddKeyValueToCtx(t *testing.T) {
	ctx := context.Background()

	newCtx := AddKeyValue(ctx, stubContextKey{}, "ctxValue")

	assert.Equal(t, newCtx.Value(stubContextKey{}), "ctxValue")
}

func TestGetKeyValueFromCtx(t *testing.T) {
	ctx := context.Background()
	ctxWithValue := context.WithValue(ctx, stubContextKey{}, "ctxValue")

	v, err := GetKeyValue[stubContextKey, string](ctxWithValue, stubContextKey{})

	assert.Nil(t, err)
	assert.Equal(t, v, "ctxValue")
}

func TestRemoveKeyFromCtx(t *testing.T) {
	ctx := context.Background()
	ctxWithValue := context.WithValue(ctx, stubContextKey{}, "ctxValue")

	newCtx := RemoveKey(ctxWithValue, stubContextKey{})

	assert.Equal(t, newCtx.Value(stubContextKey{}), nil)
}
