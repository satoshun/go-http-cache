package cache

import (
	"crypto/md5"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Response embed http.Response and Cache data
type Response struct {
	*http.Response

	Cache []byte
}

// HttpCache represents brief HTTP Response data
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

// HttpCacheClient has http.Client and Registry
type HttpCacheClient struct {
	*http.Client

	r Registry
}

var DefaultClient = &HttpCacheClient{
	Client: http.DefaultClient,
	r:      NewMemoryRegistry(),
}

// NewMemoryCacheClient returns a new HttpCacheClient from MemoryRegistry
func NewMemoryCacheClient(c *http.Client) *HttpCacheClient {
	return NewClient(c, &MemoryRegistry{cache: make(map[string]HttpCache)})
}

// NewClient returns a new HttpCacheClient
func NewClient(c *http.Client, r Registry) *HttpCacheClient {
	return &HttpCacheClient{Client: c, r: r}
}

// GetWithCache returns a new Response data by HTTP Request or Registry cache
func GetWithCache(url string) (resp *Response, err error) {
	return DefaultClient.GetWithCache(url)
}

// GetWithCache returns a new Response data by HTTP Request or Registry cache
func (client *HttpCacheClient) GetWithCache(url string) (*Response, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return client.DoWithCache(r)
}

// DoWithCache returns a new Response data by HTTP Request or Registry cache
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
