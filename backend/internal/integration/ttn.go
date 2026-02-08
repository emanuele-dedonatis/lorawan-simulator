package integration

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
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

	// Extract region from host, e.g. eu1.cloud.thethings.network
	host := parsedURL.Host
	if host == "" {
		host = parsedURL.Path
	}
	// Remove port if present
	if idx := strings.Index(host, ":"); idx > 0 {
		host = host[:idx]
	}

	// Expect host like eu1.cloud.thethings.network
	parts := strings.Split(host, ".")
	if len(parts) < 4 {
		log.Printf("[TTN] Unsupported baseURL: %s", c.baseURL)
		return ""
	}
	region := parts[0][:2] // e.g. "eu" from "eu1"
	switch region {
	case "eu", "us", "in", "au":
		uri := "wss://lns." + region + ".thethings.network:443"
		log.Printf("[TTN] Discovery URI: %s", uri)
		return uri
	default:
		log.Printf("[TTN] Unsupported region in baseURL: %s", c.baseURL)
		return ""
	}
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
				Paths: []string{"ids", "ids.eui"},
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

			allGateways = append(allGateways, gateway.GatewayInfo{
				EUI:          eui,
				DiscoveryURI: discoveryUri,
			})
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
	// TODO: Implement TTN device listing
	return []device.DeviceInfo{}, nil
}

func (c *TTNClient) CreateDevice(devEUI lorawan.EUI64, joinEUI lorawan.EUI64, appKey lorawan.AES128Key) error {
	// TODO: Implement TTN device creation
	return nil
}

func (c *TTNClient) DeleteDevice(devEUI lorawan.EUI64) error {
	// TODO: Implement TTN device deletion
	return nil
}
