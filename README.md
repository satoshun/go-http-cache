# HTTP Cache for Go

Reuse contents on Etag, Last-Modified and Expires.


## Usage

```go
client := &http.Client{}
c := cache.NewHttpCacheClient(client)
r, _ := http.NewRequest("GET", <url>, nil)
res, body, err := c.DoWithCache(r)
```

more detail [example/main.go].


## TODOs

- Implements HttpCacheClient#Get
- Writes test
