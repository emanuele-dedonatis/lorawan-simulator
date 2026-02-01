package api

import (
	"net/http"

	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
	"github.com/gin-gonic/gin"
)

func getNetworkServers(c *gin.Context) {
	networkservers := pool.List()
	res := make([]networkserver.NetworkServerInfo, 0, len(networkservers))
	for _, info := range networkservers {
		res = append(res, info.GetInfo())
	}
	c.IndentedJSON(http.StatusOK, res)
}

func getNetworkServersByName(c *gin.Context) {
	name := c.Param("name")

	ns, err := pool.Get(name)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, ns.GetInfo())
}

func postNetworkServer(c *gin.Context) {
	var json struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.Bind(&json); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ns, err := pool.Add(json.Name)
	if err != nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, ns.GetInfo())
}

func delNetworkServer(c *gin.Context) {
	name := c.Param("name")

	err := pool.Remove(name)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, nil)
}
