package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Setup test router with a fresh pool
func setupTestRouter() (*gin.Engine, *networkserver.Pool) {
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
	}

	return router, testPool
}

func TestGetNetworkServers(t *testing.T) {
	t.Run("returns empty list when no network servers", func(t *testing.T) {
		router, _ := setupTestRouter()

		req, _ := http.NewRequest("GET", "/network-servers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []networkserver.NetworkServerInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(response))
	})

	t.Run("returns list of network servers", func(t *testing.T) {
		router, testPool := setupTestRouter()

		// Add some network servers
		testPool.Add("server-1")
		testPool.Add("server-2")
		testPool.Add("server-3")

		req, _ := http.NewRequest("GET", "/network-servers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []networkserver.NetworkServerInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(response))

		// Verify all servers are present
		names := make(map[string]bool)
		for _, info := range response {
			names[info.Name] = true
		}
		assert.True(t, names["server-1"])
		assert.True(t, names["server-2"])
		assert.True(t, names["server-3"])
	})
}

func TestGetNetworkServersByName(t *testing.T) {
	t.Run("returns network server when it exists", func(t *testing.T) {
		router, testPool := setupTestRouter()
		testPool.Add("test-server")

		req, _ := http.NewRequest("GET", "/network-servers/test-server", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response networkserver.NetworkServerInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test-server", response.Name)
		assert.Equal(t, 0, response.GatewayCount)
		assert.Equal(t, 0, response.DeviceCount)
	})

	t.Run("returns 404 when network server not found", func(t *testing.T) {
		router, _ := setupTestRouter()

		req, _ := http.NewRequest("GET", "/network-servers/non-existent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "not found")
	})
}

func TestPostNetworkServer(t *testing.T) {
	t.Run("creates network server successfully", func(t *testing.T) {
		router, testPool := setupTestRouter()

		body := map[string]string{"name": "new-server"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response networkserver.NetworkServerInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "new-server", response.Name)

		// Verify it was actually added to the pool
		_, err = testPool.Get("new-server")
		assert.NoError(t, err)
	})

	t.Run("returns 409 when adding duplicate network server", func(t *testing.T) {
		router, testPool := setupTestRouter()
		testPool.Add("existing-server")

		body := map[string]string{"name": "existing-server"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "already exists")
	})

	t.Run("returns 400 when name is missing", func(t *testing.T) {
		router, _ := setupTestRouter()

		body := map[string]string{}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/network-servers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 when JSON is invalid", func(t *testing.T) {
		router, _ := setupTestRouter()

		invalidJSON := []byte(`{"name": invalid}`)

		req, _ := http.NewRequest("POST", "/network-servers", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 when Content-Type is not JSON", func(t *testing.T) {
		router, _ := setupTestRouter()

		req, _ := http.NewRequest("POST", "/network-servers", bytes.NewBufferString("name=test"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDelNetworkServer(t *testing.T) {
	t.Run("deletes network server successfully", func(t *testing.T) {
		router, testPool := setupTestRouter()
		testPool.Add("test-server")

		req, _ := http.NewRequest("DELETE", "/network-servers/test-server", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify it was actually removed
		_, err := testPool.Get("test-server")
		assert.Error(t, err)
	})

	t.Run("returns 404 when network server not found", func(t *testing.T) {
		router, _ := setupTestRouter()

		req, _ := http.NewRequest("DELETE", "/network-servers/non-existent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "not found")
	})
}

func TestIntegration_NetworkServerWorkflow(t *testing.T) {
	t.Run("complete CRUD workflow", func(t *testing.T) {
		router, _ := setupTestRouter()

		// 1. List - should be empty
		req, _ := http.NewRequest("GET", "/network-servers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var list1 []networkserver.NetworkServerInfo
		json.Unmarshal(w.Body.Bytes(), &list1)
		assert.Equal(t, 0, len(list1))

		// 2. Create server
		body := map[string]string{"name": "workflow-server"}
		jsonBody, _ := json.Marshal(body)
		req, _ = http.NewRequest("POST", "/network-servers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// 3. Get specific server
		req, _ = http.NewRequest("GET", "/network-servers/workflow-server", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 4. List - should have 1
		req, _ = http.NewRequest("GET", "/network-servers", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var list2 []networkserver.NetworkServerInfo
		json.Unmarshal(w.Body.Bytes(), &list2)
		assert.Equal(t, 1, len(list2))

		// 5. Delete server
		req, _ = http.NewRequest("DELETE", "/network-servers/workflow-server", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// 6. List - should be empty again
		req, _ = http.NewRequest("GET", "/network-servers", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var list3 []networkserver.NetworkServerInfo
		json.Unmarshal(w.Body.Bytes(), &list3)
		assert.Equal(t, 0, len(list3))
	})

	t.Run("multiple servers are independent", func(t *testing.T) {
		router, _ := setupTestRouter()

		// Add multiple servers
		servers := []string{"server-a", "server-b", "server-c"}
		for _, name := range servers {
			body := map[string]string{"name": name}
			jsonBody, _ := json.Marshal(body)
			req, _ := http.NewRequest("POST", "/network-servers", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusCreated, w.Code)
		}

		// Delete one
		req, _ := http.NewRequest("DELETE", "/network-servers/server-b", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify others still exist
		req, _ = http.NewRequest("GET", "/network-servers/server-a", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		req, _ = http.NewRequest("GET", "/network-servers/server-c", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify deleted one is gone
		req, _ = http.NewRequest("GET", "/network-servers/server-b", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
