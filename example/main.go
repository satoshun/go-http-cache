package main

import (
	"fmt"
	"net/http"

	cache "github.com/satoshun/go-http-cache/cache"
)

func main() {
	testEtagLastModified()
	testExpires()
}

func testEtagLastModified() {
	u := "https://www.google.co.jp/images/nav_logo242.png"
	client := &http.Client{}
	c := cache.NewHttpCacheClient(client)

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		panic(err)
	}

	res, err := c.DoWithCache(r)
	if res.StatusCode != http.StatusOK || err != nil {
		panic(err)
	}

	// Use cached body
	res2, err := c.DoWithCache(r)
	if res2.StatusCode != http.StatusNotModified || err != nil {
		panic(err)
	}

	// reuse body!
	fmt.Println(res2.Cache)
}

func testExpires() {
	u := "https://www.google.com/textinputassistant/tia.png"
	client := &http.Client{}
	c := cache.NewHttpCacheClient(client)

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		panic(err)
	}
	r.Header.Add("hoge", "fuga")

	res, err := c.DoWithCache(r)
	if err != nil || res.StatusCode != http.StatusOK {
		panic(err)
	}

	// Use cached body
	res2, err := c.DoWithCache(r)
	if err != nil || res2.Response != nil {
		panic(err)
	}

	// reuse body!
	fmt.Println(res2.Cache)
}
