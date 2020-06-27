package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProxyListEndpoint(t *testing.T) {
	require.NoError(t, connectDB())
	require.NoError(t, initDB())
	server := newServer()

	// populate DB to have at least 2 records
	_ = saveFetchToDB(exampleFetchRes1)
	_ = saveFetchToDB(exampleFetchRes2)
	_ = saveCheckToDB(exampleCheckRes1)
	err := saveCheckToDB(exampleCheckRes2)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/proxies", nil)

	checkResponse := func(
		testCaseName string, req *http.Request,
		f func(res *httptest.ResponseRecorder),
	) {
		t.Run(testCaseName, func(t *testing.T) {
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)
			f(rr)
		})
	}

	checkResponse("Happy path", req, func(res *httptest.ResponseRecorder) {
		require.Equal(t, 200, res.Code)
		var got []gin.H
		err = json.Unmarshal(res.Body.Bytes(), &got)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(got), 2)
	})

	// sad path
	client, _ := sql.Open("pgx", "user=foo password=bar host=baz")
	require.Error(t, client.Ping())
	db = client
	checkResponse("Sad path", req, func(res *httptest.ResponseRecorder) {
		require.Equal(t, 500, res.Code)
	})
}

func TestFetchEndpoint(t *testing.T) {
	require.NoError(t, connectDB())
	require.NoError(t, initDB())
	gin := newServer()

	cntQuery := "SELECT COUNT(*) FROM fetch_runs"
	var cntBefore, cntAfter int
	require.NoError(t, db.QueryRow(cntQuery).Scan(&cntBefore))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/fetch", nil)
	gin.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Code)

	require.NoError(t, db.QueryRow(cntQuery).Scan(&cntAfter))
	require.Greater(t, cntAfter, cntBefore)
}

func TestCheckEndpoint(t *testing.T) {
	require.NoError(t, connectDB())
	require.NoError(t, initDB())
	gin := newServer()

	cntQuery := "SELECT COUNT(*) FROM checks"
	var cntBefore int
	row := db.QueryRow(cntQuery)
	err := row.Scan(&cntBefore)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/check?limit=4", nil)
	gin.ServeHTTP(rr, req)
	require.Equal(t, http.StatusAccepted, rr.Code)
	time.Sleep(10 * time.Second)

	var cntAfter int
	row = db.QueryRow(cntQuery)
	err = row.Scan(&cntAfter)
	require.NoError(t, err)
	require.Greater(t, cntAfter, cntBefore)
}

func TestProviderDetails(t *testing.T) {
	require.NoError(t, connectDB())
	require.NoError(t, initDB())
	server := newServer()

	// populate DB to have at least 2 records
	_ = saveFetchToDB(exampleFetchRes1)
	_ = saveFetchToDB(exampleFetchRes2)
	_ = saveCheckToDB(exampleCheckRes1)
	err := saveCheckToDB(exampleCheckRes2)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/providers", nil)
	server.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var got []gin.H
	err = json.Unmarshal(rr.Body.Bytes(), &got)
	require.NoError(t, err)
	require.Greater(t, len(got), 0)
}

func TestTestURLEndpoints(t *testing.T) {
	require.NoError(t, connectDB())
	require.NoError(t, initDB())
	server := newServer()

	urls := []string{
		"https://motherfuckingwebsite.com/",
		"http://txti.es/",
	}
	payload, err := json.Marshal(urls)
	require.NoError(t, err)
	putRR := httptest.NewRecorder()
	putReq := httptest.NewRequest("PUT", "/testurls", strings.NewReader(string(payload)))
	getRR := httptest.NewRecorder()
	getReq := httptest.NewRequest("GET", "/testurls", nil)

	server.ServeHTTP(putRR, putReq)
	require.Equal(t, http.StatusNoContent, putRR.Result().StatusCode)
	server.ServeHTTP(getRR, getReq)

	var respURLs []TestURLCheckResult
	err = json.Unmarshal(getRR.Body.Bytes(), &respURLs)
	require.NoError(t, err)
	for _, u := range respURLs {
		require.Contains(t, urls, u.TestURL)
	}
}
