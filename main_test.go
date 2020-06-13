package main

import (
	"database/sql"
	"net/url"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/require"
)

func withDB(t *testing.T, f func(*sql.DB)) error {
	client, err := sql.Open("pgx", "user=postgres password=test host=localhost")
	require.NoError(t, err)
	f(client)
	return nil
}

func TestPostgresConnection(t *testing.T) {
	withDB(t, func(client *sql.DB) {
		err := client.Ping()
		require.NoError(t, err)
	})
}

func TestDBInit(t *testing.T) {
	tableExists := func(client *sql.DB, tableName string) bool {
		tableCreatedSuccessfully := false
		rows, err := client.Query(
			"SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
		require.NoError(t, err)
		defer rows.Close()
		for rows.Next() {
			var name string
			err := rows.Scan(&name)
			require.NoError(t, err)
			if name == tableName {
				tableCreatedSuccessfully = true
			}
		}
		err = rows.Err()
		require.NoError(t, err)
		return tableCreatedSuccessfully
	}
	withDB(t, func(client *sql.DB) {
		initDB(client)
		success := tableExists(client, "proxies")
		require.True(t, success)
	})
}

var providers = []string{
	"https://www.proxy-list.download/api/v1/get?type=http",
	"https://api.proxyscrape.com/?request=displayproxies&proxytype=http",
}

// This test is very flaky. Proxies can stop working any time.
func TestBasicRequestWithProxy(t *testing.T) {
	providerURL, err := url.Parse(providers[0])
	require.NoError(t, err)
	proxies, err := fetchProxyList(providerURL)
	require.NoError(t, err)
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

func TestFetchProxyList(t *testing.T) {
	for _, provider := range providers {
		providerURL, err := url.Parse(provider)
		require.NoError(t, err)
		_, err = fetchProxyList(providerURL)
		require.NoError(t, err)
	}
}
