package networkserver

import (
	"errors"
	"log"
	"sync"

	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"

	"github.com/brocaar/lorawan"
)

type NetworkServer struct {
	name              string
	devices           map[lorawan.EUI64]*device.Device
	gateways          map[lorawan.EUI64]*gateway.Gateway
	mu                sync.RWMutex
	broadcastUplink   chan<- lorawan.PHYPayload
	broadcastDownlink chan<- lorawan.PHYPayload
}

type NetworkServerInfo struct {
	Name         string `json:"name"`
	DeviceCount  int    `json:"deviceCount"`
	GatewayCount int    `json:"gatewayCount"`
}

func New(name string, broadcastUplink chan<- lorawan.PHYPayload, broadcastDownlink chan<- lorawan.PHYPayload) *NetworkServer {
	return &NetworkServer{
		name:              name,
		devices:           make(map[lorawan.EUI64]*device.Device),
		gateways:          make(map[lorawan.EUI64]*gateway.Gateway),
		broadcastUplink:   broadcastUplink,
		broadcastDownlink: broadcastDownlink,
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

	ns.gateways[EUI] = gateway.New(ns.broadcastDownlink, EUI, discoveryURI)
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

func (ns *NetworkServer) ListGateways() []gateway.GatewayInfo {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	gatewayInfos := make([]gateway.GatewayInfo, 0, len(ns.gateways))
	for _, gateway := range ns.gateways {
		gatewayInfos = append(gatewayInfos, gateway.GetInfo())
	}
	return gatewayInfos
}

func (ns *NetworkServer) RemoveGateway(EUI lorawan.EUI64) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, exists := ns.gateways[EUI]; !exists {
		return errors.New("gateway not found")
	}

	// TODO: disconnect gateway

	delete(ns.gateways, EUI)

	return nil
}

func (ns *NetworkServer) ForwardUplink(uplink lorawan.PHYPayload) error {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	// TODO: filter by location
	for _, gw := range ns.gateways {
		log.Printf("[%s] propagating uplink to gateway %s", ns.name, gw.GetInfo().EUI)
		go func(gw *gateway.Gateway) {
			err := gw.Forward(uplink)
			if err != nil {
				log.Printf("[%s] gateway %s error: %v", ns.name, gw.GetInfo().EUI, err)
			}
		}(gw)
	}

	return nil
}

// Device management methods

func (ns *NetworkServer) AddDevice(DevEUI lorawan.EUI64, JoinEUI lorawan.EUI64, AppKey lorawan.AES128Key, DevNonce lorawan.DevNonce) (*device.Device, error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, exists := ns.devices[DevEUI]; exists {
		return nil, errors.New("device already exists")
	}

	ns.devices[DevEUI] = device.New(ns.broadcastUplink, DevEUI, JoinEUI, AppKey, DevNonce)
	return ns.devices[DevEUI], nil
}

func (ns *NetworkServer) GetDevice(DevEUI lorawan.EUI64) (*device.Device, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	device, exists := ns.devices[DevEUI]
	if !exists {
		return nil, errors.New("device not found")
	}

	return device, nil
}

func (ns *NetworkServer) ListDevices() []device.DeviceInfo {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	devices := make([]device.DeviceInfo, 0, len(ns.devices))
	for _, device := range ns.devices {
		devices = append(devices, device.GetInfo())
	}
	return devices
}

func (ns *NetworkServer) RemoveDevice(DevEUI lorawan.EUI64) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, exists := ns.devices[DevEUI]; !exists {
		return errors.New("device not found")
	}

	delete(ns.devices, DevEUI)

	return nil
}

func (ns *NetworkServer) ForwardDownlink(downlink lorawan.PHYPayload) error {
	if downlink.MHDR.MType == lorawan.UnconfirmedDataDown || downlink.MHDR.MType == lorawan.ConfirmedDataDown {
		// Unconfirmed or Confirmed Downlink
		macPL, ok := downlink.MACPayload.(*lorawan.MACPayload)
		if !ok {
			log.Printf("[%s] invalid MAC payload for data downlink", ns.name)
			return errors.New("invalid MAC payload")
		}

		devAddr := macPL.FHDR.DevAddr

		ns.mu.RLock()
		defer ns.mu.RUnlock()
		for _, dev := range ns.devices {
			// TODO: filter also by location and rxw
			if dev.DevAddr == devAddr {
				// Propagate only to devices with same DevAddr
				log.Printf("[%s] propagating downlink to device %s (DevAddr: %s)", ns.name, dev.GetInfo().DevEUI, devAddr)
				go func(dev *device.Device) {
					err := dev.Downlink(downlink)
					if err != nil {
						log.Printf("[%s] device %s error: %v", ns.name, dev.GetInfo().DevEUI, err)
					}
				}(dev)
			}
		}
	} else {
		// Join Accept
		ns.mu.RLock()
		defer ns.mu.RUnlock()
		for _, dev := range ns.devices {
			// TODO: filter also by location and rxw
			log.Printf("[%s] propagating downlink to device %s", ns.name, dev.GetInfo().DevEUI)
			go func(dev *device.Device) {
				err := dev.JoinAccept(downlink)
				if err != nil {
					log.Printf("[%s] device %s error: %v", ns.name, dev.GetInfo().DevEUI, err)
				}
			}(dev)
		}
	}

	return nil
}

func (ns *NetworkServer) SendJoinRequest(DevEUI lorawan.EUI64) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	device, exists := ns.devices[DevEUI]
	if !exists {
		return errors.New("device not found")
	}

	// Prepare JoinRequest frame
	_, err := device.JoinRequest()
	if err != nil {
		return err
	}

	return nil
}

func (ns *NetworkServer) SendUplink(DevEUI lorawan.EUI64) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	device, exists := ns.devices[DevEUI]
	if !exists {
		return errors.New("device not found")
	}

	// Prepare Uplink frame
	_, err := device.Uplink()
	if err != nil {
		return err
	}

	return nil
}
