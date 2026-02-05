package integration

import (
	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"
)

type NetworkServerType string

const (
	NetworkServerTypeGeneric    NetworkServerType = "generic"
	NetworkServerTypeLORIOT     NetworkServerType = "loriot"
	NetworkServerTypeChirpStack NetworkServerType = "chirpstack"
	NetworkServerTypeTTN        NetworkServerType = "ttn"
)

type NetworkServerConfig struct {
	Type       NetworkServerType `json:"type"`
	URL        string            `json:"url,omitempty"`
	AuthHeader string            `json:"authHeader,omitempty"` // LORIOT
	APIToken   string            `json:"apiToken,omitempty"`   // ChirpStack
	APIKey     string            `json:"apiKey,omitempty"`     // TTN
}

// IntegrationClient defines the interface for network server integrations
type IntegrationClient interface {
	// Gateway operations
	ListGateways() ([]gateway.GatewayInfo, error)
	CreateGateway(eui lorawan.EUI64, discoveryURI string) error
	DeleteGateway(eui lorawan.EUI64) error

	// Device operations
	ListDevices() ([]device.DeviceInfo, error)
	CreateDevice(devEUI lorawan.EUI64, joinEUI lorawan.EUI64, appKey lorawan.AES128Key) error
	DeleteDevice(devEUI lorawan.EUI64) error
}

// NewIntegrationClient creates the appropriate integration client based on config
func NewIntegrationClient(config NetworkServerConfig) (IntegrationClient, error) {
	switch config.Type {
	case NetworkServerTypeGeneric:
		return &GenericClient{}, nil

	case NetworkServerTypeLORIOT:
		if config.URL == "" || config.AuthHeader == "" {
			return &GenericClient{}, nil // Fallback to generic if config is invalid
		}
		return NewLORIOTClient(config.URL, config.AuthHeader), nil

	case NetworkServerTypeChirpStack:
		if config.URL == "" || config.APIToken == "" {
			return &GenericClient{}, nil
		}
		return NewChirpStackClient(config.URL, config.APIToken), nil

	case NetworkServerTypeTTN:
		if config.URL == "" || config.APIKey == "" {
			return &GenericClient{}, nil
		}
		return NewTTNClient(config.URL, config.APIKey), nil

	default:
		return &GenericClient{}, nil
	}
}
