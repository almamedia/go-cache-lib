package gocachelib

import (
	"log"
	"time"

	"github.com/streamrail/concurrent-map"
)

var cache cmap.ConcurrentMap

// how many simultaneus workers should we have, default 20
var workerAmount = 20

// maximum amount of jobs buffered, default 200
var bufferedJobs = 200

// default cache size
var cacheSize = 20
var jobs chan CacheItem

// Start for setting worker amount
func Start(workers, bufferSize, cacheSizeAmount int) {
	workerAmount = workers
	bufferedJobs = bufferSize
	cacheSize = cacheSizeAmount
}

func init() {
	cache = cmap.New()
	jobs = make(chan CacheItem, bufferedJobs)
	// workers
	for w := 1; w <= workerAmount; w++ {
		go worker(w, jobs)
	}
	go doEvery(1*time.Second, checkExpiredItems)
}

// check and update near expiring items
func checkExpiredItems() {
	for _, value := range cache.Items() {
		item := value.(CacheItem)
		if time.Now().After(item.Expire.Add(-300*time.Millisecond)) && !item.Updating {
			item.Updating = true
			cache.Set(item.Key, item)
			jobs <- item
		}
	}
}

// listen to jobs channel and handle incoming items
func worker(id int, jobs <-chan CacheItem) {
	for item := range jobs {
		value := item.GetFunc(item.Key)
		if value != nil {
			d := CacheItem{
				Key:          item.Key,
				Value:        value,
				Expire:       time.Now().Add(item.UpdateLength),
				UpdateLength: item.UpdateLength,
				GetFunc:      item.GetFunc,
				Updating:     false,
			}
			AddItem(d)
		} else {
			item.Updating = false
			cache.Set(item.Key, item)
		}
	}
}

// GetItem value from cache
func GetItem(key string) []byte {
	value, ok := cache.Get(key)
	if ok {
		return value.(CacheItem).Value
	}
	return nil
}

// AddItem sets the item to cache
func AddItem(item CacheItem) {
	if cache.Count() >= cacheSize {
		log.Println("CACHE SIZE:", cache.Count(), "max size:", cacheSize)
		removingKey := cache.Keys()[0]
		log.Println("Cache full, removing key", removingKey)
		cache.Remove(removingKey)
	}
	cache.Set(item.Key, item)
}

// CacheItem for cached items
// Key cache key, for example url
// Value to be cached
// Expire time to expire item
// UpdateLength duration for next expiration
// Get function for updating the value
type CacheItem struct {
	Key          string
	Value        []byte
	Expire       time.Time
	UpdateLength time.Duration
	GetFunc      func(key string) []byte
	Updating     bool
}
