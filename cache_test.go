package gocachelib

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRevoke(t *testing.T) {
	StartWith(1, 1, 1, 1*time.Second)
	key := "xyzzy"
	i := CacheItem{
		Key:        key,
		Value:      []byte("foo"),
		Expiration: 1 * time.Second,
		GetFunc:    noopGetFunc,
	}
	AddItem(i)
	assert.True(t, string(GetValue(key)) == "foo")
	time.Sleep(2000 * time.Millisecond)
	assert.True(t, nil == GetValue(key), "Item should have been revoked by now")
}

func TestTTLLessThanExpiration(t *testing.T) {
	StartWith(1, 1, 1, 1*time.Second)
	key := "xyzzy"
	i := CacheItem{
		Key:        key,
		Value:      []byte("foo"),
		TTL:        1 * time.Second,
		Expiration: 2 * time.Second,
		GetFunc:    noopGetFunc,
	}
	AddItem(i)
	time.Sleep(1500 * time.Millisecond)
	assert.True(t, string(GetValue(key)) == "foo", "Item should still be in cache")
}

func TestGetValuePostponesRevoke(t *testing.T) {
	StartWith(1, 1, 1, 1*time.Second)
	key := "xyzzy"
	i := CacheItem{
		Key:        key,
		Value:      []byte("foo"),
		TTL:        1 * time.Second,
		Expiration: 2 * time.Second,
		GetFunc:    noopGetFunc,
	}
	now := time.Now()
	AddItem(i)
	item, ok := cache.Get(key)
	if !ok {
		t.Errorf("Should have got item %s from cache", key)
	}
	revokeAfterAdd := item.(timedCacheItem).RevokeTime
	assert.True(t, now.Before(revokeAfterAdd))
	time.Sleep(50 * time.Millisecond)
	assert.True(t, string(GetValue(key)) == "foo", "Item should be in cache")
	item, ok = cache.Get(key)
	if !ok {
		t.Errorf("Should have got item %s from cache", key)
	}
	revokeAfterGet := item.(timedCacheItem).RevokeTime
	assert.True(t, revokeAfterAdd.Before(revokeAfterGet), "Revoke time should be postponed after GetValue: %v < %v", revokeAfterAdd, revokeAfterGet)
}

func TestExpire(t *testing.T) {
	StartWith(1, 11, 1, 2*time.Second)
	url := "https://httpbin.org/ip"
	value := randomGetFunc("")
	i := CacheItem{
		Key:        url,
		Value:      value,
		Expiration: 1 * time.Second,
		GetFunc:    randomGetFunc,
	}
	AddItem(i)
	time.Sleep(2 * time.Second)
	assert.NotEqual(t, string(value), string(GetValue(url)), "Item should have new value in cache")
}

func TestItemWithShortestTTLIsRevokedWhenCacheFillsUp(t *testing.T) {
	StartWith(1, 11, 2, 1*time.Second)
	AddItem(CacheItem{
		Key:        "1",
		Value:      []byte("1"),
		Expiration: 1 * time.Second,
		TTL:        2 * time.Second,
		GetFunc:    noopGetFunc,
	})
	AddItem(CacheItem{
		Key:        "2",
		Value:      []byte("2"),
		Expiration: 1 * time.Second,
		TTL:        1 * time.Second,
		GetFunc:    noopGetFunc,
	})
	AddItem(CacheItem{
		Key:        "3",
		Value:      []byte("3"),
		Expiration: 1 * time.Second,
		TTL:        2 * time.Second,
		GetFunc:    noopGetFunc,
	})

	assert.Equal(t, 2, cache.Count(), "Cache should have 2 items")
	assert.Equal(t, "1", string(GetValue("1")))
	assert.Equal(t, "3", string(GetValue("3")))
	for k, _ := range cache.Items() {
		assert.NotEqual(t, "2", k, "Item #2 should have been revoked when cache was full")
	}
}

func noopGetFunc(s string) []byte {
	return nil
}

func randomGetFunc(s string) []byte {
	return []byte(uuid.New().String())
}
