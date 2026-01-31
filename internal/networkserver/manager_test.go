package networkserver

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	t.Run("creates manager with empty map", func(t *testing.T) {
		m := NewManager()

		assert.NotNil(t, m)
		assert.NotNil(t, m.ns)
		assert.Equal(t, 0, len(m.ns))
	})
}

func TestManager_Add(t *testing.T) {
	t.Run("adds new network server successfully", func(t *testing.T) {
		m := NewManager()
		name := "test-server"

		ns, err := m.Add(name)

		assert.NoError(t, err)
		assert.NotNil(t, ns)
		assert.Equal(t, name, ns.name)
		assert.Equal(t, 1, len(m.ns))
	})

	t.Run("returns error when adding duplicate network server", func(t *testing.T) {
		m := NewManager()
		name := "test-server"

		ns1, err1 := m.Add(name)
		assert.NoError(t, err1)
		assert.NotNil(t, ns1)

		ns2, err2 := m.Add(name)
		assert.Error(t, err2)
		assert.Nil(t, ns2)
		assert.Equal(t, "Network Server already exists", err2.Error())
		assert.Equal(t, 1, len(m.ns))
	})

	t.Run("adds multiple different network servers", func(t *testing.T) {
		m := NewManager()

		ns1, err1 := m.Add("server-1")
		ns2, err2 := m.Add("server-2")
		ns3, err3 := m.Add("server-3")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NoError(t, err3)
		assert.NotNil(t, ns1)
		assert.NotNil(t, ns2)
		assert.NotNil(t, ns3)
		assert.Equal(t, 3, len(m.ns))
	})
}

func TestManager_Get(t *testing.T) {
	t.Run("gets existing network server", func(t *testing.T) {
		m := NewManager()
		name := "test-server"
		m.Add(name)

		ns, exists := m.Get(name)

		assert.True(t, exists)
		assert.NotNil(t, ns)
		assert.Equal(t, name, ns.name)
	})

	t.Run("returns false for non-existing network server", func(t *testing.T) {
		m := NewManager()

		ns, exists := m.Get("non-existing")

		assert.False(t, exists)
		assert.Nil(t, ns)
	})

	t.Run("gets correct server from multiple servers", func(t *testing.T) {
		m := NewManager()
		m.Add("server-1")
		m.Add("server-2")
		m.Add("server-3")

		ns, exists := m.Get("server-2")

		assert.True(t, exists)
		assert.NotNil(t, ns)
		assert.Equal(t, "server-2", ns.name)
	})
}

func TestManager_List(t *testing.T) {
	t.Run("returns empty slice for empty manager", func(t *testing.T) {
		m := NewManager()

		list := m.List()

		assert.NotNil(t, list)
		assert.Equal(t, 0, len(list))
	})

	t.Run("returns all network servers info", func(t *testing.T) {
		m := NewManager()
		m.Add("server-1")
		m.Add("server-2")
		m.Add("server-3")

		list := m.List()

		assert.Equal(t, 3, len(list))

		// Verify all servers are in the list
		names := make(map[string]bool)
		for _, info := range list {
			names[info.Name] = true
			assert.Equal(t, 0, info.DeviceCount)
			assert.Equal(t, 0, info.GatewayCount)
		}

		assert.True(t, names["server-1"])
		assert.True(t, names["server-2"])
		assert.True(t, names["server-3"])
	})

	t.Run("list reflects current state after add and remove", func(t *testing.T) {
		m := NewManager()
		m.Add("server-1")
		m.Add("server-2")
		m.Add("server-3")

		list1 := m.List()
		assert.Equal(t, 3, len(list1))

		m.Remove("server-2")

		list2 := m.List()
		assert.Equal(t, 2, len(list2))

		names := make(map[string]bool)
		for _, info := range list2 {
			names[info.Name] = true
		}

		assert.True(t, names["server-1"])
		assert.False(t, names["server-2"])
		assert.True(t, names["server-3"])
	})

	t.Run("returns NetworkServerInfo not pointers", func(t *testing.T) {
		m := NewManager()
		m.Add("test-server")

		list := m.List()

		assert.Equal(t, 1, len(list))
		assert.Equal(t, "test-server", list[0].Name)
		assert.IsType(t, NetworkServerInfo{}, list[0])
	})
}

func TestManager_Remove(t *testing.T) {
	t.Run("removes existing network server", func(t *testing.T) {
		m := NewManager()
		name := "test-server"
		m.Add(name)

		m.Remove(name)

		ns, exists := m.Get(name)
		assert.False(t, exists)
		assert.Nil(t, ns)
		assert.Equal(t, 0, len(m.ns))
	})

	t.Run("removes non-existing network server without error", func(t *testing.T) {
		m := NewManager()

		// Should not panic or error
		m.Remove("non-existing")

		assert.Equal(t, 0, len(m.ns))
	})

	t.Run("removes one server from multiple servers", func(t *testing.T) {
		m := NewManager()
		m.Add("server-1")
		m.Add("server-2")
		m.Add("server-3")

		m.Remove("server-2")

		_, exists1 := m.Get("server-1")
		_, exists2 := m.Get("server-2")
		_, exists3 := m.Get("server-3")

		assert.True(t, exists1)
		assert.False(t, exists2)
		assert.True(t, exists3)
		assert.Equal(t, 2, len(m.ns))
	})
}

func TestManager_Concurrency(t *testing.T) {
	t.Run("concurrent adds are safe", func(t *testing.T) {
		m := NewManager()
		var wg sync.WaitGroup
		numGoroutines := 100

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				name := string(rune('a' + id))
				m.Add(name)
			}(i)
		}

		wg.Wait()

		// Should have added all unique servers
		assert.Equal(t, numGoroutines, len(m.ns))
	})

	t.Run("concurrent reads and writes are safe", func(t *testing.T) {
		m := NewManager()
		var wg sync.WaitGroup
		numOperations := 50

		// Pre-populate with some servers
		for i := 0; i < 10; i++ {
			m.Add(string(rune('a' + i)))
		}

		// Concurrent reads
		wg.Add(numOperations)
		for i := 0; i < numOperations; i++ {
			go func(id int) {
				defer wg.Done()
				name := string(rune('a' + (id % 10)))
				m.Get(name)
			}(i)
		}

		// Concurrent writes
		wg.Add(numOperations)
		for i := 0; i < numOperations; i++ {
			go func(id int) {
				defer wg.Done()
				name := string(rune('k' + id))
				m.Add(name)
			}(i)
		}

		wg.Wait()

		// Should not panic and should have all the original + new servers
		assert.GreaterOrEqual(t, len(m.ns), 10)
	})

	t.Run("concurrent removes are safe", func(t *testing.T) {
		m := NewManager()
		var wg sync.WaitGroup
		numServers := 50

		// Pre-populate
		for i := 0; i < numServers; i++ {
			m.Add(string(rune('a' + i)))
		}

		// Concurrent removes
		wg.Add(numServers)
		for i := 0; i < numServers; i++ {
			go func(id int) {
				defer wg.Done()
				name := string(rune('a' + id))
				m.Remove(name)
			}(i)
		}

		wg.Wait()

		// All should be removed
		assert.Equal(t, 0, len(m.ns))
	})
}
