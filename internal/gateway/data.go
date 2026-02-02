package gateway

import (
	"log"

	"github.com/gorilla/websocket"
)

func (g *Gateway) lnsDataConnect() error {
	// Connecting
	g.mu.Lock()
	g.dataState = StateConnecting
	g.mu.Unlock()
	log.Printf("[%s] data connecting", g.eui)

	conn, _, connErr := websocket.DefaultDialer.Dial(g.dataURI, nil)

	if connErr != nil {
		// Connection error
		log.Printf("[%s] data connection error: %v", g.eui, connErr)
		g.mu.Lock()
		g.dataState = StateDisconnected
		g.mu.Unlock()

		return connErr
	}

	// Connected
	g.mu.Lock()
	g.dataWs = conn
	g.dataState = StateConnected
	g.dataSendCh = make(chan string)
	g.dataDone = make(chan struct{})
	g.mu.Unlock()
	log.Printf("[%s] data connected", g.eui)

	go g.lnsDataReadLoop()
	go g.lnsDataWriteLoop()

	// Send version message to receive router_config
	versionMsg := `{"msgtype":"version","station":"lorawan-simulator","protocol":2}`
	g.lnsDataSend(versionMsg)

	// TODO: create timeout for router_config

	return nil
}

func (g *Gateway) lnsDataReadLoop() {
	defer close(g.dataDone)

	for {
		_, msg, err := g.dataWs.ReadMessage()
		if err != nil {
			log.Printf("[%s] data read error: %v", g.eui, err)
			return
		}
		log.Printf("[%s] data read: %s", g.eui, msg)
	}

}

func (g *Gateway) lnsDataWriteLoop() {
	for msg := range g.dataSendCh {
		err := g.dataWs.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Printf("[%s] data write error: %v", g.eui, err)
			return
		}
		log.Printf("[%s] data write: %s", g.eui, msg)
	}
}

func (g *Gateway) lnsDataSend(message string) {
	g.mu.RLock()
	dataSendCh := g.dataSendCh
	g.mu.RUnlock()

	if dataSendCh == nil {
		log.Printf("[%s] data write error: not allowed", g.eui)
		return
	}
	g.dataSendCh <- message
}

func (g *Gateway) lnsDataDisconnect() error {
	g.mu.Lock()
	g.dataState = StateDisconnecting
	g.mu.Unlock()

	// Don't allow sending messages anymore
	close(g.dataSendCh)

	// Close the connection
	log.Printf("[%s] data disconnecting", g.eui)
	err := g.dataWs.Close()
	if err != nil {
		g.mu.Lock()
		g.dataState = StateDisconnectionError
		g.mu.Unlock()
		log.Printf("[%s] data disconnection error: %v", g.eui, err)
		return err
	}

	// Wait for lnsDataReadLoop termination
	<-g.dataDone

	g.mu.Lock()
	g.dataWs = nil
	g.dataState = StateDisconnected
	g.mu.Unlock()
	log.Printf("[%s] data disconnected", g.eui)

	return nil
}
