package render

import (
    "fmt"
    "io/ioutil"
    "web"
    "mustache"
)

type cache struct {
    m map[string]string
}

func newCache() (* cache) {
    return &cache{ make(map[string]string) }
}

type cacheFunc func() string

func (c *cache) Get(key string, def cacheFunc) string {
    v, ok := c.m[key]
    if !ok {
        d := def()
        c.m[key] = d
        return d
    }
    return v
}

var renderedCache * cache = newCache()
var rawCache * cache = newCache()

func getViewFile(file string) cacheFunc {
    return func() string {
        bytes, _ := ioutil.ReadFile(fmt.Sprintf("views/%s.mustache", file))
        return string(bytes)
    }
}

func donate() string {
    return rawCache.Get("donate", getViewFile("donate"))
}

func Page(page string, ctx *web.Context) string {
    return renderedCache.Get(page, func() string {
        return mustache.Render(rawCache.Get("layout", getViewFile("layout")), map[string]string{
            "yield": Chunk(page),
            "donate": donate(),
            "footer": Chunk("footer"),
            "host": ctx.Host,
        })
    })
}

func Chunk(chunk string) string {
    return renderedCache.Get(chunk, func() string {
        return mustache.Render(rawCache.Get(chunk, getViewFile(chunk)), map[string]string{
            "donate": donate(),
        })
    })
}
