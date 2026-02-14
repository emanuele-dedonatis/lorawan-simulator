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
		EUI      string `json:"EUI"`
		Base     string `json:"base"`
		Location struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"location"`
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

		if loriotResp.Total == 0 {
			log.Printf("[LORIOT] no gateways found")
			break
		}

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

				gwInfo := gateway.GatewayInfo{
					EUI:          eui,
					DiscoveryURI: discoveryURI,
				}

				// Add location if available
				if gw.Location.Lat != 0 || gw.Location.Lon != 0 {
					gwInfo.Location = &gateway.Location{
						Latitude:  gw.Location.Lat,
						Longitude: gw.Location.Lon,
					}
					log.Printf("[LORIOT] gateway %s has location: lat=%f, lon=%f", eui, gw.Location.Lat, gw.Location.Lon)
				}

				allGateways = append(allGateways, gwInfo)
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
	var allDevices []device.DeviceInfo
	page := 1
	perPage := 100

	for {
		// GET /1/nwk/devices
		url := fmt.Sprintf("%s/1/nwk/devices?page=%d&perPage=%d", c.baseURL, page, perPage)
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

		var loriotResp struct {
			Devices []struct {
				DevEUI string `json:"deveui"`
				AppEUI string `json:"appeui"`
				AppKey string `json:"appkey"`
			} `json:"devices"`
			Page    int `json:"page"`
			PerPage int `json:"perPage"`
			Total   int `json:"total"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&loriotResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		if loriotResp.Total == 0 {
			log.Printf("[LORIOT] no devices found")
			break
		}

		// Convert LORIOT devices to our format
		for _, dev := range loriotResp.Devices {
			// Parse DevEUI
			devEUIStr := strings.ReplaceAll(dev.DevEUI, "-", "")
			var devEUI lorawan.EUI64
			if err := devEUI.UnmarshalText([]byte(devEUIStr)); err != nil {
				log.Printf("[LORIOT] invalid DevEUI %s: %v", dev.DevEUI, err)
				continue
			}

			// Get full device details including session keys
			deviceInfo, err := c.getDeviceDetails(devEUIStr)
			if err != nil {
				log.Printf("[LORIOT] failed to get device details for %s: %v", dev.DevEUI, err)
				continue
			}

			log.Printf("[LORIOT] found device %s", devEUI)
			allDevices = append(allDevices, deviceInfo)
		}

		// Check if we need to fetch more pages
		totalPages := (loriotResp.Total + perPage - 1) / perPage
		if page >= totalPages {
			break
		}
		page++
	}

	log.Printf("[LORIOT] listed %d devices", len(allDevices))
	return allDevices, nil
}

