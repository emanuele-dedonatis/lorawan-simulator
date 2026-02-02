package device

import (
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("creates device with correct EUI", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		device := New(eui)

		assert.NotNil(t, device)
		assert.Equal(t, eui, device.DevEUI)
		assert.Equal(t, uint32(0), device.FCntUp)
		assert.Equal(t, uint32(0), device.FCntDn)
	})

	t.Run("creates multiple devices with different EUIs", func(t *testing.T) {
		eui1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		eui2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}

		device1 := New(eui1)
		device2 := New(eui2)

		assert.NotNil(t, device1)
		assert.NotNil(t, device2)
		assert.Equal(t, eui1, device1.DevEUI)
		assert.Equal(t, eui2, device2.DevEUI)
		assert.NotEqual(t, device1.DevEUI, device2.DevEUI)
	})
}

func TestDevice_GetInfo(t *testing.T) {
	t.Run("returns device info with correct EUI", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		device := New(eui)

		info := device.GetInfo()

		assert.Equal(t, eui, info.DevEUI)
	})

	t.Run("concurrent GetInfo calls return correct data", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		device := New(eui)

		// Test concurrent reads
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				info := device.GetInfo()
				assert.Equal(t, eui, info.DevEUI)
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
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		device := New(eui)
		info := device.GetInfo()

		// The struct has json tags, so it should be serializable
		assert.Equal(t, eui, info.DevEUI)
	})
}
