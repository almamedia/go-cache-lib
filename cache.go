package gocachelib

import (
	"log"
	"time"

	cmap "github.com/streamrail/concurrent-map"
)

var cache cmap.ConcurrentMap

// how many simultaneus workers should we have, default 20
var workerAmount = 20

// maximum amount of jobs buffered, default 200
var bufferedJobs = 200

// default cache size
var cacheSize = 20

// default ttl
var ttl = 1 * time.Hour

var jobs chan CacheItem

// Start background loading cache with specified parameters
func StartWith(workers, bufferSize, cacheSizeAmount int, defaultTTL time.Duration) {
	workerAmount = workers
	bufferedJobs = bufferSize
	cacheSize = cacheSizeAmount
	ttl = defaultTTL
	Start()
}

// Start background loading cache with default parameters
func Start() {
	cache = cmap.New()
	jobs = make(chan CacheItem, bufferedJobs)
	// workers
	for w := 1; w <= workerAmount; w++ {
		go worker(w, jobs)
	}
	go doEvery(1*time.Second, refreshAndRevoke)
}

// check and update expiring items, revoke those exceeding their TTL
func refreshAndRevoke() {
	now := time.Now()
	for _, value := range cache.Items() {
		item := value.(CacheItem)
		if now.After(item.RevokeTime) && !item.Updating {
			log.Printf("Revoking item that has not been used in %v: %v", item.TTL, item.Key)
			cache.Remove(item.Key)
		} else if now.After(item.ExpireTime.Add(-300*time.Millisecond)) && !item.Updating {
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
				Key:        item.Key,
				Value:      value,
				TTL: 		item.TTL,
				RevokeTime: item.RevokeTime,
				ExpireTime: time.Now().Add(item.Expiration),
				Expiration: item.Expiration,
				GetFunc:    item.GetFunc,
				Updating:   false,
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
		item := value.(CacheItem)
		item.updateRevokeTime()
		cache.Set(item.Key, item)
		return value.(CacheItem).Value
	}
	return nil
}

// AddItem sets the item to cache
func AddItem(item CacheItem) {
	item.updateRevokeTime()
	if cache.Count() >= cacheSize {
		log.Print("Cache full")
		releaseLeastViable()
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
	Key        string
	Value      []byte
	ExpireTime time.Time
	Expiration time.Duration
	TTL        time.Duration
	RevokeTime time.Time
	GetFunc    func(key string) []byte
	Updating   bool
}

func (i *CacheItem) updateRevokeTime() {
	now := time.Now()

	if i.TTL != 0 {
		i.RevokeTime = now.Add(max(i.TTL, i.Expiration))
	} else {
		i.RevokeTime = now.Add(max(ttl, i.Expiration))
	}
}

func releaseLeastViable() {
	var earliest CacheItem
	for _, v := range cache.Items() {
		if (v.(CacheItem).RevokeTime.Before(earliest.RevokeTime) || earliest.RevokeTime == time.Time{}) {
			earliest = v.(CacheItem)
		}
	}
	log.Printf("Removing cache item %s with earliest revoke time to make room", earliest.Key)
	cache.Remove(earliest.Key)
}