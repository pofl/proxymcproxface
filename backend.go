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
	for _, str := range []string{
		"https://www.proxy-list.download/api/v1/get?type=http",
		// "https://api.proxyscrape.com/?request=displayproxies&proxytype=http",
	} {
		_ = providers.addStr(str)
	}

	testURLs = UrlList{[]*url.URL{}}
	for _, str := range []string{
		"https://motherfuckingwebsite.com/",
	} {
		_ = testURLs.addStr(str)
	}
}

type checkResult struct {
	proxy   net.Addr
	testURL *url.URL
	ts      time.Time
	worked  bool
}

func saveFetchToDB(fetch fetchResult) error {
	insertStmt := "INSERT INTO fetch_runs VALUES ($1, $2, $3)"
	_, err := db.Exec(insertStmt, fetch.providerURL.String(), fetch.proxy.String(), fetch.ts)
	return err
}

func fetchNow() error {
	for _, prov := range providers.list() {
		list, err := fetchProxiesFromProvider(prov)
		if err != nil {
			log.Fatal(err)
		}
		for _, fetch := range list {
			err := saveFetchToDB(fetch)
			if err != nil {
				log.Printf("Error during writing %+v to DB: %v", fetch, err)
			}
		}
	}
	return nil
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
				res, err := checkProxy(fetch.proxy, testURL)
				if err != nil {
					log.Print(err)
				}
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
		"INSERT INTO proxy_check_results(proxy, testURL, ts, worked) VALUES ($1, $2, $3, $4)"
	log.Printf("%v | %v | %v | %v", res.proxy.String(), res.testURL.String(), res.ts, res.worked)
	_, err := db.Exec(insertStmt, res.proxy.String(), res.testURL.String(), res.ts, res.worked)
	return err
}

func checkProxy(proxy net.Addr, testURL *url.URL) (checkResult, error) {
	res := checkResult{proxy, testURL, time.Now(), false}
	proxyURL, err := url.Parse("http://" + proxy.String())
	if err != nil {
		return res, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	response, err := client.Get(testURL.String())
	defer response.Body.Close()
	if err != nil {
		return res, err
	}
	statusCode := response.StatusCode
	if statusCode < 200 || statusCode >= 300 {
		body := []byte{}
		_, _ = response.Body.Read(body)
		return res, fmt.Errorf("Status code of response is %d, body is %v", statusCode, body)
	}
	res.worked = true
	return res, nil
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

func (list *UrlList) overwrite(newList []*url.URL) {
	list.urls = newList
}

func (list UrlList) list() []*url.URL {
	return list.urls
}

func (list *UrlList) addStr(urlStr string) error {
	url, err := url.Parse(urlStr)
	if err == nil {
		list.urls = append(list.urls, url)
	}
	return err
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
