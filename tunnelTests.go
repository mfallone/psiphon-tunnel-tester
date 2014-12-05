package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	psiphon "github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon"
)

// RunTests runs all tests to the server conatined in decodedServerEntry
func RunTests(config *psiphon.Config, decodedServerEntry *psiphon.ServerEntry) (result string, err error) {
	testsConfig := new(TestsConfig)

	pendingConns := new(psiphon.Conns)

	proxyConfig := &ProxyConfig{httpProxyAddress: "127.0.0.1",
		httpProxyPort: config.LocalHttpProxyPort,
		useHttpProxy:  false}

	// Get the untunneled IP address
	untunneledCheck, err := getSiteResource(testsConfig.ipcheck_site, proxyConfig)
	if err != nil {
		log.Println("Could not get site resource: ", err)
	}
	log.Println("Untunneled IP: ", string(untunneledCheck))

	// Build a tunnel to a psiphon server
	tunnel, err := psiphon.EstablishTunnel(config, pendingConns, decodedServerEntry)
	if err != nil {
		log.Fatalf("Could not establish tunnel: %s", err)
	}

	// Setup new HTTP proxy. Close() is handled by HttpProxy.serve()
	// and does not need to be called here.
	_, err = psiphon.NewHttpProxy(config, tunnel)
	if err != nil {
		log.Fatalf("error initializing local HTTP proxy: %s", err)
	}

	proxyConfig.useHttpProxy = true
	tunneledCheck, err := getSiteResource(testsConfig.ipcheck_site, proxyConfig)
	if err != nil {
		log.Println("Error getting resource: ", err)
	}
	log.Println("Tunneled IP: ", string(tunneledCheck))

	// NewSession test for tunneled handhsake and connected requests.
	_, err = psiphon.NewSession(config, tunnel)
	if err != nil {
		log.Println("Error getting new session: ", err)
	}
	return result, err

}

// Makes a Get request to a static site resource.
func getSiteResource(site string, proxyConfig *ProxyConfig) (body []byte, err error) {
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
	return body, nil
}
