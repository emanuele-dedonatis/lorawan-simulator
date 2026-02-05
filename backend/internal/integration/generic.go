package integration

import (
	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"
)

// Generic Network Servers (no integration)
type GenericClient struct{}

func NewGenericClient() *GenericClient {
	return &GenericClient{}
}

func (c *GenericClient) ListGateways() ([]gateway.GatewayInfo, error) {
	return []gateway.GatewayInfo{}, nil
}

func (c *GenericClient) CreateGateway(eui lorawan.EUI64, discoveryURI string) error {
	return nil
}

func (c *GenericClient) DeleteGateway(eui lorawan.EUI64) error {
	return nil
}

func (c *GenericClient) ListDevices() ([]device.DeviceInfo, error) {
	return []device.DeviceInfo{}, nil
}

func (c *GenericClient) CreateDevice(devEUI lorawan.EUI64, joinEUI lorawan.EUI64, appKey lorawan.AES128Key) error {
	return nil
}

func (c *GenericClient) DeleteDevice(devEUI lorawan.EUI64) error {
	return nil
}
