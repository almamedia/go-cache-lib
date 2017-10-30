package gocachelib

import (
	"strings"
	"testing"
	"time"

	"github.com/almamedia/go-cache-lib/client"
	"github.com/stretchr/testify/assert"
)

func TestSetItem(t *testing.T) {
	url := "https://httpbin.org/ip"
	i := CacheItem{
		Key:          url,
		Value:        nil,
		Expire:       time.Now(),
		UpdateLength: 1 * time.Second,
		GetFunc:      client.DataFetch,
	}
	AddItem(i)
	time.Sleep(2 * time.Second)
	assert.True(t, strings.Contains(string(GetItem(url)), "origin"))
	cache.Remove(url)
	assert.False(t, strings.Contains(string(GetItem(url)), "origin"))
	assert.True(t, string(GetItem(url)) == "")
}
