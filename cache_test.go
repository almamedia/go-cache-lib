package gocachelib

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/almamedia/go-cache-lib/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestExpiration(t *testing.T) {
	StartWith(1, 1, 1, 1*time.Second)
	key := "xyzzy"
	i := CacheItem{
		Key:        key,
		Value:      []byte("foo"),
		ExpireTime: time.Now(),
		Expiration: 1 * time.Second,
		GetFunc:    noopGetFunc,
	}
	AddItem(i)
	assert.True(t, string(GetItem(key)) == "foo")
	time.Sleep(1500 * time.Millisecond)
	assert.True(t, GetItem(key) == nil, "Item should have been revoked by now")
}

func TestTTLLessThanExpiration(t *testing.T) {
	StartWith(1, 1, 1, 1*time.Second)
	key := "xyzzy"
	i := CacheItem{
		Key:        key,
		Value:      []byte("foo"),
		ExpireTime: time.Now(),
		TTL:        1 * time.Second,
		Expiration: 2 * time.Second,
		GetFunc:    noopGetFunc,
	}
	AddItem(i)
	time.Sleep(1500 * time.Millisecond)
	assert.True(t, string(GetItem(key)) == "foo", "Item should still be in cache")
}
func TestGetItemPostponesRevoke(t *testing.T) {
	StartWith(1, 1, 1, 1*time.Second)
	key := "xyzzy"
	i := CacheItem{
		Key:        key,
		Value:      []byte("foo"),
		ExpireTime: time.Now(),
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
	revokeAfterAdd := item.(CacheItem).RevokeTime
	assert.True(t, now.Before(revokeAfterAdd))
	time.Sleep(50 * time.Millisecond)
	assert.True(t, string(GetItem(key)) == "foo", "Item should be in cache")
	item, ok = cache.Get(key)
	if !ok {
		t.Errorf("Should have got item %s from cache", key)
	}
	revokeAfterGet := item.(CacheItem).RevokeTime
	assert.True(t, revokeAfterAdd.Before(revokeAfterGet), "Revoke time should be postponed after GetItem: %v < %v", revokeAfterAdd, revokeAfterGet)
}

func TestExpire(t *testing.T) {
	StartWith(1, 11, 1, 2*time.Second)
	url := "https://httpbin.org/ip"
	value := randomGetFunc("")
	i := CacheItem{
		Key:        url,
		Value:      value,
		ExpireTime: time.Now(),
		Expiration: 1 * time.Second,
		GetFunc:    randomGetFunc,
	}
	AddItem(i)
	time.Sleep(2 * time.Second)
	assert.NotEqual(t, string(value), string(GetItem(url)), "Item should have new value in cache")
}


func TestSetItem(t *testing.T) {
	StartWith(1, 11, 1, 2*time.Second)
	url := "https://httpbin.org/ip"
	i := CacheItem{
		Key:        url,
		Value:      nil,
		ExpireTime: time.Now(),
		Expiration: 1 * time.Second,
		GetFunc:    client.DataFetch,
	}
	AddItem(i)
	time.Sleep(2 * time.Second)
	assert.True(t, strings.Contains(string(GetItem(url)), "origin"))
	cache.Remove(url)
	assert.False(t, strings.Contains(string(GetItem(url)), "origin"))
	assert.True(t, string(GetItem(url)) == "")
	assert.Equal(t, 1, workerAmount)
	assert.Equal(t, 11, bufferedJobs)
}

func noopGetFunc(s string) []byte {
	return nil
}

func randomGetFunc(s string) []byte {
	return []byte(uuid.New().String())
}