package cache

import (
    "github.com/bmizerany/mc.go"
)

type mcCache struct {
    conn *mc.Conn
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
        /*  If this fails, might be because of cas stuff,
            so don't worry too hard */
        c.conn.Set(key, value, cas, 0, ttl)
    }
    return value
}
