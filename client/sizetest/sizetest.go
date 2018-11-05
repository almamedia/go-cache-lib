package main

import (
	"log"
	"time"

	hc "github.com/almamedia/go-cache-lib"
	"github.com/almamedia/go-cache-lib/client"
)

func main() {
	hc.Start(1, 10, 5)
	for _, u := range urls {
		item := hc.CacheItem{
			Key:          u,
			Value:        client.DataFetch(u),
			Expire:       time.Now().Add(time.Duration(4 * time.Second)),
			UpdateLength: time.Duration(4 * time.Second),
			GetFunc:      client.DataFetch,
		}
		hc.AddItem(item)
	}
	go runURLs()
	time.Sleep(5 * time.Minute)
}

func runURLs() {
	for i := 0; i < 1000; i++ {
		for _, u := range urls {
			go fetch(u)
			time.Sleep(300 * time.Millisecond)
		}
	}
}

func fetch(url string) {
	item := hc.GetItem(url)
	if item != nil {
		log.Println("Got item from cache", url)
		return
	}

	cacheItem := hc.CacheItem{
		Key:          url,
		Value:        client.DataFetch(url),
		Expire:       time.Now().Add(time.Duration(4 * time.Second)),
		UpdateLength: time.Duration(4 * time.Second),
		GetFunc:      client.DataFetch,
	}
	hc.AddItem(cacheItem)
	log.Println(url, "not found in cache, added")
}

var urls = []string{
	"https://m.kauppalehti.fi/",
	"https://www.iltalehti.fi/",
	"https://www.aamulehti.fi/",
	"https://www.hs.fi/",
	"https://www.talouselama.fi/",
	"https://www.arvopaperi.fi/",
	"https://www.tivi.fi/",
	"https://www.mikrobitti.fi/",
	"https://www.tekniikkatalous.fi/",
	"https://www.kauppalehti.fi/",
	"https://www.marmai.fi/",
	"https://www.dagensmedia.se/",
	"https://www.affarsvarlden.se/",
	"https://www.wsj.com",
	"https://www.nyteknik.se/",
	"https://www.theguardian.com",
	"https://www.mtvuutiset.fi/",
	"https://www.reuters.com/",
	"https://www.cnbc.com/economy/",
	"https://www.bbc.com/news/business/economy",
	"https://www.wsj.com/news/economy",
	"https://www.theguardian.com/business/economics",
	"https://www.marketwatch.com/",
	"https://money.cnn.com/",
}
