package device

import (
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

	mu sync.RWMutex
}

type DeviceInfo struct {
	DevEUI   lorawan.EUI64     `json:"deveui"`
	JoinEUI  lorawan.EUI64     `json:"joineui"`
	AppKey   lorawan.AES128Key `json:"appkey"`
	DevNonce lorawan.DevNonce  `json:"devnonce"`
}

func New(DevEUI lorawan.EUI64, JoinEUI lorawan.EUI64, AppKey lorawan.AES128Key, DevNonce lorawan.DevNonce) *Device {
	return &Device{
		DevEUI:   DevEUI,
		JoinEUI:  JoinEUI,
		AppKey:   AppKey,
		DevNonce: DevNonce,
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

	return phy, nil
}
