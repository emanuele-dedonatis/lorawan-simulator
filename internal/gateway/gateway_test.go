package gateway

import (
	"sync"
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("creates gateway with valid parameters", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"

		gw := newTestGateway(eui, discoveryURI)

		assert.NotNil(t, gw)
		assert.Equal(t, eui, gw.eui)
		assert.Equal(t, discoveryURI, gw.discoveryURI)
		assert.Equal(t, StateDisconnected, gw.discoveryState)
		assert.Equal(t, StateDisconnected, gw.dataState)
		assert.Equal(t, "", gw.dataURI)
	})

	t.Run("creates multiple independent gateways", func(t *testing.T) {
		eui1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		eui2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		discoveryURI1 := "wss://gateway1.example.com:6887"
		discoveryURI2 := "wss://gateway2.example.com:6887"

		gw1 := newTestGateway(eui1, discoveryURI1)
		gw2 := newTestGateway(eui2, discoveryURI2)

		assert.NotEqual(t, gw1, gw2)
		assert.Equal(t, eui1, gw1.eui)
		assert.Equal(t, eui2, gw2.eui)
		assert.Equal(t, discoveryURI1, gw1.discoveryURI)
		assert.Equal(t, discoveryURI2, gw2.discoveryURI)
	})

	t.Run("initializes with disconnected state", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"

		gw := newTestGateway(eui, discoveryURI)

		assert.Equal(t, StateDisconnected, gw.discoveryState)
		assert.Equal(t, StateDisconnected, gw.dataState)
	})
}

func TestGateway_GetInfo(t *testing.T) {
	t.Run("returns correct gateway info", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := newTestGateway(eui, discoveryURI)

		info := gw.GetInfo()

		assert.Equal(t, eui, info.EUI)
		assert.Equal(t, discoveryURI, info.DiscoveryURI)
		assert.Equal(t, "disconnected", info.DiscoveryState)
		assert.Equal(t, "disconnected", info.DataState)
		assert.Equal(t, "", info.DataURI)
	})

	t.Run("returns updated state after connect", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "ws://localhost:3001"
		gw := newTestGateway(eui, discoveryURI)

		err := gw.Connect()
		assert.NoError(t, err)

		info := gw.GetInfo()
		assert.Equal(t, "disconnected", info.DiscoveryState) // Discovery closes after getting data URI
		assert.NotEmpty(t, info.DataURI)                     // Should have received data URI
	})

	t.Run("is thread-safe for concurrent reads", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := newTestGateway(eui, discoveryURI)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				info := gw.GetInfo()
				assert.Equal(t, eui, info.EUI)
				assert.Equal(t, discoveryURI, info.DiscoveryURI)
			}()
		}
		wg.Wait()
	})
}

func TestGateway_Connect(t *testing.T) {
	t.Run("changes state to connected", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "ws://localhost:3001"
		gw := newTestGateway(eui, discoveryURI)

		info := gw.GetInfo()
		assert.Equal(t, "disconnected", info.DiscoveryState)

		err := gw.Connect()
		assert.NoError(t, err)

		info = gw.GetInfo()
		assert.NotEmpty(t, info.DataURI) // Should have data URI
	})

	t.Run("returns error when already connected", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "ws://localhost:3001"
		gw := newTestGateway(eui, discoveryURI)

		err := gw.Connect()
		assert.NoError(t, err)

		info := gw.GetInfo()
		assert.NotEmpty(t, info.DataURI)

		err = gw.Connect()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already connect")
	})

	t.Run("is thread-safe", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "ws://localhost:3001"
		gw := newTestGateway(eui, discoveryURI)

		var wg sync.WaitGroup
		errorCount := 0
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := gw.Connect()
				mu.Lock()
				if err == nil {
					successCount++
				} else {
					errorCount++
				}
				mu.Unlock()
			}()
		}
		wg.Wait()

		// Only one should succeed, rest should get "already connecting" or "already connected" errors
		assert.Equal(t, 1, successCount)
		assert.Equal(t, 99, errorCount)
	})
}

func TestGateway_Disconnect(t *testing.T) {
	t.Run("changes state to disconnected", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "ws://localhost:3001"
		gw := newTestGateway(eui, discoveryURI)

		err := gw.Connect()
		assert.NoError(t, err)

		info := gw.GetInfo()
		assert.NotEmpty(t, info.DataURI)

		reply := gw.Disconnect()
		assert.NoError(t, reply)

		info = gw.GetInfo()
		assert.Equal(t, "disconnected", info.DataState)
		assert.Equal(t, "disconnected", info.DiscoveryState)
	})

	t.Run("returns error when already disconnected", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := newTestGateway(eui, discoveryURI)

		info := gw.GetInfo()
		assert.Equal(t, "disconnected", info.DataState)

		err := gw.Disconnect()
		assert.Error(t, err)
		assert.Equal(t, "already disconnected", err.Error())
	})

	t.Run("is thread-safe", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "ws://localhost:3001"
		gw := newTestGateway(eui, discoveryURI)

		err := gw.Connect()
		assert.NoError(t, err)

		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := gw.Disconnect()
				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		// Only one should succeed
		assert.Equal(t, 1, successCount)
		info := gw.GetInfo()
		assert.Equal(t, "disconnected", info.DiscoveryState)
	})
}

func TestGateway_StateTransitions(t *testing.T) {
	t.Run("connect and disconnect cycle", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "ws://localhost:3001"
		gw := newTestGateway(eui, discoveryURI)

		// Initial state
		info := gw.GetInfo()
		assert.Equal(t, StateDisconnected, info.DiscoveryState)

		// Connect
		err := gw.Connect()
		assert.NoError(t, err)

		info = gw.GetInfo()
		assert.Equal(t, "disconnected", info.DiscoveryState)

		// Disconnect
		err = gw.Disconnect()
		assert.NoError(t, err)

		info = gw.GetInfo()
		assert.Equal(t, "disconnected", info.DiscoveryState)

		// Reconnect
		err = gw.Connect()
		assert.NoError(t, err)

		info = gw.GetInfo()
		assert.NotEmpty(t, info.DataURI)
	})

	t.Run("concurrent connect and disconnect", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "ws://localhost:3001"
		gw := newTestGateway(eui, discoveryURI)

		var wg sync.WaitGroup

		// Concurrent connects
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				gw.Connect() // ignore errors
			}()
		}

		// Concurrent disconnects
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				gw.Disconnect() // ignore errors
			}()
		}

		wg.Wait()

		// Final state should be either connected or disconnected (no corruption)
		info := gw.GetInfo()
		assert.True(t, info.DiscoveryState == "disconnected") // Discovery always disconnected after getting data URI
	})
}

func TestGateway_ConcurrentOperations(t *testing.T) {
	t.Run("concurrent GetInfo, Connect, and Disconnect", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "ws://localhost:3001"
		gw := newTestGateway(eui, discoveryURI)

		var wg sync.WaitGroup

		// Concurrent GetInfo calls
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				info := gw.GetInfo()
				assert.Equal(t, eui, info.EUI)
				assert.Equal(t, discoveryURI, info.DiscoveryURI)
			}()
		}

		// Concurrent Connect calls
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				gw.Connect() // ignore errors
			}()
		}

		// Concurrent Disconnect calls
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				gw.Disconnect() // ignore errors
			}()
		}

		wg.Wait()

		// Verify no data corruption
		info := gw.GetInfo()
		assert.Equal(t, eui, info.EUI)
		assert.Equal(t, discoveryURI, info.DiscoveryURI)
	})
}
