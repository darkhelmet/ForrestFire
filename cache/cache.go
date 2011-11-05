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

func Set(key string, value Any, ttl int64) {
    mutex.Lock()
    defer mutex.Unlock()
    _, ok := dict[key]
    if !ok {
        // Item not in the cache, so crank up the reaper for it
        go func() {
           <-time.After(ttl)
           mutex.Lock()
           defer mutex.Unlock()
           delete(dict, key)
        }()
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
