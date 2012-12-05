package counter

import (
    "github.com/darkhelmet/env"
    "github.com/fzzbt/radix/redis"
    "net/url"
)

var client *redis.Client

func init() {
    config := redis.DefaultConfig()
    u := env.StringDefault("REDISTOGO_URL", "redis://127.0.0.1:6379")
    uri, err := url.Parse(u)
    if err != nil {
        panic(err)
    }
    config.Address = uri.Host
    if uri.User != nil {
        pw, _ := uri.User.Password()
        config.Password = pw
    }
    client = redis.NewClient(config)
}

func Get(key string) (int, error) {
    return client.Get(key).Int()
}

func Inc(key string, n int) error {
    reply := client.Incrby(key, n)
    if reply.Type == redis.ReplyError {
        return reply.Err
    }
    return nil
}
