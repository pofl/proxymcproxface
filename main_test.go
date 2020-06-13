package main

import (
	"database/sql"
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

// This test is very flaky. Proxies can stop working any time.
func TestBasicRequestWithProxy(t *testing.T) {
	providerURL := providers.list()[0]
	proxies, err := fetchProxyList(providerURL)
	require.NoError(t, err)
	foundAWorkingProxy := false
	for _, proxy := range proxies {
		proxyURL, err := net.ResolveTCPAddr("tcp4", proxy)
		require.NoError(t, err)
		testURL, _ := url.Parse("https://motherfuckingwebsite.com/")
		if err == nil {
			_, err = checkProxy(proxyURL, testURL)
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
