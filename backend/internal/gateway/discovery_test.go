package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for testing
	},
}

// Mock WebSocket server that simulates LNS discovery endpoint
func mockDiscoveryServer(t *testing.T, behavior string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/router-info" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("Failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		switch behavior {
		case "success":
			// Read router message from gateway
			_, msg, err := conn.ReadMessage()
			if err != nil {
				t.Logf("Failed to read message: %v", err)
				return
			}

			// Verify router message format
			var routerMsg struct {
				Router string `json:"router"`
			}
			if err := json.Unmarshal(msg, &routerMsg); err != nil {
				t.Logf("Failed to parse router message: %v", err)
				return
			}

			// Send back data URI
			response := map[string]string{
				"router": routerMsg.Router,
				"muxs":   routerMsg.Router,
				"uri":    "ws://localhost:3002/gateway/test",
			}
			responseBytes, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, responseBytes)

		case "timeout":
			// Read router message but never respond
			conn.ReadMessage()
			time.Sleep(10 * time.Second) // Wait longer than routerTimeout

		case "invalid_json":
			// Read router message
			conn.ReadMessage()
			// Send invalid JSON
			conn.WriteMessage(websocket.TextMessage, []byte("invalid json {"))

		case "immediate_close":
			// Close connection immediately after upgrade
			return

		case "error_after_read":
			// Read router message
			conn.ReadMessage()
			// Close with error
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "internal error"))

		default:
			t.Fatalf("Unknown behavior: %s", behavior)
		}
	}))
}

func TestLnsDiscovery_Success(t *testing.T) {
	server := mockDiscoveryServer(t, "success")
	defer server.Close()

	// Convert http://... to ws://...
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := newTestGateway(eui, wsURL)

	uri, err := gw.lnsDiscovery()

	assert.NoError(t, err)
	assert.Equal(t, "ws://localhost:3002/gateway/test", uri)

	// Verify state is back to disconnected after discovery
	info := gw.GetInfo()
	assert.Equal(t, "disconnected", info.DiscoveryState)
}

func TestLnsDiscovery_ConnectionError(t *testing.T) {
	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := newTestGateway(eui, "ws://nonexistent.local:9999")

	uri, err := gw.lnsDiscovery()

	assert.Error(t, err)
	assert.Empty(t, uri)
	assert.Contains(t, err.Error(), "dial")

	// Verify state is disconnected after error
	info := gw.GetInfo()
	assert.Equal(t, "disconnected", info.DiscoveryState)
}

func TestLnsDiscovery_Timeout(t *testing.T) {
	server := mockDiscoveryServer(t, "timeout")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := newTestGateway(eui, wsURL)

	start := time.Now()
	uri, err := gw.lnsDiscovery()
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Empty(t, uri)
	assert.Equal(t, "discovery response timeout", err.Error())

	// Should timeout around routerTimeout (5s), allow some margin
	assert.Greater(t, duration, 4*time.Second)
	assert.Less(t, duration, 7*time.Second)

	// Verify state is disconnected after timeout
	info := gw.GetInfo()
	assert.Equal(t, "disconnected", info.DiscoveryState)
}

func TestLnsDiscovery_InvalidJSON(t *testing.T) {
	server := mockDiscoveryServer(t, "invalid_json")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := newTestGateway(eui, wsURL)

	uri, err := gw.lnsDiscovery()

	assert.Error(t, err)
	assert.Empty(t, uri)
	assert.Contains(t, err.Error(), "invalid character")

	// Verify state is disconnected after error
	info := gw.GetInfo()
	assert.Equal(t, "disconnected", info.DiscoveryState)
}

func TestLnsDiscovery_ImmediateClose(t *testing.T) {
	server := mockDiscoveryServer(t, "immediate_close")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := newTestGateway(eui, wsURL)

	uri, err := gw.lnsDiscovery()

	assert.Error(t, err)
	assert.Empty(t, uri)
	// Error could be EOF or "unexpected EOF" depending on timing
	assert.True(t, strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "close"))
}

