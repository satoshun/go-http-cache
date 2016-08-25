package cache

import (
	"crypto/md5"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const KeySize = md5.Size

type Registry interface {
	Get(key [KeySize]byte) (*httpCache, error)

	Save(key [KeySize]byte, h *httpCache) error
}

type MemoryRegistry struct {
	m     sync.RWMutex
	cache map[[KeySize]byte]httpCache
}

func (r *MemoryRegistry) Get(key [KeySize]byte) (*httpCache, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	c, _ := r.cache[key]
	return &c, nil
}

func (r *MemoryRegistry) Save(key [KeySize]byte, h *httpCache) error {
	r.m.Lock()
	defer r.m.Unlock()
	r.cache[key] = *h

	return nil
}

// Response embed http.Response and Cache data
type Response struct {
	*http.Response

	Cache []byte
}

type httpCache struct {
	body []byte

	etag         string
	lastModified string
	expires      *time.Time
}

type HttpCacheClient struct {
	*http.Client

	r Registry
}

var DefaultClient = HttpCacheClient{Client: http.DefaultClient, r: &MemoryRegistry{cache: make(map[[KeySize]byte]httpCache)}}

func NewHttpCacheClient(c *http.Client) *HttpCacheClient {
	return &HttpCacheClient{Client: c, r: &MemoryRegistry{cache: make(map[[KeySize]byte]httpCache)}}
}

func GetWithCache(url string) (resp *Response, err error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return DefaultClient.DoWithCache(r)
}

func (client *HttpCacheClient) GetWithCache(url string) (*Response, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return client.DoWithCache(r)
}

func (client *HttpCacheClient) DoWithCache(req *http.Request) (*Response, error) {
	key := standardKey(req)

	c, _ := client.r.Get(key)
	if c != nil {
		if c.expires != nil && c.expires.After(time.Now()) {
			return &Response{Cache: c.body}, nil
		}
		if c.etag != "" {
			req.Header.Set("If-None-Match", c.etag)
		}
		if c.lastModified != "" {
			req.Header.Set("If-Modified-Since", c.lastModified)
		}
	}

	res, err := client.Client.Do(req)
	if err != nil {
		return &Response{Response: res}, err
	}
	if res.StatusCode == http.StatusNotModified {
		return &Response{Response: res, Cache: c.body}, nil
	}

	lm := res.Header.Get("Last-Modified")
	etag := res.Header.Get("ETag")
	expires := res.Header.Get("Expires")
	iee := isEmptyExpires(expires)
	if lm == "" && etag == "" && iee {
		return &Response{Response: res}, err
	}

	var ed *time.Time
	if !iee {
		d, err := time.Parse(time.RFC1123, expires)
		if err == nil {
			ed = &d
		}
	}
	body, _ := ioutil.ReadAll(res.Body)
	client.r.Save(key, &httpCache{body: body, lastModified: lm, etag: etag, expires: ed})

	return &Response{Response: res, Cache: body}, err
}

func isEmptyExpires(expires string) bool {
	return expires == "" || expires == "-1" || expires == "0"
}

func standardKey(req *http.Request) [KeySize]byte {
	headers := make([]string, 0, len(req.Header))
	for k, vv := range req.Header {
		for _, v := range vv {
			headers = append(headers, k+":"+v)
		}
	}
	sort.StringsAreSorted(headers)
	u := req.URL.String() + strings.Join(headers, ";")
	h := md5.New()
	io.WriteString(h, u)
	b := h.Sum(nil)
	var bb [KeySize]byte
	copy(bb[:], b[0:KeySize])
	return bb
}
