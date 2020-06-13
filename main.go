package main

import "database/sql"

func main() {

}

func initDB(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS proxies (hostpost TEXT)")
	return err
}
