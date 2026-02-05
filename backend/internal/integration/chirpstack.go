package integration

import (
	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"
)

type ChirpStackClient struct {
	baseURL  string
	apiToken string
}

func NewChirpStackClient(url, apiToken string) *ChirpStackClient {
	return &ChirpStackClient{
		baseURL:  url,
		apiToken: apiToken,
	}
}

func (c *ChirpStackClient) ListGateways() ([]gateway.GatewayInfo, error) {
	// TODO
	return []gateway.GatewayInfo{}, nil
}

func (c *ChirpStackClient) CreateGateway(eui lorawan.EUI64, discoveryURI string) error {
	// TODO
	return nil
}

func (c *ChirpStackClient) DeleteGateway(eui lorawan.EUI64) error {
	// TODO
	return nil
}

func (c *ChirpStackClient) ListDevices() ([]device.DeviceInfo, error) {
	// TODO
	return []device.DeviceInfo{}, nil
}

func (c *ChirpStackClient) CreateDevice(devEUI lorawan.EUI64, joinEUI lorawan.EUI64, appKey lorawan.AES128Key) error {
	// TODO
	return nil
}

func (c *ChirpStackClient) DeleteDevice(devEUI lorawan.EUI64) error {
	// TODO
	return nil
}
