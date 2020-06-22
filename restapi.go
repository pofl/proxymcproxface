package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func newServer() *gin.Engine {
	r := gin.Default()
	r.StaticFile("/", "index.html")
	r.POST("/fetch", invokeFetch)
	r.POST("/check", invokeCheck)
	r.POST("/clear", invokeClearDB)
	r.GET("/proxies", getProxies)
	r.GET("/providers", getProviders)
	r.PUT("/providers", putProviders)
	r.GET("/testurls", getTestURLs)
	r.PUT("/testurls", putTestURLs)
	return r
}

func getProxies(c *gin.Context) {
	list, err := getProxyList()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, list)
}

func invokeFetch(c *gin.Context) {
	fetchNow()
	c.Status(http.StatusNoContent)
}

func invokeCheck(c *gin.Context) {
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

func invokeClearDB(c *gin.Context) {
	err := truncateTables()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error(), nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func getProviders(c *gin.Context) {
	list, err := listProviders()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, list)
}

func putProviders(c *gin.Context) {
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

func getTestURLs(c *gin.Context) {
	list := testURLs.list()
	strList := []string{}
	for _, url := range list {
		strList = append(strList, url.String())
	}
	c.JSON(http.StatusOK, strList)
}

func putTestURLs(c *gin.Context) {
	var list []string
	if err := c.BindJSON(&list); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if err := testURLs.overwrite(list); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
