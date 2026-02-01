# Quick Start Demo

This guide provides simple curl commands to interact with the LoRaWAN Simulator API.

## Prerequisites

Start the simulator server:
```bash
go run cmd/lorawan-simulator/main.go
```

The API will be available at `http://localhost:2208`

## Network Server Operations

### Add a Network Server
```bash
curl -X POST http://localhost:2208/network-servers \
  -H "Content-Type: application/json" \
  -d '{"name":"localhost"}'
```

### List All Network Servers
```bash
curl http://localhost:2208/network-servers
```

### Get Specific Network Server
```bash
curl http://localhost:2208/network-servers/localhost
```

### Remove a Network Server
```bash
curl -X DELETE http://localhost:2208/network-servers/localhost
```

## Gateway Operations

### Add a Gateway
```bash
curl -X POST http://localhost:2208/network-servers/localhost/gateways \
  -H "Content-Type: application/json" \
  -d '{"eui":"AABBCCDDEEFF0011", "discoveryUri":"ws://localhost:3001"}'
```

### List Network Server's Gateways
```bash
curl http://localhost:2208/network-servers/localhost/gateways
```

### Get Specific Network Server's Gateway
```bash
curl http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011
```

### Remove a Gateway
```bash
curl -X DELETE http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011
```

### Connect a Gateway
```bash
curl -X POST http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011/connect
```

### Disconnect a Gateway
```bash
curl -X POST http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011/disconnect
```