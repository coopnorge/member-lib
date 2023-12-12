package contextutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddKeyValueToCtx(t *testing.T) {
	ctx := context.Background()

	newCtx := AddKeyValueToCtx(ctx, "ctxKey", "ctxValue")

	assert.Equal(t, newCtx.Value("ctxKey"), "ctxValue")
}

func TestGetKeyValueFromCtx(t *testing.T) {
	ctx := context.Background()
	ctxWithValue := context.WithValue(ctx, "ctxKey", "ctxValue")

	v, err := GetKeyValueFromCtx[string, string](ctxWithValue, "ctxKey")

	assert.Nil(t, err)
	assert.Equal(t, v, "ctxValue")
}

func TestRemoveKeyFromCtx(t *testing.T) {
	ctx := context.Background()
	ctxWithValue := context.WithValue(ctx, "ctxKey", "ctxValue")

	newCtx := RemoveKeyFromCtx(ctxWithValue, "ctxKey")

	assert.Equal(t, newCtx.Value("ctxKey"), nil)
}
