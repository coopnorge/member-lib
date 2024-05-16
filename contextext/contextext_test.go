package contextext

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

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
		extCtx := &ContextExtended[string]{ctx: context.Background(), values: make(map[string]string)}
		extCtx.values["unit"] = expectedValue

		result, err := SafelyExtractExtendedContextFromInterface[string](extCtx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assert.Equal(t, extCtx.values, result.values)
	})

	t.Run("IncorrectType", func(t *testing.T) {
		extCtx := &ContextExtended[int]{ctx: context.Background(), values: make(map[string]int)}
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
}
