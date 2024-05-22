package contextext

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type unitStubContextKey struct{}

func TestContextExtended(t *testing.T) {
	baseCtx := context.Background()
	cp := NewContextExtended[string](baseCtx)

	// Test adding and getting a string value
	cp.AddValue("key1", "value1")
	if value, ok := cp.GetValue("key1"); !ok || value != "value1" {
		t.Errorf("expected value1, got %v", value)
	}

	// Test removing a value
	cp.RemoveValue("key1")
	if _, ok := cp.GetValue("key1"); ok {
		t.Errorf("expected value to be removed")
	}

	// Test timeout extension
	cp.ExtendTimout(2 * time.Second)
	select {
	case <-cp.Done():
		t.Errorf("expected context to be not done yet")
	case <-time.After(1 * time.Second):
	}

	cp.ExtendTimout(1 * time.Second)
	select {
	case <-cp.Done():
		t.Errorf("expected context to be not done yet")
	case <-time.After(500 * time.Millisecond):
	}

	// Test context cancellation
	cp.Cancel()
	select {
	case <-cp.Done():
	default:
		t.Errorf("expected context to be done")
	}
}

func TestSafelyExtractExtendedContextFromInterface(t *testing.T) {
	t.Run("CorrectType", func(t *testing.T) {
		expectedValue := "test value"
		extCtx := NewContextExtended[string](context.Background())
		extCtx.AddValue("unit", expectedValue)

		result, err := SafelyExtractExtendedContextFromInterface[string](extCtx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assert.Equal(t, extCtx.values, result.values)
	})

	t.Run("IncorrectType", func(t *testing.T) {
		extCtx := NewContextExtended[int](context.Background())

		_, err := SafelyExtractExtendedContextFromInterface[string](extCtx)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		expectedError := fmt.Sprintf("given context (%v) not a type of %v", reflect.TypeOf(extCtx), reflect.TypeOf(ContextExtended[string]{}))

		if err.Error() != expectedError {
			t.Fatalf("expected error message %v, got %v", expectedError, err.Error())
		}
	})

	t.Run("NotExtendedContext", func(t *testing.T) {
		regularCtx := context.Background()
		_, err := SafelyExtractExtendedContextFromInterface[string](regularCtx)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		expectedError := fmt.Sprintf("given context (%v) not a type of %v", reflect.TypeOf(regularCtx), reflect.TypeOf(ContextExtended[string]{}))

		if err.Error() != expectedError {
			t.Fatalf("expected error message %v, got %v", expectedError, err.Error())
		}
	})

	t.Run("VanillaContext", func(t *testing.T) {
		expectedValue := "test value"
		extCtx := NewContextExtended[string](context.Background())
		extCtx.AddValue("unit", expectedValue)

		vanillaCtx, vanillaCtxCancel := context.WithDeadline(extCtx, time.Now())
		vanillaCtxCancel()

		result, err := SafelyExtractExtendedContextFromInterface[string](vanillaCtx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assert.Equal(t, extCtx.values, result.values)
	})

	t.Run("duplicatedKeyOfContextExt", func(t *testing.T) {
		expectedValue := "test value"

		extCtx := NewContextExtended[string](context.Background())
		extCtx.AddValue("unit", expectedValue)

		extCtx2 := NewContextExtended[string](context.Background())

		extCtx2Result, extCtx2ResultErr := SafelyExtractExtendedContextFromInterface[string](extCtx2)
		if extCtx2ResultErr != nil {
			t.Fatalf("expected no error, got %v", extCtx2ResultErr)
		}

		_, not := extCtx2Result.GetValue("unit")
		assert.False(t, not, "expected to be not found value")

		extCtxResult, extCtxResultErr := SafelyExtractExtendedContextFromInterface[string](extCtx)
		if extCtxResultErr != nil {
			t.Fatalf("expected no error, got %v", extCtxResultErr)
		}

		resultValue, in := extCtxResult.GetValue("unit")
		assert.True(t, in, "expected to be not found value")
		assert.True(t, resultValue == expectedValue)
	})

	t.Run("duplicatedKeyOfContextExtWithVanillaContext", func(t *testing.T) {
		expectedValue := "test value"

		extCtx := NewContextExtended[string](context.Background())
		extCtx.AddValue("unit", expectedValue)

		extCtx2 := NewContextExtended[string](context.Background())

		extCtx2Result, extCtx2ResultErr := SafelyExtractExtendedContextFromInterface[string](extCtx2)
		if extCtx2ResultErr != nil {
			t.Fatalf("expected no error, got %v", extCtx2ResultErr)
		}

		_, not := extCtx2Result.GetValue("unit")
		assert.False(t, not, "expected to be not found value")

		vanillaCtx, vanillaCtxCancel := context.WithDeadline(extCtx, time.Now())
		vanillaCtxCancel()

		extCtxResult, extCtxResultErr := SafelyExtractExtendedContextFromInterface[string](vanillaCtx)
		if extCtxResultErr != nil {
			t.Fatalf("expected no error, got %v", extCtxResultErr)
		}

		resultValue, in := extCtxResult.GetValue("unit")
		assert.True(t, in, "expected to be not found value")
		assert.True(t, resultValue == expectedValue)
	})

	t.Run("ContextExtended as int", func(t *testing.T) {
		expectedValue := 42
		extCtx := NewContextExtended[int](context.Background())
		extCtx.AddValue("unit", expectedValue)

		result, err := SafelyExtractExtendedContextFromInterface[int](extCtx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assert.Equal(t, extCtx.values, result.values)
	})

	t.Run("ContextExtended as struct", func(t *testing.T) {
		expectedValue := "42"
		extCtx := NewContextExtended[string](context.Background())
		extCtx.AddValue(unitStubContextKey{}, expectedValue)

		result, err := SafelyExtractExtendedContextFromInterface[string](extCtx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		resultValue, _ := result.GetValue(unitStubContextKey{})
		assert.True(t, resultValue == expectedValue)
		assert.Equal(t, extCtx.values, result.values)
	})

	t.Run("ContextExtended as ContextExtended", func(t *testing.T) {
		extCtxValue := "42"
		extCtx := NewContextExtended[string](context.Background())
		extCtx.AddValue(unitStubContextKey{}, extCtxValue)

		extCtxOfCtxValue := "universe"
		extCtxOfCtx := NewContextExtended[string](extCtx)
		extCtxOfCtx.AddValue("answer", extCtxOfCtxValue)

		extCtxResult, extCtxErr := SafelyExtractExtendedContextFromInterface[string](extCtx)
		if extCtxErr != nil {
			t.Fatalf("expected no error, got %v", extCtxErr)
		}

		extCtxResultValue, _ := extCtxResult.GetValue(unitStubContextKey{})
		assert.True(t, extCtxResultValue == extCtxValue)

		extCtxOfCtxResult, extCtxOfCtxErr := SafelyExtractExtendedContextFromInterface[string](extCtxOfCtx)
		if extCtxOfCtxErr != nil {
			t.Fatalf("expected no error, got %v", extCtxOfCtxErr)
		}

		extCtxOfCtxResultValue, _ := extCtxOfCtxResult.GetValue("answer")
		assert.True(t, extCtxOfCtxResultValue == extCtxOfCtxValue)
	})
}
