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
	r.GET("/providers", listProviderDetails)
	r.PUT("/providers", setProviders)
	r.POST("/fetch", triggerFetch)
	r.POST("/check", triggerCheck)
	r.POST("/clear", clearDB)
	r.StaticFile("/", "index.html")
	return r
}

func proxyList(c *gin.Context) {
	list, err := getProxyList()
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
	limitStr := c.DefaultQuery("limit", "-1")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	go func() {
		err := checkAll(limit)
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
	c.String(http.StatusNoContent, "DB cleared", nil)
}

func listProviderDetails(c *gin.Context) {
	list, err := listProviders()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, list)
}

func setProviders(c *gin.Context) {
	var list []string
	if err := c.BindJSON(&list); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if err := providers.overwrite(list); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
