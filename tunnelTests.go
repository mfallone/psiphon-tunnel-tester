package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func GetSiteResource(site string, proxyConfig *ProxyConfig) (body []byte, err error) {
	var client *http.Client

	if proxyConfig.useHttpProxy == true {
		proxyUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", proxyConfig.httpProxyAddress, proxyConfig.httpProxyPort))
		if err != nil {
			log.Fatalf("Could not parse url: ", err)
		}
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	} else {
		client = &http.Client{}
	}

	req, err := http.NewRequest("Get", site, nil)
	if err != nil {
		log.Fatalf("Could not get requested resource: ", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Could not complete request: ", err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read response body: %s", err)
	}

	fmt.Println("Body: %s", string(body))
	return body, nil
}
