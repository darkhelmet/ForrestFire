package render

import (
    "cache"
    "fmt"
    "io/ioutil"
    "web"
    "mustache"
)

const TTL = 24 * 60 * 60 * 1e9 // 1 day

func expand(path string) string {
    return fmt.Sprintf("views/%s.mustache", path)
}

func getViewFile(file string) string {
    bytes, _ := ioutil.ReadFile(expand(file))
    return string(bytes)
}

func Page(page string, ctx *web.Context) string {
    yield := Chunk(page)
    footer := Chunk(page)
    return cache.CheckAndSet(page, TTL, func() cache.Any {
        return mustache.RenderFile(expand("layout"), map[string]string{
            "yield":  yield,
            "donate": getViewFile("donate"),
            "footer": footer,
            "host":   ctx.Host,
        })
    }).(string)
}

func Chunk(chunk string) string {
    return cache.CheckAndSet(chunk, TTL, func() cache.Any {
        return mustache.RenderFile(expand(chunk), map[string]string{
            "donate": getViewFile("donate"),
        })
    }).(string)
}
