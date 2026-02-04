package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState_String(t *testing.T) {
	t.Run("StateDisconnected returns correct string", func(t *testing.T) {
		state := StateDisconnected
		assert.Equal(t, "disconnected", state.String())
	})

	t.Run("StateConnecting returns correct string", func(t *testing.T) {
		state := StateConnecting
		assert.Equal(t, "connecting", state.String())
	})

	t.Run("StateConnected returns correct string", func(t *testing.T) {
		state := StateConnected
		assert.Equal(t, "connected", state.String())
	})

	t.Run("unknown state returns correct string", func(t *testing.T) {
		state := State(999)
		assert.Equal(t, "unknown", state.String())
	})
}

func TestState_Constants(t *testing.T) {
	t.Run("state constants have expected values", func(t *testing.T) {
		assert.Equal(t, State(0), StateDisconnected)
		assert.Equal(t, State(1), StateConnecting)
		assert.Equal(t, State(2), StateConnected)
	})

	t.Run("all states are unique", func(t *testing.T) {
		states := []State{
			StateDisconnected,
			StateConnecting,
			StateConnected,
		}

		// Check uniqueness
		seen := make(map[State]bool)
		for _, state := range states {
			assert.False(t, seen[state], "State %v is not unique", state)
			seen[state] = true
		}

		assert.Equal(t, 3, len(seen), "Should have 3 unique states")
	})
}

func TestState_Ordering(t *testing.T) {
	t.Run("states follow logical progression", func(t *testing.T) {
		assert.True(t, StateDisconnected < StateConnecting)
		assert.True(t, StateConnecting < StateConnected)
	})
}
