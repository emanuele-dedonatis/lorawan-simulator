package api

import (
	"net/http"

	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/gateway"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/gin-gonic/gin"
)

// Middleware that validates gateway exists and stores it in context
func gatewayMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ns := c.MustGet("networkServer").(*networkserver.NetworkServer)

		// Parse string to EUI64
		var eui lorawan.EUI64
		if err := eui.UnmarshalText([]byte(c.Param("eui"))); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid EUI format"})
			c.Abort()
			return
		}

		gw, err := ns.GetGateway(eui)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		// Store the network server in context for handlers to use
		c.Set("gateway", gw)
		c.Next()
	}
}

func getGateways(c *gin.Context) {
	ns := c.MustGet("networkServer").(*networkserver.NetworkServer)

	c.IndentedJSON(http.StatusOK, ns.ListGateways())
}

func postGateway(c *gin.Context) {
	ns := c.MustGet("networkServer").(*networkserver.NetworkServer)

	var json struct {
		EUI          string            `json:"eui" binding:"required"`
		DiscoveryURI string            `json:"discoveryUri" binding:"required"`
		Headers      map[string]string `json:"headers"`
		Latitude     *float64          `json:"latitude"`
		Longitude    *float64          `json:"longitude"`
	}

	if err := c.Bind(&json); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Parse string to EUI64
	var eui lorawan.EUI64
	if err := eui.UnmarshalText([]byte(json.EUI)); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid EUI format"})
		return
	}

	// Convert headers map to http.Header if provided
	var headers http.Header
	if len(json.Headers) > 0 {
		headers = make(http.Header)
		for k, v := range json.Headers {
			headers.Set(k, v)
		}
	}

	// Prepare location if provided
	var location *gateway.Location
	if json.Latitude != nil && json.Longitude != nil {
		location = &gateway.Location{
			Latitude:  *json.Latitude,
			Longitude: *json.Longitude,
		}
	}

	// Create gateway with optional location and headers
	gw, err := ns.AddGateway(eui, json.DiscoveryURI, location, headers)

	if err != nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, gw.GetInfo())
}

func getGatewayByEUI(c *gin.Context) {
	gw := c.MustGet("gateway").(*gateway.Gateway)

	c.IndentedJSON(http.StatusOK, gw.GetInfo())
}

func delGateway(c *gin.Context) {
	ns := c.MustGet("networkServer").(*networkserver.NetworkServer)
	gw := c.MustGet("gateway").(*gateway.Gateway)

	err := ns.RemoveGateway(gw.GetInfo().EUI)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, nil)
}

func connectGateway(c *gin.Context) {
	gw := c.MustGet("gateway").(*gateway.Gateway)

	err := gw.Connect()

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, nil)
}

func disconnectGateway(c *gin.Context) {
	gw := c.MustGet("gateway").(*gateway.Gateway)

	err := gw.Disconnect()

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, nil)
}
