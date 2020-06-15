package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProxyListEndpoint(t *testing.T) {
	require.NoError(t, connectDB())
	require.NoError(t, initDB())

	// populate DB to have at least 2 records
	_ = saveCheckToDB(exampleCheckRes)
	anotherEx := exampleCheckRes
	anotherEx.ts = time.Now()
	err := saveCheckToDB(anotherEx)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", "localhost:5000/proxies", nil)
	require.NoError(t, err)
	limitedReq, err := http.NewRequest("GET", "localhost:5000/proxies?limit=1", nil)
	require.NoError(t, err)

	checkResponse := func(
		name string, req *http.Request,
		f func(res *httptest.ResponseRecorder),
	) {
		t.Run(name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rr)
			c.Request = req
			proxyList(c)
			f(rr)
		})
	}

	checkResponse("Happy path", req, func(res *httptest.ResponseRecorder) {
		require.Equal(t, 200, res.Result().StatusCode)
		var got []gin.H
		err = json.Unmarshal(res.Body.Bytes(), &got)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(got), 2)
	})

	checkResponse("Happy path with limit", limitedReq, func(res *httptest.ResponseRecorder) {
		require.Equal(t, 200, res.Result().StatusCode)
		var got []gin.H
		err = json.Unmarshal(res.Body.Bytes(), &got)
		require.NoError(t, err)
		require.Equal(t, 1, len(got))
	})

	// sad path
	client, _ := sql.Open("pgx", "user=foo password=bar host=baz")
	require.Error(t, client.Ping())
	db = client
	checkResponse("Sad path", req, func(res *httptest.ResponseRecorder) {
		require.Equal(t, 500, res.Result().StatusCode)
	})
}
