package cache

import (
	"github.com/xuyu/goredis"
	"net"
	"sync"
	"syscall"
)

type redisCache struct {
	redis *goredis.Redis
	mutex sync.Mutex
	url   string
}

func newRedisCache(url string) Cache {
	c := &redisCache{url: url}
	c.connect()
	return c
}

func (c *redisCache) connect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	logger.Printf("connecting")
	redis, err := goredis.DialURL(c.url)
	if err != nil {
		logger.Panicf("connecting failed: %s", err)
	}
	c.redis = redis
}

func (c *redisCache) handleError(err error, retry func()) {
	if err != nil {
		switch e := err.(type) {
		case *net.OpError:
			switch e.Err {
			case syscall.EPIPE:
				logger.Printf("broken pipe, reconnecting and retrying")
				c.connect()
				retry()
			default:
				logger.Printf("unhandled net.OpError: %s, %#v", err, err)
			}
		default:
			logger.Printf("unhandled error: %s, %#v", err)
		}
	}
}

func (c *redisCache) Get(key string) (string, error) {
	data, err := c.redis.Get(key)
	c.handleError(err, func() {
		data, err = c.redis.Get(key)
	})
	return string(data), err
}

func (c *redisCache) Set(key, data string, ttl int) error {
	err := c.redis.Set(key, data, ttl, 0, false, false)
	c.handleError(err, func() {
		err = c.redis.Set(key, data, ttl, 0, false, false)
	})
	return err
}
