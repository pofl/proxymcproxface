package main

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var db *sql.DB

func main() {
	err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	err = initDB()
	if err != nil {
		log.Fatal(err)
	}
	router := ginit()
	router.Run("localhost:5000")
}

func connectDB() error {
	client, err := sql.Open("pgx", "user=postgres password=test host=localhost")
	db = client
	return err
}

func initDB() error {
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
