# LoRaWAN Simulator API Documentation

## Overview

The LoRaWAN Simulator provides a REST API to manage network servers, gateways, and devices in a simulated LoRaWAN environment. The API allows you to:

- Create and manage multiple network servers
- Add and configure LoRaWAN gateways
- Register and control end devices
- Simulate OTAA join procedures
- Send uplink data messages

**Base URL:** `http://localhost:2208`

**Content-Type:** `application/json`

## Getting Started

### Running the Simulator

#### Using Docker Compose (Recommended)

```bash
docker-compose up -d
```

#### Using Go

```bash
go run ./cmd/lorawan-simulator/main.go
```

See [README.md](README.md) for detailed installation and configuration instructions.

---

## Table of Contents

- [Health Check](#health-check)
- [Network Servers](#network-servers)
- [Gateways](#gateways)
- [Devices](#devices)
- [Error Responses](#error-responses)
- [Examples](#examples)

---

## Health Check

### Get Health Status

Simple endpoint to verify the API is running and responsive.

```http
GET /health
```

**Response:** `200 OK`

```json
{
  "status": "ok"
}
```

**Use Cases:**
- Docker/Kubernetes health checks
- Load balancer health probes
- Monitoring systems
- Uptime checks

---

## Network Servers

Network servers are the top-level entities that manage gateways and devices.

### List All Network Servers

```http
GET /network-servers
```

**Response:** `200 OK`

```json
[
  {
    "name": "localhost",
    "deviceCount": 2,
    "gatewayCount": 1
  }
]
```

### Create Network Server

```http
POST /network-servers
```

**Request Body:**

```json
{
  "name": "localhost"
}
```

**Response:** `201 Created`

```json
{
  "name": "localhost",
  "deviceCount": 0,
  "gatewayCount": 0
}
```

### Get Network Server Details

```http
GET /network-servers/:name
```

**Parameters:**
- `name` (path) - Network server name

**Response:** `200 OK`

```json
{
  "name": "localhost",
  "deviceCount": 2,
  "gatewayCount": 1
}
```

### Delete Network Server

```http
DELETE /network-servers/:name
```

**Parameters:**
- `name` (path) - Network server name

**Response:** `204 No Content`

---

## Gateways

Gateways connect as LoRa Basics™ Station and forward messages between devices and the network server.

### List All Gateways

```http
GET /network-servers/:name/gateways
```

**Parameters:**
- `name` (path) - Network server name

**Response:** `200 OK`

```json
[
  {
    "eui": "aabbccddeeff0011",
    "discoveryuri": "ws://localhost:3001",
    "discoverystate": 2,
    "datastate": 2
  }
]
```

**Gateway States:**
- `0` - Disconnected
- `1` - Connecting
- `2` - Connected
- `3` - Disconnecting
- `4` - Disconnection Error

### Add Gateway

```http
POST /network-servers/:name/gateways
```

**Parameters:**
- `name` (path) - Network server name

**Request Body:**

```json
{
  "eui": "AABBCCDDEEFF0011",
  "discoveryUri": "ws://localhost:3001"
}
```

**Field Descriptions:**
- `eui` - Gateway EUI-64 identifier (16 hex characters)
- `discoveryUri` - WebSocket URL for LNS Discovery endpoint

**Response:** `201 Created`

```json
{
  "eui": "aabbccddeeff0011",
  "discoveryuri": "ws://localhost:3001",
  "discoverystate": 0,
  "datastate": 0
}
```

### Get Gateway Details

```http
GET /network-servers/:name/gateways/:eui
```

**Parameters:**
- `name` (path) - Network server name
- `eui` (path) - Gateway EUI-64 (16 hex characters)

**Response:** `200 OK`

```json
{
  "eui": "aabbccddeeff0011",
  "discoveryuri": "ws://localhost:3001",
  "discoverystate": 2,
  "datastate": 2
}
```

### Connect Gateway

Establishes WebSocket connection.

```http
POST /network-servers/:name/gateways/:eui/connect
```

**Parameters:**
- `name` (path) - Network server name
- `eui` (path) - Gateway EUI-64

**Response:** `204 No Content`

### Disconnect Gateway

Closes WebSocket connection.

```http
POST /network-servers/:name/gateways/:eui/disconnect
```

**Parameters:**
- `name` (path) - Network server name
- `eui` (path) - Gateway EUI-64

**Response:** `204 No Content`

### Delete Gateway

```http
DELETE /network-servers/:name/gateways/:eui
```

**Parameters:**
- `name` (path) - Network server name
- `eui` (path) - Gateway EUI-64

**Response:** `204 No Content`

---

## Devices

End devices that can perform OTAA joins and send uplink messages.

### List All Devices

```http
GET /network-servers/:name/devices
```

**Parameters:**
- `name` (path) - Network server name

**Response:** `200 OK`

```json
[
  {
    "deveui": "0011223344556677",
    "joineui": "0011223344556677",
    "appkey": "00112233445566770011223344556677",
    "devnonce": 69,
    "devaddr": "00f527f6",
    "appskey": "d3eec9397cd8af11e149ecd2fde6b531",
    "nwkskey": "d3eec9397cd8af11e149ecd2fde6b531",
    "fcntup": 2,
    "fcntdn": 0
  }
]
```

### Add Device

```http
POST /network-servers/:name/devices
```

**Parameters:**
- `name` (path) - Network server name

**Request Body:**

```json
{
  "deveui": "0011223344556677",
  "joineui": "0011223344556677",
  "appkey": "00112233445566770011223344556677",
  "devnonce": 68
}
```

**Field Descriptions:**
- `deveui` - Device EUI-64 identifier (16 hex characters)
- `joineui` - Join EUI / Application EUI (16 hex characters)
- `appkey` - Application Key for OTAA (32 hex characters)
- `devnonce` - Device nonce for join replay protection (0-65535)

**Response:** `201 Created`

```json
{
  "deveui": "0011223344556677",
  "joineui": "0011223344556677",
  "appkey": "00112233445566770011223344556677",
  "devnonce": 68,
  "devaddr": "00000000",
  "appskey": "00000000000000000000000000000000",
  "nwkskey": "00000000000000000000000000000000",
  "fcntup": 0,
  "fcntdn": 0
}
```

### Get Device Details

```http
GET /network-servers/:name/devices/:eui
```

**Parameters:**
- `name` (path) - Network server name
- `eui` (path) - Device EUI-64

**Response:** `200 OK`

```json
{
  "deveui": "0011223344556677",
  "joineui": "0011223344556677",
  "appkey": "00112233445566770011223344556677",
  "devnonce": 69,
  "devaddr": "00f527f6",
  "appskey": "d3eec9397cd8af11e149ecd2fde6b531",
  "nwkskey": "d3eec9397cd8af11e149ecd2fde6b531",
  "fcntup": 2,
  "fcntdn": 0
}
```

### Send Join Request

Initiates an OTAA join procedure for the device.

```http
POST /network-servers/:name/devices/:eui/join
```

**Parameters:**
- `name` (path) - Network server name
- `eui` (path) - Device EUI-64

**Response:** `204 No Content`

**Notes:**
- Device must be added first
- Gateway must be connected
- DevNonce is automatically incremented
- On successful join, device receives DevAddr and session keys

### Send Uplink Message

Sends a confirmed uplink data message from the device.

```http
POST /network-servers/:name/devices/:eui/uplink
```

**Parameters:**
- `name` (path) - Network server name
- `eui` (path) - Device EUI-64

**Response:** `204 No Content`

**Notes:**
- Device must be joined first (have valid DevAddr and session keys)
- Sends 4 bytes payload: `[0x01, 0x02, 0x03, 0x04]`
- Uses FPort 1
- Payload is encrypted with AppSKey
- MIC is calculated with NwkSKey
- FCntUp is automatically incremented

### Delete Device

```http
DELETE /network-servers/:name/devices/:eui
```

**Parameters:**
- `name` (path) - Network server name
- `eui` (path) - Device EUI-64

**Response:** `204 No Content`

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request

Invalid request format or parameters.

```json
{
  "message": "invalid EUI format"
}
```

### 404 Not Found

Resource does not exist.

```json
{
  "message": "device not found"
}
```

### 409 Conflict

Resource already exists.

```json
{
  "message": "device already exists"
}
```

### 500 Internal Server Error

Server-side error occurred.

```json
{
  "message": "internal server error"
}
```

### 504 Gateway Timeout

Request took too long to process (timeout: 5 seconds).

```json
{
  "message": "request timeout"
}
```

---

## Examples

### Complete Flow: Setup and Join

Here's a complete example of setting up a network server, gateway, device, and performing an OTAA join:

#### 1. Create Network Server

```bash
curl -X POST http://localhost:2208/network-servers \
  -H "Content-Type: application/json" \
  -d '{"name":"localhost"}'
```

#### 2. Add Gateway

```bash
curl -X POST http://localhost:2208/network-servers/localhost/gateways \
  -H "Content-Type: application/json" \
  -d '{
    "eui": "AABBCCDDEEFF0011",
    "discoveryUri": "ws://localhost:3001"
  }'
```

#### 3. Connect Gateway

```bash
curl -X POST http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011/connect
```

#### 4. Add Device

```bash
curl -X POST http://localhost:2208/network-servers/localhost/devices \
  -H "Content-Type: application/json" \
  -d '{
    "deveui": "0011223344556677",
    "joineui": "0011223344556677",
    "appkey": "00112233445566770011223344556677",
    "devnonce": 0
  }'
```

#### 5. Send Join Request

```bash
curl -X POST http://localhost:2208/network-servers/localhost/devices/0011223344556677/join
```

#### 6. Check Device Status (verify join succeeded)

```bash
curl http://localhost:2208/network-servers/localhost/devices/0011223344556677
```

Expected response shows non-zero DevAddr and session keys:

```json
{
  "deveui": "0011223344556677",
  "devaddr": "00f527f6",
  "appskey": "d3eec9397cd8af11e149ecd2fde6b531",
  "nwkskey": "d3eec9397cd8af11e149ecd2fde6b531",
  "fcntup": 0,
  "fcntdn": 0
}
```

#### 7. Send Uplink Data

```bash
curl -X POST http://localhost:2208/network-servers/localhost/devices/0011223344556677/uplink
```

#### 8. Send Another Uplink (FCntUp will increment)

```bash
curl -X POST http://localhost:2208/network-servers/localhost/devices/0011223344556677/uplink
```

### Cleanup

```bash
# Delete device
curl -X DELETE http://localhost:2208/network-servers/localhost/devices/0011223344556677

# Disconnect gateway
curl -X POST http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011/disconnect

# Delete gateway
curl -X DELETE http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011

# Delete network server
curl -X DELETE http://localhost:2208/network-servers/localhost
```

### Script Example

You can use the provided `test-flow.sh` script for automated testing:

```bash
# Run the complete flow
./test-flow.sh

# Run cleanup
./test-flow.sh clean
```

---

## License

See [LICENSE](LICENSE) file for details.
