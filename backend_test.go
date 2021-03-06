package main

import (
	"database/sql"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	exampleCheckRes1 = CheckResult{proxy1, testURL, time.Now(), time.Now(), true, 0, ""}
	exampleCheckRes2 = CheckResult{proxy2, testURL, time.Now(), time.Now(), true, 0, ""}
	exampleFetchRes1 = FetchResult{testURL, proxy1, time.Now()}
	exampleFetchRes2 = FetchResult{testURL, proxy2, time.Now()}

	os.Exit(m.Run())
}

func TestPostgresConnection(t *testing.T) {
	assert.NoError(t, connectDB())
	connectionPossible := assert.NoError(t, db.Ping())
	if !connectionPossible {
		t.Fatal(
			"These tests require a Postgres instance running on localhost. Run `docker-compose up`.")
	}
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
