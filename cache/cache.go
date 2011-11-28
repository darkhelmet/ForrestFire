package cache

import (
    "env"
    "fmt"
    "github.com/bmizerany/mc.go"
)

type Cache interface {
    Get(key string) (string, error)
    Set(key string, data string, ttl int)
    Fetch(key string, ttl int, f func() string) string
}

var impl Cache

func init() {
    server := env.GetDefault("MEMCACHE_SERVERS", "")
    if server == "" {
       impl = new(dictCache)
    } else {
        if cn, err := mc.Dial("tcp", fmt.Sprintf("%s:11211", server)); err != nil {
            panic(err.Error())
        } else {
            username := env.GetDefault("MEMCACHE_USERNAME", "")
            password := env.GetDefault("MEMCACHE_PASSWORD", "")
            if err = cn.Auth(username, password); err != nil {
                panic(err.Error())
            } else {
                impl = &mcCache{cn}
            }
        }
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
