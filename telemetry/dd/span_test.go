package dd

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestCtxWrapper(t *testing.T) {
	t.Run("Should support TraceID128", func(t *testing.T) {
		rawTraceID, _ := hex.DecodeString("a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6")
		var traceID [16]byte
		copy(traceID[:], rawTraceID)

		ctx := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
		})

		wrapper := ctxWrapper{otelCtx: ctx}

		assert.Equal(t, "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", wrapper.TraceID128())
	})

	t.Run("Should support TraceID128Bytes", func(t *testing.T) {
		rawTraceID, _ := hex.DecodeString("a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6")
		var traceID [16]byte
		copy(traceID[:], rawTraceID)

		ctx := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
		})

		wrapper := ctxWrapper{otelCtx: ctx}

		assert.Equal(t, traceID, wrapper.TraceID128Bytes())
	})

	t.Run("Should support SpanID", func(t *testing.T) {
		// Create a known span ID for testing
		rawSpanID, _ := hex.DecodeString("f1f2f3f4f5f6f7f8")
		var spanID [8]byte
		copy(spanID[:], rawSpanID)

		ctx := trace.NewSpanContext(trace.SpanContextConfig{
			SpanID: spanID,
		})

		wrapper := ctxWrapper{otelCtx: ctx}

		expectedSpanID := uint64(0xf1f2f3f4f5f6f7f8)
		assert.Equal(t, expectedSpanID, wrapper.SpanID())
	})

	t.Run("Should support TraceID", func(t *testing.T) {
		// Create a trace ID where we know the expected uint64 value of the last 8 bytes
		rawTraceID, _ := hex.DecodeString("0102030405060708f1f2f3f4f5f6f7f8")
		var traceID [16]byte
		copy(traceID[:], rawTraceID)

		ctx := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
		})

		wrapper := ctxWrapper{otelCtx: ctx}

		expectedID := uint64(0xf1f2f3f4f5f6f7f8)
		assert.Equal(t, expectedID, wrapper.TraceID())
	})

	t.Run("Should support ForeachBaggageItem", func(t *testing.T) {
		state := trace.TraceState{}
		state, err := state.Insert("key1", "value1")
		require.NoError(t, err)
		state, err = state.Insert("key2", "value2")
		require.NoError(t, err)

		ctx := trace.NewSpanContext(trace.SpanContextConfig{
			TraceState: state,
		})

		wrapper := ctxWrapper{otelCtx: ctx}

		items := make(map[string]string)
		wrapper.ForeachBaggageItem(func(k string, v string) bool {
			items[k] = v
			return true
		})

		assert.Equal(t, map[string]string{
			"key1": "value1",
			"key2": "value2",
		}, items)
	})
}

func TestNoParent(t *testing.T) {
	state := trace.TraceState{}
	state, err := state.Insert("foo", "bar")
	require.NoError(t, err)
	state, err = state.Insert("foo", "baz")
	require.NoError(t, err)
	wrapper := ctxWrapper{
		otelCtx: trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			TraceState: state,
			SpanID:     [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}),
	}

	t.Run("SpanID should always return 0", func(t *testing.T) {
		assert.NotEqual(t, uint64(0), wrapper.SpanID())
		assert.Equal(t, uint64(0), noParent{SpanContextW3C: &wrapper}.SpanID())
	})

	t.Run("Should delegate TraceID128 to wrapped context", func(t *testing.T) {
		assert.Equal(t, wrapper.TraceID128(), noParent{SpanContextW3C: &wrapper}.TraceID128())
	})

	t.Run("Should delegate TraceID128Bytes to wrapped context", func(t *testing.T) {
		assert.Equal(t, wrapper.TraceID128Bytes(), noParent{SpanContextW3C: &wrapper}.TraceID128Bytes())
	})

	t.Run("Should delegate ForeachBaggageItem to wrapped context", func(t *testing.T) {
		expected := map[string]string{}
		wrapper.ForeachBaggageItem(func(k string, v string) bool {
			expected[k] = v
			return true
		})
		actual := map[string]string{}
		noParent{SpanContextW3C: &wrapper}.ForeachBaggageItem(func(k string, v string) bool {
			actual[k] = v
			return true
		})
		assert.Equal(t, expected, actual)
	})
}
