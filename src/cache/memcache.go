package cache

import (
    "fmt"
    "github.com/darkhelmet/env"
    "io"
    "log"
    "os"
    "syscall"
    "vendor/github.com/bmizerany/mc"
)

type mcCache struct {
    conn     *mc.Conn
    server   string
    username string
    password string
}

var logger = log.New(os.Stdout, "[memcache] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))

func newMemcacheCache(server, username, password string) (c *mcCache) {
    c = &mcCache{nil, fmt.Sprintf("%s:11211", server), username, password}
    c.connect()
    c.auth()
    return
}

func (c *mcCache) connect() {
    if cn, err := mc.Dial("tcp", c.server); err != nil {
        logger.Panicf("Error connecting to memcached: %s", err)
    } else {
        c.conn = cn
    }
}

func (c *mcCache) auth() {
    if err := c.conn.Auth(c.username, c.password); err != nil {
        logger.Panicf("Error authenticating with memcached: %s", err)
    }
}

func (c *mcCache) handleError(action, key string, err error) bool {
    switch err {
    case io.EOF, syscall.ECONNRESET:
        // Lost connection? Try reconnecting
        c.connect()
        // And of course we have to auth again
        fallthrough
    case mc.ErrAuthRequired:
        c.auth()
        return true
    case mc.ErrNotFound:
        // Cool story bro
    default:
        logger.Panicf("memcached error in %s for key %s: %s", action, key, err)
    }
    return false
}

func (c *mcCache) Get(key string) (string, error) {
    return c.rget(key, 10)
}

func (c *mcCache) rget(key string, limit int) (string, error) {
    value, _, _, err := c.conn.Get(key)
    if err != nil {
        if c.handleError("get", key, err) && limit > 0 {
            return c.rget(key, limit-1)
        }
    }
    return value, err
}

func (c *mcCache) Set(key, data string, ttl int) {
    c.rset(key, data, ttl, 10)
}

func (c *mcCache) rset(key, data string, ttl, limit int) {
    // Don't worry about errors, live on the edge
    if err := c.conn.Set(key, data, 0, 0, ttl); err != nil {
        if c.handleError("set", key, err) && limit > 0 {
            c.rset(key, data, ttl, limit-1)
        }
    }
}

func (c *mcCache) Fetch(key string, ttl int, f func() string) string {
    value, cas, _, err := c.conn.Get(key)
    if err != nil {
        c.handleError("fetch/get", key, err)
        value = f()
        /*  If this fails, don't worry too much
            In the situations it gets used, it doesn't matter */
        if err = c.conn.Set(key, value, cas, 0, ttl); err != nil {
            c.handleError("fetch/set", key, err)
        }
    }
    return value
}
