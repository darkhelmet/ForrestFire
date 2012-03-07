package main

import (
    "bookmarklet"
    "cache"
    "encoding/json"
    "env"
    "extractor"
    "fmt"
    "github.com/garyburd/twister/expvar"
    "github.com/garyburd/twister/pprof"
    "github.com/garyburd/twister/server"
    "github.com/garyburd/twister/web"
    "io"
    "job"
    "regexp"
    "render"
)

var doneRegex *regexp.Regexp
var canonicalHost string
var port string

type JSON map[string]interface{}

func init() {
    port = env.GetDefault("PORT", "8080")
    canonicalHost = env.GetDefault("CANONICAL_HOST", fmt.Sprintf("localhost:%s", port))
    doneRegex = regexp.MustCompile("(?i:done|failed|limited|invalid|error|sorry)")
}

func handleBookmarklet(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "application/javascript; charset=utf-8")
    w.Write(bookmarklet.Javascript())
}

func pageHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
    io.WriteString(w, render.Page(req.URLParam["page"], canonicalHost))
}

func chunkHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
    io.WriteString(w, render.Chunk(req.URLParam["chunk"]))
}

func homeHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
    io.WriteString(w, render.Page("index", canonicalHost))
}

func submitHandler(req *web.Request) {
    w := req.Respond(web.StatusOK,
        web.HeaderContentType, "application/json; charset=utf-8",
        "Access-Control-Allow-Origin", "*")
    encoder := json.NewEncoder(w)
    j := job.New(req.Param.Get("email"), req.Param.Get("url"))
    if j.IsValid() {
        j.Progress("Working...")
        extractor.Extract(j)
        encoder.Encode(JSON{
            "message": "Submitted! Hang tight...",
            "id":      j.KeyString(),
        })
    } else {
        encoder.Encode(JSON{
            "message": j.ErrorMessage,
        })
    }
}

func statusHandler(req *web.Request) {
    w := req.Respond(web.StatusOK,
        web.HeaderContentType, "application/json; charset=utf-8",
        "Access-Control-Allow-Origin", "*")
    message := "No job with that ID found."
    done := true
    if v, err := cache.Get(req.URLParam["id"]); err == nil {
        message = v
        done = doneRegex.MatchString(message)
    }
    encoder := json.NewEncoder(w)
    encoder.Encode(JSON{
        "message": message,
        "done":    done,
    })
}

func redirectHandler(req *web.Request) {
    url := req.URL
    url.Host = canonicalHost
    url.Scheme = "http"
    req.Respond(web.StatusMovedPermanently, web.HeaderLocation, url.String())
}

func main() {
    submitRoute := "/ajax/submit.json"
    statusRoute := "/ajax/status/<id:[^.]+>.json"
    router := web.NewRouter().
        Register("/", "GET", homeHandler).
        Register("/static/bookmarklet.js", "GET", handleBookmarklet).
        Register("/static/<path:.*>", "GET", web.DirectoryHandler("static", nil)).
        Register("/<page:(faq|bugs|contact)>", "GET", pageHandler).
        Register("/<chunk:(firefox|safari|chrome|ie|ios|kindle-email)>", "GET", chunkHandler).
        Register(submitRoute, "GET", submitHandler).
        Register(statusRoute, "GET", statusHandler).
        Register("/debug.json", "GET", expvar.ServeWeb).
        Register("/debug/pprof/<:.*>", "*", pprof.ServeWeb)

    redirector := web.NewRouter().
        // These routes get matched in both places so they work everywhere.
        Register(submitRoute, "GET", submitHandler).
        Register(statusRoute, "GET", statusHandler).
        Register("/<splat:>", "GET", redirectHandler)

    hostRouter := web.NewHostRouter(redirector).
        Register(canonicalHost, router)

    server.Run(fmt.Sprintf("0.0.0.0:%s", port), hostRouter)
}
