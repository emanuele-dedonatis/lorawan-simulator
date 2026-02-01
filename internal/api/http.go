package api

import (
	"errors"
	"time"

	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/gin-gonic/gin"
)

const apiTimeout = 5 * time.Second

var pool *networkserver.Pool

func Init(p *networkserver.Pool) {
	pool = p
	router := gin.Default()

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
	}

	router.Run("localhost:2208")
}

var ErrTimeout = errors.New("operation timed out")

func waitForResult(reply <-chan error) error {
	timer := time.NewTimer(apiTimeout)
	defer timer.Stop()

	select {
	case err := <-reply:
		return err
	case <-timer.C:
		return ErrTimeout
	}
}
