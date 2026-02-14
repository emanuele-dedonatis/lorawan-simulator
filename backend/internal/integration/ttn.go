package integration

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type TTNClient struct {
	baseURL string
	apiKey  string
	conn    *grpc.ClientConn
}

func NewTTNClient(url, apiKey string) *TTNClient {
	return &TTNClient{
		baseURL: url,
		apiKey:  apiKey,
	}
}

func (c *TTNClient) getConnection() (*grpc.ClientConn, error) {
	if c.conn != nil {
		return c.conn, nil
	}

	// Parse the URL to extract host and determine if TLS is needed
	parsedURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Extract the scheme
	scheme := parsedURL.Scheme
	if scheme == "" {
		scheme = "https" // Default to https for TTN
	}

	// Remove scheme from the URL for gRPC
	target := strings.TrimPrefix(c.baseURL, scheme+"://")

	// Remove any trailing path
	if idx := strings.Index(target, "/"); idx > 0 {
		target = target[:idx]
	}

	// Add default port if not specified
	if !strings.Contains(target, ":") {
		if scheme == "https" {
			target = target + ":8884" // TTN default gRPC port
		} else {
			target = target + ":1884" // TTN insecure gRPC port
		}
	}

	// Determine TLS credentials based on scheme
	var creds credentials.TransportCredentials
	if scheme == "https" {
		creds = credentials.NewTLS(&tls.Config{})
	} else {
		creds = insecure.NewCredentials()
	}

	// Create gRPC connection
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TTN: %w", err)
	}

	c.conn = conn
	return conn, nil
}

func (c *TTNClient) buildDiscoveryURI() string {
	parsedURL, err := url.Parse(c.baseURL)
	if err != nil {
		log.Printf("[TTN] Error parsing URL: %v", err)
		return ""
	}

	// Extract host
	host := parsedURL.Host
	if host == "" {
		host = parsedURL.Path
	}
	// Remove port if present
	if idx := strings.Index(host, ":"); idx > 0 {
		host = host[:idx]
	}

	// Check if it's a standard TTN cloud URL (xx.cloud.thethings.network format)
	parts := strings.Split(host, ".")
	if len(parts) >= 4 && parts[1] == "cloud" && parts[2] == "thethings" && parts[3] == "network" {
		// Standard TTN cloud URL - use the cluster hostname with WebSocket port
		uri := "wss://" + host + ":8887"
		log.Printf("[TTN] Discovery URI: %s", uri)
		return uri
	}

	// Custom URL - convert scheme to websocket
	scheme := parsedURL.Scheme
	if scheme == "" {
		scheme = "https" // Default to https
	}

	wsScheme := "wss"
	port := "8887"
	if scheme == "http" {
		wsScheme = "ws"
		port = "1887"
	}

	uri := wsScheme + "://" + host + ":" + port
	log.Printf("[TTN] Discovery URI: %s", uri)
	return uri
}

