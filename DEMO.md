# Quick Start Demo

This guide provides simple curl commands to interact with the LoRaWAN Simulator API.

## Prerequisites

Start the simulator server:
```bash
go run cmd/lorawan-simulator/main.go
```

The API will be available at `http://localhost:8080`

## Network Server Operations

### Add a Network Server
```bash
curl -X POST http://localhost:8080/network-servers \
  -H "Content-Type: application/json" \
  -d '{"name":"my-lns"}'
```

### List All Network Servers
```bash
curl http://localhost:8080/network-servers
```

### Get Specific Network Server
```bash
curl http://localhost:8080/network-servers/my-lns
```

### DeleRemovete a Network Server
```bash
curl -X DELETE http://localhost:8080/network-servers/my-lns
```

## Gateway Operations

### Add a Gateway
```bash
curl -X POST http://localhost:8080/network-servers/my-lns/gateways \
  -H "Content-Type: application/json" \
  -d '{"eui":"AABBCCDDEEFF0011", "discoveryUri":"wss://localhost:1234"}'
```

### List Network Server's Gateways
```bash
curl http://localhost:8080/network-servers/my-lns/gateways
```

### Get Specific Network Server's Gateway
```bash
curl http://localhost:8080/network-servers/my-lns/gateways/AABBCCDDEEFF0011
```

### Remove a Network Server
```bash
curl -X DELETE http://localhost:8080/network-servers/my-lns/gateways/AABBCCDDEEFF0011
```