# HTTP Cache for Go

[![GoDoc](https://godoc.org/github.com/satoshun/go-http-cache?status.svg)](https://godoc.org/github.com/satoshun/go-http-cache) [![Build Status](https://travis-ci.org/satoshun/go-http-cache.svg?branch=master)](https://travis-ci.org/satoshun/go-http-cache) [![codecov](https://codecov.io/gh/satoshun/go-http-cache/branch/master/graph/badge.svg)](https://codecov.io/gh/satoshun/go-http-cache)


Reuse response body on Etag, Last-Modified, Expires and Cache-Control.


## Usage

```go
client := &http.Client{}
c := cache.NewMemoryCacheClient(client)
r, _ := http.NewRequest("GET", <url>, nil)
res, err := c.DoWithCache(r)
```

more detail [example](example/basic/main.go).


## bonus

### cachestat: check used cache or nor

```shell
## install
go get github.com/satoshun/go-http-cache/cmd/cachestat

## execute
cachestat https://www.google.co.jp/images/nav_logo242.png https://www.google.com/textinputassistant/tia.png

Not Modified - https://www.google.co.jp/images/nav_logo242.png
Use Cache - https://www.google.com/textinputassistant/tia.png
```

### implements on memcached

[memcached sample](example/memcache).
