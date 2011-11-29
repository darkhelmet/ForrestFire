package cache

import (
    "errors"
    "sync"
    "time"
)

type dictCache struct {
    dict map[string]string
    mutex sync.Mutex
}

func newDictCache() *dictCache {
    c := new(dictCache)
    c.dict = make(map[string]string)
    return c
}

func (c *dictCache) reap(key string, ttl int) {
    // Convert seconds to nanoseconds
    <-time.After(int64(ttl) * 1e9)
    c.mutex.Lock()
    defer c.mutex.Unlock()
    delete(c.dict, key)
}

func (c *dictCache) Get(key string) (string, error) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    v, ok := c.dict[key]
    if ok {
        return v, nil
    }
    return "", errors.New("not found")
}

func (c *dictCache) Set(key, data string, ttl int) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    _, ok := c.dict[key]
    if !ok {
        // Item not in the cache, so crank up the reaper for it
        go c.reap(key, ttl)
    }
    c.dict[key] = data
}

func (c *dictCache) Fetch(key string, ttl int, f func() string) string {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    value, ok := c.dict[key]
    if !ok {
        // We are setting the value, so
        go c.reap(key, ttl)
        value = f()
        c.dict[key] = value
    }
    return value
}
