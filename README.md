# HTTP Cache for Go

[![Build Status](https://travis-ci.org/satoshun/go-http-cache?branch=master)](https://travis-ci.org/satoshun/go-http-cache)

Reuse response body on Etag, Last-Modified and Expires.


## Usage

```go
client := &http.Client{}
c := cache.NewMemoryCacheClient(client)
r, _ := http.NewRequest("GET", <url>, nil)
res, err := c.DoWithCache(r)
```

more detail [example/basic/main.go].
