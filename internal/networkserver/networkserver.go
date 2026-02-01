package networkserver

import (
	"errors"
	"sync"

	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"

	"github.com/brocaar/lorawan"
)

type NetworkServer struct {
	name     string
	devices  map[lorawan.EUI64]*device.Device
	gateways map[lorawan.EUI64]*gateway.Gateway
	mu       sync.RWMutex
}

type NetworkServerInfo struct {
	Name         string `json:"name"`
	DeviceCount  int    `json:"deviceCount"`
	GatewayCount int    `json:"gatewayCount"`
}

func New(name string) *NetworkServer {
	return &NetworkServer{
		name:     name,
		devices:  make(map[lorawan.EUI64]*device.Device),
		gateways: make(map[lorawan.EUI64]*gateway.Gateway),
	}
}

func (ns *NetworkServer) GetInfo() NetworkServerInfo {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	return NetworkServerInfo{
		Name:         ns.name,
		DeviceCount:  len(ns.devices),
		GatewayCount: len(ns.gateways),
	}
}

// Gateway management methods

func (ns *NetworkServer) AddGateway(EUI lorawan.EUI64, discoveryURI string) (*gateway.Gateway, error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, exists := ns.gateways[EUI]; exists {
		return nil, errors.New("gateway already exists")
	}

	ns.gateways[EUI] = gateway.New(EUI, discoveryURI)
	return ns.gateways[EUI], nil
}

func (ns *NetworkServer) GetGateway(EUI lorawan.EUI64) (*gateway.Gateway, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	gateway, exists := ns.gateways[EUI]
	if !exists {
		return nil, errors.New("gateway not found")
	}

	return gateway, nil
}

func (ns *NetworkServer) ListGateways() []*gateway.Gateway {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	gateways := make([]*gateway.Gateway, 0, len(ns.gateways))
	for _, gateway := range ns.gateways {
		gateways = append(gateways, gateway)
	}
	return gateways
}

func (ns *NetworkServer) RemoveGateway(EUI lorawan.EUI64) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, exists := ns.gateways[EUI]; !exists {
		return errors.New("gateway not found")
	}

	delete(ns.gateways, EUI)

	return nil
}

// Device management methods

// TODO
