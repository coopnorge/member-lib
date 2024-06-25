package breaker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSate(t *testing.T) {
	assert.Equal(t, StateClosed.String(), "Closed")
	assert.Equal(t, StateHalfOpen.String(), "Half-Open")
	assert.Equal(t, StateOpen.String(), "Open")

	assert.True(t, StateClosed.IsState(StateClosed))
	assert.True(t, StateHalfOpen.IsState(StateHalfOpen))
	assert.True(t, StateOpen.IsState(StateOpen))

	assert.False(t, StateOpen.IsState(StateClosed))
	assert.False(t, StateOpen.IsState(StateHalfOpen))

	assert.Equal(t, State(42).String(), "unknown state: 42")
}
