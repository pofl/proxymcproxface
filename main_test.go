package main

import (
	"database/sql"
	"net/http"
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
	res, err := http.Get("https://blog.fefe.de")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatal("Status code of response is ", res.StatusCode)
	}
}
