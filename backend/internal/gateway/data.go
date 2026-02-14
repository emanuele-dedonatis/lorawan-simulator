package gateway

import (
	"encoding/binary"
	"encoding/hex"
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
	headers := g.headers
	g.mu.Unlock()
	log.Printf("[%s] data connecting", g.eui)

	dialer := websocket.DefaultDialer
	conn, _, connErr := dialer.Dial(g.dataURI, headers)

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
	versionMsg := `{"msgtype":"version","station":"lorawan-simulator","package":"github.com/emanuele-dedonatis/lorawan-simulator","protocol":2}`
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

		// Convert MIC to signed int32
		mic := int32(binary.LittleEndian.Uint32(frame.MIC[:]))

		// TODO: dynamic DR, freq and upinfo
		updfMsg := fmt.Sprintf(`{"msgtype":"jreq","MHdr":%d,"JoinEui":"%s","DevEui":"%s","DevNonce":%d,"MIC":%d,"DR":5,"Freq":868300000,"upinfo":{"rctx":0,"xtime":26740123065958450,"gpstime":0,"rssi":-50,"snr":9}}`,
			mhdr[0],
			formatEUI(joinReq.JoinEUI),
			formatEUI(joinReq.DevEUI),
			joinReq.DevNonce,
			mic,
		)
		return g.send(updfMsg)
	case lorawan.UnconfirmedDataUp, lorawan.ConfirmedDataUp:
		// Type assert MACPayload to lorawan.MACPayload
		macPL, ok := frame.MACPayload.(*lorawan.MACPayload)
		if !ok {
			return errors.New("invalid MAC payload")
		}

		// Convert MHDR to number
		mhdr, err := frame.MHDR.MarshalBinary()
		if err != nil {
			return errors.New("invalid MHDR")
		}

		// Convert DevAddr to signed int32
		devaddr := int32(binary.BigEndian.Uint32(macPL.FHDR.DevAddr[:]))

		// Convert FCtrl struct to byte
		fctrlByte, err := macPL.FHDR.FCtrl.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal FCtrl: %w", err)
		}

		// Convert MIC to signed int32
		mic := int32(binary.LittleEndian.Uint32(frame.MIC[:]))

		// Encode FOpts as hex string
		var fOptsHex string
		if len(macPL.FHDR.FOpts) > 0 {
			var fOptsBytes []byte
			for _, opt := range macPL.FHDR.FOpts {
				optBytes, err := opt.MarshalBinary()
				if err != nil {
					return fmt.Errorf("failed to marshal FOpts: %w", err)
				}
				fOptsBytes = append(fOptsBytes, optBytes...)
			}
			fOptsHex = hex.EncodeToString(fOptsBytes)
		}

		// Encode FRMPayload as hex string
		var frmPayloadHex string
		if len(macPL.FRMPayload) > 0 {
			if dataPayload, ok := macPL.FRMPayload[0].(*lorawan.DataPayload); ok {
				frmPayloadHex = hex.EncodeToString(dataPayload.Bytes)
			}
		}

		// TODO: dynamic DR, freq and upinfo
		updfMsg := fmt.Sprintf(`{"msgtype":"updf","MHdr":%d,"DevAddr":%d,"FCtrl":%d,"FCnt":%d,"FOpts":"%s","FPort":%d,"FRMPayload":"%s","MIC":%d,"DR":5,"Freq":868300000,"upinfo":{"rctx":0,"xtime":26740123065958450,"gpstime":0,"rssi":-50,"snr":9}}`,
			mhdr[0],
			devaddr,
			fctrlByte[0],
			macPL.FHDR.FCnt,
			fOptsHex,
			*macPL.FPort,
			frmPayloadHex,
			mic,
		)
		return g.send(updfMsg)
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

	// Decode hex string to bytes
	pduBytes, err := hex.DecodeString(dnmsg.Pdu)
	if err != nil {
		log.Printf("[%s] failed to decode PDU hex: %v", g.eui, err)
		return
	}

	// Unmarshal bytes into PHYPayload
	var phyPayload lorawan.PHYPayload
	if err := phyPayload.UnmarshalBinary(pduBytes); err != nil {
		log.Printf("[%s] failed to unmarshal PHYPayload: %v", g.eui, err)
		return
	}

	// Broadcast to devices
	g.mu.RLock()
	broadcastCh := g.broadcastDownlink
	g.mu.RUnlock()

	if broadcastCh != nil {
		log.Printf("[%s] broadcasting downlink", g.eui)
		go func() {
			broadcastCh <- phyPayload
		}()
	}
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
