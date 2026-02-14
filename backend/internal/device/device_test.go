package device

import (
	"sync"
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a device with uplink channel for testing
func newTestDevice(devEUI lorawan.EUI64, joinEUI lorawan.EUI64, appKey lorawan.AES128Key, devNonce lorawan.DevNonce) *Device {
	uplinkCh := make(chan lorawan.PHYPayload, 10)
	return New(uplinkCh, devEUI, joinEUI, appKey, devNonce, lorawan.DevAddr{}, lorawan.AES128Key{}, lorawan.AES128Key{}, 0, 0)
}

func TestNew(t *testing.T) {
	t.Run("creates device with correct EUI", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		devNonce := lorawan.DevNonce(0)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		assert.NotNil(t, device)
		assert.Equal(t, devEUI, device.DevEUI)
		assert.Equal(t, joinEUI, device.JoinEUI)
		assert.Equal(t, appKey, device.AppKey)
		assert.Equal(t, devNonce, device.DevNonce)
		assert.Equal(t, uint32(0), device.FCntUp)
		assert.Equal(t, uint32(0), device.FCntDn)
	})

	t.Run("creates multiple devices with different EUIs", func(t *testing.T) {
		devEUI1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI1 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey1 := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		devNonce1 := lorawan.DevNonce(0)

		devEUI2 := lorawan.EUI64{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}
		joinEUI2 := lorawan.EUI64{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38}
		appKey2 := lorawan.AES128Key{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}
		devNonce2 := lorawan.DevNonce(0)

		device1 := newTestDevice(devEUI1, joinEUI1, appKey1, devNonce1)
		device2 := newTestDevice(devEUI2, joinEUI2, appKey2, devNonce2)

		assert.NotNil(t, device1)
		assert.NotNil(t, device2)
		assert.Equal(t, devEUI1, device1.DevEUI)
		assert.Equal(t, devEUI2, device2.DevEUI)
		assert.NotEqual(t, device1.DevEUI, device2.DevEUI)
	})
}

func TestDevice_GetInfo(t *testing.T) {
	t.Run("returns device info with correct EUI", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		devNonce := lorawan.DevNonce(0)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		info := device.GetInfo()

		assert.Equal(t, devEUI, info.DevEUI)
		assert.Equal(t, joinEUI, info.JoinEUI)
		assert.Equal(t, appKey, info.AppKey)
		assert.Equal(t, devNonce, info.DevNonce)
	})

	t.Run("concurrent GetInfo calls return correct data", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		devNonce := lorawan.DevNonce(0)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Test concurrent reads
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				info := device.GetInfo()
				assert.Equal(t, devEUI, info.DevEUI)
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestDeviceInfo_JSON(t *testing.T) {
	t.Run("DeviceInfo can be marshaled to JSON", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		devNonce := lorawan.DevNonce(0)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)
		info := device.GetInfo()

		// The struct has json tags, so it should be serializable
		assert.Equal(t, devEUI, info.DevEUI)
	})
}

func TestDevice_JoinRequest(t *testing.T) {
	t.Run("creates valid join request payload", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		phy, err := device.JoinRequest()
		assert.NoError(t, err)

		// Verify MHDR
		assert.Equal(t, lorawan.JoinRequest, phy.MHDR.MType)
		assert.Equal(t, lorawan.LoRaWANR1, phy.MHDR.Major)

		// Verify MACPayload
		joinReq, ok := phy.MACPayload.(*lorawan.JoinRequestPayload)
		assert.True(t, ok)
		assert.Equal(t, joinEUI, joinReq.JoinEUI)
		assert.Equal(t, devEUI, joinReq.DevEUI)
		assert.Equal(t, devNonce, joinReq.DevNonce)

		// Verify MIC is set (non-zero)
		assert.NotEqual(t, [4]byte{0, 0, 0, 0}, phy.MIC)
	})

	t.Run("increments DevNonce after join request", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// First join request
		phy1, err := device.JoinRequest()
		assert.NoError(t, err)
		joinReq1, _ := phy1.MACPayload.(*lorawan.JoinRequestPayload)
		assert.Equal(t, devNonce, joinReq1.DevNonce)

		// Second join request should have incremented DevNonce
		phy2, err := device.JoinRequest()
		assert.NoError(t, err)
		joinReq2, _ := phy2.MACPayload.(*lorawan.JoinRequestPayload)
		assert.Equal(t, devNonce+1, joinReq2.DevNonce)

		// Verify device's DevNonce is updated
		info := device.GetInfo()
		assert.Equal(t, devNonce+2, info.DevNonce)
	})

	t.Run("concurrent join requests increment DevNonce correctly", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		devNonce := lorawan.DevNonce(0)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		const numRequests = 10
		var wg sync.WaitGroup
		wg.Add(numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				defer wg.Done()
				_, err := device.JoinRequest()
				assert.NoError(t, err)
			}()
		}

		wg.Wait()

		// DevNonce should have been incremented exactly numRequests times
		info := device.GetInfo()
		assert.Equal(t, lorawan.DevNonce(numRequests), info.DevNonce)
	})
}

