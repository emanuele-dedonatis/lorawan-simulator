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

	t.Run("StateDiscoveryConnecting returns correct string", func(t *testing.T) {
		state := StateDiscoveryConnecting
		assert.Equal(t, "connecting to LNS Discovery", state.String())
	})

	t.Run("StateDiscoveryConnected returns correct string", func(t *testing.T) {
		state := StateDiscoveryConnected
		assert.Equal(t, "connected to LNS Discovery", state.String())
	})

	t.Run("StateDataConnecting returns correct string", func(t *testing.T) {
		state := StateDataConnecting
		assert.Equal(t, "connecting to LNS Data", state.String())
	})

	t.Run("StateDataConnected returns correct string", func(t *testing.T) {
		state := StateDataConnected
		assert.Equal(t, "connected to LNS Data", state.String())
	})

	t.Run("unknown state returns correct string", func(t *testing.T) {
		state := State(999)
		assert.Equal(t, "unknown", state.String())
	})
}

func TestState_Constants(t *testing.T) {
	t.Run("state constants have expected values", func(t *testing.T) {
		assert.Equal(t, State(0), StateDisconnected)
		assert.Equal(t, State(1), StateDiscoveryConnecting)
		assert.Equal(t, State(2), StateDiscoveryConnected)
		assert.Equal(t, State(3), StateDataConnecting)
		assert.Equal(t, State(4), StateDataConnected)
	})

	t.Run("all states are unique", func(t *testing.T) {
		states := []State{
			StateDisconnected,
			StateDiscoveryConnecting,
			StateDiscoveryConnected,
			StateDataConnecting,
			StateDataConnected,
		}

		// Check uniqueness
		seen := make(map[State]bool)
		for _, state := range states {
			assert.False(t, seen[state], "State %v is not unique", state)
			seen[state] = true
		}

		assert.Equal(t, 5, len(seen), "Should have 5 unique states")
	})
}

func TestState_Ordering(t *testing.T) {
	t.Run("states follow logical progression", func(t *testing.T) {
		assert.True(t, StateDisconnected < StateDiscoveryConnecting)
		assert.True(t, StateDiscoveryConnecting < StateDiscoveryConnected)
		assert.True(t, StateDiscoveryConnected < StateDataConnecting)
		assert.True(t, StateDataConnecting < StateDataConnected)
	})
}
