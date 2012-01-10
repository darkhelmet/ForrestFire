package render

import (
    "cache"
    "fmt"
    "github.com/darkhelmet/web.go"
    "io/ioutil"
    "template"
)

const TTL = 24 * 60 * 60 // 1 day

func expand(path string) string {
    return fmt.Sprintf("views/%s.tmpl", path)
}

func getViewFile(file string) string {
    bytes, _ := ioutil.ReadFile(expand(file))
    return string(bytes)
}

func render(name string, data interface{}) string {
    return template.RenderToString(name, getViewFile(name), data)
}

func Page(page string, ctx *web.Context) string {
    yield := Chunk(page)
    footer := Chunk("footer")
    return cache.Fetch(fmt.Sprintf("page/%s", page), TTL, func() string {
        return render("layout", map[string]string{
            "yield":  yield,
            "donate": getViewFile("donate"),
            "footer": footer,
            "host":   ctx.Host,
        })
    })
}

func Chunk(chunk string) string {
    return cache.Fetch(fmt.Sprintf("chunk/%s", chunk), TTL, func() string {
        return render(chunk, map[string]string{
            "donate": getViewFile("donate"),
        })
    })
}
