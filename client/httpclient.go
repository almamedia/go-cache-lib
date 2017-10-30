package client

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

var client *http.Client

func init() {
	client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 64,
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 10 * time.Second,
			}).Dial,
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}
}

func httpGet(url string) ([]byte, error) {
	t := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Println("GET", resp.StatusCode, url, time.Now().Sub(t), time.Now().Format("15:04:05.000"))
	return body, nil
}

func DataFetch(url string) []byte {
	data, err := httpGet(url)
	if err != nil {
		log.Println("getting", url, "failed", err.Error())
	}
	return data
}