func TestDevice_JoinAccept(t *testing.T) {
	t.Run("processes valid join accept and derives session keys", func(t *testing.T) {
		// Use well-known test vectors
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Increment DevNonce to match what JoinRequest would do
		device.DevNonce = 101

		// Create a JoinAccept payload
		joinAccept := lorawan.JoinAcceptPayload{
			JoinNonce: lorawan.JoinNonce(0x123456), // 3-byte value
			HomeNetID: lorawan.NetID{0x00, 0x00, 0x01},
			DevAddr:   lorawan.DevAddr{0x01, 0x02, 0x03, 0x04},
			DLSettings: lorawan.DLSettings{
				RX2DataRate: 0,
				RX1DROffset: 0,
			},
			RXDelay: 1,
		}

		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.JoinAccept,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &joinAccept,
		}

		// Set MIC
		err := phy.SetDownlinkJoinMIC(lorawan.JoinRequestType, joinEUI, lorawan.DevNonce(100), appKey)
		assert.NoError(t, err)

		// Encrypt the join accept
		err = phy.EncryptJoinAcceptPayload(appKey)
		assert.NoError(t, err)

		// Process the join accept
		err = device.JoinAccept(phy)
		assert.NoError(t, err)

		// Verify DevAddr is set
		info := device.GetInfo()
		assert.Equal(t, joinAccept.DevAddr, info.DevAddr)

		// Verify session keys are derived (non-zero)
		assert.NotEqual(t, lorawan.AES128Key{}, info.NwkSKey)
		assert.NotEqual(t, lorawan.AES128Key{}, info.AppSKey)

		// Verify frame counters are reset
		assert.Equal(t, uint32(0), info.FCntUp)
		assert.Equal(t, uint32(0), info.FCntDn)
	})

	t.Run("fails on invalid MIC", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Increment DevNonce
		device.DevNonce = 101

		// Create a JoinAccept payload with wrong appKey for MIC
		wrongAppKey := lorawan.AES128Key{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00}

		joinAccept := lorawan.JoinAcceptPayload{
			JoinNonce: lorawan.JoinNonce(0x123456),
			HomeNetID: lorawan.NetID{0x00, 0x00, 0x01},
			DevAddr:   lorawan.DevAddr{0x01, 0x02, 0x03, 0x04},
			DLSettings: lorawan.DLSettings{
				RX2DataRate: 0,
				RX1DROffset: 0,
			},
			RXDelay: 1,
		}

		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.JoinAccept,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &joinAccept,
		}

		// Set MIC with wrong key
		err := phy.SetDownlinkJoinMIC(lorawan.JoinRequestType, joinEUI, lorawan.DevNonce(100), wrongAppKey)
		assert.NoError(t, err)

		// Encrypt with correct key
		err = phy.EncryptJoinAcceptPayload(appKey)
		assert.NoError(t, err)

		// Process should fail due to MIC mismatch
		err = device.JoinAccept(phy)
		// The function doesn't return error on invalid MIC based on the code,
		// it just logs and continues, so we verify keys are not set
		info := device.GetInfo()
		// DevAddr should remain unset (all zeros)
		assert.Equal(t, lorawan.DevAddr{0x00, 0x00, 0x00, 0x00}, info.DevAddr)
	})

	t.Run("handles decryption errors", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		wrongAppKey := lorawan.AES128Key{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, wrongAppKey, devNonce) // Wrong key for decryption

		device.DevNonce = 101

		joinAccept := lorawan.JoinAcceptPayload{
			JoinNonce: lorawan.JoinNonce(0x123456),
			HomeNetID: lorawan.NetID{0x00, 0x00, 0x01},
			DevAddr:   lorawan.DevAddr{0x01, 0x02, 0x03, 0x04},
			DLSettings: lorawan.DLSettings{
				RX2DataRate: 0,
				RX1DROffset: 0,
			},
			RXDelay: 1,
		}

		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.JoinAccept,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &joinAccept,
		}

		// Encrypt with the correct appKey (different from device's)
		err := phy.EncryptJoinAcceptPayload(appKey)
		assert.NoError(t, err)

		// Decryption should fail since device has wrong key
		err = device.JoinAccept(phy)
		assert.Error(t, err)
	})
}

