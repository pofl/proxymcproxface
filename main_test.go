package main

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func TestPostgresConnection(t *testing.T) {
	client, err := sql.Open("pgx", "user=postgres password=test host=localhost")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Ping()
	if err != nil {
		t.Fatal(err)
	}
}

func TestBasicRequestWithProxy(t *testing.T) {
	//creating the proxyURL
	proxyURL, err := url.Parse("http://103.28.121.58:80")
	if err != nil {
		t.Fatal(err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client := &http.Client{
		Transport: transport,
	}

	response, err := client.Get("http://blog.fefe.de")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatal("Status code of response is ", response.StatusCode)
	}
}

func TestRequestProxyList(t *testing.T) {
	res, err := http.Get("https://www.proxy-list.download/api/v1/get?type=http")
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	proxyHosts := strings.Split(string(data), "\n")

	if len(proxyHosts) <= 1 {
		t.Fatal("Response from proxy list didn't contain proxies. Response was:\n", string(data))
	}

	t.Log(proxyHosts)
}
