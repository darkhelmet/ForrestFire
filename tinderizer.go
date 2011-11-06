package main

import (
    "bookmarklet"
    "cache"
    "env"
    "extractor"
    "fmt"
    "job"
    "json"
    "render"
    "regexp"
    "web"
)

const Limit = 10
const TTL = 5 * 60 * 1e9 // 5 minutes
var done *regexp.Regexp

type JSON map[string]interface{}

func port() string {
    return env.GetDefault("PORT", "8080")
}

func startJson(ctx *web.Context) {
    ctx.SetHeader("Access-Control-Allow-Origin", "*", true)
    ctx.SetHeader("Content-Type", "application/json; charset=utf-8", true)
    ctx.StartResponse(200)
}

func renderJson(ctx *web.Context, data JSON) {
    raw, _ := json.Marshal(data)
    ctx.Write(raw)
}

func main() {
    done = regexp.MustCompile("(?i:done|failed|limited|invalid|error)")
    web.Get("/ajax/submit.json", func(ctx *web.Context) {
        // TODO: Rate limiting
        // TODO: Email checking
        // TODO: Blacklisting
        startJson(ctx)
        j := job.New(ctx.Params["email"], ctx.Params["url"])
        cache.Set(j.KeyString(), "Working...", TTL)
        extractor.Extract(j)
        renderJson(ctx, JSON{
            "message": "Submitted! Hang tight...",
            "id":      j.KeyString(),
        })
    })

    web.Get("/ajax/status/(.*).json", func(ctx *web.Context, id string) {
        startJson(ctx)

        var message string
        isDone := true

        if v, err := cache.Get(id); err == nil {
            message = v.(string)
            isDone = done.MatchString(message)
        } else {
            message = "No job with that ID found."
        }

        renderJson(ctx, JSON{
            "message": message,
            "done":    isDone,
        })
    })

    web.Get("/static/bookmarklet.js", func(ctx *web.Context) {
        ctx.SetHeader("Content-Type", "application/javascript; charset=utf-8", true)
        ctx.StartResponse(200)
        ctx.Write(bookmarklet.Javascript())
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
