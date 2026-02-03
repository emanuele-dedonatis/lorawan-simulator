# LoRaWAN Simulator

A flexible LoRaWAN network simulator to simulate multiple network servers, gateways, and end devices for testing LoRaWAN applications without physical hardware.

## Features

- ✅ **Multiple Network Servers** - Manage multiple network server instances
- ✅ **Gateway Simulation** - Simulate LoRaWAN gateways with WebSocket connections
- ✅ **Device Simulation** - Simulate end devices with OTAA join and uplink capabilities
- ✅ **REST API** - Complete HTTP API for managing simulated entities
- ✅ **LoRaWAN 1.0.x** - Full protocol support with encryption and MIC validation
- ✅ **Docker Support** - Easy deployment with Docker and docker compose

### Coming Soon

- **MAC Commands Handling** - Full support for LoRaWAN MAC commands processing
- **Custom Radio Parameters** - Configurable spreading factor, bandwidth, and frequency settings
- **Device and Gateway Channel Plans** - Support for regional channel plans and custom configurations
- **Geolocation Broadcast** - Simulate GPS coordinates and location data
- **GUI** - Web-based graphical user interface for easier management
- **Class B and Class C Support** - Beyond Class A device simulation

## Quick Start

### Using Docker Compose (Recommended)

1. **Start the simulator:**

```bash
docker compose up
```

2. **Verify it's running:**

```bash
curl http://localhost:2208/health
```

### Using Go

1. **Install dependencies:**

```bash
go mod download
```

2. **Run the simulator:**

```bash
go run ./cmd/lorawan-simulator/main.go
```

3. **Verify it's running:**

```bash
curl http://localhost:2208/health
```

## API Documentation

See [API.md](API.md) for complete API reference with examples.

### Quick Example

```bash
# Create a network server
curl -X POST http://localhost:2208/network-servers \
  -H "Content-Type: application/json" \
  -d '{"name":"localhost"}'

# Add a gateway
curl -X POST http://localhost:2208/network-servers/localhost/gateways \
  -H "Content-Type: application/json" \
  -d '{
    "eui": "AABBCCDDEEFF0011",
    "discoveryUri": "ws://localhost:3001"
  }'

# Connect gateway
curl -X POST http://localhost:2208/network-servers/localhost/gateways/AABBCCDDEEFF0011/connect

# Add a device
curl -X POST http://localhost:2208/network-servers/localhost/devices \
  -H "Content-Type: application/json" \
  -d '{
    "deveui": "0011223344556677",
    "joineui": "0011223344556677",
    "appkey": "00112233445566770011223344556677",
    "devnonce": 0
  }'

# Send join request
curl -X POST http://localhost:2208/network-servers/localhost/devices/0011223344556677/join

# Send uplink data
curl -X POST http://localhost:2208/network-servers/localhost/devices/0011223344556677/uplink
```

## Console Output Examples

The simulator provides detailed logging of all message exchanges and LoRaWAN frame processing. Here are examples from a complete test flow.

### Starting the Simulator

When you start the simulator, you'll see the registered API endpoints:

```
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /health                   --> ...healthCheck
[GIN-debug] GET    /network-servers          --> ...getNetworkServers
[GIN-debug] POST   /network-servers          --> ...postNetworkServer
[GIN-debug] GET    /network-servers/:name    --> ...getNetworkServersByName
[GIN-debug] DELETE /network-servers/:name    --> ...delNetworkServer
[GIN-debug] GET    /network-servers/:name/gateways --> ...getGateways
[GIN-debug] POST   /network-servers/:name/gateways --> ...postGateway
[GIN-debug] GET    /network-servers/:name/gateways/:eui --> ...getGatewayByEUI
[GIN-debug] DELETE /network-servers/:name/gateways/:eui --> ...delGateway
[GIN-debug] POST   /network-servers/:name/gateways/:eui/connect --> ...connectGateway
[GIN-debug] POST   /network-servers/:name/gateways/:eui/disconnect --> ...disconnectGateway
[GIN-debug] GET    /network-servers/:name/devices --> ...getDevices
[GIN-debug] POST   /network-servers/:name/devices --> ...postDevice
[GIN-debug] GET    /network-servers/:name/devices/:eui --> ...getDeviceByEUI
[GIN-debug] DELETE /network-servers/:name/devices/:eui --> ...delDevice
[GIN-debug] POST   /network-servers/:name/devices/:eui/join --> ...sendDeviceJoinRequest
[GIN-debug] POST   /network-servers/:name/devices/:eui/uplink --> ...sendDeviceUplink

[GIN-debug] Listening and serving HTTP on 0.0.0.0:2208
```

### API Request Logging

All API requests are logged with timing information:

```
[GIN] 2026/02/03 - 08:06:26 | 200 |     173.754µs |             ::1 | GET      "/health"
[GIN] 2026/02/03 - 08:06:26 | 201 |     159.676µs |             ::1 | POST     "/network-servers"
[GIN] 2026/02/03 - 08:06:26 | 201 |     150.255µs |             ::1 | POST     "/network-servers/localhost/gateways"
[GIN] 2026/02/03 - 08:06:26 | 204 |    2.459463ms |             ::1 | POST     "/network-servers/localhost/gateways/AABBCCDDEEFF0011/connect"
[GIN] 2026/02/03 - 08:06:26 | 200 |      55.802µs |             ::1 | GET      "/network-servers/localhost/gateways/AABBCCDDEEFF0011"
[GIN] 2026/02/03 - 08:06:26 | 201 |     152.089µs |             ::1 | POST     "/network-servers/localhost/devices"
[GIN] 2026/02/03 - 08:06:26 | 204 |      61.732µs |             ::1 | POST     "/network-servers/localhost/devices/0011223344556677/join"
```

