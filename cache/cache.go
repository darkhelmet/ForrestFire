package cache

import (
    "errors"
    "sync"
    "time"
)

type Any interface{}
type Cache map[string]Any

var dict Cache
var mutex sync.Mutex

func init() {
    dict = make(Cache)
}

func reap(key string, ttl int64) {
    <-time.After(ttl)
    mutex.Lock()
    defer mutex.Unlock()
    delete(dict, key)
}

func Set(key string, value Any, ttl int64) {
    mutex.Lock()
    defer mutex.Unlock()
    _, ok := dict[key]
    if !ok {
        // Item not in the cache, so crank up the reaper for it
        go reap(key, ttl)
    }
    dict[key] = value
}

func Get(key string) (Any, error) {
    mutex.Lock()
    defer mutex.Unlock()
    v, ok := dict[key]
    if ok {
        return v, nil
    }
    return nil, errors.New("not found")
}

func CheckAndSet(key string, ttl int64, f func() Any) Any {
    mutex.Lock()
    defer mutex.Unlock()
    value, ok := dict[key]
    if !ok {
        // We are setting the value, so
        go reap(key, ttl)
        value = f()
        dict[key] = value
    }
    return value
}
