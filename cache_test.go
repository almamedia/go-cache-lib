package gocachelib

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRevoke(t *testing.T) {
	// Run refresh & revoke loops quicker than usual
	defaultLoopInterval := loopInterval
	defer func() {
		stop()
		loopInterval = defaultLoopInterval
	}()
	loopInterval = 10 * time.Millisecond
	StartWith(1, 1, 1, 15*time.Millisecond)
	key := "TestRevoke"
	i := CacheItem{
		Key:        key,
		Value:      []byte("TestRevoke"),
		Expiration: 10 * time.Millisecond,
		GetFunc:    noopGetFunc,
	}
	AddItem(i)
	assert.True(t, string(GetValue(key)) == "TestRevoke")
	time.Sleep(25 * time.Millisecond)
	assert.True(t, nil == GetValue(key), "Item should have been revoked by now")
}

func TestTTLLessThanExpiration(t *testing.T) {
	// Run refresh & revoke loops quicker than usual
	defaultLoopInterval := loopInterval
	defer func() {
		stop()
		loopInterval = defaultLoopInterval
	}()
	loopInterval = 10 * time.Millisecond
	StartWith(1, 1, 1, 10*time.Millisecond)
	key := "TestTTLLessThanExpiration"
	i := CacheItem{
		Key:        key,
		Value:      []byte("TestTTLLessThanExpiration"),
		TTL:        10 * time.Millisecond,
		Expiration: 20 * time.Millisecond,
		GetFunc:    noopGetFunc,
	}
	AddItem(i)
	time.Sleep(15 * time.Millisecond)
	assert.True(t, string(GetValue(key)) == "TestTTLLessThanExpiration", "Item should still be in cache after the first revocation loop was run")
}

