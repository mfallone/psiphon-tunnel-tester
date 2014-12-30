package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	psiphon "github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon"
)

func sendGETRequest(site string, client *http.Client) (*http.Response, error) {
	if client == nil {
		client = &http.Client{}
	}

	req, err := http.NewRequest("GET", site, nil)
	if err != nil {
		log.Fatalf("Could not make new GET request: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Could not complete GET request: %s")
	}

	return resp, err
}

func getExternalIPAddress(site string, client *http.Client) (ip net.IP, err error) {

	resp, err := sendGETRequest(site, client)
	if err != nil {
		log.Fatalf("Could not send request: %s", err)
	}
	defer resp.Body.Close()

	// Print out a notification if a 200 isn't received
	if resp.StatusCode != 200 {
		log.Printf("Status Code: %s", resp.StatusCode)
	}

	//
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read response body: %s", err)
	}

	ip = net.ParseIP(strings.TrimSpace(string(body)))
	if ip == nil {
		log.Println("Could not parse IP")
	}

	return
}

// downloadFile will do exactly that.  It will take a website URL and
// download it to a local file in the current directory.
// The intent of this test is to check the timing of a tunneled vs untunneled
// file download
// TODO the `done` chan should be replaced with one that reports the time duration.  Possibly []bytes
func downloadFile(site string, outpath string, client *http.Client, done chan string) {
	// a channel will be needed to signal when complete

	startTime := time.Now()

	resp, err := sendGETRequest(site, client)
	if err != nil {
		log.Printf("Error sending request: %s", err)
	}
	defer resp.Body.Close()

	if outpath == "" {
		outpath = "./tmp/LRGOutFile.bin"
	}
	outfile, err := os.Create(outpath)
	if err != nil {
		log.Printf("Error creating file: %s", err)
	}
	io.Copy(outfile, resp.Body)
	duration := time.Now().Sub(startTime).String()
	done <- duration
}

// SetupTasks is called by the main function.  It prepares and runs the tasks
// TODO have tasks run concurrently.
func SetupTasks(config *psiphon.Config, decodedServerEntry *psiphon.ServerEntry, tasksConfig TasksConfig) {
	log.Println("Setting up Tasks")

	untunneled := new(TasksResults)
	untunneled.Label = "UNTUNNELED"
	untunneled.done = make(chan bool, 1)

	httpTunneled := new(TasksResults)
	httpTunneled.Label = "HTTPTUNNEL"
	httpTunneled.done = make(chan bool, 1)

	proxyConfig := setProxyConfig("127.0.0.1", 8080, false)
	untunneled.useProxy = false
	fmt.Println(proxyConfig.httpProxyAddress)

	// start Psiphon session
	log.Print("Starting Psiphon Session...")
	pendingConns := new(psiphon.Conns)
	tunnel, err := psiphon.EstablishTunnel(config, pendingConns, decodedServerEntry)
	log.Println("Psiphon Tunnel Connected")
	// Setup new HTTP proxy. Close() is handled by HttpProxy.serve()
	// and does not need to be called here.
	log.Println("Setting HTTP Proxy")
	_, err = psiphon.NewHttpProxy(config, tunnel)
	if err != nil {
		log.Fatalf("error initializing local HTTP proxy: %s", err)
	}
	httpTunneled.useProxy = true

	log.Println("Running tests")
	go untunneled.Run(tasksConfig, proxyConfig)
	go httpTunneled.Run(tasksConfig, proxyConfig)

	<-untunneled.done
	<-httpTunneled.done

	log.Println("Tests Completed")

	log.Printf("Untunneled IP: %s", untunneled.externalIP)
	log.Printf("Tunneled IP: %s", httpTunneled.externalIP)

}

func (tasks *TasksResults) Run(tasksConfig TasksConfig, proxyConfig ProxyConfig) {
	// Set client connection to be proxied or not.  Run Tasks.

	// check tasks.useProxy to determine if a proxy should be set.
	var client *http.Client

	if tasks.useProxy == true { // Set a http proxy.  Some smarter port selection would be good.
		proxyUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", proxyConfig.httpProxyAddress, proxyConfig.httpProxyPort))
		if err != nil {
			log.Fatalf("Could not parse url: ", err)
		}
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	} else {
		client = &http.Client{}
	}

	// Get a web resource we know serves up our external IP address as a page
	log.Printf("%s: Checking external IP address..", tasks.Label)
	external_ip, err := getExternalIPAddress(tasksConfig.ExternalIPCheckSite, client)
	if err != nil {
		log.Fatalf("%s: Error getting Exterinal IP: %s", tasks.Label, err)
	}
	tasks.externalIP = external_ip

	// Start file download test
	tasks.largeDownloadFileTime = make(chan string, 1)
	outfile := "./tmp/" + tasks.Label + ".bin"

	go downloadFile(tasksConfig.LRGDownloadFile, outfile, client, tasks.largeDownloadFileTime)

	log.Printf("%s: Large file download started", tasks.Label)
	log.Printf("%s: Large file download complete: %s", tasks.Label, <-tasks.largeDownloadFileTime)

	tasks.done <- true
}
