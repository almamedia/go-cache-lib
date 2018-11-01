package gocachelib

import (
	"time"

	"github.com/streamrail/concurrent-map"
)

var cache cmap.ConcurrentMap

// how many simultaneus workers should we have, default 20
var workerAmount = 20

// maximum amount of jobs buffered, default 200
var bufferedJobs = 200
var jobs chan CacheItem

// Start for setting worker amount
func Start(workers int, bufferSize int) {
	workerAmount = workers
	bufferedJobs = bufferSize
}

func init() {
	cache = cmap.New()
	jobs = make(chan CacheItem, bufferedJobs)
	// workers
	for w := 1; w <= workerAmount; w++ {
		go worker(w, jobs)
	}
	go doEvery(time.Second, checkExpiredItems)
}

// check and update near expiring items
func checkExpiredItems() {
	for _, value := range cache.Items() {
		item := value.(CacheItem)
		if time.Now().After(item.Expire.Add(-1 * time.Second)) {
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
			}
			cache.Set(item.Key, d)
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
}
