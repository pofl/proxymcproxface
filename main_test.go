package main

import (
	"database/sql"
	"net/http"
	"net/url"
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
	providers := []string{
		"https://www.proxy-list.download/api/v1/get?type=http",
		"https://api.proxyscrape.com/?request=displayproxies&proxytype=http",
	}
	for _, provider := range providers {
		_, err := fetchProxyList(provider)
		if err != nil {
			t.Fatal(err)
		}
	}
}
