# HTTP Cache for Go

[![GoDoc](https://godoc.org/github.com/satoshun/go-http-cache/cache?status.svg)](https://godoc.org/github.com/satoshun/go-http-cache/cache) [![Build Status](https://travis-ci.org/satoshun/go-http-cache?branch=master)](https://travis-ci.org/satoshun/go-http-cache)

Reuse response body on Etag, Last-Modified and Expires.


## Usage

```go
client := &http.Client{}
c := cache.NewMemoryCacheClient(client)
r, _ := http.NewRequest("GET", <url>, nil)
res, err := c.DoWithCache(r)
```

more detail [example/basic/main.go].


## bonus

cache checker cmd

```shell
## install
go get github.com/satoshun/go-http-cache/cmd/cachestat

## execute
cachestat https://www.google.co.jp/images/nav_logo242.png https://www.google.com/textinputassistant/tia.png

Not Modified - https://www.google.co.jp/images/nav_logo242.png
Use Cache - https://www.google.com/textinputassistant/tia.png
```
