package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var testURLs UrlList
var providers UrlList

func init() {
	providers = UrlList{[]*url.URL{}}
	providers.overwrite([]string{
		"https://www.proxy-list.download/api/v1/get?type=http",
		// "https://api.proxyscrape.com/?request=displayproxies&proxytype=http",
	})

	testURLs = UrlList{[]*url.URL{}}
	testURLs.overwrite([]string{
		"https://motherfuckingwebsite.com/",
	})
}

type checkResult struct {
	proxy      net.Addr
	testURL    *url.URL
	ts         time.Time
	worked     bool
	statusCode int
	errorMsg   string
}

func saveFetchToDB(fetch fetchResult) error {
	insertStmt := "INSERT INTO fetch_runs VALUES ($1, $2, $3)"
	_, err := db.Exec(insertStmt, fetch.providerURL.String(), fetch.proxy.String(), fetch.ts)
	return err
}

func fetchNow() {
	for _, prov := range providers.list() {
		list, err := fetchProxiesFromProvider(prov)
		if err != nil {
			log.Printf("can't fetch from %v, err: %v", prov, err)
		}
		for _, fetch := range list {
			err := saveFetchToDB(fetch)
			if err != nil {
				log.Printf("Error during writing %+v to DB: %v", fetch, err)
			}
		}
	}
}

func updateNow() error {
	viper.SetDefault("proxies_take_first", 0)
	limit := viper.GetInt("proxies_take_first")
	for _, prov := range providers.list() {
		list, err := fetchProxiesFromProvider(prov)
		if err != nil {
			log.Fatal(err)
		}
		var actualList []fetchResult
		if limit > 0 {
			actualList = list[:limit]
		} else {
			actualList = list
		}
		for _, fetch := range actualList {
			for _, testURL := range testURLs.list() {
				res := checkProxy(fetch.proxy, testURL)
				err = saveCheckToDB(res)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}
	return nil
}

func saveCheckToDB(res checkResult) error {
	insertStmt :=
		"INSERT INTO checks VALUES ($1, $2, $3, $4, $5, $6)"
	log.Printf(
		"%v | %v | %v | %v | %v | %v",
		res.proxy.String(), res.testURL.String(), res.ts, res.worked, res.statusCode, res.errorMsg,
	)
	_, err := db.Exec(
		insertStmt,
		res.proxy.String(), res.testURL.String(), res.ts, res.worked, res.statusCode, res.errorMsg,
	)
	return err
}

func retrieveDistinctProxies() ([]net.Addr, error) {
	getProxyList := "SELECT DISTINCT proxy FROM fetch_runs"
	rows, err := db.Query(getProxyList)
	if err != nil {
		log.Fatal(err)
	}
	proxyList := []net.Addr{}
	for rows.Next() {
		var proxy string
		rows.Scan(&proxy)
		parsed, err := net.ResolveTCPAddr("tcp4", proxy)
		if err != nil {
			log.Fatal(
				"got an invalid proxy address from the DB although that should be impossible")
		}
		proxyList = append(proxyList, parsed)
	}
	return proxyList, rows.Err()
}

func checkAll() error {
	proxies, err := retrieveDistinctProxies()
	for _, proxy := range proxies {
		for _, testURL := range testURLs.list() {
			checkRes := checkProxy(proxy, testURL)
			_ = saveCheckToDB(checkRes) // just drop it if it can't be saved
		}
	}
	return err
}

func checkProxy(proxy net.Addr, testURL *url.URL) checkResult {
	res := checkResult{proxy, testURL, time.Now(), true, 0, ""}
	proxyURL, err := url.Parse("http://" + proxy.String())
	if err != nil {
		res.worked = false
		res.errorMsg = err.Error()
		return res
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	response, err := client.Get(testURL.String())
	if err != nil {
		res.worked = false
		res.errorMsg = err.Error()
		return res
	}
	defer response.Body.Close()
	statusCode := response.StatusCode
	if statusCode < 200 || statusCode >= 300 {
		if body, err := ioutil.ReadAll(response.Body); err == nil {
			res.errorMsg = string(body)
		} else {
			res.errorMsg = err.Error()
		}
		res.worked = false
	}
	return res
}

type fetchResult struct {
	providerURL *url.URL
	proxy       net.Addr
	ts          time.Time
}

func fetchProxiesFromProvider(prov *url.URL) ([]fetchResult, error) {
	res := []fetchResult{}
	list, err := fetchProxyList(prov)
	if err != nil {
		return nil, err
	}
	for _, p := range list {
		addr, err := net.ResolveTCPAddr("tcp4", p)
		if err == nil {
			fetch := fetchResult{prov, addr, time.Now()}
			res = append(res, fetch)
		}
	}
	return res, nil
}

type UrlList struct{ urls []*url.URL }

func (list *UrlList) overwrite(newList []string) error {
	newURLs := []*url.URL{}
	for _, str := range newList {
		url, err := url.Parse(str)
		if err != nil {
			return err
		}
		newURLs = append(newURLs, url)
	}
	list.urls = newURLs
	return nil
}

func (list UrlList) list() []*url.URL {
	return list.urls
}

func fetchProxyList(provider *url.URL) ([]string, error) {
	res, err := http.Get(provider.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyContent, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	proxyHostsLF := strings.Split(string(bodyContent), "\n")
	proxyHostsCRLF := strings.Split(string(bodyContent), "\r\n")

	if len(proxyHostsLF) <= 1 && len(proxyHostsCRLF) <= 1 {
		return nil, fmt.Errorf(
			"Response from proxy list didn't contain proxies. Response was:\n%s", string(bodyContent))
	}

	lineEndingsAreCRLF := len(proxyHostsCRLF) == len(proxyHostsLF)
	if lineEndingsAreCRLF {
		return proxyHostsCRLF, nil
	}
	lineEndingsAreLF := len(proxyHostsLF) > len(proxyHostsCRLF) && len(proxyHostsCRLF) <= 1
	if lineEndingsAreLF {
		return proxyHostsLF, nil
	}

	return nil, fmt.Errorf("No idea how we got to this point in the code ...")
}

type ProxyListItem struct {
	Proxy   string
	TestURL string
	TS      string
	Worked  bool
}

func getProxyList(limit int) ([]ProxyListItem, error) {
	query := "SELECT * FROM proxy_check_results"
	if limit > 0 {
		query = query + fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []ProxyListItem{}
	for rows.Next() {
		item := ProxyListItem{}
		// var proxy, testURL, ts string
		// var worked bool
		err = rows.Scan(&item.Proxy, &item.TestURL, &item.TS, &item.Worked)
		if err != nil {
			log.Fatal("Scan didn't work")
		}
		res = append(res, item)
	}
	err = rows.Err()
	return res, err
}
