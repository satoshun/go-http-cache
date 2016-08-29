package cache

import (
	"crypto/md5"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Registry interface {
	Get(key []byte) (*HttpCache, error)

	Save(key []byte, h *HttpCache) error
}

type MemoryRegistry struct {
	m     sync.RWMutex
	cache map[string]HttpCache
}

func (r *MemoryRegistry) Get(key []byte) (*HttpCache, error) {
	r.m.RLock()
	c, _ := r.cache[string(key)]
	r.m.RUnlock()
	if c.invalidate() {
		r.m.Lock()
		delete(r.cache, string(key))
		r.m.Unlock()
		return nil, nil
	}
	return &c, nil
}

func (r *MemoryRegistry) Save(key []byte, h *HttpCache) error {
	r.m.Lock()
	defer r.m.Unlock()
	r.cache[string(key)] = *h

	return nil
}

// Response embed http.Response and Cache data
type Response struct {
	*http.Response

	Cache []byte
}

type HttpCache struct {
	Body []byte `json:"body"`

	Etag         string     `json:"etag"`
	LastModified string     `json:"last_modified"`
	Expires      *time.Time `json:"expires"`
}

func (c *HttpCache) invalidate() bool {
	if c.Etag != "" {
		return false
	}
	if c.LastModified != "" {
		return false
	}
	return c.Expires != nil && c.Expires.Before(time.Now())
}

type HttpCacheClient struct {
	*http.Client

	r Registry
}

var DefaultClient = HttpCacheClient{Client: http.DefaultClient, r: &MemoryRegistry{cache: make(map[string]HttpCache)}}

func NewMemoryCacheClient(c *http.Client) *HttpCacheClient {
	return NewClient(c, &MemoryRegistry{cache: make(map[string]HttpCache)})
}

func NewClient(c *http.Client, r Registry) *HttpCacheClient {
	return &HttpCacheClient{Client: c, r: r}
}

func GetWithCache(url string) (resp *Response, err error) {
	return DefaultClient.GetWithCache(url)
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
		if c.Expires != nil && c.Expires.After(time.Now()) {
			return &Response{Cache: c.Body}, nil
		}
		if c.Etag != "" {
			req.Header.Set("If-None-Match", c.Etag)
		}
		if c.LastModified != "" {
			req.Header.Set("If-Modified-Since", c.LastModified)
		}
	}

	res, err := client.Client.Do(req)
	if err != nil {
		return &Response{Response: res}, err
	}
	if res.StatusCode == http.StatusNotModified {
		return &Response{Response: res, Cache: c.Body}, nil
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
	client.r.Save(key, &HttpCache{Body: body, LastModified: lm, Etag: etag, Expires: ed})

	return &Response{Response: res, Cache: body}, err
}

func isEmptyExpires(expires string) bool {
	return expires == "" || expires == "-1" || expires == "0"
}

func standardKey(req *http.Request) []byte {
	headers := make([]string, 0, len(req.Header))
	for k, vv := range req.Header {
		for _, v := range vv {
			headers = append(headers, k+":"+v)
		}
	}
	sort.StringsAreSorted(headers)
	u := req.URL.String() + strings.Join(headers, ";")
	d := md5.Sum([]byte(u))
	return digest(d[:])
}

const ldigsts = "0123456789abcdef"

func digest(b []byte) []byte {
	buf := make([]byte, 0, len(b)*2)
	for _, c := range b {
		buf = append(buf, ldigsts[c>>4], ldigsts[c&0x0F])
	}

	return buf
}
