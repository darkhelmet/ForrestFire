package cache

import (
	"github.com/darkhelmet/env"
	"log"
	"os"
	"regexp"
)

type Cache interface {
	Get(key string) (string, error)
	Set(key string, data string, ttl int) error
}

var impl Cache = newDictCache()
var logger *log.Logger

func SetupRedis(url, options string) {
	url = regexp.MustCompile(`^redis:`).ReplaceAllString(url, "tcp:")
	url += "/0?" + options
	logger = log.New(os.Stdout, "[redis] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
	impl = newRedisCache(url)
}

func Get(key string) (string, error) {
	return impl.Get(key)
}

func Set(key, data string, ttl int) error {
	return impl.Set(key, data, ttl)
}
