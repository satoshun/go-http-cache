package cache

import "sync"

// Registry stores data
type Registry interface {
	Get(key []byte) (*HttpCache, error)

	Save(key []byte, h *HttpCache) error
}

// MemoryRegistry stores data on memory, not persistent data
//
// implements Registry
type MemoryRegistry struct {
	m     sync.RWMutex
	cache map[string]HttpCache
}

// NewMemoryRegistry returns a new MemoryRegistry
func NewMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{cache: make(map[string]HttpCache)}
}

// Get gets data from memory cache
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

// Save saves data to memory cache
func (r *MemoryRegistry) Save(key []byte, h *HttpCache) error {
	r.m.Lock()
	defer r.m.Unlock()
	r.cache[string(key)] = *h

	return nil
}
