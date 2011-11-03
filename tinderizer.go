package main

import (
    "fmt"
    "os"
    "render"
    "web"
)

func port() string {
    tmp := os.Getenv("PORT")
    if tmp == "" {
        tmp = "8080"
    }
    return tmp
}

func forJson(ctx *web.Context) {
    ctx.SetHeader("Access-Control-Allow-Origin", "*", true)
    ctx.SetHeader("Content-Type", "application/json; charset=utf-8", true)
}

func main() {
    web.Get("/ajax/submit.json", func(ctx *web.Context) string {
        forJson(ctx)
        return "submit"
    })

    web.Get("/ajax/status/(.*)", func(ctx *web.Context, id string) string {
        forJson(ctx)
        return "status"
    })

    web.Get("/", func(ctx *web.Context) string {
        return render.Page("index", ctx)
    })

    web.Get("/kindle-email", func() string {
        return render.Chunk("kindle_email")
    })

    web.Get("/(firefox|safari|chrome|ie|ios)", func(page string) string {
        return render.Chunk(page)
    })

    web.Get("/(faq|bugs|contact)", func(ctx *web.Context, page string) string {
        return render.Page(page, ctx)
    })

    web.Run(fmt.Sprintf("0.0.0.0:%s", port()))
}
