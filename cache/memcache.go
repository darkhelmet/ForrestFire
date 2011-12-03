package cache

import (
    "fmt"
    "github.com/bmizerany/mc.go"
)

type mcCache struct {
    conn *mc.Conn
}

func log(action, key string, err error) {
    println(fmt.Sprintf("memcached error in %s for key %s: %s", action, key, err.Error()))
}

func newMemcacheCache(server, username, password string) (c *mcCache) {
    if cn, err := mc.Dial("tcp", fmt.Sprintf("%s:11211", server)); err != nil {
        panic(err.Error())
    } else {
        if err = cn.Auth(username, password); err != nil {
            panic(err.Error())
        } else {
            c = &mcCache{cn}
        }
    }
    return
}

func (c *mcCache) Get(key string) (string, error) {
    value, _, _, err := c.conn.Get(key)
    if err != nil {
        log("get", key, err)
    }
    return value, err
}

func (c *mcCache) Set(key, data string, ttl int) {
    // Don't worry about errors, live on the edge
    if err := c.conn.Set(key, data, 0, 0, ttl); err != nil {
        log("set", key, err)
    }
}

func (c *mcCache) Fetch(key string, ttl int, f func() string) string {
    value, cas, _, err := c.conn.Get(key)
    if err != nil {
        log("fetch/get", key, err)
        value = f()
        /*  If this fails, don't worry too much
            In the situations it gets used, it doesn't matter */
        if err = c.conn.Set(key, value, cas, 0, ttl); err != nil {
            log("fetch/set", key, err)
        }
    }
    return value
}
