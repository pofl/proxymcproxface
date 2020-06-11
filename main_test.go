package main

import (
	"database/sql"
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

var providers = []string{
	"https://www.proxy-list.download/api/v1/get?type=http",
	"https://api.proxyscrape.com/?request=displayproxies&proxytype=http",
}

// This test is very flaky. Proxies can stop working any time.
func TestBasicRequestWithProxy(t *testing.T) {
	providerURL, err := url.Parse(providers[0])
	if err != nil {
		t.Fatal(err)
	}
	proxies, err := fetchProxyList(providerURL)
	if err != nil {
		t.Fatal(err)
	}
	foundAWorkingProxy := false
	for _, proxy := range proxies {
		proxyURL, err := url.Parse("http://" + proxy)
		if err == nil {
			err = testProxy(proxyURL)
			if err == nil {
				foundAWorkingProxy = true
				break
			}
		}
	}
	if !foundAWorkingProxy {
		t.Fatal("No working proxy found")
	}
}

func TestRequestProxyList(t *testing.T) {
	for _, provider := range providers {
		providerURL, err := url.Parse(provider)
		if err != nil {
			t.Fatal(err)
		}
		_, err = fetchProxyList(providerURL)
		if err != nil {
			t.Fatal(err)
		}
	}
}
