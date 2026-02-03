package device

import (
	"crypto/aes"
	"encoding/binary"
	"errors"
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

	DevAddr lorawan.DevAddr   `json:"devaddr"`
	AppSKey lorawan.AES128Key `json:"appskey"`
	NwkSKey lorawan.AES128Key `json:"nwkskey"`

	FCntUp uint32 `json:"fcntup"`
	FCntDn uint32 `json:"fcntdn"`
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
		DevAddr:  d.DevAddr,
		AppSKey:  d.AppSKey,
		NwkSKey:  d.NwkSKey,
		FCntUp:   d.FCntUp,
		FCntDn:   d.FCntDn,
	}
}

func (d *Device) JoinAccept(frame lorawan.PHYPayload) error {
	log.Printf("[%s] received join accept", d.DevEUI)

	err := frame.DecryptJoinAcceptPayload(d.AppKey)
	if err != nil {
		log.Printf("[%s] decryption error %v", d.DevEUI, err)
		return err
	}

	ok, err := frame.ValidateDownlinkJoinMIC(lorawan.JoinRequestType, d.JoinEUI, d.DevNonce-1, d.AppKey)
	if err != nil {
		log.Printf("[%s] MIC error %v", d.DevEUI, err)
		return err
	}
	if !ok {
		// Join Accept is for another device
		log.Printf("[%s] invalid MIC", d.DevEUI)
		return errors.New("invalid MIC")
	} else {
		// Join Accept is for this device
		joinAccept, ok := frame.MACPayload.(*lorawan.JoinAcceptPayload)
		if !ok {
			log.Printf("[%s] invalid MAC payload for data downlink", d.DevEUI)
			return errors.New("invalid MAC payload")
		}

		d.mu.Lock()
		defer d.mu.Unlock()

		// DevAddr
		d.DevAddr = joinAccept.DevAddr

		// Derive NwkSKey
		d.NwkSKey, err = deriveSessionKey(0x01, d.AppKey, joinAccept.JoinNonce, joinAccept.HomeNetID, d.DevNonce-1)
		if err != nil {
			log.Printf("[%s] failed to derive NwkSKey: %v", d.DevEUI, err)
			return err
		}

		// Derive AppSKey
		d.AppSKey, err = deriveSessionKey(0x02, d.AppKey, joinAccept.JoinNonce, joinAccept.HomeNetID, d.DevNonce-1)
		if err != nil {
			log.Printf("[%s] failed to derive AppSKey: %v", d.DevEUI, err)
			return err
		}

		// Reset frame counters
		d.FCntUp = 0
		d.FCntDn = 0

		log.Printf("[%s] join successful - DevAddr: %s", d.DevEUI, d.DevAddr)

	}
	return nil
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
	d.DevNonce++

	// Prepare AppKey for MIC
	appkey := d.AppKey

	d.mu.Unlock()

	if err := phy.SetUplinkJoinMIC(appkey); err != nil {
		return lorawan.PHYPayload{}, err
	}

	d.broadcast(phy)

	return phy, nil
}

func (d *Device) Uplink() (lorawan.PHYPayload, error) {
	fPort := uint8(1)

	d.mu.Lock()
	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.ConfirmedDataUp,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: d.DevAddr, // Use the actual DevAddr from the device
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
				},
				FCnt: d.FCntUp,
			},
			FPort:      &fPort,
			FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{1, 2, 3, 4}}},
		},
	}

	// Increment FCntup
	d.FCntUp++

	// Prepare session keys for encryption and MIC
	appskey := d.AppSKey
	nwkskey := d.NwkSKey

	d.mu.Unlock()

	if err := phy.EncryptFRMPayload(appskey); err != nil {
		return lorawan.PHYPayload{}, err
	}

	if err := phy.SetUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, nwkskey, lorawan.AES128Key{}); err != nil {
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
		// Marshal PHYPayload to bytes and encode as hex
		phyBytes, err := phy.MarshalBinary()
		if err != nil {
			log.Printf("[%s] failed to marshal PHYPayload: %v", d.DevEUI, err)
			return
		}

		log.Printf("[%s] broadcasting uplink: %x", d.DevEUI, phyBytes)

		go func() {
			broadcastCh <- phy
		}()
	}
}

// deriveSessionKey derives NwkSKey (typ=0x01) or AppSKey (typ=0x02)
// Following LoRaWAN 1.0.x specification
func deriveSessionKey(typ byte, appKey lorawan.AES128Key, joinNonce lorawan.JoinNonce, netID lorawan.NetID, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key

	// Build the plaintext: type | JoinNonce | NetID | DevNonce | pad
	plaintext := make([]byte, 16)
	plaintext[0] = typ

	// Convert JoinNonce (uint32) to 3 bytes (little-endian, only lower 3 bytes)
	joinNonceBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(joinNonceBytes, uint32(joinNonce))
	copy(plaintext[1:4], joinNonceBytes[:3])

	// Copy NetID (3 bytes)
	copy(plaintext[4:7], netID[:])

	// Convert DevNonce (uint16) to 2 bytes (little-endian)
	devNonceBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(devNonceBytes, uint16(devNonce))
	copy(plaintext[7:9], devNonceBytes)

	// bytes 9-15 are zero padding (already initialized)

	// Encrypt with AppKey
	block, err := aes.NewCipher(appKey[:])
	if err != nil {
		return key, err
	}

	block.Encrypt(key[:], plaintext)
	return key, nil
}
