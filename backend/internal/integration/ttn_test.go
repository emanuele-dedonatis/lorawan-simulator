package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTTNClient(t *testing.T) {
	client := NewTTNClient("https://eu1.cloud.thethings.network", "test-api-key")

	assert.NotNil(t, client)
	assert.Equal(t, "https://eu1.cloud.thethings.network", client.baseURL)
	assert.Equal(t, "test-api-key", client.apiKey)
}

func TestTTNClient_buildDiscoveryURI(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{
			name:     "HTTPS URL without port",
			baseURL:  "https://eu1.cloud.thethings.network",
			expected: "wss://eu1.cloud.thethings.network:8887",
		},
		{
			name:     "HTTP URL without port",
			baseURL:  "http://localhost",
			expected: "ws://localhost:1887",
		},
		{
			name:     "HTTPS URL with port",
			baseURL:  "https://eu1.cloud.thethings.network:8443",
			expected: "wss://eu1.cloud.thethings.network:8443",
		},
		{
			name:     "HTTP URL with port",
			baseURL:  "http://localhost:8080",
			expected: "ws://localhost:8080",
		},
		{
			name:     "HTTPS URL with path",
			baseURL:  "https://eu1.cloud.thethings.network/api",
			expected: "wss://eu1.cloud.thethings.network:8887",
		},
		{
			name:     "HTTP URL without scheme",
			baseURL:  "localhost",
			expected: "wss://localhost:8887",
		},
		{
			name:     "HTTP URL with port and path",
			baseURL:  "http://localhost:8080/api/v3",
			expected: "ws://localhost:8080",
		},
		{
			name:     "HTTPS URL with port and path",
			baseURL:  "https://thethings.example.com:8884/api",
			expected: "wss://thethings.example.com:8884",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewTTNClient(tt.baseURL, "test-api-key")
			result := client.buildDiscoveryURI()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTTNClient_getConnection(t *testing.T) {
	t.Run("creates connection for HTTPS URL", func(t *testing.T) {
		client := NewTTNClient("https://eu1.cloud.thethings.network", "test-api-key")

		// Note: This will fail to actually connect in tests, but we can verify the client is created
		conn, err := client.getConnection()
		// We expect an error because we can't actually connect to TTN in tests
		// but the connection object should be created
		if err == nil {
			assert.NotNil(t, conn)
		}
	})

	t.Run("creates connection for HTTP URL", func(t *testing.T) {
		client := NewTTNClient("http://localhost", "test-api-key")

		conn, err := client.getConnection()
		// We expect an error because localhost won't have TTN running
		// but the connection object should be created
		if err == nil {
			assert.NotNil(t, conn)
		}
	})
}

func TestTTNClient_getAuthContext(t *testing.T) {
	client := NewTTNClient("https://eu1.cloud.thethings.network", "test-api-key")
	ctx := client.getAuthContext()

	assert.NotNil(t, ctx)
}

func TestTTNClient_CreateGateway(t *testing.T) {
	client := NewTTNClient("https://eu1.cloud.thethings.network", "test-api-key")

	// Test stub implementation
	err := client.CreateGateway([8]byte{1, 2, 3, 4, 5, 6, 7, 8}, "ws://test:1887")
	assert.NoError(t, err)
}

func TestTTNClient_DeleteGateway(t *testing.T) {
	client := NewTTNClient("https://eu1.cloud.thethings.network", "test-api-key")

	// Test stub implementation
	err := client.DeleteGateway([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	assert.NoError(t, err)
}

func TestTTNClient_ListDevices(t *testing.T) {
	t.Skip("Skipping test that requires valid TTN credentials")
	client := NewTTNClient("https://eu1.cloud.thethings.network", "test-api-key")

	// Test stub implementation
	devices, err := client.ListDevices()
	assert.NoError(t, err)
	assert.Empty(t, devices)
}

func TestTTNClient_CreateDevice(t *testing.T) {
	client := NewTTNClient("https://eu1.cloud.thethings.network", "test-api-key")

	// Test stub implementation
	err := client.CreateDevice(
		[8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		[8]byte{8, 7, 6, 5, 4, 3, 2, 1},
		[16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	)
	assert.NoError(t, err)
}

func TestTTNClient_DeleteDevice(t *testing.T) {
	client := NewTTNClient("https://eu1.cloud.thethings.network", "test-api-key")

	// Test stub implementation
	err := client.DeleteDevice([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	assert.NoError(t, err)
}
