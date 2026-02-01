package gateway

import (
	"sync"

	"github.com/brocaar/lorawan"
)

type Gateway struct {
	eui          lorawan.EUI64
	discoveryURI string
	state        State
	mu           sync.RWMutex
}

type GatewayInfo struct {
	EUI          lorawan.EUI64 `json:"eui"`
	DiscoveryURI string        `json:"discoveryUri"`
	State        State         `json:"state"`
}

func New(EUI lorawan.EUI64, discoveryURI string) *Gateway {
	return &Gateway{
		eui:          EUI,
		discoveryURI: discoveryURI,
		state:        StateDisconnected,
	}
}

func (g *Gateway) GetInfo() GatewayInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return GatewayInfo{
		EUI:          g.eui,
		DiscoveryURI: g.discoveryURI,
		State:        g.state,
	}
}

func (g *Gateway) Connect() {
	g.mu.Lock()
	defer g.mu.Unlock()

	// TODO

	g.state = StateDataConnected
}

func (g *Gateway) Disconnect() {
	g.mu.Lock()
	defer g.mu.Unlock()

	// TODO

	g.state = StateDisconnected
}