func (c *TTNClient) getAuthContext() context.Context {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + c.apiKey,
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

func (c *TTNClient) ListGateways() ([]gateway.GatewayInfo, error) {
	discoveryUri := c.buildDiscoveryURI()
	log.Printf("[TTN] Discovery URI %s", discoveryUri)

	conn, err := c.getConnection()
	if err != nil {
		return nil, err
	}

	// Create gateway registry client
	client := ttnpb.NewGatewayRegistryClient(conn)
	ctx := c.getAuthContext()

	var allGateways []gateway.GatewayInfo
	var page uint32 = 1
	limit := uint32(100)

	// Paginate through all gateways
	for {
		req := &ttnpb.ListGatewaysRequest{
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"ids", "ids.eui", "antennas"},
			},
			Page:  page,
			Limit: limit,
		}

		resp, err := client.List(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to list gateways: %w", err)
		}

		// Convert TTN gateways to GatewayInfo
		for _, gw := range resp.Gateways {
			if gw.Ids == nil || len(gw.Ids.Eui) == 0 {
				continue
			}

			if len(gw.Ids.Eui) != 8 {
				log.Printf("[TTN] Invalid EUI length %d: %v", len(gw.Ids.Eui), gw.Ids.Eui)
				continue
			}

			var eui lorawan.EUI64
			copy(eui[:], gw.Ids.Eui)

			// Create headers with Authorization bearer token
			headers := http.Header{}
			headers.Add("Authorization", "Bearer "+c.apiKey)

			gwInfo := gateway.GatewayInfo{
				EUI:          eui,
				DiscoveryURI: discoveryUri,
				Headers:      headers,
			}

			// Extract location from antennas (TTN stores location in first antenna)
			if len(gw.Antennas) > 0 && gw.Antennas[0].Location != nil {
				loc := gw.Antennas[0].Location
				if loc.Latitude != 0 || loc.Longitude != 0 {
					gwInfo.Location = &gateway.Location{
						Latitude:  loc.Latitude,
						Longitude: loc.Longitude,
					}
					log.Printf("[TTN] gateway %s has location: lat=%f, lon=%f", eui, loc.Latitude, loc.Longitude)
				}
			}

			allGateways = append(allGateways, gwInfo)
		}

		// Check if there are more pages
		if len(resp.Gateways) < int(limit) {
			break
		}
		page++
	}

	log.Printf("[TTN] Found %d gateways", len(allGateways))
	return allGateways, nil
}

func (c *TTNClient) CreateGateway(eui lorawan.EUI64, discoveryURI string) error {
	// TODO: Implement TTN gateway creation
	return nil
}

func (c *TTNClient) DeleteGateway(eui lorawan.EUI64) error {
	// TODO: Implement TTN gateway deletion
	return nil
}

func (c *TTNClient) ListDevices() ([]device.DeviceInfo, error) {
	conn, err := c.getConnection()
	if err != nil {
		return nil, err
	}

	ctx := c.getAuthContext()
	var allDevices []device.DeviceInfo

	// First, list all applications
	appClient := ttnpb.NewApplicationRegistryClient(conn)
	var appPage uint32 = 1
	appLimit := uint32(100)

	for {
		appReq := &ttnpb.ListApplicationsRequest{
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"ids"},
			},
			Page:  appPage,
			Limit: appLimit,
		}

		appResp, err := appClient.List(ctx, appReq)
		if err != nil {
			return nil, fmt.Errorf("failed to list applications: %w", err)
		}

		// For each application, list its devices
		endDeviceClient := ttnpb.NewEndDeviceRegistryClient(conn)
		for _, app := range appResp.Applications {
			if app.Ids == nil {
				continue
			}

			log.Printf("[TTN] listing devices for application %s", app.Ids.ApplicationId)

			var devPage uint32 = 1
			devLimit := uint32(100)

			// Paginate through devices in this application
			for {
				devReq := &ttnpb.ListEndDevicesRequest{
					ApplicationIds: app.Ids,
					FieldMask: &fieldmaskpb.FieldMask{
						Paths: []string{
							"ids.device_id",
							"ids.dev_eui",
							"ids.join_eui",
							"locations",
						},
					},
					Page:  devPage,
					Limit: devLimit,
				}

				devResp, err := endDeviceClient.List(ctx, devReq)
				if err != nil {
					log.Printf("[TTN] failed to list devices for application %s: %v", app.Ids.ApplicationId, err)
					break
				}

				// Process each device - need to fetch full details individually
				for _, dev := range devResp.EndDevices {
					if dev.Ids == nil || len(dev.Ids.DevEui) == 0 {
						continue
					}

					// Parse DevEUI from the list response
					var devEUI lorawan.EUI64
					if len(dev.Ids.DevEui) != 8 {
						log.Printf("[TTN] invalid DevEUI length for device: %d", len(dev.Ids.DevEui))
						continue
					}
					copy(devEUI[:], dev.Ids.DevEui)

					// Parse JoinEUI from the list response
					var joinEUI lorawan.EUI64
					if len(dev.Ids.JoinEui) == 8 {
						copy(joinEUI[:], dev.Ids.JoinEui)
					}

					// Try to get full device details, but don't fail if we can't
					// Some users may not have permission to read sensitive fields
					deviceInfo := device.DeviceInfo{
						DevEUI:  devEUI,
						JoinEUI: joinEUI,
						// Other fields will be zero values if we can't fetch them
					}

					// Extract location if available from the list response
					if len(dev.Locations) > 0 {
						// TTN can have multiple locations (user-set, frm-payload, etc.)
						// Try "user" location first, then fall back to any available location
						var location *ttnpb.Location
						for source, loc := range dev.Locations {
							if source == "user" {
								location = loc
								break
							}
							// Keep first non-nil location as fallback
							if location == nil && loc != nil {
								location = loc
							}
						}

						if location != nil && (location.Latitude != 0 || location.Longitude != 0) {
							deviceInfo.Location = &device.Location{
								Latitude:  location.Latitude,
								Longitude: location.Longitude,
							}
							log.Printf("[TTN] device %s has location: lat=%f, lon=%f", devEUI, location.Latitude, location.Longitude)
						}
					}

					// Attempt to get full details
					fullInfo, err := c.getDeviceDetails(ctx, dev.Ids)
					if err != nil {
						log.Printf("[TTN] could not get full details for device %s: %v (using basic info only)", dev.Ids.DeviceId, err)
					} else {
						// Preserve location from list response if not in full details
						if fullInfo.Location == nil && deviceInfo.Location != nil {
							fullInfo.Location = deviceInfo.Location
						}
						deviceInfo = fullInfo
					}

					log.Printf("[TTN] found device %s", deviceInfo.DevEUI)
					allDevices = append(allDevices, deviceInfo)
				}

				// Check if we've retrieved all devices for this application
				if len(devResp.EndDevices) < int(devLimit) {
					break
				}

				devPage++
			}
		}

		// Check if we've retrieved all applications
		if len(appResp.Applications) < int(appLimit) {
			break
		}

		appPage++
	}

	log.Printf("[TTN] listed %d devices across all applications", len(allDevices))
	return allDevices, nil
}

