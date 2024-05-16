package contextext

import (
	"context"
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