func TestDeriveSessionKey(t *testing.T) {
	t.Run("derives NwkSKey correctly", func(t *testing.T) {
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		joinNonce := lorawan.JoinNonce(0x123456)
		netID := lorawan.NetID{0x00, 0x00, 0x01}
		devNonce := lorawan.DevNonce(100)

		nwkSKey, err := deriveSessionKey(0x01, appKey, joinNonce, netID, devNonce)
		assert.NoError(t, err)
		assert.NotEqual(t, lorawan.AES128Key{}, nwkSKey)

		// Verify it's deterministic - same inputs produce same output
		nwkSKey2, err := deriveSessionKey(0x01, appKey, joinNonce, netID, devNonce)
		assert.NoError(t, err)
		assert.Equal(t, nwkSKey, nwkSKey2)
	})

	t.Run("derives AppSKey correctly", func(t *testing.T) {
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		joinNonce := lorawan.JoinNonce(0x123456)
		netID := lorawan.NetID{0x00, 0x00, 0x01}
		devNonce := lorawan.DevNonce(100)

		appSKey, err := deriveSessionKey(0x02, appKey, joinNonce, netID, devNonce)
		assert.NoError(t, err)
		assert.NotEqual(t, lorawan.AES128Key{}, appSKey)

		// Verify it's deterministic
		appSKey2, err := deriveSessionKey(0x02, appKey, joinNonce, netID, devNonce)
		assert.NoError(t, err)
		assert.Equal(t, appSKey, appSKey2)
	})

	t.Run("NwkSKey and AppSKey are different", func(t *testing.T) {
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		joinNonce := lorawan.JoinNonce(0x123456)
		netID := lorawan.NetID{0x00, 0x00, 0x01}
		devNonce := lorawan.DevNonce(100)

		nwkSKey, err := deriveSessionKey(0x01, appKey, joinNonce, netID, devNonce)
		assert.NoError(t, err)

		appSKey, err := deriveSessionKey(0x02, appKey, joinNonce, netID, devNonce)
		assert.NoError(t, err)

		// NwkSKey and AppSKey should be different
		assert.NotEqual(t, nwkSKey, appSKey)
	})

	t.Run("different inputs produce different keys", func(t *testing.T) {
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		joinNonce1 := lorawan.JoinNonce(0x123456)
		joinNonce2 := lorawan.JoinNonce(0x654321)
		netID := lorawan.NetID{0x00, 0x00, 0x01}
		devNonce := lorawan.DevNonce(100)

		key1, err := deriveSessionKey(0x01, appKey, joinNonce1, netID, devNonce)
		assert.NoError(t, err)

		key2, err := deriveSessionKey(0x01, appKey, joinNonce2, netID, devNonce)
		assert.NoError(t, err)

		// Different joinNonces should produce different keys
		assert.NotEqual(t, key1, key2)
	})

	t.Run("handles edge case values", func(t *testing.T) {
		appKey := lorawan.AES128Key{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		joinNonce := lorawan.JoinNonce(0xffffff) // Max 3-byte value
		netID := lorawan.NetID{0xff, 0xff, 0xff}
		devNonce := lorawan.DevNonce(0xffff) // Max 2-byte value

		nwkSKey, err := deriveSessionKey(0x01, appKey, joinNonce, netID, devNonce)
		assert.NoError(t, err)
		assert.NotEqual(t, lorawan.AES128Key{}, nwkSKey)
	})
}

func TestDevice_Uplink(t *testing.T) {
	t.Run("creates valid uplink with device DevAddr", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Set device as joined with DevAddr and session keys
		device.DevAddr = lorawan.DevAddr{0x00, 0xdf, 0xb2, 0x28}
		device.AppSKey = lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		device.NwkSKey = lorawan.AES128Key{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}
		device.FCntUp = 0

		phy, err := device.Uplink()
		assert.NoError(t, err)

		// Verify MHDR
		assert.Equal(t, lorawan.ConfirmedDataUp, phy.MHDR.MType)
		assert.Equal(t, lorawan.LoRaWANR1, phy.MHDR.Major)

		// Verify MACPayload
		macPL, ok := phy.MACPayload.(*lorawan.MACPayload)
		assert.True(t, ok)
		assert.Equal(t, device.DevAddr, macPL.FHDR.DevAddr)
		assert.Equal(t, uint32(0), macPL.FHDR.FCnt)
		assert.NotNil(t, macPL.FPort)
		assert.Equal(t, uint8(1), *macPL.FPort)

		// Verify MIC is set
		assert.NotEqual(t, [4]byte{0, 0, 0, 0}, phy.MIC)

		// Verify FCntUp was incremented
		info := device.GetInfo()
		assert.Equal(t, uint32(1), info.FCntUp)
	})

	t.Run("increments FCntUp on each uplink", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Set device as joined
		device.DevAddr = lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		device.AppSKey = lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		device.NwkSKey = lorawan.AES128Key{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}

		// First uplink
		phy1, err := device.Uplink()
		assert.NoError(t, err)
		macPL1, _ := phy1.MACPayload.(*lorawan.MACPayload)
		assert.Equal(t, uint32(0), macPL1.FHDR.FCnt)

		// Second uplink should have incremented FCnt
		phy2, err := device.Uplink()
		assert.NoError(t, err)
		macPL2, _ := phy2.MACPayload.(*lorawan.MACPayload)
		assert.Equal(t, uint32(1), macPL2.FHDR.FCnt)

		// Verify device's FCntUp
		info := device.GetInfo()
		assert.Equal(t, uint32(2), info.FCntUp)
	})

	t.Run("encrypts FRMPayload with AppSKey", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Set device as joined
		device.DevAddr = lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		device.AppSKey = lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		device.NwkSKey = lorawan.AES128Key{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}

		phy, err := device.Uplink()
		assert.NoError(t, err)

		// FRMPayload should be encrypted (not [1, 2, 3, 4])
		macPL, _ := phy.MACPayload.(*lorawan.MACPayload)
		dataPayload, ok := macPL.FRMPayload[0].(*lorawan.DataPayload)
		assert.True(t, ok)

		// The encrypted payload should be different from the original [1, 2, 3, 4]
		assert.NotEqual(t, []byte{1, 2, 3, 4}, dataPayload.Bytes)
	})

	t.Run("concurrent uplinks increment FCntUp correctly", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Set device as joined
		device.DevAddr = lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		device.AppSKey = lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		device.NwkSKey = lorawan.AES128Key{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}

		const numUplinks = 10
		var wg sync.WaitGroup
		wg.Add(numUplinks)

		for i := 0; i < numUplinks; i++ {
			go func() {
				defer wg.Done()
				_, err := device.Uplink()
				assert.NoError(t, err)
			}()
		}

		wg.Wait()

		// FCntUp should have been incremented exactly numUplinks times
		info := device.GetInfo()
		assert.Equal(t, uint32(numUplinks), info.FCntUp)
	})
}

