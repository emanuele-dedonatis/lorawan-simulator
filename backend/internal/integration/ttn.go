package integration

import (
	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"
)

type TTNClient struct {
	baseURL string
	apiKey  string
}

func NewTTNClient(url, apiKey string) *TTNClient {
	return &TTNClient{
		baseURL: url,
		apiKey:  apiKey,
	}
}

func (c *TTNClient) ListGateways() ([]gateway.GatewayInfo, error) {
	// TODO
	return []gateway.GatewayInfo{}, nil
}

func (c *TTNClient) CreateGateway(eui lorawan.EUI64, discoveryURI string) error {
	// TODO
	return nil
}

func (c *TTNClient) DeleteGateway(eui lorawan.EUI64) error {
	// TODO
	return nil
}

func (c *TTNClient) ListDevices() ([]device.DeviceInfo, error) {
	// TODO
	return []device.DeviceInfo{}, nil
}

func (c *TTNClient) CreateDevice(devEUI lorawan.EUI64, joinEUI lorawan.EUI64, appKey lorawan.AES128Key) error {
	// TODO
	return nil
}

func (c *TTNClient) DeleteDevice(devEUI lorawan.EUI64) error {
	// TODO
	return nil
}