### Gateway WebSocket Connection

When a gateway connects to the network server, you'll see the WebSocket handshake:

```
2026/02/03 08:06:26 [aabbccddeeff0011] discovery connected
2026/02/03 08:06:26 [aabbccddeeff0011] discovery sent: {"router":"aa-bb-cc-dd-ee-ff-00-11"}
2026/02/03 08:06:26 [aabbccddeeff0011] discovery response: {"router":"aabb:ccdd:eeff:0011","muxs":"aabb:ccdd:eeff:0011","uri":"ws://localhost:3001/gateway/aabbccddeeff0011"}
2026/02/03 08:06:26 [aabbccddeeff0011] discovery disconnected
2026/02/03 08:06:26 [aabbccddeeff0011] data connecting
2026/02/03 08:06:26 [aabbccddeeff0011] data connected
2026/02/03 08:06:26 [aabbccddeeff0011] data write: {"msgtype":"version","station":"lorawan-simulator","protocol":2}
```

### Router Configuration

The gateway receives configuration from the network server:

```
2026/02/03 08:06:26 [aabbccddeeff0011] data read: {"msgtype":"router_config","NetID":null,"JoinEui":null,"region":"EU863","hwspec":"sx1301/1","freq_range":[863000000,870000000],...}
2026/02/03 08:06:26 [aabbccddeeff0011] unknown msgtype: router_config
```

### OTAA Join Request

When a device sends a join request, you'll see the frame broadcast and WebSocket message:

```
2026/02/03 08:06:26 [0011223344556677] broadcasting uplink: 00776655443322110077665544332211000000393a1d36
2026/02/03 08:06:26 [pool] propagating uplink to network server localhost
2026/02/03 08:06:26 [localhost] propagating uplink to gateway aabbccddeeff0011
2026/02/03 08:06:26 [aabbccddeeff0011] data write: {"msgtype":"jreq","MHdr":0,"JoinEui":"00-11-22-33-44-55-66-77","DevEui":"00-11-22-33-44-55-66-77","DevNonce":0,"MIC":907885113,"DR":5,"Freq":868300000,"upinfo":{"rctx":0,"xtime":26740123065958450,"gpstime":0,"rssi":-50,"snr":9}}
```

### Join Accept Response

When the network server sends a join accept downlink:

```
2026/02/03 08:06:27 [aabbccddeeff0011] data read: {"msgtype":"dntxed","DevEui":"01-01-01-01-01-01-01-01","diid":5768,"pdu":"209c8f0f1a8f7c2ed3af2e4a8c7e95c9cb63fe3f2b","rctx":0,"xtime":26740123065958450}
2026/02/03 08:06:27 [localhost] received downlink message: 209c8f0f1a8f7c2ed3af2e4a8c7e95c9cb63fe3f2b
2026/02/03 08:06:27 [localhost] propagating downlink to device 0101010101010101
2026/02/03 08:06:27 [0101010101010101] downlink FCnt 0 - FPort: 0 - FRMPayload: 1a8f7c2ed3af2e4a8c7e95c9cb63fe3f
```

### Uplink Data Message

When a device sends data uplink:

```
2026/02/03 08:06:27 [0011223344556677] broadcasting uplink: 80f627f600a0000001010203049997a7ab
2026/02/03 08:06:27 [pool] propagating uplink to network server localhost
2026/02/03 08:06:27 [localhost] propagating uplink to gateway aabbccddeeff0011
2026/02/03 08:06:27 [aabbccddeeff0011] data write: {"msgtype":"updf","MHdr":128,"DevAddr":16066550,"FCtrl":0,"FCnt":0,"FOpts":"","FPort":1,"FRMPayload":"9997a7ab","MIC":-1974718544,"DR":5,"Freq":868300000,"upinfo":{"rctx":0,"xtime":26740123065958450,"rssi":-50,"snr":9}}
```

### Downlink Data Message

When receiving a downlink from the network server:

```
2026/02/03 08:06:27 [aabbccddeeff0011] data read: {"msgtype":"dnmsg","DevEui":"01-01-01-01-01-01-01-01","pdu":"60f627f600a000000146fa7a872fd7053f","RxDelay":1,"RX1DR":5,"RX1Freq":868300000}
2026/02/03 08:06:27 [localhost] received downlink message: 60f627f600a000000146fa7a872fd7053f
2026/02/03 08:06:27 [localhost] propagating downlink to device 0101010101010101
2026/02/03 08:06:27 [0101010101010101] downlink FCnt 0 - FPort: 1 - FRMPayload: 46
```

### Running the Complete Test Flow

You can see all these logs by running the test script:

```bash
./test-flow.sh
```

The script will:
1. Check API health
2. Create a network server
3. Add and connect a gateway
4. Add a device
5. Send a join request
6. Send uplink data
7. Verify the device state

All messages exchanged between gateways, devices, and the network server are logged with full details for debugging and learning purposes.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

See [LICENSE](LICENSE) file for details.

## Disclaimer

LoRaWAN® is a registered trademark of the LoRa Alliance®. This project is not affiliated with, endorsed by, or sponsored by the LoRa Alliance. This simulator is developed independently for educational and testing purposes only.

