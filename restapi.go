package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ginit() *gin.Engine {
	r := gin.Default()
	r.GET("/proxies", proxyList)
	return r
}

func proxyList(c *gin.Context) {
	list, err := getProxyList()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	c.JSON(http.StatusOK, list)
}
