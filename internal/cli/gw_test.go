package cli

import (
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/stretchr/testify/assert"
)

func TestGwAddCmd(t *testing.T) {
	t.Run("adds a gateway successfully", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"gw", "add", "test-server", "0102030405060708", "wss://gateway.example.com:6887"})
		err := rootCmd.Execute()

		// Assert
		assert.NoError(t, err)

		// Verify gateway was actually added
		ns, _ := testPool.Get("test-server")
		var eui lorawan.EUI64
		eui.UnmarshalText([]byte("0102030405060708"))
		gw, exists := ns.GetGateway(eui)
		assert.True(t, exists)
		assert.NotNil(t, gw)

		// Verify gateway info
		info := gw.GetInfo()
		assert.Equal(t, eui, info.EUI)
		assert.Equal(t, "wss://gateway.example.com:6887", info.DiscoveryURI)
	})

	t.Run("returns error when network server not found", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"gw", "add", "non-existent", "0102030405060708", "wss://gateway.example.com:6887"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error when adding duplicate gateway", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		ns, _ := testPool.Add("test-server")
		var eui lorawan.EUI64
		eui.UnmarshalText([]byte("0102030405060708"))
		ns.AddGateway(eui, "wss://gateway1.example.com:6887")
		rootCmd := InitRootCmd(testPool)

		// Execute command with same EUI
		rootCmd.SetArgs([]string{"gw", "add", "test-server", "0102030405060708", "wss://gateway2.example.com:6887"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("returns error with invalid EUI format", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")
		rootCmd := InitRootCmd(testPool)

		// Execute command with invalid EUI
		rootCmd.SetArgs([]string{"gw", "add", "test-server", "invalid-eui", "wss://gateway.example.com:6887"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid EUI format")
	})

	t.Run("requires all arguments", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")
		rootCmd := InitRootCmd(testPool)

		// Execute command without enough arguments
		rootCmd.SetArgs([]string{"gw", "add", "test-server"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
	})

	t.Run("adds multiple gateways to same server", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")

		// Add first gateway
		rootCmd := InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"gw", "add", "test-server", "0102030405060708", "wss://gateway1.example.com:6887"})
		err := rootCmd.Execute()
		assert.NoError(t, err)

		// Add second gateway
		rootCmd = InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"gw", "add", "test-server", "1112131415161718", "wss://gateway2.example.com:6887"})
		err = rootCmd.Execute()
		assert.NoError(t, err)

		// Verify both gateways exist
		ns, _ := testPool.Get("test-server")
		gateways := ns.ListGateways()
		assert.Equal(t, 2, len(gateways))
	})
}

