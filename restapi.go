package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ginit() *gin.Engine {
	r := gin.Default()
	r.GET("/proxies", proxyList)
	r.POST("/fetch", triggerFetch)
	r.POST("/check", triggerCheck)
	r.POST("/clear", clearDB)
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

func triggerFetch(c *gin.Context) {
	fetchNow()
	c.Status(http.StatusNoContent)
}

func triggerCheck(c *gin.Context) {
	go func() {
		err := checkAll()
		if err != nil {
			log.Printf("Error while checking all known proxies: %v", err)
		}
	}()
	c.Status(http.StatusAccepted)
}

func clearDB(c *gin.Context) {
	err := truncateTables()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error(), nil)
		return
	}
	c.String(http.StatusOK, "DB cleared", nil)
}
