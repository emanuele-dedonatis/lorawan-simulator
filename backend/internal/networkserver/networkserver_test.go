package networkserver

import (
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/integration"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a network server with channels for testing
func newTestNetworkServer(name string) *NetworkServer {
	uplinkCh := make(chan lorawan.PHYPayload, 10)
	downlinkCh := make(chan lorawan.PHYPayload, 10)
	config := integration.NetworkServerConfig{
		Type: integration.NetworkServerTypeGeneric,
	}
	return New(name, config, uplinkCh, downlinkCh)
}

func TestNew(t *testing.T) {
	t.Run("creates network server with valid name", func(t *testing.T) {
		name := "my-network-server"
		uplinkCh := make(chan lorawan.PHYPayload)
		downlinkCh := make(chan lorawan.PHYPayload)
		config := integration.NetworkServerConfig{
			Type: integration.NetworkServerTypeGeneric,
		}
		ns := New(name, config, uplinkCh, downlinkCh)

		assert.NotNil(t, ns)
		assert.Equal(t, name, ns.name)
		assert.NotNil(t, ns.gateways)
		assert.NotNil(t, ns.devices)
		assert.Equal(t, 0, len(ns.gateways))
		assert.Equal(t, 0, len(ns.devices))
	})

	t.Run("multiple instances are independent", func(t *testing.T) {
		name1 := "server-1"
		uplinkCh1 := make(chan lorawan.PHYPayload)
		downlinkCh1 := make(chan lorawan.PHYPayload)
		config := integration.NetworkServerConfig{
			Type: integration.NetworkServerTypeGeneric,
		}
		ns1 := New(name1, config, uplinkCh1, downlinkCh1)
		name2 := "server-2"
		uplinkCh2 := make(chan lorawan.PHYPayload)
		downlinkCh2 := make(chan lorawan.PHYPayload)
		ns2 := New(name2, config, uplinkCh2, downlinkCh2)

		assert.NotEqual(t, ns1, ns2)
		assert.Equal(t, name1, ns1.name)
		assert.Equal(t, name2, ns2.name)
	})
}

func TestNetworkServer_AddGateway(t *testing.T) {
	t.Run("adds gateway successfully", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://example.com:6887"

		gw, err := ns.AddGateway(eui, discoveryURI)

		assert.NoError(t, err)
		assert.NotNil(t, gw)
		assert.Equal(t, eui, gw.GetInfo().EUI)

		// Verify it's in the map
		info := ns.GetInfo()
		assert.Equal(t, 1, info.GatewayCount)
	})

	t.Run("returns error when adding duplicate gateway", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://example.com:6887"

		ns.AddGateway(eui, discoveryURI)
		gw2, err := ns.AddGateway(eui, discoveryURI)

		assert.Error(t, err)
		assert.Nil(t, gw2)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("can add multiple different gateways", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")
		eui1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		eui2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}

		gw1, err1 := ns.AddGateway(eui1, "wss://gw1.com")
		gw2, err2 := ns.AddGateway(eui2, "wss://gw2.com")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotNil(t, gw1)
		assert.NotNil(t, gw2)

		info := ns.GetInfo()
		assert.Equal(t, 2, info.GatewayCount)
	})
}

func TestNetworkServer_GetGateway(t *testing.T) {
	t.Run("gets existing gateway", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		ns.AddGateway(eui, "wss://example.com")

		gw, err := ns.GetGateway(eui)

		assert.Nil(t, err)
		assert.NotNil(t, gw)
		assert.Equal(t, eui, gw.GetInfo().EUI)
	})

	t.Run("returns err for non-existing gateway", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		gw, err := ns.GetGateway(eui)

		assert.NotNil(t, err)
		assert.Equal(t, "gateway not found", err.Error())
		assert.Nil(t, gw)
	})
}

func TestNetworkServer_ListGateways(t *testing.T) {
	t.Run("returns empty list when no gateways", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		gateways := ns.ListGateways()

		assert.NotNil(t, gateways)
		assert.Equal(t, 0, len(gateways))
	})

	t.Run("returns all gateways", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")
		eui1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		eui2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		eui3 := lorawan.EUI64{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}

		ns.AddGateway(eui1, "wss://gw1.com")
		ns.AddGateway(eui2, "wss://gw2.com")
		ns.AddGateway(eui3, "wss://gw3.com")

		gateways := ns.ListGateways()

		assert.Equal(t, 3, len(gateways))

		// Verify all EUIs are present
		euiMap := make(map[lorawan.EUI64]bool)
		for _, gw := range gateways {
			euiMap[gw.EUI] = true
		}

		assert.True(t, euiMap[eui1])
		assert.True(t, euiMap[eui2])
		assert.True(t, euiMap[eui3])
	})
}

