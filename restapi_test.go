package main

import (
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProxyListEndpoint(t *testing.T) {
	require.NoError(t, connectDB())
	require.NoError(t, initDB())

	// happy path
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	proxyList(c)
	require.Equal(t, 200, w.Result().StatusCode)

	var got []gin.H
	err := json.Unmarshal(w.Body.Bytes(), &got)
	require.NoError(t, err)

	// sad path
	client, _ := sql.Open("pgx", "user=foo password=bar host=baz")
	require.Error(t, client.Ping())
	db = client

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	proxyList(c2)
	require.Equal(t, 500, w2.Result().StatusCode)
}
