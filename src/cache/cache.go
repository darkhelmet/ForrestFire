package cache

import (
    "github.com/darkhelmet/env"
)

type Cache interface {
    Get(key string) (string, error)
    Set(key string, data string, ttl int)
    Fetch(key string, ttl int, f func() string) string
}

var impl Cache

func init() {
    server := env.StringDefault("MEMCACHE_SERVERS", "")
    if server == "" {
        impl = newDictCache()
    } else {
        impl = newMemcacheCache(server, env.StringDefault("MEMCACHE_USERNAME", ""), env.StringDefault("MEMCACHE_PASSWORD", ""))
    }
}

func Get(key string) (string, error) {
    return impl.Get(key)
}

func Set(key, data string, ttl int) {
    impl.Set(key, data, ttl)
}

func Fetch(key string, ttl int, f func() string) string {
    return impl.Fetch(key, ttl, f)
}
