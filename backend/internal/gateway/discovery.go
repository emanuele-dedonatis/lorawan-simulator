package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const routerTimeout = 5 * time.Second

type discoveryResponse struct {
	uri string
	err error
}

func (g *Gateway) lnsDiscovery() (string, error) {
	// Connecting
	g.mu.Lock()
	g.discoveryState = StateConnecting
	g.mu.Unlock()

	conn, _, connErr := websocket.DefaultDialer.Dial(g.discoveryURI+"/router-info", nil)

	if connErr != nil {
		// Connection error
		log.Printf("[%s] discovery connection error: %v", g.eui, connErr)
		g.mu.Lock()
		g.discoveryState = StateDisconnected
		g.mu.Unlock()

		return "", connErr
	}

	// Connected
	log.Printf("[%s] discovery connected", g.eui)
	g.mu.Lock()
	g.discoveryState = StateConnected
	g.mu.Unlock()

	// LNS Discovery connection not needed anymore when exiting this function
	defer func() {
		g.mu.Lock()
		g.discoveryState = StateDisconnecting
		g.mu.Unlock()

		err := conn.Close()
		if err != nil {
			g.mu.Lock()
			g.discoveryState = StateDisconnectionError
			g.mu.Unlock()
			log.Printf("[%s] discovery disconnection error: %s", g.eui, err)
			return
		}
		log.Printf("[%s] discovery disconnected", g.eui)
		g.mu.Lock()
		g.discoveryState = StateDisconnected
		g.mu.Unlock()
	}()

	// Send router message
	routerMsg := fmt.Sprintf(`{"router":"%s"}`, formatEUI(g.eui))
	if routerErr := conn.WriteMessage(websocket.TextMessage, []byte(routerMsg)); routerErr != nil {
		log.Printf("[%s] discovery router error: %v", g.eui, routerErr)

		return "", routerErr
	}
	log.Printf("[%s] discovery sent: %s", g.eui, routerMsg)

	// Wait for LNS Data Uri
	timer := time.NewTimer(routerTimeout)

	res := make(chan discoveryResponse, 1)
	go func() {
		defer timer.Stop()

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[%s] discovery response error: %s", g.eui, err)
			res <- discoveryResponse{
				uri: "",
				err: err,
			}
			return
		}

		log.Printf("[%s] discovery response: %s", g.eui, msg)

		// Parse JSON response to extract URI field
		var response struct {
			Router string `json:"router"`
			Muxs   string `json:"muxs"`
			URI    string `json:"uri"`
		}

		if parseErr := json.Unmarshal(msg, &response); parseErr != nil {
			log.Printf("[%s] discovery response parse error: %s", g.eui, parseErr)
			res <- discoveryResponse{
				uri: "",
				err: parseErr,
			}
			return
		}

		res <- discoveryResponse{
			uri: response.URI,
			err: nil,
		}
	}()

	select {
	case r := <-res:
		return r.uri, r.err
	case <-timer.C:
		log.Printf("[%s] discovery response timeout", g.eui)
		return "", errors.New("discovery response timeout")
	}
}
