package integration

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"
)

type LORIOTClient struct {
	baseURL    string
	authHeader string
	httpClient *http.Client
}

func NewLORIOTClient(url, authHeader string) *LORIOTClient {
	return &LORIOTClient{
		baseURL:    url,
		authHeader: authHeader,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type loriotStatusResponse struct {
	DiscoveryURL  string `json:"basicsStationUrl"`
	DiscoveryPort int    `json:"basicsStationDiscoveryPort"`
}

type loriotGatewayResponse struct {
	Gateways []struct {
		EUI  string `json:"EUI"`
		Base string `json:"base"`
	} `json:"gateways"`
	Page    int `json:"page"`
	PerPage int `json:"perPage"`
	Total   int `json:"total"`
}

func (c *LORIOTClient) ListGateways() ([]gateway.GatewayInfo, error) {
	// Get Discovery URI
	discoveryURI, err := c.getBasicsStationURI()
	if err != nil {
		return nil, err

	}

	// Get gateways
	var allGateways []gateway.GatewayInfo
	page := 1
	perPage := 100
	for {
		// GET /1/nwk/gateways
		url := fmt.Sprintf("%s/1/nwk/gateways?page=%d&perPage=%d", c.baseURL, page, perPage)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", c.authHeader)

		log.Printf("[LORIOT] GET %s", url)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("LORIOT API returned status %d", resp.StatusCode)
		}

		var loriotResp loriotGatewayResponse
		if err := json.NewDecoder(resp.Body).Decode(&loriotResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		// Convert LORIOT gateways to our format
		for _, gw := range loriotResp.Gateways {
			// Sync only Basics Station
			if gw.Base == "basics-station" {
				// Parse EUI from the _id field (format: "AABBCCDDEEFF0011" or "AA-BB-CC-DD-EE-FF-00-11")
				euiStr := strings.ReplaceAll(gw.EUI, "-", "")
				var eui lorawan.EUI64
				if err := eui.UnmarshalText([]byte(euiStr)); err != nil {
					// Skip invalid EUIs
					continue
				}
				log.Printf("[LORIOT] found gateway %s", eui)

				allGateways = append(allGateways, gateway.GatewayInfo{
					EUI:          eui,
					DiscoveryURI: discoveryURI,
				})
			}
		}

		// Check if we need to fetch more pages
		totalPages := (loriotResp.Total + perPage - 1) / perPage
		if page >= totalPages {
			break
		}
		page++
	}

	return allGateways, nil
}

func (c *LORIOTClient) getBasicsStationURI() (string, error) {
	// GET /1/nwk/status
	url := fmt.Sprintf("%s/1/nwk/status", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)

	log.Printf("[LORIOT] GET %s", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return "", fmt.Errorf("LORIOT API returned status %d", resp.StatusCode)
	}

	var loriotResp loriotStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&loriotResp); err != nil {
		resp.Body.Close()
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	resp.Body.Close()

	discoveryUri := fmt.Sprintf("%s:%d", loriotResp.DiscoveryURL, loriotResp.DiscoveryPort)
	log.Printf("[LORIOT] Discovery URI %s", discoveryUri)
	return discoveryUri, nil
}

func (c *LORIOTClient) CreateGateway(eui lorawan.EUI64, discoveryURI string) error {
	// TODO
	return nil
}

func (c *LORIOTClient) DeleteGateway(eui lorawan.EUI64) error {
	// TODO
	return nil
}

func (c *LORIOTClient) ListDevices() ([]device.DeviceInfo, error) {
	// TODO
	return []device.DeviceInfo{}, nil
}

func (c *LORIOTClient) CreateDevice(devEUI lorawan.EUI64, joinEUI lorawan.EUI64, appKey lorawan.AES128Key) error {
	// TODO
	return nil
}

func (c *LORIOTClient) DeleteDevice(devEUI lorawan.EUI64) error {
	// TODO
	return nil
}