func TestNetworkServer_RemoveGateway(t *testing.T) {
	t.Run("removes existing gateway", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		ns.AddGateway(eui, "wss://example.com")

		err := ns.RemoveGateway(eui)

		assert.NoError(t, err)

		// Verify it's removed
		_, err2 := ns.GetGateway(eui)
		assert.NotNil(t, err2)
		assert.Equal(t, "gateway not found", err2.Error())

		info := ns.GetInfo()
		assert.Equal(t, 0, info.GatewayCount)
	})

	t.Run("returns error when removing non-existing gateway", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		err := ns.RemoveGateway(eui)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("removes one gateway from multiple", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")
		eui1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		eui2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		eui3 := lorawan.EUI64{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}

		ns.AddGateway(eui1, "wss://gw1.com")
		ns.AddGateway(eui2, "wss://gw2.com")
		ns.AddGateway(eui3, "wss://gw3.com")

		err := ns.RemoveGateway(eui2)

		assert.NoError(t, err)

		// Verify eui2 is removed but others remain
		_, err1 := ns.GetGateway(eui1)
		_, err2 := ns.GetGateway(eui2)
		_, err3 := ns.GetGateway(eui3)

		assert.Nil(t, err1)
		assert.NotNil(t, err2)
		assert.Equal(t, "gateway not found", err2.Error())
		assert.Nil(t, err3)

		info := ns.GetInfo()
		assert.Equal(t, 2, info.GatewayCount)
	})
}

func TestNetworkServer_GetInfo(t *testing.T) {
	t.Run("returns correct counts", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		// Add gateways
		eui1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		eui2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		ns.AddGateway(eui1, "wss://gw1.com")
		ns.AddGateway(eui2, "wss://gw2.com")

		info := ns.GetInfo()

		assert.Equal(t, "test-server", info.Name)
		assert.Equal(t, 2, info.GatewayCount)
		assert.Equal(t, 0, info.DeviceCount)
	})
}

func TestNetworkServer_ForwardDownlink(t *testing.T) {
	t.Run("broadcasts JoinAccept to all devices", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		// Add devices
		devEUI1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		devEUI2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		joinEUI := lorawan.EUI64{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)

		dev1, _ := ns.AddDevice(devEUI1, joinEUI, appKey, devNonce)
		dev2, _ := ns.AddDevice(devEUI2, joinEUI, appKey, devNonce)

		// Create a JoinAccept frame
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

		// Forward should broadcast to all devices (JoinAccept type)
		err := ns.ForwardDownlink(phy)
		assert.NoError(t, err)

		// Both devices should exist
		assert.NotNil(t, dev1)
		assert.NotNil(t, dev2)
	})

	t.Run("forwards data downlink only to device with matching DevAddr", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		// Add devices with different DevAddrs
		devEUI1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		devEUI2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		joinEUI := lorawan.EUI64{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)

		dev1, _ := ns.AddDevice(devEUI1, joinEUI, appKey, devNonce)
		dev2, _ := ns.AddDevice(devEUI2, joinEUI, appKey, devNonce)

		// Simulate that dev1 has completed join and has a DevAddr
		targetDevAddr := lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		dev1.DevAddr = targetDevAddr
		dev2.DevAddr = lorawan.DevAddr{0x05, 0x06, 0x07, 0x08} // Different DevAddr

		// Create a data downlink frame for dev1's DevAddr
		macPayload := lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: targetDevAddr,
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
					FPending:  false,
				},
				FCnt: 1,
			},
			FPort:      nil,
			FRMPayload: []lorawan.Payload{},
		}

		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.UnconfirmedDataDown,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &macPayload,
		}

		// Forward downlink - should only go to dev1 (matching DevAddr)
		err := ns.ForwardDownlink(phy)
		assert.NoError(t, err)

		// Verify both devices still exist
		assert.NotNil(t, dev1)
		assert.NotNil(t, dev2)
	})

	t.Run("handles empty device list gracefully", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.JoinAccept,
				Major: lorawan.LoRaWANR1,
			},
		}

		// Should not panic when no devices exist
		err := ns.ForwardDownlink(phy)
		assert.NoError(t, err)
	})

	t.Run("filters data downlink by DevAddr correctly", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		// Add multiple devices
		devEUI1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		devEUI2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		devEUI3 := lorawan.EUI64{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}
		joinEUI := lorawan.EUI64{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)

		dev1, _ := ns.AddDevice(devEUI1, joinEUI, appKey, devNonce)
		dev2, _ := ns.AddDevice(devEUI2, joinEUI, appKey, devNonce)
		dev3, _ := ns.AddDevice(devEUI3, joinEUI, appKey, devNonce)

		// Assign different DevAddrs
		dev1.DevAddr = lorawan.DevAddr{0x01, 0x00, 0x00, 0x00}
		dev2.DevAddr = lorawan.DevAddr{0x02, 0x00, 0x00, 0x00}
		dev3.DevAddr = lorawan.DevAddr{0x03, 0x00, 0x00, 0x00}

		// Create downlink for dev2's DevAddr
		targetDevAddr := lorawan.DevAddr{0x02, 0x00, 0x00, 0x00}
		macPayload := lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: targetDevAddr,
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
					FPending:  false,
				},
				FCnt: 1,
			},
			FPort:      nil,
			FRMPayload: []lorawan.Payload{},
		}

		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.UnconfirmedDataDown,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &macPayload,
		}

		// Forward - should only match dev2
		err := ns.ForwardDownlink(phy)
		assert.NoError(t, err)

		// All devices should still exist
		assert.NotNil(t, dev1)
		assert.NotNil(t, dev2)
		assert.NotNil(t, dev3)
	})

	t.Run("handles ConfirmedDataDown same as UnconfirmedDataDown", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		appKey := lorawan.AES128Key{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		devNonce := lorawan.DevNonce(100)

		dev, _ := ns.AddDevice(devEUI, joinEUI, appKey, devNonce)
		targetDevAddr := lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		dev.DevAddr = targetDevAddr

		// Create a confirmed data downlink frame
		macPayload := lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: targetDevAddr,
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
					FPending:  false,
				},
				FCnt: 1,
			},
			FPort:      nil,
			FRMPayload: []lorawan.Payload{},
		}

		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.ConfirmedDataDown,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &macPayload,
		}

		// Should handle ConfirmedDataDown with DevAddr filtering
		err := ns.ForwardDownlink(phy)
		assert.NoError(t, err)
	})
}