func TestGetValuePostponesRevoke(t *testing.T) {
	StartWith(1, 1, 1, 1*time.Second)
	defer stop()
	key := "TestGetValuePostponesRevoke"
	i := CacheItem{
		Key:        key,
		Value:      []byte("TestGetValuePostponesRevoke"),
		TTL:        10 * time.Millisecond,
		Expiration: 10 * time.Millisecond,
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
	assert.True(t, string(GetValue(key)) == "TestGetValuePostponesRevoke", "Item should be in cache")
	time.Sleep(5 * time.Millisecond)
	item, ok = cache.Get(key)
	if !ok {
		t.Errorf("Should have got item %s from cache after first GetValue", key)
	}
	revokeAfterGet := item.(timedCacheItem).RevokeTime
	assert.True(t, revokeAfterAdd.Before(revokeAfterGet), "Revoke time should be postponed after GetValue: %v < %v", revokeAfterAdd, revokeAfterGet)
}

func TestExpire(t *testing.T) {
	// Run refresh & revoke loops quicker than usual
	defaultLoopInterval := loopInterval
	defer func() {
		stop()
		loopInterval = defaultLoopInterval
	}()
	loopInterval = 10 * time.Millisecond
	StartWith(1, 11, 1, 1*time.Second)
	key := "TestExpire"
	value := randomGetFunc("")
	i := CacheItem{
		Key:        key,
		Value:      value,
		Expiration: 10 * time.Millisecond,
		GetFunc:    randomGetFunc,
	}
	AddItem(i)
	time.Sleep(20 * time.Millisecond)
	assert.NotEqual(t, string(value), string(GetValue(key)), "Item should have new value in cache")
}

func TestItemWithShortestTTLIsRevokedWhenCacheFillsUp(t *testing.T) {
	StartWith(1, 11, 2, 1*time.Second)
	defer stop()
	AddItem(CacheItem{
		Key:        "TestItemWithShortestTTLIsRevokedWhenCacheFillsUp1",
		Value:      []byte("TestItemWithShortestTTLIsRevokedWhenCacheFillsUp1"),
		Expiration: 1 * time.Second,
		TTL:        2 * time.Second,
		GetFunc:    noopGetFunc,
	})
	AddItem(CacheItem{
		Key:        "TestItemWithShortestTTLIsRevokedWhenCacheFillsUp2",
		Value:      []byte("TestItemWithShortestTTLIsRevokedWhenCacheFillsUp2"),
		Expiration: 1 * time.Second,
		TTL:        1 * time.Second,
		GetFunc:    noopGetFunc,
	})
	AddItem(CacheItem{
		Key:        "TestItemWithShortestTTLIsRevokedWhenCacheFillsUp3",
		Value:      []byte("TestItemWithShortestTTLIsRevokedWhenCacheFillsUp3"),
		Expiration: 1 * time.Second,
		TTL:        2 * time.Second,
		GetFunc:    noopGetFunc,
	})

	assert.Equal(t, "TestItemWithShortestTTLIsRevokedWhenCacheFillsUp1", string(GetValue("TestItemWithShortestTTLIsRevokedWhenCacheFillsUp1")))
	assert.Equal(t, "TestItemWithShortestTTLIsRevokedWhenCacheFillsUp3", string(GetValue("TestItemWithShortestTTLIsRevokedWhenCacheFillsUp3")))
	for k, _ := range cache.Items() {
		assert.NotEqual(t, "TestItemWithShortestTTLIsRevokedWhenCacheFillsUp2", k, "Item #2 should have been revoked when cache was full")
	}
}

func TestConcurrentRefreshAndGetValueBug(t *testing.T) {
	// Run refresh & revoke loops quicker than usual
	defaultLoopInterval := loopInterval
	defer func() {
		stop()
		loopInterval = defaultLoopInterval
	}()
	loopInterval = 10 * time.Millisecond
	// Make sure revoke does not interfere here
	StartWith(1, 11, 1, 5*time.Second)
	// continuously spamming GetValue should manifest the bug
	c := make(chan []byte, 1)
	key := "TestConcurrentRefreshAndGetValueBug"
	go busyGet(t, c, key)
	value := randomGetFunc("")
	i := CacheItem{
		Key:        key,
		Value:      value,
		Expiration: 10 * time.Millisecond,
		GetFunc:    randomGetFunc,
	}
	AddItem(i)
	// sleep long enough for the bug to kick in repeatably
	time.Sleep(1013 * time.Millisecond)
	c <- []byte("stop")
	v, _ := cache.Get(key)
	ci := v.(timedCacheItem)
	assert.NotEqual(t, ci.Updating, true, "Item Should not be in updating state")
}

func TestConcurrentRevokeAndGetValueBug(t *testing.T) {
	// Run refresh & revoke loops quicker than usual
	defaultLoopInterval := loopInterval
	defer func() {
		stop()
		loopInterval = defaultLoopInterval
	}()
	loopInterval = 10 * time.Millisecond
	StartWith(1, 11, 1, 20*time.Millisecond)
	// Continuously spamming GetValue should force ConcurrentRefreshAndGetBug to manifest if it is present
	c := make(chan []byte, 1)
	key := "TestConcurrentRevokeAndGetValueBug"
	go busyGet(t, c, key)
	value := randomGetFunc("")
	i := CacheItem{
		Key:        key,
		Value:      value,
		Expiration: 15 * time.Millisecond,
		GetFunc:    randomGetFunc,
	}
	AddItem(i)
	// sleep long enough for the bug to kick in repeatably
	time.Sleep(1 * time.Second)
	c <- []byte("stop")
	time.Sleep(53 * time.Millisecond)
	assert.True(t, nil == GetValue(key), "Item should have been revoked by now")
}

func noopGetFunc(s string) []byte {
	return nil
}

func randomGetFunc(s string) []byte {
	return []byte(uuid.New().String())
}

func busyGet(t *testing.T, c chan []byte, k string) {
	defer func() {
		_ = <-c
	}()
	for len(c) == 0 {
		_ = GetValue(k)
	}
}
