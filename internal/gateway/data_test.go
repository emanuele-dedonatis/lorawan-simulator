package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// Mock WebSocket server that simulates LNS data endpoint
func mockDataServer(t *testing.T, behavior string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("Failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		switch behavior {
		case "success":
			// Read version message
			_, msg, err := conn.ReadMessage()
			if err != nil {
				t.Logf("Failed to read message: %v", err)
				return
			}

			// Verify version message format
			var versionMsg struct {
				Msgtype  string `json:"msgtype"`
				Station  string `json:"station"`
				Protocol int    `json:"protocol"`
			}
			if err := json.Unmarshal(msg, &versionMsg); err != nil {
				t.Logf("Failed to parse version message: %v", err)
				return
			}

			// Send back router_config
			routerConfig := map[string]interface{}{
				"msgtype":    "router_config",
				"NetID":      []int{1, 2, 3},
				"JoinEui":    [][]int{{1, 2, 3, 4, 5, 6, 7, 8}},
				"region":     "EU863",
				"hwspec":     "sx1301/1",
				"freq_range": []int{863000000, 870000000},
				"DRs":        [][]int{{7, 125, 0}, {8, 125, 0}},
			}
			responseBytes, _ := json.Marshal(routerConfig)
			conn.WriteMessage(websocket.TextMessage, responseBytes)

			// Keep connection alive for additional messages
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
				// Echo back any additional messages
				conn.WriteMessage(websocket.TextMessage, []byte(`{"msgtype":"ack"}`))
			}

		case "immediate_close":
			// Close connection immediately after upgrade
			return

		case "read_error":
			// Close connection after first message
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
			return

		case "write_error":
			// Read version message
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// Close connection to cause write error
			conn.Close()
			time.Sleep(50 * time.Millisecond) // Give time for close to propagate

		case "echo":
			// Simple echo server for testing writes
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}
				conn.WriteMessage(websocket.TextMessage, msg)
			}

		default:
			t.Fatalf("Unknown behavior: %s", behavior)
		}
	}))
}

func TestLnsDataConnect_Success(t *testing.T) {
	server := mockDataServer(t, "success")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = wsURL

	err := gw.lnsDataConnect()
	assert.NoError(t, err)

	// Give goroutines time to start
	time.Sleep(50 * time.Millisecond)

	// Verify state is connected
	info := gw.GetInfo()
	assert.Equal(t, "connected", info.DataState)
	assert.NotNil(t, gw.dataWs)
	assert.NotNil(t, gw.dataSendCh)
}

func TestLnsDataConnect_ConnectionError(t *testing.T) {
	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = "ws://nonexistent.local:9999/invalid"

	err := gw.lnsDataConnect()
	assert.Error(t, err)

	// Verify state is disconnected
	info := gw.GetInfo()
	assert.Equal(t, "disconnected", info.DataState)
}

func TestLnsDataConnect_StateTransitions(t *testing.T) {
	server := mockDataServer(t, "success")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = wsURL

	// Initial state
	assert.Equal(t, "disconnected", gw.GetInfo().DataState)

	// Connect
	err := gw.lnsDataConnect()
	assert.NoError(t, err)

	// Give time for connection to establish
	time.Sleep(50 * time.Millisecond)

	// Final state should be Connected
	assert.Equal(t, "connected", gw.GetInfo().DataState, "Final state should be StateConnected")
}

func TestLnsDataConnect_SendsVersionMessage(t *testing.T) {
	receivedMsg := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read version message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		receivedMsg <- string(msg)

		// Keep connection alive
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = wsURL

	err := gw.lnsDataConnect()
	assert.NoError(t, err)

	// Wait for version message
	select {
	case msg := <-receivedMsg:
		var versionMsg struct {
			Msgtype  string `json:"msgtype"`
			Station  string `json:"station"`
			Protocol int    `json:"protocol"`
		}
		err := json.Unmarshal([]byte(msg), &versionMsg)
		assert.NoError(t, err)
		assert.Equal(t, "version", versionMsg.Msgtype)
		assert.Equal(t, "lorawan-simulator", versionMsg.Station)
		assert.Equal(t, 2, versionMsg.Protocol)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for version message")
	}
}

func TestLnsDataReadLoop_ReceivesMessages(t *testing.T) {
	server := mockDataServer(t, "success")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = wsURL

	err := gw.lnsDataConnect()
	assert.NoError(t, err)

	// Give time to receive router_config
	time.Sleep(100 * time.Millisecond)

	// The read loop should have received the router_config message
	// (we can't directly observe it, but we can verify the connection stays alive)
	info := gw.GetInfo()
	assert.Equal(t, "connected", info.DataState)
}

