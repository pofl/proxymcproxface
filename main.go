package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

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
	return execSQLFile("schema.sql")
}

func execSQLFile(path string) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	queries := strings.Split(string(file), ";")

	for _, query := range queries {
		log.Print("executing query ", query)
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("error {%w} during execution of query {%s}", err, query)
		}
	}
	return nil
}

func truncateTables() error {
	if _, err := db.Exec("TRUNCATE TABLE fetch_runs"); err != nil {
		return err
	}
	_, err := db.Exec("TRUNCATE TABLE checks")
	return err
}
