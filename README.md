# go-cache-lib

## background cache

- set expired time for cache items
- update "to be expired" items in the background
- always answer only from cache to user requests
- log background fetches and errors, especially timeouts clearly
- in the case of error, do not update cache with erronous or empty data

## cache number two coming up ...

# usage

see cache_test.go and client/main/loadtest.go