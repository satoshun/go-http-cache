package cache

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPackageGetWithCache(t *testing.T) {
	n := time.Now().Add(time.Second * 10).Format(time.RFC1123)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Expires", n)
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	resp1, err := GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp2, err := GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp2.Response.StatusCode != StatusCacheContent {
		t.Fatal("should use cache data on ETag")
	}
	if !bytes.Equal(resp1.Cache, resp2.Cache) {
		t.Fatal("cache is invalid")
	}
}

func TestNoHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	c := NewMemoryCacheClient(&http.Client{})
	_, err := c.GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := c.r.(*MemoryRegistry)
	if len(r.cache) != 0 {
		t.Fatalf("expect len == 0, but actual %d", len(r.cache))
	}
}

func TestCacheControl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "public, max-age=86400, public")
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	c := NewMemoryCacheClient(&http.Client{})
	_, err := c.GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := c.r.(*MemoryRegistry)
	if len(r.cache) != 1 {
		t.Fatalf("expect len == 1, but actual %d", len(r.cache))
	}
	for _, v := range r.cache {
		if v.Body == nil {
			t.Fatal("body is nil")
		}
		if v.LastModified != "" {
			t.Fatal("lastModified is not empty")
		}
		if v.Etag != "" {
			t.Fatal("etag is not empty")
		}
		if v.Expires == nil {
			t.Fatal("expires is not empty")
		}
	}
}
func TestLastModified(t *testing.T) {
	n := time.Now().String()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Last-Modified", n)
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	c := NewMemoryCacheClient(&http.Client{})
	_, err := c.GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := c.r.(*MemoryRegistry)
	if len(r.cache) != 1 {
		t.Fatalf("expect len == 1, but actual %d", len(r.cache))
	}
	for _, v := range r.cache {
		if v.Body == nil {
			t.Fatal("body is nil")
		}
		if v.LastModified != n {
			t.Fatal("lastModified is empty")
		}
		if v.Etag != "" {
			t.Fatal("etag is not empty")
		}
		if v.Expires != nil {
			t.Fatal("expires is not empty")
		}
	}
}

func TestETag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("ETag", "0123456789")
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	c := NewMemoryCacheClient(&http.Client{})
	_, err := c.GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := c.r.(*MemoryRegistry)
	if len(r.cache) != 1 {
		t.Fatalf("expect len == 1, but actual %d", len(r.cache))
	}
	for _, v := range r.cache {
		if v.Body == nil {
			t.Fatal("body is nil")
		}
		if v.Etag != "0123456789" {
			t.Fatal("etag is empty")
		}
		if v.LastModified != "" {
			t.Fatal("lastModified is not empty")
		}
		if v.Expires != nil {
			t.Fatal("expires is not empty")
		}
	}
}

func TestExpires(t *testing.T) {
	n := time.Now().Add(time.Second * 10).Format(time.RFC1123)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Expires", n)
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	c := NewMemoryCacheClient(&http.Client{})
	_, err := c.GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := c.r.(*MemoryRegistry)
	if len(r.cache) != 1 {
		t.Fatalf("expect len == 1, but actual %d", len(r.cache))
	}
	for _, v := range r.cache {
		if v.Body == nil {
			t.Fatal("body is nil")
		}
		if v.Etag != "" {
			t.Fatal("etag is not empty")
		}
		if v.LastModified != "" {
			t.Fatal("lastModified is not empty")
		}
		if v.Expires == nil {
			t.Fatal("expires is empty")
		}
	}

	res, err := c.GetWithCache(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.cache) != 1 {
		t.Fatalf("expect len == 1, but actual %d", len(r.cache))
	}
	if string(res.Cache) != "Hello, client" {
		t.Fatalf("actual: %s", string(res.Cache))
	}
}
