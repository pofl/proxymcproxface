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
)

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

type CheckResult struct {
	proxy      net.Addr
	testURL    *url.URL
	ts         time.Time
	worked     bool
	statusCode int
	errorMsg   string
}

// limit < 0 means go all the way
func checkAll(limit int) error {
	checkOne := func(proxy net.Addr, testURL *url.URL) {
		checkRes := checkProxy(proxy, testURL)
		_ = saveCheckToDB(checkRes) // just drop it if it can't be saved
	}

	checkLoop := func(proxies []net.Addr) {
		cnt := 0
		for _, proxy := range proxies {
			for _, testURL := range testURLs.list() {
				go checkOne(proxy, testURL)
				cnt++
				if limit >= 0 && cnt >= limit {
					return
				}
				time.Sleep(1 * time.Second)
			}
		}
	}

	proxies, err := retrieveDistinctProxies()
	go checkLoop(proxies)
	return err
}

func checkProxy(proxy net.Addr, testURL *url.URL) CheckResult {
	res := CheckResult{proxy, testURL, time.Now(), false, 0, ""}
	proxyURL, err := url.Parse("http://" + proxy.String())
	if err != nil {
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
		res.errorMsg = err.Error()
		return res
	}
	defer response.Body.Close()
	statusCode := response.StatusCode
	if statusCode >= 200 && statusCode < 300 {
		res.worked = true
	} else {
		res.worked = false
		if body, err := ioutil.ReadAll(response.Body); err == nil {
			res.errorMsg = string(body)
		} else {
			res.errorMsg = err.Error()
		}
	}
	return res
}

func saveCheckToDB(res CheckResult) error {
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

type FetchResult struct {
	providerURL *url.URL
	proxy       net.Addr
	ts          time.Time
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

func fetchProxiesFromProvider(prov *url.URL) ([]FetchResult, error) {
	res := []FetchResult{}
	proxies, err := fetchProxyList(prov)
	if err != nil {
		return nil, err
	}
	for _, proxy := range proxies {
		addr, err := net.ResolveTCPAddr("tcp4", proxy)
		if err == nil {
			fetch := FetchResult{prov, addr, time.Now()}
			res = append(res, fetch)
		}
	}
	return res, nil
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

func saveFetchToDB(fetch FetchResult) error {
	insertStmt := "INSERT INTO fetch_runs VALUES ($1, $2, $3)"
	_, err := db.Exec(insertStmt, fetch.providerURL.String(), fetch.proxy.String(), fetch.ts)
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

type ProxyListItem struct {
	Proxy       string
	LastSuccess time.Time
	LastSeen    time.Time
	FirstSeen   time.Time
	// ErrorMsg
	// Success
}

func getProxyList() ([]ProxyListItem, error) {
	query, err := ioutil.ReadFile("sql/query_proxy_details.sql")
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(string(query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []ProxyListItem{}
	for rows.Next() {
		item := ProxyListItem{}
		err = rows.Scan(&item.Proxy, &item.LastSuccess, &item.LastSeen, &item.FirstSeen)
		if err != nil {
			return res, err
		}
		res = append(res, item)
	}
	err = rows.Err()
	return res, err
}

type ProviderDetails struct {
	Provider  string
	LastFetch time.Time
}

func listProviders() ([]ProviderDetails, error) {
	res := []ProviderDetails{}

	query, err := ioutil.ReadFile("sql/query_provider_details.sql")
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(string(query))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var prov ProviderDetails
		err := rows.Scan(&prov.Provider, &prov.LastFetch)
		if err != nil {
			// there really shouldn't be an error here so make it very visible if there ever is
			log.Fatal(err)
		}
		res = append(res, prov)
	}
	return res, rows.Err()
}
