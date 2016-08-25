package main

import (
	"encoding/json"
	"net/http"

	"github.com/bradfitz/gomemcache/memcache"
	cache "github.com/satoshun/go-http-cache/cache"
)

var mc = memcache.New("127.0.0.1:11211")

type MemcacheRegistry struct {
}

func (r *MemcacheRegistry) Get(key []byte) (*cache.HttpCache, error) {
	it, err := mc.Get(string(key))
	if err != nil {
		return nil, err
	}
	var v cache.HttpCache
	err = json.Unmarshal(it.Value, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *MemcacheRegistry) Save(key []byte, h *cache.HttpCache) error {
	v, err := json.Marshal(&h)
	if err != nil {
		return err
	}
	err = mc.Set(&memcache.Item{
		Key:   string(key),
		Value: v,
	})
	return err
}

func NewMemcacheClient(c *http.Client) *cache.HttpCacheClient {
	return &cache.HttpCacheClient{
		Client: c,
		R:      &MemcacheRegistry{},
	}
}
