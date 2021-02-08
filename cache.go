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

var jobs chan timedCacheItem

// StartWith background loading cache with specified parameters
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
	jobs = make(chan timedCacheItem, bufferedJobs)
	// workers
	for w := 1; w <= workerAmount; w++ {
		go worker(w, jobs)
	}
	go doEvery(1*time.Second, refresh)
	go doEvery(1*time.Second, revoke)
}

// check and update expiring items
func refresh() {
	now := time.Now()
	for _, value := range cache.Items() {
		item := value.(timedCacheItem)
		if now.After(item.ExpireTime.Add(-300*time.Millisecond)) && !item.Updating {
			item.Updating = true
			cache.Set(item.Key, item)
			jobs <- item
		}
	}
}

// revoke those exceeding their TTL
func revoke() {
	now := time.Now()
	for _, value := range cache.Items() {
		item := value.(timedCacheItem)
		if now.After(item.RevokeTime) && !item.Updating {
			log.Printf("Revoking item that has not been used in %v: %v", item.TTL, item.Key)
			cache.Remove(item.Key)
		}
	}
}

// listen to jobs channel and handle incoming items
func worker(id int, jobs <-chan timedCacheItem) {
	for item := range jobs {
		value := item.GetFunc(item.Key)
		if value != nil {
			item.Value = value
			item.UpdateExpireTime()
		}
		item.Updating = false
		cache.Set(item.Key, item)
	}
}

// GetValue value from cache
func GetValue(key string) []byte {
	value, ok := cache.Get(key)
	if ok {
		item := value.(timedCacheItem)
		item.UpdateRevokeTime()
		cache.Set(item.Key, item)
		return value.(timedCacheItem).Value
	}
	return nil
}

// AddItem sets the item to cache and updates its revoke and expire times
func AddItem(item CacheItem) {
	i := timedCacheItem{CacheItem: item}
	i.UpdateRevokeTime()
	i.UpdateExpireTime()
	if cache.Count() >= cacheSize {
		log.Print("Cache full")
		revokeLeastViable()
	}
	cache.Set(i.Key, i)
}

// CacheItem for cached items
// Key cache key, for example url
// Value to be cached
// Expiration Time to expire item. Item is refreshed using GetFunc after it expires
// TTL Time to revocation from cache after last access
// GetFunc function for updating the value
type CacheItem struct {
	Key        string
	Value      []byte
	Expiration time.Duration
	TTL        time.Duration
	GetFunc    func(key string) []byte
}

type timedCacheItem struct {
	CacheItem
	RevokeTime time.Time
	ExpireTime time.Time
	Updating   bool
}

func (i *timedCacheItem) UpdateRevokeTime() {
	now := time.Now()
	if i.TTL == 0 {
		i.TTL = ttl
	}
	i.RevokeTime = now.Add(max(i.TTL, i.Expiration))
}

func (i *timedCacheItem) UpdateExpireTime() {
	now := time.Now()
	i.ExpireTime = now.Add(i.Expiration)
}

func revokeLeastViable() {
	var earliest timedCacheItem
	for _, v := range cache.Items() {
		if (v.(timedCacheItem).RevokeTime.Before(earliest.RevokeTime) || earliest.RevokeTime == time.Time{}) {
			earliest = v.(timedCacheItem)
		}
	}
	log.Printf("Removing cache item %s with earliest revoke time to make room", earliest.Key)
	cache.Remove(earliest.Key)
}
