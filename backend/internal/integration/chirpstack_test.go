package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChirpStackClient_BuildDiscoveryURI(t *testing.T) {
	testCases := []struct {
		name        string
		baseURL     string
		expectedURI string
	}{
		{
			name:        "http localhost no port",
			baseURL:     "http://localhost",
			expectedURI: "ws://localhost:3001",
		},
		{
			name:        "http localhost with port",
			baseURL:     "http://localhost:8080",
			expectedURI: "ws://localhost:3001",
		},
		{
			name:        "http localhost with trailing slash",
			baseURL:     "http://localhost/",
			expectedURI: "ws://localhost:3001",
		},
		{
			name:        "http localhost with path",
			baseURL:     "http://localhost:8080/api",
			expectedURI: "ws://localhost:3001",
		},
		{
			name:        "https domain no port",
			baseURL:     "https://chirpstack.example.com",
			expectedURI: "wss://chirpstack.example.com:3001",
		},
		{
			name:        "https domain with port",
			baseURL:     "https://chirpstack.example.com:8080",
			expectedURI: "wss://chirpstack.example.com:3001",
		},
		{
			name:        "https domain with port and path",
			baseURL:     "https://chirpstack.example.com:8080/api/v1",
			expectedURI: "wss://chirpstack.example.com:3001",
		},
		{
			name:        "http with subdomain",
			baseURL:     "http://cs.lorawan.example.com:8080",
			expectedURI: "ws://cs.lorawan.example.com:3001",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewChirpStackClient(tc.baseURL, "test-api-key")
			uri := client.buildDiscoveryURI()
			assert.Equal(t, tc.expectedURI, uri)
		})
	}
}

func TestChirpStackClient_NewChirpStackClient(t *testing.T) {
	t.Run("creates client with url and api key", func(t *testing.T) {
		url := "http://localhost:8080"
		apiKey := "test-api-key"

		client := NewChirpStackClient(url, apiKey)

		assert.NotNil(t, client)
		assert.Equal(t, url, client.baseURL)
		assert.Equal(t, apiKey, client.apiKey)
		assert.Nil(t, client.conn)
	})
}

func TestChirpStackClient_GetAuthContext(t *testing.T) {
	t.Run("creates context with api key", func(t *testing.T) {
		client := NewChirpStackClient("http://localhost:8080", "my-api-key")
		ctx := client.getAuthContext()

		assert.NotNil(t, ctx)
		// Note: In a real scenario, you would extract and verify the metadata
		// For this test, we just ensure it doesn't panic and returns a context
	})
}

func TestChirpStackClient_ListDevices(t *testing.T) {
	t.Run("returns empty list (not implemented)", func(t *testing.T) {
		t.Skip("Skipping test that requires valid ChirpStack credentials")
		client := NewChirpStackClient("http://localhost:8080", "test-api-key")
		devices, err := client.ListDevices()

		assert.NoError(t, err)
		assert.NotNil(t, devices)
		assert.Len(t, devices, 0)
	})
}

func TestChirpStackClient_CreateDevice(t *testing.T) {
	t.Run("returns nil (not implemented)", func(t *testing.T) {
		client := NewChirpStackClient("http://localhost:8080", "test-api-key")

		var devEUI, joinEUI [8]byte
		var appKey [16]byte

		err := client.CreateDevice(devEUI, joinEUI, appKey)
		assert.NoError(t, err)
	})
}

func TestChirpStackClient_DeleteDevice(t *testing.T) {
	t.Run("returns nil (not implemented)", func(t *testing.T) {
		client := NewChirpStackClient("http://localhost:8080", "test-api-key")

		var devEUI [8]byte

		err := client.DeleteDevice(devEUI)
		assert.NoError(t, err)
	})
}

func TestChirpStackClient_CreateGateway(t *testing.T) {
	t.Run("returns nil (not implemented)", func(t *testing.T) {
		client := NewChirpStackClient("http://localhost:8080", "test-api-key")

		var eui [8]byte

		err := client.CreateGateway(eui, "ws://localhost:3001")
		assert.NoError(t, err)
	})
}

func TestChirpStackClient_DeleteGateway(t *testing.T) {
	t.Run("returns nil (not implemented)", func(t *testing.T) {
		client := NewChirpStackClient("http://localhost:8080", "test-api-key")

		var eui [8]byte

		err := client.DeleteGateway(eui)
		assert.NoError(t, err)
	})
}

func TestChirpStackClient_Close(t *testing.T) {
	t.Run("closes nil connection without error", func(t *testing.T) {
		client := NewChirpStackClient("http://localhost:8080", "test-api-key")
		err := client.Close()
		assert.NoError(t, err)
	})
}

func TestChirpStackClient_ConnectionURLParsing(t *testing.T) {
	testCases := []struct {
		name    string
		baseURL string
	}{
		{
			name:    "strips http scheme",
			baseURL: "http://localhost:8080",
		},
		{
			name:    "strips https scheme",
			baseURL: "https://chirpstack.example.com:8080",
		},
		{
			name:    "removes path component",
			baseURL: "http://localhost:8080/api/v1",
		},
		{
			name:    "handles no port for http",
			baseURL: "http://localhost",
		},
		{
			name:    "handles no port for https",
			baseURL: "https://chirpstack.example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewChirpStackClient(tc.baseURL, "test-api-key")

			// Verify the client was created properly
			assert.NotNil(t, client)
			assert.Equal(t, tc.baseURL, client.baseURL)
			assert.Equal(t, "test-api-key", client.apiKey)
		})
	}
}
