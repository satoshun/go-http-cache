package cache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setUp() {
	cache = make(map[[KeySize]byte]httpCache)
}

func TestNoHeader(t *testing.T) {
	setUp()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	_, err := GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(cache) != 0 {
		t.Fatalf("expect len == 0, but actual %d", len(cache))
	}
}

func TestLastModified(t *testing.T) {
	setUp()
	n := time.Now().String()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Last-Modified", n)
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	_, err := GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(cache) != 1 {
		t.Fatalf("expect len == 1, but actual %d", len(cache))
	}
	for _, v := range cache {
		if v.body == nil {
			t.Fatal("body is nil")
		}
		if v.lastModified != n {
			t.Fatal("lastModified is empty")
		}
		if v.etag != "" {
			t.Fatal("etag is not empty")
		}
		if v.expires != nil {
			t.Fatal("expires is not empty")
		}
	}
}

func TestETag(t *testing.T) {
	setUp()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("ETag", "0123456789")
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	_, err := GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(cache) != 1 {
		t.Fatalf("expect len == 1, but actual %d", len(cache))
	}
	for _, v := range cache {
		if v.body == nil {
			t.Fatal("body is nil")
		}
		if v.etag != "0123456789" {
			t.Fatal("etag is empty")
		}
		if v.lastModified != "" {
			t.Fatal("lastModified is not empty")
		}
		if v.expires != nil {
			t.Fatal("expires is not empty")
		}
	}
}

func TestExpires(t *testing.T) {
	setUp()

	n := time.Now().Add(time.Second * 10).Format(time.RFC1123)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Expires", n)
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	_, err := GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(cache) != 1 {
		t.Fatalf("expect len == 1, but actual %d", len(cache))
	}
	for _, v := range cache {
		if v.body == nil {
			t.Fatal("body is nil")
		}
		if v.etag != "" {
			t.Fatal("etag is not empty")
		}
		if v.lastModified != "" {
			t.Fatal("lastModified is not empty")
		}
		if v.expires == nil {
			t.Fatal("expires is empty")
		}
	}

	res, err := GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(cache) != 1 {
		t.Fatalf("expect len == 1, but actual %d", len(cache))
	}
	if string(res.Cache) != "Hello, client" {
		t.Fatalf("actual: %s", string(res.Cache))
	}
}
