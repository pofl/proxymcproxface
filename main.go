package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

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
	router.Run(":5000")
}

func connectDB() error {
	client, err := sql.Open("pgx", "user=postgres password=test host=localhost")
	db = client
	return err
}

func initDB() error {
	sqlFilesToRun := []string{
		"sql/table_fetch_runs.sql",
		"sql/table_checks.sql",
	}
	for _, path := range sqlFilesToRun {
		query, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = db.Exec(string(query))
		if err != nil {
			return fmt.Errorf("error {%w} during execution of query {%s}", err, query)
		}
	}
	return nil
}

type UrlList struct{ urls []*url.URL }

func (list *UrlList) overwrite(newList []string) error {
	newURLs := []*url.URL{}
	for _, str := range newList {
		url, err := url.Parse(str)
		if err != nil {
			return err
		}
		if url.Hostname() == "" {
			return fmt.Errorf("%s is not a URL", str)
		}
		newURLs = append(newURLs, url)
	}
	list.urls = newURLs
	return nil
}

func (list UrlList) list() []*url.URL {
	return list.urls
}

func execSQLFile(path string) error {
	query, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(query))
	if err != nil {
		return fmt.Errorf("error {%w} during execution of query {%s}", err, query)
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
