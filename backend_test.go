package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var exampleCheckRes1 CheckResult
var exampleCheckRes2 CheckResult
var exampleFetchRes1 FetchResult
var exampleFetchRes2 FetchResult

func TestMain(m *testing.M) {
	proxy1, _ := net.ResolveTCPAddr("tcp4", "1.2.3.4:5")
	proxy2, _ := net.ResolveTCPAddr("tcp4", "5.6.7.8:9")
	testURL, _ := url.Parse("https://motherfuckingwebsite.com/")
	exampleCheckRes1 = CheckResult{proxy1, testURL, time.Now(), true, 0, ""}
	exampleCheckRes2 = CheckResult{proxy2, testURL, time.Now(), true, 0, ""}
	exampleFetchRes1 = FetchResult{testURL, proxy1, time.Now()}
	exampleFetchRes2 = FetchResult{testURL, proxy2, time.Now()}

	os.Exit(m.Run())
}

func TestPostgresConnection(t *testing.T) {
	require.NoError(t, connectDB())
	require.NoError(t, db.Ping())
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

	require.NoError(t, connectDB())
	require.NoError(t, initDB())
	require.True(t, tableExists(db, "checks"))
	require.True(t, tableExists(db, "fetch_runs"))
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
			res := checkProxy(proxyAddr, testURL)
			if res.worked == true {
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

func TestProxyList(t *testing.T) {
	require.NoError(t, connectDB())
	require.NoError(t, initDB())
	list, err := getProxyList()
	require.NoError(t, err)
	require.NotEmpty(t, list)
}

func TestFetch(t *testing.T) {
	require.NoError(t, connectDB())
	require.NoError(t, initDB())

	var cntBefore, cntAfter int
	row := db.QueryRow("SELECT COUNT(*) FROM fetch_runs")
	err := row.Scan(&cntBefore)
	require.NoError(t, err)

	fetchNow()

	err = row.Scan(&cntAfter)
	require.NoError(t, err)
	require.Greater(t, cntAfter, cntBefore)
}
