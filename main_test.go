package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"testing"

	"github.com/spf13/viper"
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
		err := initDB(client)
		require.NoError(t, err)
		success := tableExists(client, "proxy_check_results")
		require.True(t, success)
	})
}

func getNWorkingProxies(n int) ([]net.Addr, error) {
	providerURL := providers.list()[0]
	proxies, err := fetchProxyList(providerURL)
	if err != nil {
		return nil, err
	}
	workingProxies := []net.Addr{}
	for _, proxy := range proxies {
		proxyAddr, err := net.ResolveTCPAddr("tcp4", proxy)
		if err != nil {
			return nil, err
		}
		testURL, _ := url.Parse("https://motherfuckingwebsite.com/")
		if err == nil {
			_, err = checkProxy(proxyAddr, testURL)
			if err == nil {
				workingProxies = append(workingProxies, proxyAddr)
				if !(len(workingProxies) < n) {
					break
				}
			}
		}
	}
	if len(workingProxies) != n {
		return nil, fmt.Errorf("Not enough working proxies found")
	}
	return workingProxies, nil
}

// This test is very flaky. Proxies can stop working any time.
func TestBasicRequestWithProxy(t *testing.T) {
	_, err := getNWorkingProxies(1)
	require.NoError(t, err)
}

func TestFetchProxyList(t *testing.T) {
	for _, provider := range providers.list() {
		_, err := fetchProxyList(provider)
		require.NoError(t, err)
	}
}

func TestUpdate(t *testing.T) {
	viper.Set("proxies_take_first", 2)
	client, err := sql.Open("pgx", "user=postgres password=test host=localhost")
	require.NoError(t, err)
	db = client
	initDB(db)
	err = updateNow()
	require.NoError(t, err)
}
