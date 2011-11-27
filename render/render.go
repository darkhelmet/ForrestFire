package render

import (
    "bytes"
    "cache"
    "fmt"
    "github.com/darkhelmet/web.go"
    "io/ioutil"
    "text/template"
)

const TTL = 24 * 60 * 60 * 1e9 // 1 day

func expand(path string) string {
    return fmt.Sprintf("views/%s.tmpl", path)
}

func getViewFile(file string) string {
    bytes, _ := ioutil.ReadFile(expand(file))
    return string(bytes)
}

func getTemplate(name string) *template.Template {
    tmpl, err := template.ParseFile(expand(name))
    if err != nil {
        panic(err.Error())
    }
    return tmpl
}

func renderToString(name string, data map[string]string) string {
    var buffer bytes.Buffer
    tmpl := getTemplate(name)
    err := tmpl.Execute(&buffer, data)
    if err != nil {
        panic(err.Error())
    }
    return buffer.String()
}

func Page(page string, ctx *web.Context) string {
    yield := Chunk(page)
    footer := Chunk("footer")
    return cache.CheckAndSet(fmt.Sprintf("page/%s", page), TTL, func() cache.Any {
        return renderToString("layout", map[string]string{
            "yield":  yield,
            "donate": getViewFile("donate"),
            "footer": footer,
            "host":   ctx.Host,
        })
    }).(string)
}

func Chunk(chunk string) string {
    return cache.CheckAndSet(fmt.Sprintf("chunk/%s", chunk), TTL, func() cache.Any {
        return renderToString(chunk, map[string]string{
            "donate": getViewFile("donate"),
        })
    }).(string)
}
