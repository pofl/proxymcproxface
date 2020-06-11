package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func fetchProxyList(url *url.URL) ([]string, error) {
	res, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}

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
