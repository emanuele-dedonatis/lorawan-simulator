package integration

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	"github.com/brocaar/lorawan"
	"github.com/chirpstack/chirpstack/api/go/v4/api"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ChirpStackClient struct {
	baseURL string
	apiKey  string
	conn    *grpc.ClientConn
}

func NewChirpStackClient(url, apiKey string) *ChirpStackClient {
	return &ChirpStackClient{
		baseURL: url,
		apiKey:  apiKey,
	}
}

func (c *ChirpStackClient) ListGateways() ([]gateway.GatewayInfo, error) {
	discoveryUri := c.buildDiscoveryURI()
	log.Printf("[CHIRPSTACK] Discovery URI %s", discoveryUri)

	conn, err := c.getConnection()
	if err != nil {
		return nil, err
	}

	// Create gateway service client
	client := api.NewGatewayServiceClient(conn)
	ctx := c.getAuthContext()

	var allGateways []gateway.GatewayInfo
	var offset uint32 = 0
	limit := uint32(100)

	// Paginate through all gateways
	for {
		// List gateways request
		req := &api.ListGatewaysRequest{
			Limit:  limit,
			Offset: offset,
		}

		resp, err := client.List(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to list gateways: %w", err)
		}

		// Process each gateway
		for _, gw := range resp.Result {
			// Parse EUI from gateway ID (ChirpStack stores as hex string)
			var eui lorawan.EUI64
			if err := eui.UnmarshalText([]byte(gw.GatewayId)); err != nil {
				log.Printf("[CHIRPSTACK] invalid gateway EUI %s: %v", gw.GatewayId, err)
				continue
			}

			log.Printf("[CHIRPSTACK] found gateway %s", eui)
			allGateways = append(allGateways, gateway.GatewayInfo{
				EUI:          eui,
				DiscoveryURI: c.buildDiscoveryURI(),
			})
		}

		// Check if we've retrieved all gateways
		if uint32(len(resp.Result)) < limit {
			break
		}

		offset += limit
	}

	log.Printf("[CHIRPSTACK] listed %d gateways", len(allGateways))
	return allGateways, nil
}

func (c *ChirpStackClient) CreateGateway(eui lorawan.EUI64, discoveryURI string) error {
	// TODO
	return nil
}

func (c *ChirpStackClient) DeleteGateway(eui lorawan.EUI64) error {
	// TODO
	return nil
}

func (c *ChirpStackClient) ListDevices() ([]device.DeviceInfo, error) {
	// TODO
	return []device.DeviceInfo{}, nil
}

func (c *ChirpStackClient) CreateDevice(devEUI lorawan.EUI64, joinEUI lorawan.EUI64, appKey lorawan.AES128Key) error {
	// TODO
	return nil
}

func (c *ChirpStackClient) DeleteDevice(devEUI lorawan.EUI64) error {
	// TODO
	return nil
}

// Establish a gRPC connection
func (c *ChirpStackClient) getConnection() (*grpc.ClientConn, error) {
	if c.conn != nil {
		return c.conn, nil
	}

	// Extract host:port from baseURL (remove http:// or https:// and any path)
	grpcAddr := c.baseURL
	grpcAddr = strings.TrimPrefix(grpcAddr, "https://")
	grpcAddr = strings.TrimPrefix(grpcAddr, "http://")

	// Remove any path component (everything after the first /)
	if idx := strings.Index(grpcAddr, "/"); idx != -1 {
		grpcAddr = grpcAddr[:idx]
	}

	// If no port specified, add default port based on scheme
	if !strings.Contains(grpcAddr, ":") {
		if strings.HasPrefix(c.baseURL, "https://") {
			grpcAddr = grpcAddr + ":443"
		} else {
			grpcAddr = grpcAddr + ":8080" // ChirpStack default gRPC port
		}
	}

	log.Printf("[CHIRPSTACK] connecting to gRPC server at %s", grpcAddr)

	// Determine if we should use TLS based on original URL
	var opts []grpc.DialOption
	if strings.HasPrefix(c.baseURL, "https://") {
		// Use TLS for https URLs
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		}
	} else {
		// Use insecure connection for http URLs (no TLS)
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
	}

	// Establish connection without blocking
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ChirpStack at %s: %w", grpcAddr, err)
	}

	log.Printf("[CHIRPSTACK] connected to gRPC server at %s", grpcAddr)
	c.conn = conn
	return conn, nil
}

// Create a context with API key authentication
func (c *ChirpStackClient) getAuthContext() context.Context {
	ctx := context.Background()
	md := metadata.Pairs("authorization", "Bearer "+c.apiKey)
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *ChirpStackClient) buildDiscoveryURI() string {
	url := c.baseURL

	// Remove http:// or https:// scheme first
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Remove any path component (everything after the first /)
	// Now we can safely check for / since scheme is already removed
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	// Remove port if present (gRPC typically uses 8080)
	if idx := strings.LastIndex(url, ":"); idx > 0 {
		url = url[:idx]
	}

	// Determine if original URL was https or http
	if strings.HasPrefix(c.baseURL, "https://") {
		url = "wss://" + url
	} else {
		url = "ws://" + url
	}

	// Add port 3001 for Basics Station
	return fmt.Sprintf("%s:3001", url)
}

// Close closes the gRPC connection
func (c *ChirpStackClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
