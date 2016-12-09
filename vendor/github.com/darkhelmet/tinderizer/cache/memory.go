package cache

import (
	"errors"
	"sync"
	"time"
)

type dictCache struct {
	dict  map[string]string
	mutex sync.Mutex
}

func newDictCache() Cache {
	c := new(dictCache)
	c.dict = make(map[string]string)
	return c
}

func (c *dictCache) lock() {
	c.mutex.Lock()
}

func (c *dictCache) unlock() {
	c.mutex.Unlock()
}

func (c *dictCache) reap(key string, ttl int) {
	// Convert seconds to nanoseconds
	<-time.After(time.Duration(int64(ttl) * 1e9))
	c.delete(key)
}

func (c *dictCache) delete(key string) {
	c.lock()
	defer c.unlock()
	delete(c.dict, key)
}

func (c *dictCache) Get(key string) (string, error) {
	c.lock()
	defer c.unlock()
	if v, ok := c.dict[key]; ok {
		return v, nil
	}
	return "", errors.New("not found")
}

func (c *dictCache) Set(key, data string, ttl int) error {
	c.lock()
	defer c.unlock()
	if _, ok := c.dict[key]; !ok {
		// Item not in the cache, so crank up the reaper for it
		go c.reap(key, ttl)
	}
	c.dict[key] = data
	return nil
}
