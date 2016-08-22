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

type httpCache struct {
	body []byte

	etag         string
	lastModified string
	expires      *time.Time
}

const KeySize = md5.Size

var (
	cache map[[KeySize]byte]httpCache
	m     sync.RWMutex
)

type HttpCacheClient struct {
	*http.Client
}

func NewHttpCacheClient(c *http.Client) *HttpCacheClient {
	return &HttpCacheClient{
		c,
	}
}

func (client *HttpCacheClient) DoWithCache(req *http.Request) (*http.Response, []byte, error) {
	key := standardKey(req)

	m.RLock()
	c, ok := cache[key]
	if ok {
		if c.expires != nil && c.expires.After(time.Now()) {
			return nil, c.body, nil
		}
		if c.etag != "" {
			req.Header.Set("If-None-Match", c.etag)
		}
		if c.lastModified != "" {
			req.Header.Set("If-Modified-Since", c.lastModified)
		}
	}
	m.RUnlock()

	res, err := client.Client.Do(req)
	if err != nil {
		return res, nil, err
	}
	if res.StatusCode == http.StatusNotModified {
		return res, c.body, nil
	}

	lm := res.Header.Get("Last-Modified")
	etag := res.Header.Get("Etag")
	expires := res.Header.Get("Expires")
	iee := isEmptyExpires(expires)
	if lm == "" && etag == "" && iee {
		return res, nil, err
	}

	var ed *time.Time
	if !iee {
		d, err := time.Parse(time.RFC1123, expires)
		if err == nil {
			ed = &d
		}
	}

	m.Lock()
	body, _ := ioutil.ReadAll(res.Body)
	cache[key] = httpCache{body: body, lastModified: lm, etag: etag, expires: ed}
	m.Unlock()

	return res, body, err
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

func init() {
	cache = make(map[[KeySize]byte]httpCache)
}
