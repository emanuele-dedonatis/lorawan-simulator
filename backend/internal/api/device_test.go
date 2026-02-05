package api

import (
        "github.com/emanuele-dedonatis/lorawan-simulator/internal/integration"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Setup test router with device routes
func setupDeviceTestRouter() (*gin.Engine, *networkserver.Pool) {
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

		// Device operations
		ns.GET("/devices", getDevices)
		ns.POST("/devices", postDevice)

		// Device by EUI operations
		dev := ns.Group("/devices/:eui")
		dev.Use(deviceMiddleware())
		{
			dev.GET("", getDeviceByEUI)
			dev.DELETE("", delDevice)
		}
	}

	return router, testPool
}

func TestGetDevices(t *testing.T) {
	t.Run("returns empty list when no devices", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		req, _ := http.NewRequest("GET", "/network-servers/test-server/devices", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []device.DeviceInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(response))
	})

	t.Run("returns list of devices", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		ns, _ := testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		// Add some devices
		devEUI1 := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		devEUI2 := lorawan.EUI64{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
		devEUI3 := lorawan.EUI64{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}
		joinEUI := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}

		ns.AddDevice(devEUI1, joinEUI, appKey, 0)
		ns.AddDevice(devEUI2, joinEUI, appKey, 0)
		ns.AddDevice(devEUI3, joinEUI, appKey, 0)

		req, _ := http.NewRequest("GET", "/network-servers/test-server/devices", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []device.DeviceInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(response))
	})

	t.Run("returns 404 when network server not found", func(t *testing.T) {
		router, _ := setupDeviceTestRouter()

		req, _ := http.NewRequest("GET", "/network-servers/non-existent/devices", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestPostDevice(t *testing.T) {
	t.Run("creates device successfully", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		body := map[string]interface{}{
			"deveui":   "0102030405060708",
			"joineui":  "aabbccddeeff0011",
			"appkey":   "0102030405060708090a0b0c0d0e0f10",
			"devnonce": 0,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/devices", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response device.DeviceInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, response.DevEUI)
	})

	t.Run("returns 409 when adding duplicate device", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		ns, _ := testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		ns.AddDevice(devEUI, joinEUI, appKey, 0)

		body := map[string]string{
			"deveui":  "0102030405060708",
			"joineui": "aabbccddeeff0011",
			"appkey":  "0102030405060708090a0b0c0d0e0f10",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/devices", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "already exists")
	})

	t.Run("returns 400 when DevEUI is missing", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		body := map[string]string{}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/devices", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 when DevEUI format is invalid", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		body := map[string]string{
			"deveui":  "invalid-eui",
			"joineui": "aabbccddeeff0011",
			"appkey":  "0102030405060708090a0b0c0d0e0f10",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/devices", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "invalid DevEui format")
	})

	t.Run("returns 400 when JSON is invalid", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		invalidJSON := []byte(`{"deveui": invalid}`)

		req, _ := http.NewRequest("POST", "/network-servers/test-server/devices", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 404 when network server not found", func(t *testing.T) {
		router, _ := setupDeviceTestRouter()

		body := map[string]string{
			"deveui":  "0102030405060708",
			"joineui": "aabbccddeeff0011",
			"appkey":  "0102030405060708090a0b0c0d0e0f10",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers/non-existent/devices", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetDeviceByEUI(t *testing.T) {
	t.Run("returns device when it exists", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		ns, _ := testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		ns.AddDevice(devEUI, joinEUI, appKey, 0)

		req, _ := http.NewRequest("GET", "/network-servers/test-server/devices/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response device.DeviceInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, devEUI, response.DevEUI)
	})

	t.Run("returns 404 when device not found", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		req, _ := http.NewRequest("GET", "/network-servers/test-server/devices/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "not found")
	})

	t.Run("returns 400 when EUI format is invalid", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		req, _ := http.NewRequest("GET", "/network-servers/test-server/devices/invalid-eui", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "invalid EUI format")
	})

	t.Run("returns 404 when network server not found", func(t *testing.T) {
		router, _ := setupDeviceTestRouter()

		req, _ := http.NewRequest("GET", "/network-servers/non-existent/devices/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDelDevice(t *testing.T) {
	t.Run("deletes device successfully", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		ns, _ := testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		devEUI := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		joinEUI := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
		appKey := lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
		ns.AddDevice(devEUI, joinEUI, appKey, 0)

		req, _ := http.NewRequest("DELETE", "/network-servers/test-server/devices/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify it was actually removed
		_, err := ns.GetDevice(devEUI)
		assert.Error(t, err)
	})

	t.Run("returns 404 when device not found", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		req, _ := http.NewRequest("DELETE", "/network-servers/test-server/devices/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "not found")
	})

	t.Run("returns 400 when EUI format is invalid", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		req, _ := http.NewRequest("DELETE", "/network-servers/test-server/devices/invalid-eui", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "invalid EUI format")
	})

	t.Run("returns 404 when network server not found", func(t *testing.T) {
		router, _ := setupDeviceTestRouter()

		req, _ := http.NewRequest("DELETE", "/network-servers/non-existent/devices/0102030405060708", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestIntegration_DeviceWorkflow(t *testing.T) {
	t.Run("complete CRUD workflow for devices", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("test-server", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		// 1. List devices - should be empty
		req, _ := http.NewRequest("GET", "/network-servers/test-server/devices", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var list1 []device.DeviceInfo
		json.Unmarshal(w.Body.Bytes(), &list1)
		assert.Equal(t, 0, len(list1))

		// 2. Create device
		body := map[string]string{
			"deveui":  "0102030405060708",
			"joineui": "aabbccddeeff0011",
			"appkey":  "0102030405060708090a0b0c0d0e0f10",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ = http.NewRequest("POST", "/network-servers/test-server/devices", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// 3. List devices - should have 1
		req, _ = http.NewRequest("GET", "/network-servers/test-server/devices", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var list2 []device.DeviceInfo
		json.Unmarshal(w.Body.Bytes(), &list2)
		assert.Equal(t, 1, len(list2))

		// 4. Get device by EUI
		req, _ = http.NewRequest("GET", "/network-servers/test-server/devices/0102030405060708", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var deviceInfo device.DeviceInfo
		json.Unmarshal(w.Body.Bytes(), &deviceInfo)
		assert.Equal(t, lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, deviceInfo.DevEUI)

		// 5. Delete device
		req, _ = http.NewRequest("DELETE", "/network-servers/test-server/devices/0102030405060708", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// 6. List devices - should be empty again
		req, _ = http.NewRequest("GET", "/network-servers/test-server/devices", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var list3 []device.DeviceInfo
		json.Unmarshal(w.Body.Bytes(), &list3)
		assert.Equal(t, 0, len(list3))
	})

	t.Run("device operations isolated per network server", func(t *testing.T) {
		router, testPool := setupDeviceTestRouter()
		testPool.Add("server-1", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})
		testPool.Add("server-2", integration.NetworkServerConfig{Type: integration.NetworkServerTypeGeneric})

		// Add device to server-1
		body := map[string]string{
			"deveui":  "0102030405060708",
			"joineui": "aabbccddeeff0011",
			"appkey":  "0102030405060708090a0b0c0d0e0f10",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/network-servers/server-1/devices", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Verify device exists in server-1
		req, _ = http.NewRequest("GET", "/network-servers/server-1/devices/0102030405060708", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify device does NOT exist in server-2
		req, _ = http.NewRequest("GET", "/network-servers/server-2/devices/0102030405060708", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
