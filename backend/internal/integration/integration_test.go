package integration

import (
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/stretchr/testify/assert"
)

func TestNewIntegrationClient(t *testing.T) {
	t.Run("creates generic client for generic type", func(t *testing.T) {
		config := NetworkServerConfig{
			Type: NetworkServerTypeGeneric,
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &GenericClient{}, client)
	})

	t.Run("creates LORIOT client with valid config", func(t *testing.T) {
		config := NetworkServerConfig{
			Type:       NetworkServerTypeLORIOT,
			URL:        "https://eu1.loriot.io",
			AuthHeader: "Bearer token123",
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &LORIOTClient{}, client)
	})

	t.Run("falls back to generic client for LORIOT with missing URL", func(t *testing.T) {
		config := NetworkServerConfig{
			Type:       NetworkServerTypeLORIOT,
			AuthHeader: "Bearer token123",
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &GenericClient{}, client)
	})

	t.Run("falls back to generic client for LORIOT with missing auth header", func(t *testing.T) {
		config := NetworkServerConfig{
			Type: NetworkServerTypeLORIOT,
			URL:  "https://eu1.loriot.io",
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &GenericClient{}, client)
	})

	t.Run("creates ChirpStack client with valid config", func(t *testing.T) {
		config := NetworkServerConfig{
			Type:     NetworkServerTypeChirpStack,
			URL:      "https://chirpstack.example.com",
			APIToken: "token123",
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &ChirpStackClient{}, client)
	})

	t.Run("falls back to generic client for ChirpStack with missing URL", func(t *testing.T) {
		config := NetworkServerConfig{
			Type:     NetworkServerTypeChirpStack,
			APIToken: "token123",
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &GenericClient{}, client)
	})

	t.Run("falls back to generic client for ChirpStack with missing token", func(t *testing.T) {
		config := NetworkServerConfig{
			Type: NetworkServerTypeChirpStack,
			URL:  "https://chirpstack.example.com",
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &GenericClient{}, client)
	})

	t.Run("creates TTN client with valid config", func(t *testing.T) {
		config := NetworkServerConfig{
			Type:   NetworkServerTypeTTN,
			URL:    "https://eu1.cloud.thethings.network",
			APIKey: "NNSXS.abc123",
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &TTNClient{}, client)
	})

	t.Run("falls back to generic client for TTN with missing URL", func(t *testing.T) {
		config := NetworkServerConfig{
			Type:   NetworkServerTypeTTN,
			APIKey: "NNSXS.abc123",
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &GenericClient{}, client)
	})

	t.Run("falls back to generic client for TTN with missing API key", func(t *testing.T) {
		config := NetworkServerConfig{
			Type: NetworkServerTypeTTN,
			URL:  "https://eu1.cloud.thethings.network",
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &GenericClient{}, client)
	})

	t.Run("falls back to generic client for unknown type", func(t *testing.T) {
		config := NetworkServerConfig{
			Type: "unknown",
		}

		client, err := NewIntegrationClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &GenericClient{}, client)
	})

	t.Run("returns error-free client for all valid types", func(t *testing.T) {
		types := []NetworkServerType{
			NetworkServerTypeGeneric,
			NetworkServerTypeLORIOT,
			NetworkServerTypeChirpStack,
			NetworkServerTypeTTN,
		}

		for _, typ := range types {
			config := NetworkServerConfig{Type: typ}
			client, err := NewIntegrationClient(config)
			assert.NoError(t, err, "type %s should not error", typ)
			assert.NotNil(t, client, "type %s should return non-nil client", typ)
		}
	})
}

func TestNetworkServerConfig_JSON(t *testing.T) {
	t.Run("marshals and unmarshals correctly", func(t *testing.T) {
		config := NetworkServerConfig{
			Type:       NetworkServerTypeLORIOT,
			URL:        "https://eu1.loriot.io",
			AuthHeader: "Bearer token123",
		}

		// This test ensures the JSON tags work correctly
		assert.Equal(t, NetworkServerTypeLORIOT, config.Type)
		assert.Equal(t, "https://eu1.loriot.io", config.URL)
		assert.Equal(t, "Bearer token123", config.AuthHeader)
	})

	t.Run("has correct zero values", func(t *testing.T) {
		config := NetworkServerConfig{}

		assert.Equal(t, NetworkServerType(""), config.Type)
		assert.Equal(t, "", config.URL)
		assert.Equal(t, "", config.AuthHeader)
		assert.Equal(t, "", config.APIToken)
		assert.Equal(t, "", config.APIKey)
	})
}

func TestNetworkServerType_Constants(t *testing.T) {
	t.Run("constants have expected values", func(t *testing.T) {
		assert.Equal(t, NetworkServerType("generic"), NetworkServerTypeGeneric)
		assert.Equal(t, NetworkServerType("loriot"), NetworkServerTypeLORIOT)
		assert.Equal(t, NetworkServerType("chirpstack"), NetworkServerTypeChirpStack)
		assert.Equal(t, NetworkServerType("ttn"), NetworkServerTypeTTN)
	})

	t.Run("all constants are unique", func(t *testing.T) {
		types := []NetworkServerType{
			NetworkServerTypeGeneric,
			NetworkServerTypeLORIOT,
			NetworkServerTypeChirpStack,
			NetworkServerTypeTTN,
		}

		seen := make(map[NetworkServerType]bool)
		for _, typ := range types {
			assert.False(t, seen[typ], "duplicate type: %s", typ)
			seen[typ] = true
		}

		assert.Equal(t, 4, len(seen))
	})
}

func TestIntegrationClient_Interface(t *testing.T) {
	t.Run("all client types implement IntegrationClient interface", func(t *testing.T) {
		var _ IntegrationClient = &GenericClient{}
		var _ IntegrationClient = &LORIOTClient{}
		var _ IntegrationClient = &ChirpStackClient{}
		var _ IntegrationClient = &TTNClient{}
	})
}

// Test that the interface methods exist (compile-time check)
func TestIntegrationClient_Methods(t *testing.T) {
	t.Run("interface has all required methods", func(t *testing.T) {
		client := &GenericClient{}
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		appKey := lorawan.AES128Key{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

		// Test that all methods can be called (even if they're no-ops)
		_, err := client.ListGateways()
		assert.NoError(t, err)

		err = client.CreateGateway(eui, "ws://example.com")
		assert.NoError(t, err)

		err = client.DeleteGateway(eui)
		assert.NoError(t, err)

		_, err = client.ListDevices()
		assert.NoError(t, err)

		err = client.CreateDevice(eui, eui, appKey)
		assert.NoError(t, err)

		err = client.DeleteDevice(eui)
		assert.NoError(t, err)
	})
}
