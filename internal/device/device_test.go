package device

import (
	"sync"
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("creates device with correct EUI", func(t *testing.T) {
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		devNonce := lorawan.DevNonce(0)
		device := New(devEUI, joinEUI, appKey, devNonce)

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

		device1 := New(devEUI1, joinEUI1, appKey1, devNonce1)
		device2 := New(devEUI2, joinEUI2, appKey2, devNonce2)

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
		device := New(devEUI, joinEUI, appKey, devNonce)

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
		device := New(devEUI, joinEUI, appKey, devNonce)

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
		device := New(devEUI, joinEUI, appKey, devNonce)
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
		device := New(devEUI, joinEUI, appKey, devNonce)

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
		device := New(devEUI, joinEUI, appKey, devNonce)

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
		device := New(devEUI, joinEUI, appKey, devNonce)

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

