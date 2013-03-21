package main

import (
    "bookmarklet"
    "bytes"
    "cache"
    "cleaner"
    "counter"
    "emailer"
    "encoding/hex"
    "encoding/json"
    "errors"
    "extractor"
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/postmark"
    "github.com/darkhelmet/webutil"
    "github.com/garyburd/twister/adapter"
    "github.com/garyburd/twister/expvar"
    "github.com/garyburd/twister/pprof"
    "github.com/garyburd/twister/web"
    "html/template"
    "io"
    J "job"
    "kindlegen"
    "log"
    "looper"
    "net/http"
    "os"
    "regexp"
    "stat"
    "strings"
    "time"
)

const (
    HeaderAccessControlAllowOrigin = "Access-Control-Allow-Origin"
    QueueSize                      = 10
)

var (
    doneRegex     = regexp.MustCompile("(?i:done|failed|limited|invalid|error|sorry)")
    port          = env.IntDefault("PORT", 8080)
    canonicalHost = env.StringDefaultF("CANONICAL_HOST", func() string { return fmt.Sprintf("localhost:%d", port) })
    logger        = log.New(os.Stdout, "[server] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
    templates     = template.Must(template.ParseGlob("views/*.tmpl"))
    newJobs       chan J.Job
)

type JSON map[string]interface{}

func init() {
    stat.Count(stat.RuntimeBoot, 1)
    newJobs = RunApp()
}

func RunApp() chan J.Job {
    input := make(chan J.Job, QueueSize)
    conversion := make(chan J.Job, QueueSize)
    emailing := make(chan J.Job, QueueSize)
    cleaning := make(chan J.Job, QueueSize)

    go extractor.New(input, conversion, cleaning).Run()
    go kindlegen.New(conversion, emailing, cleaning).Run()
    go emailer.New(emailing, cleaning, cleaning).Run()
    go cleaner.New(cleaning).Run()

    return input
}

func renderPage(w io.Writer, page, host string) error {
    key := time.Now().Format("2006:01")
    count, err := counter.Get(key)
    if err != nil {
        logger.Printf("failed getting count: %s", err)
        count, _ = counter.Get(key)
    }

    var buffer bytes.Buffer
    if err := templates.ExecuteTemplate(&buffer, page, nil); err != nil {
        return err
    }
    return templates.ExecuteTemplate(w, "layout.tmpl", JSON{
        "host":  host,
        "yield": template.HTML(buffer.String()),
        "count": count,
    })
}

func handleBookmarklet(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "application/javascript; charset=utf-8")
    w.Write(bookmarklet.Javascript())
}

func pageHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
    tmpl := fmt.Sprintf("%s.tmpl", req.URLParam["page"])
    if err := renderPage(w, tmpl, canonicalHost); err != nil {
        logger.Printf("failed rendering page: %s", err)
    }
}

func chunkHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
    tmpl := fmt.Sprintf("%s.tmpl", req.URLParam["chunk"])
    if err := templates.ExecuteTemplate(w, tmpl, nil); err != nil {
        logger.Printf("failed rendering chunk: %s", err)
    }
}

func homeHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
    if err := renderPage(w, "index.tmpl", canonicalHost); err != nil {
        logger.Printf("failed rendering index: %s", err)
    }
}

type EmailHeader struct {
    Name, Value string
}

type EmailToFull struct {
    Email, Name string
}

type InboundEmail struct {
    From, To, CC, ReplyTo, Subject string
    ToFull                         []EmailToFull
    MessageId, Date, MailboxHash   string
    TextBody, HtmlBody             string
    Tag                            string
    Headers                        []EmailHeader
}

func extractParts(e *InboundEmail) (email string, url string, err error) {
    parts := strings.Split(e.ToFull[0].Email, "@")
    if len(parts) == 0 {
        return "", "", errors.New("failed splitting email on '@'")
    }
    emailBytes, err := hex.DecodeString(parts[0])
    if err != nil {
        return "", "", fmt.Errorf("failed decoding email from hex: %s", err)
    }
    email = string(emailBytes)
    buffer := bytes.NewBufferString(strings.TrimSpace(e.TextBody))
    url, err = buffer.ReadString('\n')
    if len(url) == 0 && err != nil {
        return "", "", fmt.Errorf("failed reading line from email body: %s", err)
    }
    err = nil
    url = strings.TrimSpace(url)
    return
}

func inboundHandler(req *web.Request) {
    decoder := json.NewDecoder(req.Body)
    var inbound InboundEmail
    err := decoder.Decode(&inbound)
    if err != nil {
        logger.Printf("failed decoding inbound email: %s", err)
    } else {
        email, url, err := extractParts(&inbound)
        if err != nil {
            logger.Printf("failed extracting needed parts from email: %s", err)
        } else {
            logger.Printf("email submission of %#v to %#v", url, email)
            if job, err := J.New(email, url, ""); err == nil {
                newJobs <- *job
                stat.Count(stat.SubmitEmail, 1)
            }
        }
    }
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/plain; charset=utf-8")
    io.WriteString(w, "ok")
}

