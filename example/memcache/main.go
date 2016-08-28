package main

import (
	"fmt"
	"net/http"
)

func main() {
	mc.FlushAll()

	testEtagLastModified()
	testExpires()
}

func testEtagLastModified() {
	u := "https://www.google.co.jp/images/nav_logo242.png"
	client := &http.Client{}
	c := NewMemcacheClient(client)

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		panic(err)
	}

	res, err := c.DoWithCache(r)
	if err != nil || res.StatusCode != http.StatusOK {
		panic(err)
	}

	// Use cached body
	res, err = c.DoWithCache(r)
	if res.StatusCode != http.StatusNotModified || err != nil {
		panic(err)
	}

	// reuse body!
	fmt.Println(res.Cache)
}

func testExpires() {
	u := "https://www.google.com/textinputassistant/tia.png"
	client := &http.Client{}
	c := NewMemcacheClient(client)

	res, err := c.GetWithCache(u)
	if err != nil || res.StatusCode != http.StatusOK {
		panic(err)
	}

	// Use cached body
	res, err = c.GetWithCache(u)
	if err != nil || res.Response != nil {
		panic(err)
	}

	// reuse body!
	fmt.Println(res.Cache)
}