func TestGwRemoveCmd(t *testing.T) {
	t.Run("removes an existing gateway", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		ns, _ := testPool.Add("test-server")
		var eui lorawan.EUI64
		eui.UnmarshalText([]byte("0102030405060708"))
		ns.AddGateway(eui, "wss://gateway.example.com:6887")
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"gw", "remove", "test-server", "0102030405060708"})
		err := rootCmd.Execute()

		// Assert
		assert.NoError(t, err)

		// Verify gateway was actually removed
		_, exists := ns.GetGateway(eui)
		assert.False(t, exists)
	})

	t.Run("returns error when network server not found", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"gw", "remove", "non-existent", "0102030405060708"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error when removing non-existent gateway", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"gw", "remove", "test-server", "0102030405060708"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error with invalid EUI format", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")
		rootCmd := InitRootCmd(testPool)

		// Execute command with invalid EUI
		rootCmd.SetArgs([]string{"gw", "remove", "test-server", "invalid-eui"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid EUI format")
	})

	t.Run("requires all arguments", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")
		rootCmd := InitRootCmd(testPool)

		// Execute command without enough arguments
		rootCmd.SetArgs([]string{"gw", "remove", "test-server"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
	})
}

func TestGwListCmd(t *testing.T) {
	t.Run("lists all gateways in a network server", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		ns, _ := testPool.Add("test-server")

		var eui1 lorawan.EUI64
		eui1.UnmarshalText([]byte("0102030405060708"))
		ns.AddGateway(eui1, "wss://gateway1.example.com:6887")

		var eui2 lorawan.EUI64
		eui2.UnmarshalText([]byte("1112131415161718"))
		ns.AddGateway(eui2, "wss://gateway2.example.com:6887")

		var eui3 lorawan.EUI64
		eui3.UnmarshalText([]byte("2122232425262728"))
		ns.AddGateway(eui3, "wss://gateway3.example.com:6887")

		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"gw", "list", "test-server"})
		err := rootCmd.Execute()

		// Assert
		assert.NoError(t, err)

		// Verify gateways are in the server
		gateways := ns.ListGateways()
		assert.Equal(t, 3, len(gateways))
	})

	t.Run("returns error when network server not found", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"gw", "list", "non-existent"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("shows empty list when no gateways", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"gw", "list", "test-server"})
		err := rootCmd.Execute()

		// Assert
		assert.NoError(t, err)

		// Verify no gateways
		ns, _ := testPool.Get("test-server")
		gateways := ns.ListGateways()
		assert.Equal(t, 0, len(gateways))
	})

	t.Run("requires network server argument", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		rootCmd := InitRootCmd(testPool)

		// Execute command without arguments
		rootCmd.SetArgs([]string{"gw", "list"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
	})
}

func TestIntegration_GwCommands(t *testing.T) {
	t.Run("add, list, remove workflow", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")

		// Add gateway
		rootCmd := InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"gw", "add", "test-server", "0102030405060708", "wss://gateway.example.com:6887"})
		err := rootCmd.Execute()
		assert.NoError(t, err)

		// Verify added
		ns, _ := testPool.Get("test-server")
		gateways := ns.ListGateways()
		assert.Equal(t, 1, len(gateways))

		// Remove gateway
		rootCmd = InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"gw", "remove", "test-server", "0102030405060708"})
		err = rootCmd.Execute()
		assert.NoError(t, err)

		// Verify empty
		gateways = ns.ListGateways()
		assert.Equal(t, 0, len(gateways))
	})

	t.Run("add multiple gateways and list them", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")

		// Add first gateway
		rootCmd := InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"gw", "add", "test-server", "0102030405060708", "wss://gateway1.example.com:6887"})
		err := rootCmd.Execute()
		assert.NoError(t, err)

		// Add second gateway
		rootCmd = InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"gw", "add", "test-server", "1112131415161718", "wss://gateway2.example.com:6887"})
		err = rootCmd.Execute()
		assert.NoError(t, err)

		// List gateways
		rootCmd = InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"gw", "list", "test-server"})
		err = rootCmd.Execute()
		assert.NoError(t, err)

		// Verify both exist
		ns, _ := testPool.Get("test-server")
		gateways := ns.ListGateways()
		assert.Equal(t, 2, len(gateways))
	})

	t.Run("gateways in different network servers are independent", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("server-1")
		testPool.Add("server-2")

		// Add gateway to server-1
		rootCmd := InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"gw", "add", "server-1", "0102030405060708", "wss://gateway1.example.com:6887"})
		err := rootCmd.Execute()
		assert.NoError(t, err)

		// Add gateway to server-2
		rootCmd = InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"gw", "add", "server-2", "1112131415161718", "wss://gateway2.example.com:6887"})
		err = rootCmd.Execute()
		assert.NoError(t, err)

		// Verify server-1 has 1 gateway
		ns1, _ := testPool.Get("server-1")
		gateways1 := ns1.ListGateways()
		assert.Equal(t, 1, len(gateways1))

		// Verify server-2 has 1 gateway
		ns2, _ := testPool.Get("server-2")
		gateways2 := ns2.ListGateways()
		assert.Equal(t, 1, len(gateways2))

		// Verify they have different EUIs
		assert.NotEqual(t, gateways1[0].GetInfo().EUI, gateways2[0].GetInfo().EUI)
	})
}
