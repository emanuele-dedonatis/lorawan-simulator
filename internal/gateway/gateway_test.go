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

		gw := New(eui, discoveryURI)

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

		gw1 := New(eui1, discoveryURI1)
		gw2 := New(eui2, discoveryURI2)

		assert.NotEqual(t, gw1, gw2)
		assert.Equal(t, eui1, gw1.eui)
		assert.Equal(t, eui2, gw2.eui)
		assert.Equal(t, discoveryURI1, gw1.discoveryURI)
		assert.Equal(t, discoveryURI2, gw2.discoveryURI)
	})

	t.Run("initializes with disconnected state", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"

		gw := New(eui, discoveryURI)

		assert.Equal(t, StateDisconnected, gw.discoveryState)
		assert.Equal(t, StateDisconnected, gw.dataState)
	})
}

func TestGateway_GetInfo(t *testing.T) {
	t.Run("returns correct gateway info", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

		info := gw.GetInfo()

		assert.Equal(t, eui, info.EUI)
		assert.Equal(t, discoveryURI, info.DiscoveryURI)
		assert.Equal(t, StateDisconnected, info.DiscoveryState)
		assert.Equal(t, StateDisconnected, info.DataState)
		assert.Equal(t, "", info.DataURI)
	})

	t.Run("returns updated state after connect", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

		reply := gw.ConnectAsync()
		err := <-reply
		assert.NoError(t, err)

		info := gw.GetInfo()
		assert.Equal(t, StateConnected, info.DiscoveryState)
	})

	t.Run("is thread-safe for concurrent reads", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

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
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

		info := gw.GetInfo()
		assert.Equal(t, StateDisconnected, info.DiscoveryState)

		reply := gw.ConnectAsync()
		err := <-reply
		assert.NoError(t, err)

		info = gw.GetInfo()
		assert.Equal(t, StateConnected, info.DiscoveryState)
	})

	t.Run("returns error when already connected", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

		reply := gw.ConnectAsync()
		err := <-reply
		assert.NoError(t, err)

		info := gw.GetInfo()
		assert.Equal(t, StateConnected, info.DiscoveryState)

		reply = gw.ConnectAsync()
		err = <-reply
		assert.Error(t, err)
		assert.Equal(t, "already connected", err.Error())
	})

	t.Run("is thread-safe", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				reply := gw.ConnectAsync()
				err := <-reply
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
		assert.Equal(t, StateConnected, info.DiscoveryState)
	})
}

func TestGateway_Disconnect(t *testing.T) {
	t.Run("changes state to disconnected", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

		reply := gw.ConnectAsync()
		err := <-reply
		assert.NoError(t, err)

		info := gw.GetInfo()
		assert.Equal(t, StateConnected, info.DiscoveryState)

		reply = gw.DisconnectAsync()
		err = <-reply
		assert.NoError(t, err)

		info = gw.GetInfo()
		assert.Equal(t, StateDisconnected, info.DataState)
		assert.Equal(t, StateDisconnected, info.DiscoveryState)
	})

	t.Run("returns error when already disconnected", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

		info := gw.GetInfo()
		assert.Equal(t, StateDisconnected, info.DataState)

		reply := gw.DisconnectAsync()
		err := <-reply
		assert.Error(t, err)
		assert.Equal(t, "already disconnected", err.Error())
	})

	t.Run("is thread-safe", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

		reply := gw.ConnectAsync()
		err := <-reply
		assert.NoError(t, err)

		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				reply := gw.DisconnectAsync()
				err := <-reply
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
		assert.Equal(t, StateDisconnected, info.DiscoveryState)
	})
}

func TestGateway_StateTransitions(t *testing.T) {
	t.Run("connect and disconnect cycle", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

		// Initial state
		info := gw.GetInfo()
		assert.Equal(t, StateDisconnected, info.DiscoveryState)

		// Connect
		reply := gw.ConnectAsync()
		err := <-reply
		assert.NoError(t, err)

		info = gw.GetInfo()
		assert.Equal(t, StateConnected, info.DiscoveryState)

		// Disconnect
		reply = gw.DisconnectAsync()
		err = <-reply
		assert.NoError(t, err)

		info = gw.GetInfo()
		assert.Equal(t, StateDisconnected, info.DiscoveryState)

		// Reconnect
		reply = gw.ConnectAsync()
		err = <-reply
		assert.NoError(t, err)

		info = gw.GetInfo()
		assert.Equal(t, StateConnected, info.DiscoveryState)
	})

	t.Run("concurrent connect and disconnect", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

		var wg sync.WaitGroup

		// Concurrent connects
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				reply := gw.ConnectAsync()
				<-reply // ignore errors
			}()
		}

		// Concurrent disconnects
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				reply := gw.DisconnectAsync()
				<-reply // ignore errors
			}()
		}

		wg.Wait()

		// Final state should be either connected or disconnected (no corruption)
		info := gw.GetInfo()
		assert.True(t, info.DiscoveryState == StateConnected || info.DiscoveryState == StateDisconnected)
	})
}

func TestGateway_ConcurrentOperations(t *testing.T) {
	t.Run("concurrent GetInfo, Connect, and Disconnect", func(t *testing.T) {
		t.Skip("Skipping test that requires real WebSocket server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://gateway.example.com:6887"
		gw := New(eui, discoveryURI)

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
				reply := gw.ConnectAsync()
				<-reply // ignore errors
			}()
		}

		// Concurrent Disconnect calls
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				reply := gw.DisconnectAsync()
				<-reply // ignore errors
			}()
		}

		wg.Wait()

		// Verify no data corruption
		info := gw.GetInfo()
		assert.Equal(t, eui, info.EUI)
		assert.Equal(t, discoveryURI, info.DiscoveryURI)
	})
}
