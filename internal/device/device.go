package device

import "github.com/brocaar/lorawan"

type Device struct {
	DevEUI  lorawan.EUI64
	JoinEUI lorawan.EUI64

	DevAddr lorawan.DevAddr
	AppKey  lorawan.AES128Key
	AppSKey lorawan.AES128Key
	NwkSKey lorawan.AES128Key

	FCntUp uint32
	FCntDn uint32
}
