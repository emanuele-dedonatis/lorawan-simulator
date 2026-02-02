package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/gin-gonic/gin"
)

const apiTimeout = 5 * time.Second

var pool *networkserver.Pool

func Init(p *networkserver.Pool) {
	pool = p
	router := gin.Default()

	// Add timeout middleware to all routes
	router.Use(timeoutMiddleware(apiTimeout))

	// GET /network-servers
	router.GET("/network-servers", getNetworkServers)

	// POST /network-servers
	router.POST("/network-servers", postNetworkServer)

	ns := router.Group("/network-servers/:name")
	ns.Use(networkServerMiddleware())
	{
		// GET /network-servers/:name
		ns.GET("", getNetworkServersByName)

		// DELETE /network-servers/:name
		ns.DELETE("", delNetworkServer)

		/*
		 *	GATEWAYS
		 */
		// GET /network-servers/:name/gateways
		ns.GET("/gateways", getGateways)

		// POST /network-servers/:name/gateways
		ns.POST("/gateways", postGateway)

		gw := ns.Group("/gateways/:eui")
		gw.Use(gatewayMiddleware())
		{
			// GET /network-servers/:name/gateways/:eui
			gw.GET("", getGatewayByEUI)

			// DELETE /network-servers/:name/gateways/:eui
			gw.DELETE("", delGateway)

			// POST /network-servers/:name/gateways/:eui/connect
			gw.POST("/connect", connectGateway)

			// POST /network-servers/:name/gateways/:eui/disconnect
			gw.POST("/disconnect", disconnectGateway)
		}

		/*
		 *	DEVICES
		 */
		// GET /network-servers/:name/devices
		ns.GET("/devices", getDevices)

		// POST /network-servers/:name/devices
		ns.POST("/devices", postDevice)

		dev := ns.Group("/devices/:eui")
		dev.Use(deviceMiddleware())
		{
			// GET /network-servers/:name/devices/:eui
			dev.GET("", getDeviceByEUI)

			// DELETE /network-servers/:name/devices/:eui
			dev.DELETE("", delDevice)

			// POST /network-servers/:name/devices/:eui/join
			dev.POST("/join", sendDeviceJoinRequest)
		}
	}

	router.Run("localhost:2208")
}

var ErrTimeout = errors.New("operation timed out")

// timeoutMiddleware adds a timeout to all HTTP requests
func timeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context with timeout context
		c.Request = c.Request.WithContext(ctx)

		// Channel to signal when handler completes
		finished := make(chan struct{})

		go func() {
			c.Next()
			close(finished)
		}()

		select {
		case <-finished:
			// Handler completed successfully
			return
		case <-ctx.Done():
			// Timeout occurred
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"message": "request timeout",
			})
			return
		}
	}
}