func TestLnsDataReadLoop_HandlesReadError(t *testing.T) {
	server := mockDataServer(t, "read_error")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = wsURL
	gw.dataDone = make(chan struct{})

	err := gw.lnsDataConnect()
	assert.NoError(t, err)

	// Wait for read loop to exit (dataDone should close)
	select {
	case <-gw.dataDone:
		// Expected behavior - read loop closed the channel
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for read loop to exit")
	}
}

func TestLnsDataWriteLoop_SendsMessages(t *testing.T) {
	messagesReceived := make(chan string, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Collect all messages
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			messagesReceived <- string(msg)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = wsURL

	err := gw.lnsDataConnect()
	assert.NoError(t, err)

	// Send additional test messages
	testMsg1 := `{"msgtype":"test1"}`
	testMsg2 := `{"msgtype":"test2"}`

	gw.lnsDataSend(testMsg1)
	gw.lnsDataSend(testMsg2)

	// Collect messages
	var messages []string
	timeout := time.After(1 * time.Second)

	for i := 0; i < 3; i++ { // version + 2 test messages
		select {
		case msg := <-messagesReceived:
			messages = append(messages, msg)
		case <-timeout:
			t.Fatal("Timeout waiting for messages")
		}
	}

	// Verify all messages were sent
	assert.Contains(t, messages, `{"msgtype":"version","station":"lorawan-simulator","protocol":2}`)
	assert.Contains(t, messages, testMsg1)
	assert.Contains(t, messages, testMsg2)
}

func TestLnsDataWriteLoop_HandlesWriteError(t *testing.T) {
	server := mockDataServer(t, "write_error")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = wsURL

	err := gw.lnsDataConnect()
	assert.NoError(t, err)

	// Give time for connection to be established and then closed by server
	time.Sleep(100 * time.Millisecond)

	// Try to send a message after connection is closed
	// This should cause a write error (logged but not panicking)
	gw.lnsDataSend(`{"msgtype":"test"}`)

	// Give time for write to be attempted
	time.Sleep(50 * time.Millisecond)

	// The write loop should have exited gracefully (no panic)
}

func TestLnsDataSend(t *testing.T) {
	server := mockDataServer(t, "echo")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = wsURL

	err := gw.lnsDataConnect()
	assert.NoError(t, err)

	// Give time for connection to establish
	time.Sleep(50 * time.Millisecond)

	// Send a test message
	testMsg := `{"msgtype":"uplink","data":"test123"}`
	gw.lnsDataSend(testMsg)

	// Give time for message to be sent and echoed
	time.Sleep(50 * time.Millisecond)

	// If we got here without blocking or panicking, the send worked
}

func TestLnsDataConnect_ConcurrentSends(t *testing.T) {
	messagesReceived := make(chan string, 100)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Collect all messages
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			messagesReceived <- string(msg)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = wsURL

	err := gw.lnsDataConnect()
	assert.NoError(t, err)

	// Give time for connection to establish
	time.Sleep(50 * time.Millisecond)

	// Send messages concurrently
	const numGoroutines = 10
	const messagesPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := fmt.Sprintf(`{"msgtype":"test","id":%d,"seq":%d}`, id, j)
				gw.lnsDataSend(msg)
			}
		}(i)
	}

	wg.Wait()

	// Collect messages
	timeout := time.After(2 * time.Second)
	messageCount := 0
	expectedMessages := numGoroutines*messagesPerGoroutine + 1 // +1 for version message

	for messageCount < expectedMessages {
		select {
		case <-messagesReceived:
			messageCount++
		case <-timeout:
			t.Fatalf("Timeout: received %d messages, expected %d", messageCount, expectedMessages)
		}
	}

	assert.Equal(t, expectedMessages, messageCount, "Should receive all concurrent messages")
}

func TestLnsDataConnect_ImmediateClose(t *testing.T) {
	server := mockDataServer(t, "immediate_close")
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
	gw := New(eui, "ws://discovery.test")
	gw.dataURI = wsURL
	gw.dataDone = make(chan struct{})

	err := gw.lnsDataConnect()
	assert.NoError(t, err)

	// Wait for read loop to detect closure
	select {
	case <-gw.dataDone:
		// Expected - read loop should exit when connection closes
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for read loop to exit after immediate close")
	}
}
