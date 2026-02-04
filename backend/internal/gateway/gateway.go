package gateway

import (
	"errors"
	"sync"

	"github.com/brocaar/lorawan"
	"github.com/gorilla/websocket"
)

type Gateway struct {
	eui               lorawan.EUI64
	discoveryURI      string
	discoveryState    State
	dataURI           string
	dataState         State
	dataWs            *websocket.Conn
	dataDone          chan struct{}
	dataSendCh        chan string
	mu                sync.RWMutex
	broadcastDownlink chan<- lorawan.PHYPayload
}

type GatewayInfo struct {
	EUI            lorawan.EUI64 `json:"eui"`
	DiscoveryURI   string        `json:"discoveryUri"`
	DiscoveryState string        `json:"discoveryState"`
	DataURI        string        `json:"dataUri"`
	DataState      string        `json:"dataState"`
}

func New(broadcastDownlink chan<- lorawan.PHYPayload, EUI lorawan.EUI64, discoveryURI string) *Gateway {
	return &Gateway{
		eui:               EUI,
		discoveryURI:      discoveryURI,
		discoveryState:    StateDisconnected,
		dataURI:           "",
		dataState:         StateDisconnected,
		broadcastDownlink: broadcastDownlink,
	}
}

func (g *Gateway) GetInfo() GatewayInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return GatewayInfo{
		EUI:            g.eui,
		DiscoveryURI:   g.discoveryURI,
		DiscoveryState: g.discoveryState.String(),
		DataURI:        g.dataURI,
		DataState:      g.dataState.String(),
	}
}

func (g *Gateway) Connect() error {
	g.mu.RLock()

	// Check if already connected to LNS Data
	if g.dataState == StateConnected {
		g.mu.RUnlock()
		return errors.New("already connected")
	}

	// Check if connection is in progress
	if g.discoveryState == StateConnecting || g.dataState == StateConnecting {
		g.mu.RUnlock()
		return errors.New("already connecting")
	}
	g.mu.RUnlock()

	// Get LNS Data URI from LNS Discovery
	uri, discoveryErr := g.lnsDiscovery()
	if discoveryErr != nil {
		return discoveryErr
	}
	g.mu.Lock()
	g.dataURI = uri
	g.mu.Unlock()

	// Connect to LNS Data
	dataErr := g.lnsDataConnect()

	return dataErr
}

func (g *Gateway) Disconnect() error {
	g.mu.RLock()
	if g.discoveryState == StateDisconnected && g.dataState == StateDisconnected {
		g.mu.RUnlock()
		return errors.New("already disconnected")
	}
	g.mu.RUnlock()

	return g.lnsDataDisconnect()
}
