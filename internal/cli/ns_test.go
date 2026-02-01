package cli

import (
	"testing"

	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/stretchr/testify/assert"
)

func TestNsAddCmd(t *testing.T) {
	t.Run("adds a network server successfully", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"ns", "add", "test-server"})
		err := rootCmd.Execute()

		// Assert
		assert.NoError(t, err)

		// Verify server was actually added
		_, exists := testPool.Get("test-server")
		assert.True(t, exists)

		// Verify pool has 1 server
		servers := testPool.List()
		assert.Equal(t, 1, len(servers))
		assert.Equal(t, "test-server", servers[0].GetInfo().Name)
	})

	t.Run("returns error when adding duplicate server", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("existing-server")
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"ns", "add", "existing-server"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("requires server name argument", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		rootCmd := InitRootCmd(testPool)

		// Execute command without arguments
		rootCmd.SetArgs([]string{"ns", "add"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
	})
}

func TestNsRemoveCmd(t *testing.T) {
	t.Run("removes an existing network server", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("test-server")
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"ns", "remove", "test-server"})
		err := rootCmd.Execute()

		// Assert
		assert.NoError(t, err)

		// Verify server was actually removed
		_, exists := testPool.Get("test-server")
		assert.False(t, exists)
	})

	t.Run("returns error when removing non-existent server", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"ns", "remove", "non-existent"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("requires server name argument", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		rootCmd := InitRootCmd(testPool)

		// Execute command without arguments
		rootCmd.SetArgs([]string{"ns", "remove"})
		err := rootCmd.Execute()

		// Assert
		assert.Error(t, err)
	})
}

func TestNsListCmd(t *testing.T) {
	t.Run("lists all network servers", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		testPool.Add("server-1")
		testPool.Add("server-2")
		testPool.Add("server-3")
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"ns", "list"})
		err := rootCmd.Execute()

		// Assert
		assert.NoError(t, err)

		// Verify servers are in the pool
		servers := testPool.List()
		assert.Equal(t, 3, len(servers))
	})

	t.Run("shows empty list when no servers", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()
		rootCmd := InitRootCmd(testPool)

		// Execute command
		rootCmd.SetArgs([]string{"ns", "list"})
		err := rootCmd.Execute()

		// Assert
		assert.NoError(t, err)

		// Verify pool is empty
		servers := testPool.List()
		assert.Equal(t, 0, len(servers))
	})
}

func TestIntegration_NsCommands(t *testing.T) {
	t.Run("add, list, remove workflow", func(t *testing.T) {
		// Setup
		testPool := networkserver.NewPool()

		// Add server
		rootCmd := InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"ns", "add", "test-server"})
		err := rootCmd.Execute()
		assert.NoError(t, err)

		// Verify added
		servers := testPool.List()
		assert.Equal(t, 1, len(servers))
		assert.Equal(t, "test-server", servers[0].GetInfo().Name)

		// Remove server
		rootCmd = InitRootCmd(testPool)
		rootCmd.SetArgs([]string{"ns", "remove", "test-server"})
		err = rootCmd.Execute()
		assert.NoError(t, err)

		// Verify empty
		servers = testPool.List()
		assert.Equal(t, 0, len(servers))
	})
}
