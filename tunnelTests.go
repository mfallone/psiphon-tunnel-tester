package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func GetWebResource(site string, useProxy bool, localHttpProxyAddress string, localHttpProxyPort int) (body []byte, err error) {
	var client *http.Client
	if useProxy == true {
		proxyUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", localHttpProxyAddress, localHttpProxyPort))
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