func TestLnsDiscovery_EUIFormatting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read router message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var routerMsg struct {
			Router string `json:"router"`
		}
		json.Unmarshal(msg, &routerMsg)

		// Verify EUI format is ID6 format (colon-separated 16-bit blocks)
		// Format: HHHH:HHHH:HHHH:HHHH where H is hex digit
		assert.Regexp(t, `^[0-9a-f]{1,4}:[0-9a-f]{1,4}:[0-9a-f]{1,4}:[0-9a-f]{1,4}$`, routerMsg.Router)

		// For EUI aa:bb:cc:dd:ee:ff:00:11, expect aabb:ccdd:eeff:11 (ID6 format)
		assert.Equal(t, "aabb:ccdd:eeff:11", routerMsg.Router)

		// Send response
		response := map[string]string{
			"uri": "ws://test",
		}
		responseBytes, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, responseBytes)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := newTestGateway(eui, wsURL)

	uri, err := gw.lnsDiscovery()

	assert.NoError(t, err)
	assert.Equal(t, "ws://test", uri)
}

func TestLnsDiscovery_StateTransitions(t *testing.T) {
	// Use a server that adds a small delay to make state transitions observable
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read router message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// Add small delay before responding to allow state observation
		time.Sleep(100 * time.Millisecond)

		var routerMsg struct {
			Router string `json:"router"`
		}
		json.Unmarshal(msg, &routerMsg)

		// Send response
		response := map[string]string{
			"router": routerMsg.Router,
			"muxs":   routerMsg.Router,
			"uri":    "ws://localhost:3002/gateway/test",
		}
		responseBytes, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, responseBytes)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := newTestGateway(eui, wsURL)

	// Initial state
	info := gw.GetInfo()
	assert.Equal(t, "disconnected", info.DiscoveryState)

	// Start discovery in goroutine to observe state changes
	done := make(chan struct{})
	var observedConnecting bool

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				info := gw.GetInfo()
				if info.DiscoveryState == "connecting" {
					observedConnecting = true
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	uri, err := gw.lnsDiscovery()
	close(done)

	assert.NoError(t, err)
	assert.NotEmpty(t, uri)

	// Should have observed connecting state (connected state is very brief, may or may not be observed)
	assert.True(t, observedConnecting, "Should observe StateConnecting")

	// Final state should be disconnected
	info = gw.GetInfo()
	assert.Equal(t, "disconnected", info.DiscoveryState)
}

func TestLnsDiscovery_ConcurrentCalls(t *testing.T) {
	server := mockDiscoveryServer(t, "success")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := newTestGateway(eui, wsURL)

	// Try concurrent discovery calls
	const numCalls = 5
	results := make(chan error, numCalls)

	for i := 0; i < numCalls; i++ {
		go func() {
			_, err := gw.lnsDiscovery()
			results <- err
		}()
	}

	// Collect results
	var successCount, errorCount int
	for i := 0; i < numCalls; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	// All calls should complete (either success or error, but no deadlock)
	assert.Equal(t, numCalls, successCount+errorCount)

	// Final state should be disconnected
	info := gw.GetInfo()
	assert.Equal(t, "disconnected", info.DiscoveryState)
}

func TestDiscoveryResponse_Structure(t *testing.T) {
	// Test the internal discoveryResponse struct
	tests := []struct {
		name string
		resp discoveryResponse
	}{
		{
			name: "success response",
			resp: discoveryResponse{
				uri: "ws://test",
				err: nil,
			},
		},
		{
			name: "error response",
			resp: discoveryResponse{
				uri: "",
				err: fmt.Errorf("test error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resp.err == nil {
				assert.NotEmpty(t, tt.resp.uri)
			} else {
				assert.Empty(t, tt.resp.uri)
				assert.Error(t, tt.resp.err)
			}
		})
	}
}

func TestRouterTimeout_Constant(t *testing.T) {
	assert.Equal(t, 5*time.Second, routerTimeout)
}
