package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"math/rand"

	_ "net/http/pprof"

	hc "github.com/almamedia/go-cache-lib"
	"github.com/almamedia/go-cache-lib/client"
	"github.com/julienschmidt/httprouter"
)

func main() {
	hc.Start(3, 10, 22)
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	for _, u := range urls {
		item := hc.CacheItem{
			Key:          u,
			Value:        client.DataFetch(u),
			Expire:       time.Now().Add(time.Duration(r1.Intn(20)+5) * time.Second),
			UpdateLength: time.Duration(r1.Intn(20)+5) * time.Second,
			GetFunc:      client.DataFetch,
		}
		hc.AddItem(item)
	}
	router := httprouter.New()
	router.RedirectFixedPath = true
	router.RedirectTrailingSlash = true
	for _, u := range urls {
		router.GET("/"+strings.Replace(strings.Replace(u, "http://", "", -1), "https://", "", -1), func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			item := hc.GetItem("http:/" + r.URL.Path)
			if item != nil {
				w.Header().Add("Content-Type", "text/html")
				w.WriteHeader(200)
				w.Write(item)
			} else {
				w.WriteHeader(404)
				w.Write([]byte("Not found"))
			}
		})
	}
	go runURLs()
	http.ListenAndServe(":8800", router)
}

func runURLs() {
	for _, u := range urls {
		go fetch("http://localhost:8800/"+strings.Replace(strings.Replace(u, "http://", "", -1), "https://", "", -1), "title")
	}
}

func fetch(url, match string) {
	t := time.Now()
	const howMany = 5
	for i := 0; i < howMany; i++ {
		item := client.DataFetch(url)
		if item == nil || !strings.Contains(string(item), match) {
			log.Fatalln("ERROR page not found", string(item), item == nil, match)
		}
		time.Sleep(300 * time.Millisecond)
	}
	log.Println("fetched", howMany, "times from cache successfully in", time.Now().Sub(t))
}

var urls = []string{
	"http://m.kauppalehti.fi/",
	"http://www.iltalehti.fi/",
	"http://www.aamulehti.fi/",
	"http://www.hs.fi/",
	"http://www.talouselama.fi/",
	"http://www.arvopaperi.fi/",
	"http://www.tivi.fi/",
	"http://www.mikrobitti.fi/",
	"http://www.tekniikkatalous.fi/",
	"http://www.kauppalehti.fi/",
	"http://www.marmai.fi/",
	"http://www.dagensmedia.se/",
	"http://www.affarsvarlden.se/",
	"http://www.nyteknik.se/",
	"http://www.mtvuutiset.fi/",
	"http://www.reuters.com/",
	"http://www.cnbc.com/economy/",
	"http://www.bbc.com/news/business/economy",
	"http://www.wsj.com/news/economy",
	"http://www.theguardian.com/business/economics",
	"http://www.marketwatch.com/",
	"http://money.cnn.com/",
}
