package gateway

import (
	"errors"
	"sync"

	"github.com/brocaar/lorawan"
)

type Gateway struct {
	eui            lorawan.EUI64
	discoveryURI   string
	discoveryState State
	dataURI        string
	dataState      State
	mu             sync.RWMutex
}

type GatewayInfo struct {
	EUI            lorawan.EUI64 `json:"eui"`
	DiscoveryURI   string        `json:"discoveryUri"`
	DiscoveryState State         `json:"discoveryState"`
	DataURI        string        `json:"dataUri"`
	DataState      State         `json:"dataState"`
}

func New(EUI lorawan.EUI64, discoveryURI string) *Gateway {
	return &Gateway{
		eui:            EUI,
		discoveryURI:   discoveryURI,
		discoveryState: StateDisconnected,
		dataURI:        "",
		dataState:      StateDisconnected,
	}
}

func (g *Gateway) GetInfo() GatewayInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return GatewayInfo{
		EUI:            g.eui,
		DiscoveryURI:   g.discoveryURI,
		DiscoveryState: g.discoveryState,
		DataURI:        g.dataURI,
		DataState:      g.dataState,
	}
}

func (g *Gateway) ConnectAsync() <-chan error {
	g.mu.Lock()
	defer g.mu.Unlock()

	reply := make(chan error, 1)

	if g.dataState == StateConnected {
		reply <- errors.New("already connected")
		return reply
	} else if g.discoveryState != StateDisconnected && g.dataState != StateDisconnected {
		reply <- errors.New("already connecting")
		return reply
	}

	// TODO:
	// - connect to LNS Discovery
	// - receive LNS Data URI
	// - disconnect LNS Discovery
	// - connect LNS Data

	g.dataState = StateConnected

	reply <- nil

	return reply
}

func (g *Gateway) DisconnectAsync() <-chan error {
	g.mu.Lock()
	defer g.mu.Unlock()

	reply := make(chan error, 1)

	if g.discoveryState == StateDisconnected && g.dataState == StateDisconnected {
		reply <- errors.New("already disconnected")
		return reply
	}

	// TODO
	// - disconnect both LNS Discovery and LNS Data

	g.discoveryState = StateDisconnected
	g.dataState = StateDisconnected

	reply <- nil

	return reply
}
