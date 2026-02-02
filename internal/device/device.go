package device

import (
	"log"
	"sync"

	"github.com/brocaar/lorawan"
)

type Device struct {
	DevEUI   lorawan.EUI64
	JoinEUI  lorawan.EUI64
	AppKey   lorawan.AES128Key
	DevNonce lorawan.DevNonce

	DevAddr lorawan.DevAddr
	AppSKey lorawan.AES128Key
	NwkSKey lorawan.AES128Key

	FCntUp uint32
	FCntDn uint32

	mu              sync.RWMutex
	broadcastUplink chan<- lorawan.PHYPayload
}

type DeviceInfo struct {
	DevEUI   lorawan.EUI64     `json:"deveui"`
	JoinEUI  lorawan.EUI64     `json:"joineui"`
	AppKey   lorawan.AES128Key `json:"appkey"`
	DevNonce lorawan.DevNonce  `json:"devnonce"`
}

func New(broadcastUplink chan<- lorawan.PHYPayload, DevEUI lorawan.EUI64, JoinEUI lorawan.EUI64, AppKey lorawan.AES128Key, DevNonce lorawan.DevNonce) *Device {
	return &Device{
		broadcastUplink: broadcastUplink,
		DevEUI:          DevEUI,
		JoinEUI:         JoinEUI,
		AppKey:          AppKey,
		DevNonce:        DevNonce,
	}
}

func (d *Device) GetInfo() DeviceInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return DeviceInfo{
		DevEUI:   d.DevEUI,
		JoinEUI:  d.JoinEUI,
		AppKey:   d.AppKey,
		DevNonce: d.DevNonce,
	}
}

func (d *Device) Downlink(frame lorawan.PHYPayload) error {
	// TODO: handle received downlink
	log.Printf("[%s] received downlink", d.DevEUI)

	return nil
}

func (d *Device) JoinRequest() (lorawan.PHYPayload, error) {
	d.mu.Lock()
	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			JoinEUI:  d.JoinEUI,
			DevEUI:   d.DevEUI,
			DevNonce: d.DevNonce,
		},
	}
	// Increment DevNonce for next Join Request
	d.DevNonce = d.DevNonce + 1
	// Compute MIC
	appkey := d.AppKey
	d.mu.Unlock()
	// TODO: crypto task, use goroutine
	if err := phy.SetUplinkJoinMIC(appkey); err != nil {
		return lorawan.PHYPayload{}, err
	}

	d.broadcast(phy)

	return phy, nil
}

func (d *Device) broadcast(phy lorawan.PHYPayload) {
	// Broadcast to gateways
	d.mu.RLock()
	broadcastCh := d.broadcastUplink
	d.mu.RUnlock()

	if broadcastCh != nil {
		log.Printf("[%s] broadcasting uplink", d.DevEUI)
		go func() {
			broadcastCh <- phy
		}()
	}
}
