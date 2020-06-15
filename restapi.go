package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ginit() *gin.Engine {
	r := gin.Default()
	r.GET("/proxies", proxyList)
	return r
}

func proxyList(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "-1")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	list, err := getProxyList(limit)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, list)
}
