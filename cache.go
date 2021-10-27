package gocachelib

import (
	"log"
	"sync"
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

// revoke & refresh loop interval
var loopInterval = 1 * time.Second

var jobs chan timedCacheItem

var refreshTicker *time.Ticker
var revokeTicker *time.Ticker

var loopMutex = sync.Mutex{}

var workerWg = sync.WaitGroup{}

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
	log.Printf("Starting in-memory cache with %d workers, %d job queue size, %d cache maximum and %d default TTL", workerAmount, bufferedJobs, cacheSize, ttl)
	cache = cmap.New()
	jobs = make(chan timedCacheItem, bufferedJobs)
	// workers
	for w := 1; w <= workerAmount; w++ {
		go worker(w, jobs)
	}
	refreshTicker = doEvery(loopInterval, refresh)
	revokeTicker = doEvery(loopInterval, revoke)
}

// Stop background tickers
func Stop() {
	log.Printf("Stop in-memory cache background processing")
	refreshTicker.Stop()
	revokeTicker.Stop()
	loopMutex.Lock()
	defer loopMutex.Unlock()
	close(jobs)
	workerWg.Wait()
	cache = cmap.New()
}

// check and update expiring items
func refresh() {
	loopMutex.Lock()
	defer loopMutex.Unlock()
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
	loopMutex.Lock()
	defer loopMutex.Unlock()
	now := time.Now()
	for _, value := range cache.Items() {
		item := value.(timedCacheItem)
		if now.After(item.RevokeTime) {
			log.Printf("Revoking item that has not been used in %v: %v", item.TTL, item.Key)
			cache.Remove(item.Key)
		}
	}
}

// listen to jobs channel and handle incoming items
func worker(id int, jobs <-chan timedCacheItem) {
	workerWg.Add(1)
	defer workerWg.Done()
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
		item.Updating = false
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
	loopMutex.Lock()
	defer loopMutex.Unlock()
	var earliest timedCacheItem
	for _, v := range cache.Items() {
		if (v.(timedCacheItem).RevokeTime.Before(earliest.RevokeTime) || earliest.RevokeTime == time.Time{}) {
			earliest = v.(timedCacheItem)
		}
	}
	log.Printf("Removing cache item %s with earliest revoke time to make room", earliest.Key)
	cache.Remove(earliest.Key)
}
