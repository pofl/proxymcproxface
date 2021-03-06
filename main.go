package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"

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
	router := newServer()
	addr := os.Getenv("SERVER_ADDRESS")
	if addr == "" {
		addr = "127.0.0.1:80"
	}
	router.Run(addr)
}

func connectDB() error {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "test"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	database := os.Getenv("POSTGRES_DB")
	if database == "" {
		database = "postgres"
	}
	connStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s database=%s",
		user, password, host, port, database,
	)
	client, err := sql.Open("pgx", connStr)
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

var testURLs UrlList
var providers UrlList

func init() {
	providers = UrlList{[]*url.URL{}}
	providers.overwrite([]string{
		"https://www.proxy-list.download/api/v1/get?type=http",
		"https://api.proxyscrape.com/?request=displayproxies&proxytype=http",
	})

	testURLs = UrlList{[]*url.URL{}}
	testURLs.overwrite([]string{
		"https://motherfuckingwebsite.com/",
	})
}