func TestNetworkServer_ForwardUplink(t *testing.T) {
	t.Run("forwards uplink to all gateways", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		// Add gateways
		eui1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		eui2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}

		gw1, _ := ns.AddGateway(eui1, "wss://gw1.com")
		gw2, _ := ns.AddGateway(eui2, "wss://gw2.com")

		// Create an uplink frame
		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.UnconfirmedDataUp,
				Major: lorawan.LoRaWANR1,
			},
		}

		// Forward should send to all gateways
		assert.NotPanics(t, func() {
			ns.ForwardUplink(phy)
		})

		// Verify gateways still exist
		assert.NotNil(t, gw1)
		assert.NotNil(t, gw2)
	})

	t.Run("handles empty gateway list gracefully", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.UnconfirmedDataUp,
				Major: lorawan.LoRaWANR1,
			},
		}

		// Should not panic when no gateways exist
		assert.NotPanics(t, func() {
			ns.ForwardUplink(phy)
		})
	})
}

func TestNetworkServer_SendUplink(t *testing.T) {
	t.Run("triggers uplink for existing device", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		// Add a device
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		appKey := lorawan.AES128Key{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
		devNonce := lorawan.DevNonce(100)
		dev, err := ns.AddDevice(devEUI, joinEUI, appKey, devNonce)
		assert.NoError(t, err)
		assert.NotNil(t, dev)

		// Set DevAddr and session keys (simulating joined device)
		dev.DevAddr = lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		dev.NwkSKey = lorawan.AES128Key{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
		dev.AppSKey = lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
		dev.FCntUp = 0

		// Call SendUplink should not panic
		err = ns.SendUplink(devEUI)
		assert.NoError(t, err)

		// Verify FCnt incremented
		assert.Equal(t, uint32(1), dev.FCntUp)
	})

	t.Run("returns error for non-existing device", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		// Try to send uplink for device that doesn't exist
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		err := ns.SendUplink(devEUI)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("handles device without DevAddr", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		// Add a device but don't set DevAddr (not joined)
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		appKey := lorawan.AES128Key{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
		devNonce := lorawan.DevNonce(100)
		dev, err := ns.AddDevice(devEUI, joinEUI, appKey, devNonce)
		assert.NoError(t, err)
		assert.NotNil(t, dev)

		// Device has zero DevAddr (not joined)
		// This should still work but the uplink will have empty DevAddr
		err = ns.SendUplink(devEUI)
		assert.NoError(t, err)
	})

	t.Run("increments FCnt correctly across multiple uplinks", func(t *testing.T) {
		ns := newTestNetworkServer("test-server")

		// Add a device
		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		appKey := lorawan.AES128Key{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
		devNonce := lorawan.DevNonce(100)
		dev, err := ns.AddDevice(devEUI, joinEUI, appKey, devNonce)
		assert.NoError(t, err)

		// Set DevAddr and session keys
		dev.DevAddr = lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}
		dev.NwkSKey = lorawan.AES128Key{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
		dev.AppSKey = lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

		// Send multiple uplinks
		for i := uint32(0); i < 5; i++ {
			err := ns.SendUplink(devEUI)
			assert.NoError(t, err)
			assert.Equal(t, i+1, dev.FCntUp)
		}

		assert.Equal(t, uint32(5), dev.FCntUp)
	})
}