func (c *TTNClient) getDeviceDetails(ctx context.Context, devIds *ttnpb.EndDeviceIdentifiers) (device.DeviceInfo, error) {
	// Get connection for Network Server client
	conn, err := c.getConnection()
	if err != nil {
		return device.DeviceInfo{}, fmt.Errorf("failed to get connection: %w", err)
	}

	// Use Network Server client to get session information
	nsClient := ttnpb.NewNsEndDeviceRegistryClient(conn)

	// Request device details with session information from Network Server
	getReq := &ttnpb.GetEndDeviceRequest{
		EndDeviceIds: devIds,
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"ids",
				"session.dev_addr",
				"session.keys.f_nwk_s_int_key.key",
				"session.keys.nwk_s_enc_key.key",
				"session.keys.s_nwk_s_int_key.key",
				"session.last_f_cnt_up",
				"session.last_n_f_cnt_down",
				"session.started_at",
			},
		},
	}

	dev, err := nsClient.Get(ctx, getReq)
	if err != nil {
		return device.DeviceInfo{}, fmt.Errorf("failed to get device from Network Server: %w", err)
	}

	// Also get Application Server session information
	asClient := ttnpb.NewAsEndDeviceRegistryClient(conn)
	asGetReq := &ttnpb.GetEndDeviceRequest{
		EndDeviceIds: devIds,
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"ids",
				"session.dev_addr",
				"session.keys.app_s_key.key",
				"session.last_a_f_cnt_down",
			},
		},
	}

	asDev, err := asClient.Get(ctx, asGetReq)
	if err != nil {
		log.Printf("[TTN] Warning: could not get device from Application Server: %v", err)
		// Continue without AS data
	}

	// Also get Join Server information to retrieve root keys (AppKey) and DevNonce
	jsClient := ttnpb.NewJsEndDeviceRegistryClient(conn)
	jsGetReq := &ttnpb.GetEndDeviceRequest{
		EndDeviceIds: devIds,
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"ids",
				"root_keys.app_key.key",
				"root_keys.nwk_key.key",
				"last_dev_nonce",
			},
		},
	}

	jsDev, err := jsClient.Get(ctx, jsGetReq)
	if err != nil {
		log.Printf("[TTN] Warning: could not get device from Join Server: %v", err)
		// Continue without JS data
	}

	// Parse DevEUI
	var devEUI lorawan.EUI64
	if len(dev.Ids.DevEui) != 8 {
		return device.DeviceInfo{}, fmt.Errorf("invalid DevEUI length: %d", len(dev.Ids.DevEui))
	}
	copy(devEUI[:], dev.Ids.DevEui)

	// Parse JoinEUI
	var joinEUI lorawan.EUI64
	if len(dev.Ids.JoinEui) == 8 {
		copy(joinEUI[:], dev.Ids.JoinEui)
	}

	// DevAddr - prefer from session
	var devAddr lorawan.DevAddr
	if dev.Session != nil && len(dev.Session.DevAddr) == 4 {
		copy(devAddr[:], dev.Session.DevAddr)
	} else if len(dev.Ids.DevAddr) == 4 {
		copy(devAddr[:], dev.Ids.DevAddr)
	}

	// Parse AppKey from Join Server
	var appKey lorawan.AES128Key
	if jsDev != nil && jsDev.RootKeys != nil && jsDev.RootKeys.AppKey != nil && len(jsDev.RootKeys.AppKey.Key) == 16 {
		copy(appKey[:], jsDev.RootKeys.AppKey.Key)
	}

	// Parse DevNonce from Join Server
	var devNonce lorawan.DevNonce
	if jsDev != nil {
		// LastDevNonce is a uint32, but DevNonce is uint16 - take the lower 16 bits
		devNonce = lorawan.DevNonce(jsDev.LastDevNonce & 0xFFFF)
	}

	// Parse session information (if device has joined)
	var appSKey lorawan.AES128Key
	var nwkSKey lorawan.AES128Key
	var fcntUp uint32
	var fcntDn uint32

	// Get Network Server session keys
	if dev.Session != nil {
		// Session keys
		if dev.Session.Keys != nil {
			// TTN uses NwkSEncKey as the network session key for LoRaWAN 1.1+
			if dev.Session.Keys.NwkSEncKey != nil && len(dev.Session.Keys.NwkSEncKey.Key) == 16 {
				copy(nwkSKey[:], dev.Session.Keys.NwkSEncKey.Key)
			}
		}

		// Frame counters
		fcntUp = dev.Session.LastFCntUp
		fcntDn = dev.Session.LastNFCntDown
	}

	// Get Application Server session key
	if asDev != nil && asDev.Session != nil && asDev.Session.Keys != nil {
		if asDev.Session.Keys.AppSKey != nil && len(asDev.Session.Keys.AppSKey.Key) == 16 {
			copy(appSKey[:], asDev.Session.Keys.AppSKey.Key)
		}
	}

	return device.DeviceInfo{
		DevEUI:   devEUI,
		JoinEUI:  joinEUI,
		AppKey:   appKey,
		DevNonce: devNonce,
		DevAddr:  devAddr,
		AppSKey:  appSKey,
		NwkSKey:  nwkSKey,
		FCntUp:   fcntUp,
		FCntDn:   fcntDn,
	}, nil
}

func (c *TTNClient) CreateDevice(devEUI lorawan.EUI64, joinEUI lorawan.EUI64, appKey lorawan.AES128Key) error {
	// TODO: Implement TTN device creation
	return nil
}

func (c *TTNClient) DeleteDevice(devEUI lorawan.EUI64) error {
	// TODO: Implement TTN device deletion
	return nil
}
