package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func GetWebResource(site string) (body []byte, err error) {
	client := &http.Client{}

	req, err := http.NewRequest("Get", site, nil)
	if err != nil {
		log.Fatalf("Could not get resource: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read response body: %s", err)
	}

	fmt.Println("Body: %s", string(body))
	return body, nil
}
