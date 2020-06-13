package main

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var db *sql.DB

func main() {}

func initDB(db *sql.DB) error {
	ddl := `
		CREATE TABLE IF NOT EXISTS proxy_check_results (
			proxy   TEXT      NOT NULL,
			testURL TEXT      NOT NULL,
			ts      TIMESTAMP NOT NULL,
			worked  BOOLEAN   NOT NULL
		)`
	_, err := db.Exec(ddl)
	return err
}