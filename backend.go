package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func fetchProxyList(url string) ([]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	bodyContent, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	proxyHosts := strings.Split(string(bodyContent), "\n")

	if len(proxyHosts) <= 1 {
		return nil, fmt.Errorf(
			"Response from proxy list didn't contain proxies. Response was:\n%s", string(bodyContent))
	}

	return proxyHosts, nil
}
