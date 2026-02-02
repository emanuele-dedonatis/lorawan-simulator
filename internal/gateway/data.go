package gateway

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/brocaar/lorawan"
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
	g.send(versionMsg)

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

		go g.parseIncomingMessage(string(msg))
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

func (g *Gateway) send(message string) error {
	g.mu.RLock()
	dataSendCh := g.dataSendCh
	g.mu.RUnlock()

	if dataSendCh == nil {
		log.Printf("[%s] data write error: not allowed", g.eui)
		return errors.New("not allowed")
	}
	dataSendCh <- message

	return nil
}

func (g *Gateway) Forward(frame lorawan.PHYPayload) error {
	switch frame.MHDR.MType {
	case lorawan.JoinRequest:
		// Type assert MACPayload to JoinRequestPayload
		joinReq, ok := frame.MACPayload.(*lorawan.JoinRequestPayload)
		if !ok {
			return errors.New("invalid join request payload")
		}

		// Convert MHDR to number
		mhdr, err := frame.MHDR.MarshalBinary()
		if err != nil {
			return errors.New("invalid MHDR")
		}

		// Convert MIC to number
		mic := binary.LittleEndian.Uint32(frame.MIC[:])

		// TODO: dynamic DR, freq and upinfo
		updfMsg := fmt.Sprintf(`{"msgtype":"jreq","MHdr":%d,"JoinEui":"%s","DevEui":"%s","DevNonce":%d,"MIC":%d,"DR":5,"Freq":868300000,"upinfo":{"rctx":0,"xtime":26740123065958450,"gpstime":0,"rssi":-50,"snr":9}}`,
			mhdr[0],
			formatEUI(joinReq.JoinEUI),
			formatEUI(joinReq.DevEUI),
			joinReq.DevNonce,
			mic,
		)
		return g.send(updfMsg)
	case lorawan.UnconfirmedDataUp:
		// TODO: implement unconfirmed uplink
		return errors.New("unconfirmed uplink not implemented")
	case lorawan.ConfirmedDataUp:
		// TODO: implement confirmed uplink
		return errors.New("confirmed uplink not implemented")
	default:
		return errors.New("unsupported uplink message type")
	}
}

func (g *Gateway) parseIncomingMessage(msg string) {
	// Parse JSON to extract msgtype
	var baseMsg struct {
		MsgType string `json:"msgtype"`
	}

	if err := json.Unmarshal([]byte(msg), &baseMsg); err != nil {
		log.Printf("[%s] failed to parse message: %v", g.eui, err)
		return
	}

	// Switch based on msgtype
	switch baseMsg.MsgType {
	case "dnmsg":
		g.handleDownlinkMessage(msg)
	default:
		log.Printf("[%s] unknown msgtype: %s", g.eui, baseMsg.MsgType)
	}
}

func (g *Gateway) handleDownlinkMessage(msg string) {
	var dnmsg struct {
		DevEui string `json:"DevEui"`
		Pdu    string `json:"pdu"`
	}

	if err := json.Unmarshal([]byte(msg), &dnmsg); err != nil {
		log.Printf("[%s] failed to parse message: %v", g.eui, err)
		return
	}

	log.Printf("[%s] downlink message for DevEui %s", g.eui, dnmsg.DevEui)

	// TODO: dispatch to corresponding devices
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
