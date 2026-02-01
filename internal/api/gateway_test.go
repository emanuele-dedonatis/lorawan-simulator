package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Setup test router with gateway routes
func setupGatewayTestRouter() (*gin.Engine, *networkserver.Pool) {
	gin.SetMode(gin.TestMode)

	testPool := networkserver.NewPool()
	pool = testPool

	router := gin.Default()

	// Network Servers collection
	router.GET("/network-servers", getNetworkServers)
	router.POST("/network-servers", postNetworkServer)

	// All routes with :name parameter share middleware
	ns := router.Group("/network-servers/:name")
	ns.Use(networkServerMiddleware())
	{
		// Network Server operations
		ns.GET("", getNetworkServersByName)
		ns.DELETE("", delNetworkServer)

		// Gateway operations
		ns.GET("/gateways", getGateways)
		ns.POST("/gateways", postGateway)

		// Gateway by EUI operations
		gw := ns.Group("/gateways/:eui")
		gw.Use(gatewayMiddleware())
		{
			gw.GET("", getGatewayByEUI)
			gw.DELETE("", delGateway)
		}
	}

	return router, testPool
}

func TestGetGateways(t *testing.T) {
	t.Run("returns empty list when no gateways", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		req, _ := http.NewRequest("GET", "/network-servers/test-server/gateways", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []gateway.GatewayInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(response))
	})

	t.Run("returns list of gateways", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		ns, _ := testPool.Add("test-server")

		// Add some gateways
		eui1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		eui2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		eui3 := lorawan.EUI64{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}

		ns.AddGateway(eui1, "http://discovery1.example.com")
		ns.AddGateway(eui2, "http://discovery2.example.com")
		ns.AddGateway(eui3, "http://discovery3.example.com")

		req, _ := http.NewRequest("GET", "/network-servers/test-server/gateways", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []gateway.GatewayInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(response))
	})

	t.Run("returns 404 when network server not found", func(t *testing.T) {
		router, _ := setupGatewayTestRouter()

		req, _ := http.NewRequest("GET", "/network-servers/non-existent/gateways", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestPostGateway(t *testing.T) {
	t.Run("creates gateway successfully", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		body := map[string]string{
			"eui":          "0102030405060708",
			"discoveryUri": "http://discovery.example.com",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/gateways", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response gateway.GatewayInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, response.EUI)
		assert.Equal(t, "http://discovery.example.com", response.DiscoveryURI)
	})

	t.Run("returns 409 when adding duplicate gateway", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		ns, _ := testPool.Add("test-server")

		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		ns.AddGateway(eui, "http://discovery.example.com")

		body := map[string]string{
			"eui":          "0102030405060708",
			"discoveryUri": "http://discovery.example.com",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/gateways", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "already exists")
	})

	t.Run("returns 400 when EUI is missing", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		body := map[string]string{
			"discoveryUri": "http://discovery.example.com",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/gateways", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 when discoveryUri is missing", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		body := map[string]string{
			"eui": "0102030405060708",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/gateways", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 when EUI format is invalid", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		body := map[string]string{
			"eui":          "invalid-eui",
			"discoveryUri": "http://discovery.example.com",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/gateways", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "invalid EUI format")
	})

	t.Run("returns 400 when JSON is invalid", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		invalidJSON := []byte(`{"eui": invalid}`)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/gateways", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 404 when network server not found", func(t *testing.T) {
		router, _ := setupGatewayTestRouter()

		body := map[string]string{
			"eui":          "0102030405060708",
			"discoveryUri": "http://discovery.example.com",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/non-existent/gateways", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetGatewayByEUI(t *testing.T) {
	t.Run("returns gateway when it exists", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		ns, _ := testPool.Add("test-server")

		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		ns.AddGateway(eui, "http://discovery.example.com")

		req, _ := http.NewRequest("GET", "/network-servers/test-server/gateways/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response gateway.GatewayInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, eui, response.EUI)
		assert.Equal(t, "http://discovery.example.com", response.DiscoveryURI)
	})

	t.Run("returns 404 when gateway not found", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		req, _ := http.NewRequest("GET", "/network-servers/test-server/gateways/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "not found")
	})

	t.Run("returns 400 when EUI format is invalid", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		req, _ := http.NewRequest("GET", "/network-servers/test-server/gateways/invalid-eui", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "invalid EUI format")
	})

	t.Run("returns 404 when network server not found", func(t *testing.T) {
		router, _ := setupGatewayTestRouter()

		req, _ := http.NewRequest("GET", "/network-servers/non-existent/gateways/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDelGateway(t *testing.T) {
	t.Run("deletes gateway successfully", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		ns, _ := testPool.Add("test-server")

		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		ns.AddGateway(eui, "http://discovery.example.com")

		req, _ := http.NewRequest("DELETE", "/network-servers/test-server/gateways/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify it was actually removed
		_, err := ns.GetGateway(eui)
		assert.Error(t, err)
	})

	t.Run("returns 404 when gateway not found", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		req, _ := http.NewRequest("DELETE", "/network-servers/test-server/gateways/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "not found")
	})

	t.Run("returns 400 when EUI format is invalid", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		req, _ := http.NewRequest("DELETE", "/network-servers/test-server/gateways/invalid-eui", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "invalid EUI format")
	})

	t.Run("returns 404 when network server not found", func(t *testing.T) {
		router, _ := setupGatewayTestRouter()

		req, _ := http.NewRequest("DELETE", "/network-servers/non-existent/gateways/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestIntegration_GatewayWorkflow(t *testing.T) {
	t.Run("complete CRUD workflow for gateways", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		// 1. List gateways - should be empty
		req, _ := http.NewRequest("GET", "/network-servers/test-server/gateways", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var list1 []gateway.GatewayInfo
		json.Unmarshal(w.Body.Bytes(), &list1)
		assert.Equal(t, 0, len(list1))

		// 2. Create gateway
		body := map[string]string{
			"eui":          "0102030405060708",
			"discoveryUri": "http://discovery.example.com",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ = http.NewRequest("POST", "/network-servers/test-server/gateways", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// 3. Get specific gateway
		req, _ = http.NewRequest("GET", "/network-servers/test-server/gateways/0102030405060708", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 4. List gateways - should have 1
		req, _ = http.NewRequest("GET", "/network-servers/test-server/gateways", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var list2 []gateway.GatewayInfo
		json.Unmarshal(w.Body.Bytes(), &list2)
		assert.Equal(t, 1, len(list2))

		// 5. Delete gateway
		req, _ = http.NewRequest("DELETE", "/network-servers/test-server/gateways/0102030405060708", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// 6. List gateways - should be empty again
		req, _ = http.NewRequest("GET", "/network-servers/test-server/gateways", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var list3 []gateway.GatewayInfo
		json.Unmarshal(w.Body.Bytes(), &list3)
		assert.Equal(t, 0, len(list3))
	})

	t.Run("multiple gateways are independent", func(t *testing.T) {
		router, testPool := setupGatewayTestRouter()
		testPool.Add("test-server")

		// Add multiple gateways
		gateways := []map[string]string{
			{"eui": "0102030405060708", "discoveryUri": "http://discovery1.example.com"},
			{"eui": "1112131415161718", "discoveryUri": "http://discovery2.example.com"},
			{"eui": "2122232425262728", "discoveryUri": "http://discovery3.example.com"},
		}

		for _, gw := range gateways {
			jsonBody, _ := json.Marshal(gw)
			req, _ := http.NewRequest("POST", "/network-servers/test-server/gateways", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusCreated, w.Code)
		}

		// Delete one
		req, _ := http.NewRequest("DELETE", "/network-servers/test-server/gateways/1112131415161718", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify others still exist
		req, _ = http.NewRequest("GET", "/network-servers/test-server/gateways/0102030405060708", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		req, _ = http.NewRequest("GET", "/network-servers/test-server/gateways/2122232425262728", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify deleted one is gone
		req, _ = http.NewRequest("GET", "/network-servers/test-server/gateways/1112131415161718", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
