# go-cache-lib

Cache library for in-memory loading cache.

- Cachitems have expire and TTL
- Update expired items in background
- Revoke items whose TTL has been exceeded
- When cache is full, revoke items nearest to their TTL
- If item fetch causes error, keep old item
- Log background refreshes, special attention to error logging
## Go version

- 1.13 or newer

## Initial setup

Run setup.sh. It installs a commit-msg hook to enforce semantic commit messages

## Versioning & releases

Semantic vesioning, period. New version is released by github workflow whenever new commits are pushed to main branch.

Commit message format needed is enforced by commit-msg git hook. See https://github.com/angular/angular.js/blob/master/DEVELOPERS.md#commit-message-format

Commit messages must follow convention:

```
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

Scope, body and footer are optional, type and subject mandatory

Type can be one of the following:

- feat: A new feature
- fix: A bug fix
- docs: Documentation only changes
- style: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- refactor: A code change that neither fixes a bug nor adds a feature
- perf: A code change that improves performance
- test: Adding missing or correcting existing tests
- chore: Changes to the build process or auxiliary tools and libraries such as documentation generation

New major versions are released when there is commit with message footer starts with BREAKING CHANGE. The rest of the message after that should detail the changes.

Examples:

- test: add unit test for foobar
- feat(postgresql): support postgresql
- fix(mysql): error handling

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
    TTL:          1* time.Hour,
    UpdateLength: expire,
    GetFunc:      DataFetch,
}
hc.AddItem(cacheItem)
```

see cache_test.go and client/main/loadtest.go
