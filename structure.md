lorawan-sim/
├── go.mod
├── cmd/
│   └── simulator/
│       └── main.go
├── internal/
│   ├── server/
│   │   ├── server.go          # NetworkServer struct + public API
│   │   ├── uplink.go          # SendUplink logic
│   │   ├── downlink.go        # HandleDownlink logic
│   │   └── gateways.go        # Gateway registry helpers
│   ├── gateway/
│   │   ├── gateway.go         # Gateway struct, state, public methods
│   │   ├── discovery.go       # CUPS (Discovery Server) flow
│   │   ├── dataserver.go      # Data Server WS lifecycle
│   │   ├── messages.go        # Basics Station JSON structs
│   │   └── websocket.go       # readLoop / writeLoop
│   ├── device/
│   │   └── device.go          # Device struct
│   ├── lorawan/
│   │   ├── uplink.go          # Build uplink (brocaar)
│   │   ├── downlink.go        # MIC validation + decrypt
│   │   └── types.go           # Shared LoRaWAN helpers
│   ├── geo/
│   │   └── distance.go        # Distance calculation
│   ├── api/
│   │   ├── http.go            # HTTP router
│   │   ├── gateways.go        # Gateway endpoints
│   │   └── uplinks.go         # Uplink endpoints
│   └── logging/
│       └── logger.go
└── README.md