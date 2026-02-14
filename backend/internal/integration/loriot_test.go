package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLORIOTClient_ListGateways_Pagination(t *testing.T) {
	t.Run("fetches single page of gateways", func(t *testing.T) {
		// Mock server that returns 1 page
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle status endpoint first
			if r.URL.Path == "/1/nwk/status" {
				statusResp := loriotStatusResponse{
					DiscoveryURL:  "wss://eu1.loriot.io",
					DiscoveryPort: 6001,
				}
				json.NewEncoder(w).Encode(statusResp)
				return
			}

			assert.Equal(t, "/1/nwk/gateways", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			page := r.URL.Query().Get("page")
			assert.Equal(t, "1", page)

			response := loriotGatewayResponse{
				Gateways: []struct {
				EUI      string `json:"EUI"`
				Base     string `json:"base"`
				Location struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"location"`
			}{
					{EUI: "AA-BB-CC-DD-EE-FF-00-11", Base: "basics-station"},
					{EUI: "00-0A-0F-FF-FF-26-0A-0D", Base: "pktfwd"}, // Should be filtered out
					{EUI: "11-22-33-44-55-66-77-88", Base: "basics-station"},
				},
				Page:    1,
				PerPage: 100,
				Total:   3,
			}

			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewLORIOTClient(server.URL, "Bearer test-token")
		gateways, err := client.ListGateways()

		assert.NoError(t, err)
		assert.Len(t, gateways, 2) // Only basics-station gateways
		assert.Equal(t, "aabbccddeeff0011", gateways[0].EUI.String())
		assert.Equal(t, "1122334455667788", gateways[1].EUI.String())
		assert.Equal(t, "wss://eu1.loriot.io:6001", gateways[0].DiscoveryURI)
	})

	t.Run("fetches multiple pages of gateways", func(t *testing.T) {
		// Mock server that returns 2 pages
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle status endpoint
			if r.URL.Path == "/1/nwk/status" {
				statusResp := loriotStatusResponse{
					DiscoveryURL:  "wss://eu1.loriot.io",
					DiscoveryPort: 6001,
				}
				json.NewEncoder(w).Encode(statusResp)
				return
			}

			page := r.URL.Query().Get("page")

			var response loriotGatewayResponse
			if page == "1" {
				response = loriotGatewayResponse{
					Gateways: []struct {
				EUI      string `json:"EUI"`
				Base     string `json:"base"`
				Location struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"location"`
			}{
						{EUI: "AA-BB-CC-DD-EE-FF-00-11", Base: "basics-station"},
					},
					Page:    1,
					PerPage: 100,
					Total:   150, // More than 1 page
				}
			} else if page == "2" {
				response = loriotGatewayResponse{
					Gateways: []struct {
				EUI      string `json:"EUI"`
				Base     string `json:"base"`
				Location struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"location"`
			}{
						{EUI: "00-0A-0F-FF-FF-26-0A-0D", Base: "basics-station"},
					},
					Page:    2,
					PerPage: 100,
					Total:   150,
				}
			}

			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewLORIOTClient(server.URL, "Bearer test-token")
		gateways, err := client.ListGateways()

		assert.NoError(t, err)
		assert.Len(t, gateways, 2)
	})

	t.Run("handles EUI with dashes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle status endpoint
			if r.URL.Path == "/1/nwk/status" {
				statusResp := loriotStatusResponse{
					DiscoveryURL:  "wss://eu1.loriot.io",
					DiscoveryPort: 6001,
				}
				json.NewEncoder(w).Encode(statusResp)
				return
			}

			response := loriotGatewayResponse{
				Gateways: []struct {
				EUI      string `json:"EUI"`
				Base     string `json:"base"`
				Location struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"location"`
			}{
					{EUI: "FC-C2-3D-FF-FE-0B-9F-98", Base: "basics-station"},
				},
				Page:    1,
				PerPage: 100,
				Total:   1,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewLORIOTClient(server.URL, "Bearer test-token")
		gateways, err := client.ListGateways()

		assert.NoError(t, err)
		assert.Len(t, gateways, 1)
		assert.Equal(t, "fcc23dfffe0b9f98", gateways[0].EUI.String())
	})

	t.Run("skips invalid EUIs", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle status endpoint
			if r.URL.Path == "/1/nwk/status" {
				statusResp := loriotStatusResponse{
					DiscoveryURL:  "wss://eu1.loriot.io",
					DiscoveryPort: 6001,
				}
				json.NewEncoder(w).Encode(statusResp)
				return
			}

			response := loriotGatewayResponse{
				Gateways: []struct {
				EUI      string `json:"EUI"`
				Base     string `json:"base"`
				Location struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"location"`
			}{
					{EUI: "AA-BB-CC-DD-EE-FF-00-11", Base: "basics-station"},
					{EUI: "INVALID", Base: "basics-station"},
					{EUI: "00-0A-0F-FF-FF-26-0A-0D", Base: "basics-station"},
				},
				Page:    1,
				PerPage: 100,
				Total:   3,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewLORIOTClient(server.URL, "Bearer test-token")
		gateways, err := client.ListGateways()

		assert.NoError(t, err)
		assert.Len(t, gateways, 2) // Invalid EUI is skipped
	})

	t.Run("handles HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client := NewLORIOTClient(server.URL, "Bearer test-token")
		gateways, err := client.ListGateways()

		assert.Error(t, err)
		assert.Nil(t, gateways)
		assert.Contains(t, err.Error(), "401")
	})

	t.Run("handles empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle status endpoint
			if r.URL.Path == "/1/nwk/status" {
				statusResp := loriotStatusResponse{
					DiscoveryURL:  "wss://eu1.loriot.io",
					DiscoveryPort: 6001,
				}
				json.NewEncoder(w).Encode(statusResp)
				return
			}

			response := loriotGatewayResponse{
				Gateways: []struct {
				EUI      string `json:"EUI"`
				Base     string `json:"base"`
				Location struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"location"`
			}{},
				Page:    1,
				PerPage: 100,
				Total:   0,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewLORIOTClient(server.URL, "Bearer test-token")
		gateways, err := client.ListGateways()

		assert.NoError(t, err)
		assert.Len(t, gateways, 0)
	})
}