func (c *LORIOTClient) getDeviceDetails(devEUI string) (device.DeviceInfo, error) {
	// GET /1/nwk/device/<deveui>
	url := fmt.Sprintf("%s/1/nwk/device/%s", c.baseURL, devEUI)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return device.DeviceInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)

	log.Printf("[LORIOT] GET %s", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return device.DeviceInfo{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return device.DeviceInfo{}, fmt.Errorf("LORIOT API returned status %d", resp.StatusCode)
	}

	var loriotDevice struct {
		DevEUI        string   `json:"deveui"`
		JoinEUI       string   `json:"joineui"`
		AppKey        string   `json:"appkey"`
		AppSKey       string   `json:"appskey"`
		NwkSKey       string   `json:"nwkskey"`
		DevAddr       string   `json:"devaddr"`
		LastDevNonces []uint16 `json:"lastDevNonces"`
		SeqNo         int      `json:"seqno"`
		SeqDn         int      `json:"seqdn"`
		Location      struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"location"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loriotDevice); err != nil {
		return device.DeviceInfo{}, fmt.Errorf("failed to decode response: %w", err)
	}
	// Parse DevEUI
	var parsedDevEUI lorawan.EUI64
	if err := parsedDevEUI.UnmarshalText([]byte(loriotDevice.DevEUI)); err != nil {
		return device.DeviceInfo{}, fmt.Errorf("invalid DevEUI: %w", err)
	}

	// Parse JoinEUI
	var joinEUI lorawan.EUI64
	if err := joinEUI.UnmarshalText([]byte(loriotDevice.JoinEUI)); err != nil {
		return device.DeviceInfo{}, fmt.Errorf("invalid JoinEUI: %w", err)
	}

	// Parse AppKey
	var appKey lorawan.AES128Key
	if err := appKey.UnmarshalText([]byte(loriotDevice.AppKey)); err != nil {
		return device.DeviceInfo{}, fmt.Errorf("invalid AppKey: %w", err)
	}

	// Parse session keys (may be empty for devices that haven't joined)
	var appSKey lorawan.AES128Key
	if loriotDevice.AppSKey != "" {
		if err := appSKey.UnmarshalText([]byte(loriotDevice.AppSKey)); err != nil {
			log.Printf("[LORIOT] invalid AppSKey for device %s: %v", devEUI, err)
		}
	}

	var nwkSKey lorawan.AES128Key
	if loriotDevice.NwkSKey != "" {
		if err := nwkSKey.UnmarshalText([]byte(loriotDevice.NwkSKey)); err != nil {
			log.Printf("[LORIOT] invalid NwkSKey for device %s: %v", devEUI, err)
		}
	}

	// Parse DevAddr (may be empty for devices that haven't joined)
	var devAddr lorawan.DevAddr
	if loriotDevice.DevAddr != "" {
		if err := devAddr.UnmarshalText([]byte(loriotDevice.DevAddr)); err != nil {
			log.Printf("[LORIOT] invalid DevAddr for device %s: %v", devEUI, err)
		}
	}

	// Get the maximum DevNonce
	var devNonce lorawan.DevNonce
	if len(loriotDevice.LastDevNonces) > 0 {
		maxNonce := loriotDevice.LastDevNonces[0]
		for _, nonce := range loriotDevice.LastDevNonces {
			if nonce > maxNonce {
				maxNonce = nonce
			}
		}
		devNonce = lorawan.DevNonce(maxNonce)
	}

	// Convert frame counters (LORIOT uses -1 for uninitialized seqno)
	var fcntUp uint32
	if loriotDevice.SeqNo >= 0 {
		fcntUp = uint32(loriotDevice.SeqNo) + 1 // LORIOT seqno is the last used FCntUp, so we add 1 to get the next expected value
	}

	var fcntDn uint32
	if loriotDevice.SeqDn >= 0 {
		fcntDn = uint32(loriotDevice.SeqDn)
	}

	log.Printf("[LORIOT] Returning device info - DevAddr: %s, AppSKey: %s (parsed: %x), NwkSKey: %s (parsed: %x), FCntUp: %d, FCntDn: %d",
		devAddr, loriotDevice.AppSKey, appSKey[:], loriotDevice.NwkSKey, nwkSKey[:], fcntUp, fcntDn)

	deviceInfo := device.DeviceInfo{
		DevEUI:   parsedDevEUI,
		JoinEUI:  joinEUI,
		AppKey:   appKey,
		DevNonce: devNonce,
		DevAddr:  devAddr,
		AppSKey:  appSKey,
		NwkSKey:  nwkSKey,
		FCntUp:   fcntUp,
		FCntDn:   fcntDn,
	}

	// Add location if available
	if loriotDevice.Location.Lat != 0 || loriotDevice.Location.Lon != 0 {
		deviceInfo.Location = &device.Location{
			Latitude:  loriotDevice.Location.Lat,
			Longitude: loriotDevice.Location.Lon,
		}
		log.Printf("[LORIOT] device %s has location: lat=%f, lon=%f", devEUI, loriotDevice.Location.Lat, loriotDevice.Location.Lon)
	}

	return deviceInfo, nil
}

func (c *LORIOTClient) CreateDevice(devEUI lorawan.EUI64, joinEUI lorawan.EUI64, appKey lorawan.AES128Key) error {
	// TODO
	return nil
}

func (c *LORIOTClient) DeleteDevice(devEUI lorawan.EUI64) error {
	// TODO
	return nil
}