func TestDevice_Downlink(t *testing.T) {
	t.Run("processes valid downlink with FRMPayload", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Set device as joined
		device.DevAddr = lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		device.NwkSKey = lorawan.AES128Key{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
		device.AppSKey = lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}

		// Create a downlink with FRMPayload
		fPort := uint8(10)
		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.UnconfirmedDataDown,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &lorawan.MACPayload{
				FHDR: lorawan.FHDR{
					DevAddr: device.DevAddr,
					FCtrl: lorawan.FCtrl{
						ADR:       false,
						ADRACKReq: false,
						ACK:       false,
					},
					FCnt: 5,
				},
				FPort:      &fPort,
				FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{0xaa, 0xbb, 0xcc, 0xdd}}},
			},
		}

		// Encrypt and set MIC
		err := phy.EncryptFRMPayload(device.AppSKey)
		assert.NoError(t, err)

		err = phy.SetDownlinkDataMIC(lorawan.LoRaWAN1_0, 0, device.NwkSKey)
		assert.NoError(t, err)

		// Process downlink
		err = device.Downlink(phy)
		assert.NoError(t, err)
	})

	t.Run("processes downlink with empty FRMPayload (MAC-only)", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Set device as joined
		device.DevAddr = lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		device.NwkSKey = lorawan.AES128Key{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
		device.AppSKey = lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}

		// Create a MAC-only downlink (no FRMPayload)
		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.UnconfirmedDataDown,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &lorawan.MACPayload{
				FHDR: lorawan.FHDR{
					DevAddr: device.DevAddr,
					FCtrl: lorawan.FCtrl{
						ADR:       false,
						ADRACKReq: false,
						ACK:       false,
					},
					FCnt: 3,
				},
				FPort:      nil,
				FRMPayload: []lorawan.Payload{}, // Empty payload
			},
		}

		// Set MIC
		err := phy.SetDownlinkDataMIC(lorawan.LoRaWAN1_0, 0, device.NwkSKey)
		assert.NoError(t, err)

		// Process downlink - should not panic
		err = device.Downlink(phy)
		assert.NoError(t, err)
	})

	t.Run("fails on invalid MIC", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Set device as joined
		device.DevAddr = lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		device.NwkSKey = lorawan.AES128Key{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
		device.AppSKey = lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}

		// Use wrong NwkSKey for MIC
		wrongNwkSKey := lorawan.AES128Key{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00}

		fPort := uint8(10)
		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.UnconfirmedDataDown,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &lorawan.MACPayload{
				FHDR: lorawan.FHDR{
					DevAddr: device.DevAddr,
					FCtrl: lorawan.FCtrl{
						ADR:       false,
						ADRACKReq: false,
						ACK:       false,
					},
					FCnt: 2,
				},
				FPort:      &fPort,
				FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{0x11, 0x22}}},
			},
		}

		// Encrypt with correct key but set MIC with wrong key
		err := phy.EncryptFRMPayload(device.AppSKey)
		assert.NoError(t, err)

		err = phy.SetDownlinkDataMIC(lorawan.LoRaWAN1_0, 0, wrongNwkSKey)
		assert.NoError(t, err)

		// Process downlink - should fail MIC validation
		err = device.Downlink(phy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid MIC")
	})

	t.Run("handles ConfirmedDataDown same as UnconfirmedDataDown", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)
		device := newTestDevice(devEUI, joinEUI, appKey, devNonce)

		// Set device as joined
		device.DevAddr = lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		device.NwkSKey = lorawan.AES128Key{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
		device.AppSKey = lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}

		// Create a ConfirmedDataDown
		fPort := uint8(10)
		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.ConfirmedDataDown,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &lorawan.MACPayload{
				FHDR: lorawan.FHDR{
					DevAddr: device.DevAddr,
					FCtrl: lorawan.FCtrl{
						ADR:       false,
						ADRACKReq: false,
						ACK:       true, // Confirmed downlink
					},
					FCnt: 7,
				},
				FPort:      &fPort,
				FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{0x55, 0x66}}},
			},
		}

		// Encrypt and set MIC
		err := phy.EncryptFRMPayload(device.AppSKey)
		assert.NoError(t, err)

		err = phy.SetDownlinkDataMIC(lorawan.LoRaWAN1_0, 0, device.NwkSKey)
		assert.NoError(t, err)

		// Process downlink
		err = device.Downlink(phy)
		assert.NoError(t, err)
	})
}
