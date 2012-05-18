package main

import (
    "bookmarklet"
    "bytes"
    "cache"
    "encoding/json"
    "extractor"
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/garyburd/twister/expvar"
    "github.com/garyburd/twister/pprof"
    "github.com/garyburd/twister/server"
    "github.com/garyburd/twister/web"
    "html/template"
    "io"
    "job"
    "log"
    "net"
    "os"
    "regexp"
    "stat"
)

type JSON map[string]interface{}

var (
    doneRegex     = regexp.MustCompile("(?i:done|failed|limited|invalid|error|sorry)")
    port          = env.IntDefault("PORT", 8080)
    canonicalHost = env.StringDefaultF("CANONICAL_HOST", func() string { return fmt.Sprintf("localhost:%d", port) })
    logger        = log.New(os.Stdout, "[server] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
    templates     = template.Must(template.ParseGlob("views/*.tmpl"))
)

func init() {
    stat.Count(stat.RuntimeBoot, 1)
}

func renderPage(w io.Writer, page, host string) error {
    buffer := new(bytes.Buffer)
    if err := templates.ExecuteTemplate(buffer, page, nil); err != nil {
        return err
    }
    return templates.ExecuteTemplate(w, "layout.tmpl", JSON{
        "host":  host,
        "yield": template.HTML(buffer.String()),
    })
}

func handleBookmarklet(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "application/javascript; charset=utf-8")
    w.Write(bookmarklet.Javascript())
}

func pageHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
    if err := renderPage(w, fmt.Sprintf("%s.tmpl", req.URLParam["page"]), canonicalHost); err != nil {
        logger.Printf("Failed rendering page: %s", err)
    }
}

func chunkHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
    if err := templates.ExecuteTemplate(w, fmt.Sprintf("%s.tmpl", req.URLParam["chunk"]), nil); err != nil {
        logger.Printf("Failed rendering chunk: %s", err)
    }
}

func homeHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
    if err := renderPage(w, "index.tmpl", canonicalHost); err != nil {
        logger.Printf("Failed rendering index: %s", err)
    }
}

func submitHandler(req *web.Request) {
    w := req.Respond(web.StatusOK,
        web.HeaderContentType, "application/json; charset=utf-8",
        "Access-Control-Allow-Origin", "*")
    encoder := json.NewEncoder(w)
    j := job.New(req.Param.Get("email"), req.Param.Get("url"))
    if err, ok := j.IsValid(); ok {
        stat.Count(stat.SubmitSuccess, 1)
        j.Progress("Working...")
        extractor.Extract(j)
        encoder.Encode(JSON{
            "message": "Submitted! Hang tight...",
            "id":      j.KeyString(),
        })
    } else {
        stat.Count(stat.SubmitBlacklist, 1)
        encoder.Encode(JSON{
            "message": err,
        })
    }
    stat.Debug()
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
    stat.Count(stat.HttpRedirect, 1)
    url := req.URL
    url.Host = canonicalHost
    url.Scheme = "http"
    req.Respond(web.StatusMovedPermanently, web.HeaderLocation, url.String())
}

func ShortLogger(lr *server.LogRecord) {
    if lr.Error != nil {
        logger.Printf("%d %s %s %s\n", lr.Status, lr.Request.Method, lr.Request.URL, lr.Error)
    } else {
        logger.Printf("%d %s %s\n", lr.Status, lr.Request.Method, lr.Request.URL)
    }
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

    listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
    if err != nil {
        logger.Fatalf("Failed to listen: %s", err)
    }
    defer listener.Close()
    server := &server.Server{
        Listener: listener,
        Handler:  hostRouter,
        Logger:   server.LoggerFunc(ShortLogger),
    }
    err = server.Serve()
    if err != nil {
        logger.Fatalf("Failed to server: %s", err)
    }
}
