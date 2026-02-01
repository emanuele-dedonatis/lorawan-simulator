# Quick Start Demo

This guide provides simple curl commands to interact with the LoRaWAN Simulator API.

## Prerequisites

Start the simulator server:
```bash
go run cmd/lorawan-simulator/main.go
```

The API will be available at `http://localhost:8080`

## Basic Operations

### 1. Create a Network Server
```bash
curl -X POST http://localhost:8080/network-servers \
  -H "Content-Type: application/json" \
  -d '{"name":"my-server"}'
```

### 2. List All Network Servers
```bash
curl http://localhost:8080/network-servers
```

### 3. Get Specific Network Server
```bash
curl http://localhost:8080/network-servers/my-server
```

### 4. Delete a Network Server
```bash
curl -X DELETE http://localhost:8080/network-servers/my-server
```