package api

import (
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/gin-gonic/gin"
)

var pool *networkserver.Pool

func Init(p *networkserver.Pool) {
	pool = p

	router := gin.Default()

	// Network Servers
	router.GET("/network-servers", getNetworkServers)
	router.GET("/network-servers/:name", getNetworkServersByName)
	router.POST("/network-servers", postNetworkServer)
	router.DELETE("/network-servers/:name", delNetworkServer)

	router.Run("localhost:8080")
}
