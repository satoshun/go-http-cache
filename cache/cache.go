package cache

import (
	"crypto/md5"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const StatusCacheContent = 999

// Response embed http.Response and Cache data
type Response struct {
	*http.Response

	Cache []byte
}

// HTTPCache represents brief HTTP Response data
type HTTPCache struct {
	Body []byte `json:"body"`

	Etag         string     `json:"etag"`
	LastModified string     `json:"last_modified"`
	Expires      *time.Time `json:"expires"`
}

func (c *HTTPCache) invalidate() bool {
	if c.Etag != "" {
		return false
	}
	if c.LastModified != "" {
		return false
	}
	return c.Expires != nil && c.Expires.Before(time.Now())
}

// HTTPCacheClient has http.Client and Registry
type HTTPCacheClient struct {
	*http.Client

	r Registry
}

var DefaultClient = &HTTPCacheClient{
	Client: http.DefaultClient,
	r:      NewMemoryRegistry(),
}

// NewMemoryCacheClient returns a new HTTPCacheClient from MemoryRegistry
func NewMemoryCacheClient(c *http.Client) *HTTPCacheClient {
	return NewClient(c, &MemoryRegistry{cache: make(map[string]HTTPCache)})
}

// NewClient returns a new HTTPCacheClient
func NewClient(c *http.Client, r Registry) *HTTPCacheClient {
	return &HTTPCacheClient{Client: c, r: r}
}

// GetWithCache returns a new Response data by HTTP Request or Registry cache
func GetWithCache(url string) (resp *Response, err error) {
	return DefaultClient.GetWithCache(url)
}

// GetWithCache returns a new Response data by HTTP Request or Registry cache
func (client *HTTPCacheClient) GetWithCache(url string) (*Response, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return client.DoWithCache(r)
}

// DoWithCache returns a new Response data by HTTP Request or Registry cache
func (client *HTTPCacheClient) DoWithCache(req *http.Request) (*Response, error) {
	key := standardKey(req)

	c, _ := client.r.Get(key)
	if c != nil {
		if c.Expires != nil && c.Expires.After(time.Now()) {
			return &Response{
				Response: &http.Response{StatusCode: StatusCacheContent},
				Cache:    c.Body,
			}, nil
		}
		if c.Etag != "" {
			req.Header.Set("If-None-Match", c.Etag)
		}
		if c.LastModified != "" {
			req.Header.Set("If-Modified-Since", c.LastModified)
		}
	}

	res, err := client.Do(req)
	if err != nil {
		return &Response{Response: res}, err
	}
	if res.StatusCode == http.StatusNotModified {
		return &Response{Response: res, Cache: c.Body}, nil
	}

	cc := res.Header.Get("Cache-Control")
	var cci int
	if cc != "" {
		i := strings.Index(cc, "max-age=")
		if i != -1 {
			i += 8
			j := i + 1
			for ; j < len(cc); j++ {
				if cc[j] >= '0' && cc[j] <= '9' {
					continue
				}
				break
			}
			cci, _ = strconv.Atoi(cc[i:j])
		}
	}

	var expires string
	if cci == 0 {
		expires = res.Header.Get("Expires")
	}
	iee := cci == 0 && isEmptyExpires(expires)
	lm := res.Header.Get("Last-Modified")
	etag := res.Header.Get("ETag")
	if lm == "" && etag == "" && iee {
		return &Response{Response: res}, err
	}

	var ed *time.Time
	if !iee {
		if cci != 0 {
			e := time.Now().Add(time.Duration(cci) * time.Second)
			ed = &e
		} else {
			d, err := time.Parse(time.RFC1123, expires)
			if err == nil {
				ed = &d
			}
		}
	}
	body, _ := ioutil.ReadAll(res.Body)
	client.r.Save(key, &HTTPCache{Body: body, LastModified: lm, Etag: etag, Expires: ed})

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
