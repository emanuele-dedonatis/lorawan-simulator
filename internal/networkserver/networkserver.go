package networkserver

import (
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"

	"github.com/brocaar/lorawan"
)

type NetworkServer struct {
	name string

	devices  map[lorawan.EUI64]*device.Device
	gateways map[lorawan.EUI64]*gateway.Gateway
}

type NetworkServerInfo struct {
	Name         string
	DeviceCount  int
	GatewayCount int
}

func New(name string) *NetworkServer {
	return &NetworkServer{
		name:     name,
		devices:  make(map[lorawan.EUI64]*device.Device),
		gateways: make(map[lorawan.EUI64]*gateway.Gateway),
	}
}

func (ns *NetworkServer) GetInfo() NetworkServerInfo {
	return NetworkServerInfo{
		Name:         ns.name,
		DeviceCount:  len(ns.devices),
		GatewayCount: len(ns.gateways),
	}
}
