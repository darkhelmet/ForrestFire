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

func expand(path string) string {
    return fmt.Sprintf("views/%s.mustache", path)
}

func getViewFile(file string) string {
    bytes, _ := ioutil.ReadFile(expand(file))
    return string(bytes)
}

func Page(page string, ctx *web.Context) string {
    return renderedCache.Get(page, func() string {
        return mustache.RenderFile(expand("layout"), map[string]string{
            "yield": Chunk(page),
            "donate": getViewFile("donate"),
            "footer": Chunk("footer"),
            "host": ctx.Host,
        })
    })
}

func Chunk(chunk string) string {
    return renderedCache.Get(chunk, func() string {
        return mustache.RenderFile(expand(chunk), map[string]string{
            "donate": getViewFile("donate"),
        })
    })
}
