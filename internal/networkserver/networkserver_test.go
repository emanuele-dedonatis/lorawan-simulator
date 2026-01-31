package networkserver

import (
	"testing"

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
