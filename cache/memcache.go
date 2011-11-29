package cache

import (
    "fmt"
    "github.com/bmizerany/mc.go"
)

type mcCache struct {
    conn *mc.Conn
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
    return value, err
}

func (c *mcCache) Set(key, data string, ttl int) {
    // Don't worry about errors, live on the edge
    c.conn.Set(key, data, 0, 0, ttl)
}

func (c *mcCache) Fetch(key string, ttl int, f func() string) string {
    value, cas, _, err := c.conn.Get(key)
    if err != nil {
        value = f()
        /*  If this fails, don't worry too much
            In the situations it gets used, it doesn't matter */
        c.conn.Set(key, value, cas, 0, ttl)
    }
    return value
}
