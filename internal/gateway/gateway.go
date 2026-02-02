package gateway

import (
	"errors"
	"sync"

	"github.com/brocaar/lorawan"
	"github.com/gorilla/websocket"
)

type Gateway struct {
	eui            lorawan.EUI64
	discoveryURI   string
	discoveryState State
	dataURI        string
	dataState      State
	dataWs         *websocket.Conn
	dataDone       chan struct{}
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

func (g *Gateway) Connect() error {
	g.mu.RLock()

	if g.dataState == StateConnected {
		g.mu.RUnlock()
		return errors.New("already connected")
	} else if g.discoveryState != StateDisconnected || g.dataState != StateDisconnected {
		g.mu.RUnlock()
		return errors.New("already connecting")
	}
	g.mu.RUnlock()

	uri, err := g.lnsDiscovery()

	g.mu.Lock()
	g.dataURI = uri
	g.mu.Unlock()

	return err
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
	// - disconnect LNS Data

	reply <- errors.New("not yet implemented")
	return reply
}
