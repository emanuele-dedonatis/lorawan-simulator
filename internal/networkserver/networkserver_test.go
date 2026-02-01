package networkserver

import (
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("creates network server with valid name", func(t *testing.T) {
		name := "my-network-server"
		ns := New(name)

		assert.NotNil(t, ns)
		assert.Equal(t, name, ns.name)
		assert.NotNil(t, ns.gateways)
		assert.NotNil(t, ns.devices)
		assert.Equal(t, 0, len(ns.gateways))
		assert.Equal(t, 0, len(ns.devices))
	})

	t.Run("multiple instances are independent", func(t *testing.T) {
		name1 := "server-1"
		ns1 := New(name1)
		name2 := "server-2"
		ns2 := New(name2)

		assert.NotEqual(t, ns1, ns2)
		assert.Equal(t, name1, ns1.name)
		assert.Equal(t, name2, ns2.name)
	})
}

func TestNetworkServer_AddGateway(t *testing.T) {
	t.Run("adds gateway successfully", func(t *testing.T) {
		ns := New("test-server")
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
		ns := New("test-server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		discoveryURI := "wss://example.com:6887"

		ns.AddGateway(eui, discoveryURI)
		gw2, err := ns.AddGateway(eui, discoveryURI)

		assert.Error(t, err)
		assert.Nil(t, gw2)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("can add multiple different gateways", func(t *testing.T) {
		ns := New("test-server")
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
		ns := New("test-server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		ns.AddGateway(eui, "wss://example.com")

		gw, err := ns.GetGateway(eui)

		assert.Nil(t, err)
		assert.NotNil(t, gw)
		assert.Equal(t, eui, gw.GetInfo().EUI)
	})

	t.Run("returns err for non-existing gateway", func(t *testing.T) {
		ns := New("test-server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		gw, err := ns.GetGateway(eui)

		assert.NotNil(t, err)
		assert.Equal(t, "gateway not found", err.Error())
		assert.Nil(t, gw)
	})
}

func TestNetworkServer_ListGateways(t *testing.T) {
	t.Run("returns empty list when no gateways", func(t *testing.T) {
		ns := New("test-server")

		gateways := ns.ListGateways()

		assert.NotNil(t, gateways)
		assert.Equal(t, 0, len(gateways))
	})

	t.Run("returns all gateways", func(t *testing.T) {
		ns := New("test-server")
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
			euiMap[gw.GetInfo().EUI] = true
		}

		assert.True(t, euiMap[eui1])
		assert.True(t, euiMap[eui2])
		assert.True(t, euiMap[eui3])
	})
}

func TestNetworkServer_RemoveGateway(t *testing.T) {
	t.Run("removes existing gateway", func(t *testing.T) {
		ns := New("test-server")
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
		ns := New("test-server")
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		err := ns.RemoveGateway(eui)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("removes one gateway from multiple", func(t *testing.T) {
		ns := New("test-server")
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
		ns := New("test-server")

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
