package device

import (
	"sync"

	"github.com/brocaar/lorawan"
)

type Device struct {
	DevEUI  lorawan.EUI64
	JoinEUI lorawan.EUI64

	DevAddr lorawan.DevAddr
	AppKey  lorawan.AES128Key
	AppSKey lorawan.AES128Key
	NwkSKey lorawan.AES128Key

	FCntUp uint32
	FCntDn uint32

	mu sync.RWMutex
}

type DeviceInfo struct {
	DevEUI lorawan.EUI64 `json:"deveui"`
}

func New(EUI lorawan.EUI64) *Device {
	return &Device{
		DevEUI: EUI,
	}
}

func (d *Device) GetInfo() DeviceInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return DeviceInfo{
		DevEUI: d.DevEUI,
	}
}
