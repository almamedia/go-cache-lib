# go-cache-lib

## Go version

- 1.11

## background cache

- set expired time for cache items
- update "to be expired" items in the background
- always answer only from cache to user requests
- log background fetches and errors, especially timeouts clearly
- in the case of error, do not update cache with erronous or empty data

## cache number two coming up ...

## testing

`cd client/main`

`go build && ./main`

## usage

```go
hc "github.com/almamedia/go-cache-lib"
```

set worker amount and buffered job amount if needed, defaults 20 and 200:

```go
hc.Start(5, 20)
```

get item:

```go
item := hc.GetItem(url)
```

add item:

```go
cacheItem := hc.CacheItem{
    Key:          url,
    Value:        res,
    Expire:       time.Now().Add(expire),
    UpdateLength: expire,
    GetFunc:      DataFetch,
}
hc.AddItem(cacheItem)
```

see cache_test.go and client/main/loadtest.go
