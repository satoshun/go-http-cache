package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/satoshun/go-http-cache"
)

func main() {
	urls := os.Args[1:]

	for _, u := range urls {
		r, e := cache.GetWithCache(u)
		if e != nil {
			fmt.Printf("%v, %s\n", e, u)
			continue
		}
		if r.StatusCode != http.StatusOK {
			fmt.Printf("not success HTTP request - %d - %s\n", r.StatusCode, u)
			continue
		}

		r, _ = cache.GetWithCache(u)
		switch r.StatusCode {
		case cache.StatusCacheContent:
			fmt.Printf("%s - %s\n", "Use Cache", u)
		default:
			fmt.Printf("%s - %s\n", http.StatusText(r.StatusCode), u)
		}
	}
}
