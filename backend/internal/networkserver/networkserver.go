package networkserver

import (
	"errors"
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/integration"

	"github.com/brocaar/lorawan"
)

type NetworkServer struct {
	name              string
	config            integration.NetworkServerConfig
	integrationClient integration.IntegrationClient
	devices           map[lorawan.EUI64]*device.Device
	gateways          map[lorawan.EUI64]*gateway.Gateway
	mu                sync.RWMutex
	broadcastUplink   chan<- lorawan.PHYPayload
	broadcastDownlink chan<- lorawan.PHYPayload
}

type NetworkServerInfo struct {
	Name         string                          `json:"name"`
	Config       integration.NetworkServerConfig `json:"config"`
	DeviceCount  int                             `json:"deviceCount"`
	GatewayCount int                             `json:"gatewayCount"`
}

func New(name string, config integration.NetworkServerConfig, broadcastUplink chan<- lorawan.PHYPayload, broadcastDownlink chan<- lorawan.PHYPayload) *NetworkServer {
	integrationClient, err := integration.NewIntegrationClient(config)
	if err != nil {
		return nil
	}

	return &NetworkServer{
		name:              name,
		config:            config,
		integrationClient: integrationClient,
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
		Config:       ns.config,
		DeviceCount:  len(ns.devices),
		GatewayCount: len(ns.gateways),
	}
}

// Gateway management methods

func (ns *NetworkServer) AddGateway(EUI lorawan.EUI64, discoveryURI string, headers ...http.Header) (*gateway.Gateway, error) {
	return ns.AddGatewayWithLocation(EUI, discoveryURI, nil, headers...)
}

func (ns *NetworkServer) AddGatewayWithLocation(EUI lorawan.EUI64, discoveryURI string, location *gateway.Location, headers ...http.Header) (*gateway.Gateway, error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, exists := ns.gateways[EUI]; exists {
		return nil, errors.New("gateway already exists")
	}

	// Create gateway with optional headers (nil if not provided)
	var h http.Header
	if len(headers) > 0 {
		h = headers[0]
	}

	ns.gateways[EUI] = gateway.NewWithLocation(ns.broadcastDownlink, EUI, discoveryURI, h, location)
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

	// Sort alphabetically by EUI
	sort.Slice(gatewayInfos, func(i, j int) bool {
		return gatewayInfos[i].EUI.String() < gatewayInfos[j].EUI.String()
	})

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

func (ns *NetworkServer) AddDevice(
	DevEUI lorawan.EUI64,
	JoinEUI lorawan.EUI64,
	AppKey lorawan.AES128Key,
	DevNonce lorawan.DevNonce,
	DevAddr lorawan.DevAddr,
	AppSKey lorawan.AES128Key,
	NwkSKey lorawan.AES128Key,
	FCntUp uint32,
	FCntDn uint32,
) (*device.Device, error) {
	return ns.AddDeviceWithLocation(DevEUI, JoinEUI, AppKey, DevNonce, DevAddr, AppSKey, NwkSKey, FCntUp, FCntDn, nil)
}

func (ns *NetworkServer) AddDeviceWithLocation(
	DevEUI lorawan.EUI64,
	JoinEUI lorawan.EUI64,
	AppKey lorawan.AES128Key,
	DevNonce lorawan.DevNonce,
	DevAddr lorawan.DevAddr,
	AppSKey lorawan.AES128Key,
	NwkSKey lorawan.AES128Key,
	FCntUp uint32,
	FCntDn uint32,
	location *device.Location,
) (*device.Device, error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, exists := ns.devices[DevEUI]; exists {
		return nil, errors.New("device already exists")
	}

	ns.devices[DevEUI] = device.NewWithLocation(ns.broadcastUplink, DevEUI, JoinEUI, AppKey, DevNonce, DevAddr, AppSKey, NwkSKey, FCntUp, FCntDn, location)
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

	// Sort alphabetically by DevEUI
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].DevEUI.String() < devices[j].DevEUI.String()
	})

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

// Sync syncs gateways and devices from the remote network server
func (ns *NetworkServer) Sync() error {

	// Sync gateways
	nsGws, err := ns.integrationClient.ListGateways()
	if err != nil {
		return err
	}

	// Collect gateways to remove and to add
	ns.mu.RLock()
	var gwsToRemove, gwsToAdd []gateway.GatewayInfo
	for _, nsGw := range nsGws {
		gw, exists := ns.gateways[nsGw.EUI]
		if exists {
			// Gateway already exists
			if gw.GetInfo().DiscoveryURI == nsGw.DiscoveryURI {
				// Nothing to update
				log.Printf("[%s] gateway %s already exists", ns.name, nsGw.EUI)
				continue
			} else {
				// Remove current gateway
				log.Printf("[%s] gateway %s exists but with different discovery URI", ns.name, nsGw.EUI)
				gwsToRemove = append(gwsToRemove, nsGw)
			}
		}

		log.Printf("[%s] new gateway %s", ns.name, nsGw.EUI)
		gwsToAdd = append(gwsToAdd, nsGw)
	}
	ns.mu.RUnlock()

	// Remove not existing gateways
	for _, gw := range gwsToRemove {
		err := ns.RemoveGateway(gw.EUI)
		if err != nil {
			log.Printf("[%s] unable to remove gateway %s: %v", ns.name, gw.EUI, err)
		}
	}

	// Add new gateways
	for _, gw := range gwsToAdd {
		_, err := ns.AddGatewayWithLocation(gw.EUI, gw.DiscoveryURI, gw.Location, gw.Headers)
		if err != nil {
			log.Printf("[%s] unable to add gateway %s: %v", ns.name, gw.EUI, err)
		}
	}

	// Sync devices
	nsDevs, err := ns.integrationClient.ListDevices()
	if err != nil {
		return err
	}

	// Collect devices to remove and to add
	ns.mu.RLock()
	var devsToRemove, devsToAdd []device.DeviceInfo
	for _, nsDev := range nsDevs {
		_, exists := ns.devices[nsDev.DevEUI]
		if exists {
			// Device already exists - for simplicity, we skip checking if it changed
			// because that would require calling GetInfo() which could cause a deadlock
			log.Printf("[%s] device %s already exists", ns.name, nsDev.DevEUI)
			continue
		}

		log.Printf("[%s] new device %s", ns.name, nsDev.DevEUI)
		devsToAdd = append(devsToAdd, nsDev)
	}
	ns.mu.RUnlock()

	// Remove not existing devices
	for _, dev := range devsToRemove {
		err := ns.RemoveDevice(dev.DevEUI)
		if err != nil {
			log.Printf("[%s] unable to remove device %s: %v", ns.name, dev.DevEUI, err)
		}
	}

	// Add new devices
	for _, dev := range devsToAdd {
		_, err := ns.AddDeviceWithLocation(dev.DevEUI, dev.JoinEUI, dev.AppKey, dev.DevNonce, dev.DevAddr, dev.AppSKey, dev.NwkSKey, dev.FCntUp, dev.FCntDn, dev.Location)
		if err != nil {
			log.Printf("[%s] unable to add device %s: %v", ns.name, dev.DevEUI, err)
		}
	}

	return nil
}
