# LoRaWAN® Simulator API Documentation

This document provides complete API reference for the LoRaWAN® Simulator REST API.

## Base URL

```
http://localhost:2208
```

## Table of Contents

- [Health Check](#health-check)
- [Network Servers](#network-servers)
  - [List Network Servers](#list-network-servers)
  - [Create Network Server](#create-network-server)
  - [Get Network Server](#get-network-server)
  - [Delete Network Server](#delete-network-server)
  - [Sync Network Server](#sync-network-server)
- [Gateways](#gateways)
  - [List Gateways](#list-gateways)
  - [Create Gateway](#create-gateway)
  - [Get Gateway](#get-gateway)
  - [Delete Gateway](#delete-gateway)
  - [Connect Gateway](#connect-gateway)
  - [Disconnect Gateway](#disconnect-gateway)
- [Devices](#devices)
  - [List Devices](#list-devices)
  - [Create Device](#create-device)
  - [Get Device](#get-device)
  - [Delete Device](#delete-device)
  - [Send Join Request](#send-join-request)
  - [Send Uplink](#send-uplink)
- [Network Server Types](#network-server-types)
- [Error Responses](#error-responses)

---

## Health Check

### GET /health

Check if the API is running and healthy.

**Response:**
```json
{
  "status": "ok"
}
```

**Example:**
```bash
curl http://localhost:2208/health
```

---

## Network Servers

Network servers represent LoRaWAN® network server instances. Each network server can have multiple gateways and devices.

### List Network Servers

**GET** `/network-servers`

Returns a list of all network servers.

**Response:**
```json
[
  {
    "name": "localhost",
    "config": {
      "type": "generic"
    },
    "deviceCount": 2,
    "gatewayCount": 1
  },
  {
    "name": "loriot-eu1",
    "config": {
      "type": "loriot",
      "url": "https://eu1.loriot.io",
      "authHeader": "Bearer ..."
    },
    "deviceCount": 0,
    "gatewayCount": 5
  }
]
```

**Example:**
```bash
curl http://localhost:2208/network-servers
```

### Create Network Server

**POST** `/network-servers`

Creates a new network server instance.

**Request Body:**
```json
{
  "name": "my-server",
  "config": {
    "type": "generic"
  }
}
```

**Network Server Types:**

#### Generic (No Integration)
```json
{
  "name": "localhost",
  "config": {
    "type": "generic"
  }
}
```

#### LORIOT Integration
```json
{
  "name": "loriot-eu1",
  "config": {
    "type": "loriot",
    "url": "https://eu1.loriot.io",
    "authHeader": "Bearer YOUR_TOKEN_HERE"
  }
}
```

#### ChirpStack Integration
```json
{
  "name": "chirpstack",
  "config": {
    "type": "chirpstack",
    "url": "https://chirpstack.example.com",
    "apiToken": "YOUR_API_TOKEN"
  }
}
```

#### The Things Network (TTN) Integration
```json
{
  "name": "ttn-eu1",
  "config": {
    "type": "ttn",
    "url": "https://eu1.cloud.thethings.network",
    "apiKey": "YOUR_API_KEY"
  }
}
```

**Response:** `201 Created`
```json
{
  "name": "my-server",
  "config": {
    "type": "generic"
  },
  "deviceCount": 0,
  "gatewayCount": 0
}
```

**Example:**
```bash
curl -X POST http://localhost:2208/network-servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "localhost",
    "config": {
      "type": "generic"
    }
  }'
```

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `409 Conflict` - Network server with this name already exists

### Get Network Server

**GET** `/network-servers/:name`

Returns information about a specific network server.

**Response:**
```json
{
  "name": "localhost",
  "config": {
    "type": "generic"
  },
  "deviceCount": 2,
  "gatewayCount": 1
}
```

**Example:**
```bash
curl http://localhost:2208/network-servers/localhost
```

**Error Responses:**
- `404 Not Found` - Network server not found

### Delete Network Server

**DELETE** `/network-servers/:name`

Deletes a network server and all its associated gateways and devices.

**Response:** `204 No Content`

**Example:**
```bash
curl -X DELETE http://localhost:2208/network-servers/localhost
```

**Error Responses:**
- `404 Not Found` - Network server not found

### Sync Network Server

**POST** `/network-servers/:name/sync`

Synchronizes gateways and devices from the remote network server (only for LORIOT, ChirpStack, and TTN integrations).

This endpoint:
- Fetches gateways from the remote network server
- Adds new gateways that don't exist locally
- Updates gateways with changed discovery URIs
- (Future) Syncs devices as well

**Response:** `204 No Content`

**Example:**
```bash
curl -X POST http://localhost:2208/network-servers/loriot-eu1/sync
```

**Error Responses:**
- `400 Bad Request` - Sync failed (e.g., network error, authentication error)
- `404 Not Found` - Network server not found

**Note:** For generic network servers, this endpoint does nothing and returns success.

---

## Gateways

Gateways simulate LoRa Basics™ Station gateways that connect to network servers via WebSocket.

### List Gateways

**GET** `/network-servers/:name/gateways`

Returns a list of all gateways for a network server.

**Response:**
```json
[
  {
    "eui": "aabbccddeeff0011",
    "discoveryUri": "ws://localhost:3001",
    "discoveryState": "connected",
    "dataState": "connected"
  },
  {
    "eui": "1122334455667788",
    "discoveryUri": "wss://eu1.loriot.io:6001",
    "discoveryState": "disconnected",
    "dataState": "disconnected"
  }
]
```

**States:**
- `disconnected` - Gateway is not connected
- `connecting` - Gateway is attempting to connect
- `connected` - Gateway is connected and operational
- `disconnecting` - Gateway is disconnecting

**Example:**
```bash
curl http://localhost:2208/network-servers/localhost/gateways
```

### Create Gateway

**POST** `/network-servers/:name/gateways`

Creates a new gateway. The gateway is created in a disconnected state.

**Request Body:**
```json
{
  "eui": "AABBCCDDEEFF0011",
  "discoveryUri": "ws://localhost:3001"
}
```

**EUI Format:** 16 hex characters (8 bytes), case-insensitive, with or without dashes/colons

**Discovery URI Format:**
- WebSocket: `ws://host:port` or `ws://host:port/path`
- Secure WebSocket: `wss://host:port` or `wss://host:port/path`

**Response:** `201 Created`
```json
{
  "eui": "aabbccddeeff0011",
  "discoveryUri": "ws://localhost:3001",
  "discoveryState": "disconnected",
  "dataState": "disconnected"
}
```

**Example:**
```bash
curl -X POST http://localhost:2208/network-servers/localhost/gateways \
  -H "Content-Type: application/json" \
  -d '{
    "eui": "AABBCCDDEEFF0011",
    "discoveryUri": "ws://localhost:3001"
  }'
```

**Error Responses:**
- `400 Bad Request` - Invalid EUI format or missing required fields
- `404 Not Found` - Network server not found
- `409 Conflict` - Gateway with this EUI already exists

### Get Gateway

**GET** `/network-servers/:name/gateways/:eui`

Returns information about a specific gateway.

**Response:**
```json
{
  "eui": "aabbccddeeff0011",
  "discoveryUri": "ws://localhost:3001",
  "discoveryState": "connected",
  "dataState": "connected"
}
```

**Example:**
```bash
curl http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011
```

**Error Responses:**
- `400 Bad Request` - Invalid EUI format
- `404 Not Found` - Network server or gateway not found

### Delete Gateway

**DELETE** `/network-servers/:name/gateways/:eui`

Deletes a gateway. If the gateway is connected, it will be disconnected first.

**Response:** `204 No Content`

**Example:**
```bash
curl -X DELETE http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011
```

**Error Responses:**
- `400 Bad Request` - Invalid EUI format
- `404 Not Found` - Network server or gateway not found

### Connect Gateway

**POST** `/network-servers/:name/gateways/:eui/connect`

Connects a gateway to the network server via WebSocket.

The gateway will:
1. Perform discovery handshake to obtain data connection URI
2. Connect to the data WebSocket endpoint
3. Send version information
4. Begin listening for uplink/downlink messages

**Response:** `204 No Content`

**Example:**
```bash
curl -X POST http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011/connect
```

**Error Responses:**
- `400 Bad Request` - Invalid EUI format or gateway already connected
- `404 Not Found` - Network server or gateway not found
- `500 Internal Server Error` - Connection failed

### Disconnect Gateway

**POST** `/network-servers/:name/gateways/:eui/disconnect`

Disconnects a gateway from the network server.

**Response:** `204 No Content`

**Example:**
```bash
curl -X POST http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011/disconnect
```

**Error Responses:**
- `400 Bad Request` - Invalid EUI format or gateway already disconnected
- `404 Not Found` - Network server or gateway not found

---

## Devices

Devices simulate LoRaWAN® end devices that can perform OTAA join and send uplink messages.

### List Devices

**GET** `/network-servers/:name/devices`

Returns a list of all devices for a network server.

**Response:**
```json
[
  {
    "deveui": "0011223344556677",
    "joineui": "0011223344556677",
    "devaddr": "00f627f6",
    "fcntUp": 1,
    "fcntDown": 0
  },
  {
    "deveui": "1122334455667788",
    "joineui": "1122334455667788",
    "devaddr": "00000000",
    "fcntUp": 0,
    "fcntDown": 0
  }
]
```

**DevAddr:** `00000000` indicates device has not joined yet

**Example:**
```bash
curl http://localhost:2208/network-servers/localhost/devices
```

### Create Device

**POST** `/network-servers/:name/devices`

Creates a new device with OTAA credentials.

**Request Body:**
```json
{
  "deveui": "0011223344556677",
  "joineui": "0011223344556677",
  "appkey": "00112233445566770011223344556677",
  "devnonce": 0
}
```

**Field Formats:**
- `deveui`: 16 hex characters (8 bytes)
- `joineui`: 16 hex characters (8 bytes)
- `appkey`: 32 hex characters (16 bytes)
- `devnonce`: Integer (0-65535)

**Response:** `201 Created`
```json
{
  "deveui": "0011223344556677",
  "joineui": "0011223344556677",
  "devaddr": "00000000",
  "fcntUp": 0,
  "fcntDown": 0
}
```

**Example:**
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

**Error Responses:**
- `400 Bad Request` - Invalid format or missing required fields
- `404 Not Found` - Network server not found
- `409 Conflict` - Device with this DevEUI already exists

### Get Device

**GET** `/network-servers/:name/devices/:eui`

Returns information about a specific device.

**Response:**
```json
{
  "deveui": "0011223344556677",
  "joineui": "0011223344556677",
  "devaddr": "00f627f6",
  "fcntUp": 5,
  "fcntDown": 2
}
```

**Example:**
```bash
curl http://localhost:2208/network-servers/localhost/devices/0011223344556677
```

**Error Responses:**
- `400 Bad Request` - Invalid EUI format
- `404 Not Found` - Network server or device not found

### Delete Device

**DELETE** `/network-servers/:name/devices/:eui`

Deletes a device.

**Response:** `204 No Content`

**Example:**
```bash
curl -X DELETE http://localhost:2208/network-servers/localhost/devices/0011223344556677
```

**Error Responses:**
- `400 Bad Request` - Invalid EUI format
- `404 Not Found` - Network server or device not found

### Send Join Request

**POST** `/network-servers/:name/devices/:eui/join`

Sends an OTAA join request from the device to the network server.

The device will:
1. Generate a JoinRequest frame
2. Broadcast it to all connected gateways
3. Wait for a JoinAccept response
4. Process the JoinAccept and derive session keys
5. Update its DevAddr and frame counters

**Response:** `204 No Content`

**Example:**
```bash
curl -X POST http://localhost:2208/network-servers/localhost/devices/0011223344556677/join
```

**Error Responses:**
- `400 Bad Request` - Invalid EUI format or device operation failed
- `404 Not Found` - Network server or device not found

**Console Output Example:**
```
[0011223344556677] broadcasting uplink: 00776655443322110077665544332211000000393a1d36
[pool] propagating uplink to network server localhost
[localhost] propagating uplink to gateway aabbccddeeff0011
[aabbccddeeff0011] data write: {"msgtype":"jreq","MHdr":0,"JoinEui":"00-11-22-33-44-55-66-77",...}
```

### Send Uplink

**POST** `/network-servers/:name/devices/:eui/uplink`

Sends an uplink data message from the device to the network server.

The device must have joined (has a valid DevAddr) before sending uplink.

The device will:
1. Generate an uplink data frame with dummy payload
2. Broadcast it to all connected gateways
3. Wait for potential downlink response

**Response:** `204 No Content`

**Example:**
```bash
curl -X POST http://localhost:2208/network-servers/localhost/devices/0011223344556677/uplink
```

**Error Responses:**
- `400 Bad Request` - Invalid EUI format, device not joined, or operation failed
- `404 Not Found` - Network server or device not found

**Console Output Example:**
```
[0011223344556677] broadcasting uplink: 80f627f600a0000001010203049997a7ab
[pool] propagating uplink to network server localhost
[localhost] propagating uplink to gateway aabbccddeeff0011
[aabbccddeeff0011] data write: {"msgtype":"updf","MHdr":128,"DevAddr":16066550,...}
```

---

## Network Server Types

The simulator supports different network server integrations:

### Generic

No external integration. Gateways and devices are managed manually through the API.

```json
{
  "type": "generic"
}
```

### LORIOT

Integrates with LORIOT network servers to automatically sync gateways.

**Features:**
- Automatic gateway discovery from LORIOT API
- Pagination support for large gateway lists
- Basics Station gateway filtering
- Discovery URI retrieval from network status

**Configuration:**
```json
{
  "type": "loriot",
  "url": "https://eu1.loriot.io",
  "authHeader": "Bearer YOUR_TOKEN_HERE"
}
```

**API Requirements:**
- LORIOT API v1
- Bearer token authentication
- Required endpoints:
  - `GET /1/nwk/status` - Get Basics Station URL and port
  - `GET /1/nwk/gateways?page=X&perPage=100` - List gateways

### ChirpStack

Integrates with ChirpStack network servers (planned).

```json
{
  "type": "chirpstack",
  "url": "https://chirpstack.example.com",
  "apiToken": "YOUR_API_TOKEN"
}
```

**Status:** Implementation planned

### The Things Network (TTN)

Integrates with The Things Network (planned).

```json
{
  "type": "ttn",
  "url": "https://eu1.cloud.thethings.network",
  "apiKey": "YOUR_API_KEY"
}
```

**Status:** Implementation planned

---

## Error Responses

All error responses follow this format:

```json
{
  "message": "error description"
}
```

### HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `204 No Content` - Request successful with no response body
- `400 Bad Request` - Invalid request format or parameters
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `500 Internal Server Error` - Server error
- `504 Gateway Timeout` - Request timeout (default: 5 seconds)

### Common Error Messages

**Network Server Errors:**
- `"network server not found"` - The specified network server doesn't exist
- `"network server already exists"` - A network server with this name already exists

**Gateway Errors:**
- `"gateway not found"` - The specified gateway doesn't exist
- `"gateway already exists"` - A gateway with this EUI already exists
- `"invalid EUI format"` - The EUI is not in valid hex format
- `"gateway is already connected"` - Cannot connect an already connected gateway
- `"gateway is already disconnected"` - Cannot disconnect an already disconnected gateway

**Device Errors:**
- `"device not found"` - The specified device doesn't exist
- `"device already exists"` - A device with this DevEUI already exists
- `"invalid EUI format"` - The DevEUI or JoinEUI is not in valid hex format
- `"invalid key format"` - The AppKey is not in valid hex format
- `"device has not joined"` - Cannot send uplink before device joins

---

## Complete Workflow Example

Here's a complete example showing how to set up and use the simulator:

### 1. Check Health
```bash
curl http://localhost:2208/health
```

### 2. Create Network Server
```bash
curl -X POST http://localhost:2208/network-servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "localhost",
    "config": {
      "type": "generic"
    }
  }'
```

### 3. Add Gateway
```bash
curl -X POST http://localhost:2208/network-servers/localhost/gateways \
  -H "Content-Type: application/json" \
  -d '{
    "eui": "AABBCCDDEEFF0011",
    "discoveryUri": "ws://localhost:3001"
  }'
```

### 4. Connect Gateway
```bash
curl -X POST http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011/connect
```

### 5. Verify Gateway Status
```bash
curl http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011
```

### 6. Add Device
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

### 7. Send Join Request
```bash
curl -X POST http://localhost:2208/network-servers/localhost/devices/0011223344556677/join
```

### 8. Send Uplink Data
```bash
curl -X POST http://localhost:2208/network-servers/localhost/devices/0011223344556677/uplink
```

### 9. Check Device Status
```bash
curl http://localhost:2208/network-servers/localhost/devices/0011223344556677
```

### 10. List All Resources
```bash
# List network servers
curl http://localhost:2208/network-servers

# List gateways
curl http://localhost:2208/network-servers/localhost/gateways

# List devices
curl http://localhost:2208/network-servers/localhost/devices
```

---

## LORIOT Integration Example

Here's how to use the LORIOT integration:

### 1. Create LORIOT Network Server
```bash
curl -X POST http://localhost:2208/network-servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "loriot-eu1",
    "config": {
      "type": "loriot",
      "url": "https://eu1.loriot.io",
      "authHeader": "Bearer YOUR_LORIOT_TOKEN"
    }
  }'
```

**Note:** The network server will automatically sync gateways on creation.

### 2. Manually Sync Gateways (Optional)
```bash
curl -X POST http://localhost:2208/network-servers/loriot-eu1/sync
```

### 3. List Synced Gateways
```bash
curl http://localhost:2208/network-servers/loriot-eu1/gateways
```

### 4. Connect a Synced Gateway
```bash
curl -X POST http://localhost:2208/network-servers/loriot-eu1/gateways/FCC23DFFFE0B9F98/connect
```

**Note:** Only Basics Station gateways are synced from LORIOT. The discovery URI is automatically obtained from the LORIOT network status endpoint.

---

## Rate Limiting

Currently, there is no rate limiting on the API. This may be added in future versions.

## WebSocket Protocol

The simulator uses the LoRa Basics™ Station protocol for gateway communication. For details on the WebSocket message format, refer to the [Basics Station documentation](https://doc.sm.tc/station/).

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/emanuele-dedonatis/lorawan-simulator).
