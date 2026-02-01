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
	discoveryWs    *websocket.Conn
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

	reply := make(chan error, 1)

	if g.dataState == StateConnected {
		reply <- errors.New("already connected")
		g.mu.Unlock()
		return reply
	} else if g.discoveryState != StateDisconnected || g.dataState != StateDisconnected {
		reply <- errors.New("already connecting")
		g.mu.Unlock()
		return reply
	}

	g.mu.Unlock()
	go g.connectDiscoveryWs(reply)

	// TODO:
	// - receive LNS Data URI
	// - disconnect LNS Discovery
	// - connect LNS Data

	return reply
}

func (g *Gateway) connectDiscoveryWs(reply chan error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.discoveryState = StateConnecting

	conn, _, err := websocket.DefaultDialer.Dial(g.discoveryURI, nil)

	if err != nil {
		g.discoveryState = StateDisconnected
		reply <- err
		return
	}

	g.discoveryWs = conn
	g.discoveryState = StateConnected
	reply <- nil
}

func (g *Gateway) DisconnectAsync() <-chan error {
	g.mu.Lock()

	reply := make(chan error, 1)

	if g.discoveryState == StateDisconnected && g.dataState == StateDisconnected {
		reply <- errors.New("already disconnected")
		g.mu.Unlock()
		return reply
	}

	g.mu.Unlock()
	go g.disconnectDiscoveryWs(reply)

	// TODO
	// - disconnect LNS Data

	return reply
}

func (g *Gateway) disconnectDiscoveryWs(reply chan error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.discoveryWs != nil {
		if err := g.discoveryWs.Close(); err != nil {
			reply <- err
			return
		}
	}

	g.discoveryWs = nil
	g.discoveryState = StateDisconnected
	reply <- nil
}
