package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	psiphon "github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon"
)

// RunTests runs all tests to the server conatined in decodedServerEntry
func RunTests(config *psiphon.Config, decodedServerEntry *psiphon.ServerEntry, tasksConfig *TasksConfig) (result string, err error) {
	runStartTime := time.Now()

	pendingConns := new(psiphon.Conns)

	proxyConfig := &ProxyConfig{httpProxyAddress: "127.0.0.1",
		httpProxyPort: config.LocalHttpProxyPort,
		useHttpProxy:  false}

	// Get the untunneled IP address
	siteResponse, err := getSiteResource(tasksConfig.ExternalIPCheckSite, proxyConfig)
	if err != nil {
		log.Println("Could not get site resource: ", err)
	}

	untunneledCheck, err := readResponseBody(siteResponse)
	if err != nil {
		log.Println("Could not parse body")
	}
	siteResponse.Body.Close()
	log.Println("Untunneled IP: ", string(untunneledCheck))

	/* Download Files */
	startTime := time.Now()
	siteResponse, err = getSiteResource(tasksConfig.LRGDownloadFile, proxyConfig)
	outfile, err := os.Create("LRGOutFile.bin")
	io.Copy(outfile, siteResponse.Body)
	siteResponse.Body.Close()
	endTime := time.Now()
	untunneledDuration := endTime.Sub(startTime)

	/* END TEST */

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
	siteResponse, err = getSiteResource(tasksConfig.ExternalIPCheckSite, proxyConfig)
	if err != nil {
		log.Println("Error getting resource: ", err)
	}

	tunneledCheck, err := readResponseBody(siteResponse)
	if err != nil {
		log.Println("Could not read respone body")
	}
	siteResponse.Body.Close()

	log.Println("Tunneled IP: ", string(tunneledCheck))

	// NewSession test for tunneled handhsake and connected requests.
	_, err = psiphon.NewSession(config, tunnel)
	if err != nil {
		log.Println("Error getting new session: ", err)
	}

	// Download 100MB file
	startTime = time.Now()
	siteResponse, err = getSiteResource(tasksConfig.LRGDownloadFile, proxyConfig)
	outfile, err = os.Create("LRGOutFile.bin")
	io.Copy(outfile, siteResponse.Body)
	siteResponse.Body.Close()
	endTime = time.Now()
	tunneledDuration := endTime.Sub(startTime)
	log.Printf("100MB File Download\nUntunneled: %s\nTunneled: %s", untunneledDuration, tunneledDuration)

	runEndTime := time.Now()
	log.Printf("Run Duration: %v", runEndTime.Sub(runStartTime))
	return result, err

}

// Makes a Get request to a static site resource.
func getSiteResource(site string, proxyConfig *ProxyConfig) (*http.Response, error) {
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

	req, err := http.NewRequest("GET", site, nil)
	if err != nil {
		log.Fatalf("Could not get requested resource: ", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Could not complete request: ", err)
	}

	return resp, nil
}

func readResponseBody(resp *http.Response) ([]byte, error) {

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read response body: %s", err)
	}

	return body, nil
}