func bounceHandler(req *web.Request) {
    decoder := json.NewDecoder(req.Body)
    var bounce postmark.Bounce
    err := decoder.Decode(&bounce)
    if err != nil {
        logger.Printf("failed decoding bounce: %s", err)
        return
    }

    if looper.AlreadyResent(bounce.MessageID, bounce.Email) {
        logger.Printf("skipping resend of message ID %s", bounce.MessageID)
    } else {
        err = emailer.Pm.Reactivate(bounce)
        if err != nil {
            logger.Printf("failed reactivating bounce: %s", err)
            return
        }
        uri := looper.MarkResent(bounce.MessageID, bounce.Email)
        if job, err := J.New(bounce.Email, uri, ""); err != nil {
            logger.Printf("bounced email failed to validate as a job: %s", err)
        } else {
            newJobs <- *job
            logger.Printf("resending %#v to %#v after bounce", uri, bounce.Email)
            stat.Count(stat.PostmarkBounce, 1)
        }
    }
    w := req.Respond(web.StatusOK, web.HeaderContentType, "text/plain; charset=utf-8")
    io.WriteString(w, "ok")
}

type Submission struct {
    Url     string `json:"url"`
    Email   string `json:"email"`
    Content string `json:"content"`
}

func submitHandler(req *web.Request) {
    decoder := json.NewDecoder(req.Body)
    var submission Submission
    err := decoder.Decode(&submission)
    if err != nil {
        logger.Printf("failed decoding submission: %s", err)
    }
    logger.Printf("submission of %#v to %#v", submission.Url, submission.Email)

    w := req.Respond(web.StatusOK, web.HeaderContentType, "application/json; charset=utf-8")
    encoder := json.NewEncoder(w)
    submit(encoder, submission.Email, submission.Url, submission.Content)
}

func oldSubmitHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "application/json; charset=utf-8")
    encoder := json.NewEncoder(w)

    submit(encoder, req.Param.Get("email"), req.Param.Get("url"), "")
    stat.Count(stat.SubmitOld, 1)
}

func submitError(encoder *json.Encoder, err error) {
    stat.Count(stat.SubmitError, 1)
    encoder.Encode(JSON{"message": err.Error()})
}

func submit(encoder *json.Encoder, email, url, content string) {
    stat.Debug()
    job, err := J.New(email, url, content)
    if err != nil {
        submitError(encoder, err)
        return
    }

    job.Progress("Working...")
    newJobs <- *job
    encoder.Encode(JSON{
        "message": "Submitted! Hang tight...",
        "id":      job.Key.String(),
    })
    stat.Count(stat.SubmitSuccess, 1)
}

func statusHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "application/json; charset=utf-8")
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

func main() {
    submitRoute := "/ajax/submit.json"
    statusRoute := "/ajax/status/<id:[^.]+>.json"
    router := web.NewRouter().
        Register("/", "GET", homeHandler).
        Register("/inbound", "POST", inboundHandler).
        Register("/bounce", "POST", bounceHandler).
        Register("/static/bookmarklet.js", "GET", handleBookmarklet).
        Register("/<page:(faq|bugs|contact)>", "GET", pageHandler).
        Register("/<chunk:(firefox|safari|chrome|ie|ios|kindle-email)>", "GET", chunkHandler).
        Register(submitRoute, "POST", submitHandler, "GET", oldSubmitHandler).
        Register(statusRoute, "GET", statusHandler).
        Register("/debug.json", "GET", expvar.ServeWeb).
        Register("/debug/pprof/<:.*>", "*", pprof.ServeWeb).
        Register("/<path:.*>", "GET", web.DirectoryHandler("public", nil))

    redirector := web.NewRouter().
        // These routes get matched in both places so they work everywhere.
        Register(submitRoute, "POST", submitHandler, "GET", oldSubmitHandler).
        Register(statusRoute, "GET", statusHandler).
        Register("/<splat:>", "GET", redirectHandler)

    hostRouter := web.NewHostRouter(redirector).
        Register(canonicalHost, router)

    var handler http.Handler = adapter.HTTPHandler{hostRouter}
    handler = webutil.AlwaysHeaderHandler{handler, http.Header{HeaderAccessControlAllowOrigin: {"*"}}}
    handler = webutil.GzipHandler{handler}
    handler = webutil.LoggerHandler{handler, logger}
    handler = webutil.EnsureRequestBodyClosedHandler{handler}

    http.Handle("/", handler)

    logger.Printf("Tinderizer is starting on 0.0.0.0:%d", port)
    err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
    if err != nil {
        logger.Fatalf("failed to serve: %s", err)
    }
}
