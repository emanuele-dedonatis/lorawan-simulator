package api

import (
	"net/http"

	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/device"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/gin-gonic/gin"
)

// Middleware that validates device exists and stores it in context
func deviceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ns := c.MustGet("networkServer").(*networkserver.NetworkServer)

		// Parse string to EUI64
		var eui lorawan.EUI64
		if err := eui.UnmarshalText([]byte(c.Param("eui"))); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid EUI format"})
			c.Abort()
			return
		}

		dev, err := ns.GetDevice(eui)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		// Store the device in context for handlers to use
		c.Set("device", dev)
		c.Next()
	}
}

func getDevices(c *gin.Context) {
	ns := c.MustGet("networkServer").(*networkserver.NetworkServer)

	c.IndentedJSON(http.StatusOK, ns.ListDevices())
}

func postDevice(c *gin.Context) {
	ns := c.MustGet("networkServer").(*networkserver.NetworkServer)

	var json struct {
		DevEui   string `json:"deveui" binding:"required"`
		JoinEUI  string `json:"joineui" binding:"required"`
		AppKey   string `json:"appkey" binding:"required"`
		DevNonce uint16 `json:"devnonce"`
	}

	if err := c.Bind(&json); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Parse DevEui to EUI64
	var deveui lorawan.EUI64
	if err := deveui.UnmarshalText([]byte(json.DevEui)); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid DevEui format"})
		return
	}

	// Parse JoinEUI to EUI64
	var joineui lorawan.EUI64
	if err := joineui.UnmarshalText([]byte(json.JoinEUI)); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JoinEUI format"})
		return
	}

	// Parse AppKey to AES128Key
	var appkey lorawan.AES128Key
	if err := appkey.UnmarshalText([]byte(json.AppKey)); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid AppKey format"})
		return
	}

	dev, err := ns.AddDevice(deveui, joineui, appkey, lorawan.DevNonce(json.DevNonce))
	if err != nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, dev.GetInfo())
}

func getDeviceByEUI(c *gin.Context) {
	dev := c.MustGet("device").(*device.Device)

	c.IndentedJSON(http.StatusOK, dev.GetInfo())
}

func delDevice(c *gin.Context) {
	ns := c.MustGet("networkServer").(*networkserver.NetworkServer)
	dev := c.MustGet("device").(*device.Device)

	err := ns.RemoveDevice(dev.GetInfo().DevEUI)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, nil)
}

func sendDeviceJoinRequest(c *gin.Context) {
	ns := c.MustGet("networkServer").(*networkserver.NetworkServer)
	dev := c.MustGet("device").(*device.Device)

	err := ns.SendJoinRequest(dev.GetInfo().DevEUI)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, nil)
}
