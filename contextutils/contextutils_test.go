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
	var testCases = []struct {
		description  string
		isSuccessful bool
	}{
		{
			description:  "should return an error because key is not found",
			isSuccessful: false,
		},
		{
			description:  "should return no error because key is found",
			isSuccessful: true,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		ctxWithValue := context.WithValue(ctx, stubContextKey{}, "ctxValue")

		if testCase.isSuccessful {
			v, err := GetKeyValue[stubContextKey, string](ctxWithValue, stubContextKey{})
			assert.Nil(t, err)
			assert.Equal(t, v, "ctxValue")
		} else {
			v, err := GetKeyValue[stubContextKey, string](ctx, stubContextKey{})
			assert.NotNil(t, err)
			assert.Equal(t, v, "")
		}
	}
}

func TestRemoveKeyFromCtx(t *testing.T) {
	ctx := context.Background()
	ctxWithValue := context.WithValue(ctx, stubContextKey{}, "ctxValue")

	newCtx := RemoveKey(ctxWithValue, stubContextKey{})

	assert.Equal(t, newCtx.Value(stubContextKey{}), nil)
}
