package cache

import "sync"

// Registry stores data
type Registry interface {
	Get(key []byte) (*HTTPCache, error)

	Save(key []byte, h *HTTPCache) error
}

// MemoryRegistry stores data on memory, not persistent data
//
// implements Registry
type MemoryRegistry struct {
	m     sync.RWMutex
	cache map[string]HTTPCache
}

// NewMemoryRegistry returns a new MemoryRegistry
func NewMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{cache: make(map[string]HTTPCache)}
}

// Get gets data from memory cache
func (r *MemoryRegistry) Get(key []byte) (*HTTPCache, error) {
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
func (r *MemoryRegistry) Save(key []byte, h *HTTPCache) error {
	r.m.Lock()
	defer r.m.Unlock()
	r.cache[string(key)] = *h

	return nil
}
