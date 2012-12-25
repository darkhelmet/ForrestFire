package stat

import (
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/go-librato"
    "runtime"
)

const (
    SubmitOld     = "submit.old"
    SubmitSuccess = "submit.success"
    SubmitError   = "submit.error"
    SubmitBounce  = "submit.bounce"

    HttpRedirect = "http.redirect"

    RuntimeGoroutines = "runtime.goroutines"
    RuntimeMemory     = "runtime.memory"
    RuntimeBoot       = "runtime.boot"

    PostmarkBounce       = "postmark.bounce"
    PostmarkSuccess      = "postmark.success"
    PostmarkTooBig       = "postmark.toobig"
    PostmarkError        = "postmark.error"
    PostmarkInvalidEmail = "postmark.invalidemail"
    PostmarkDeactivated  = "postmark.deactivated"

    ExtractorAuthor     = "extractor.author"
    ExtractorImage      = "extractor.image"
    ExtractorError      = "extractor.error"
    ExtractorImageError = "extractor.imageerror"

    KindlegenError = "kindlegen.error"

    OneMillion = 1000000
)

var (
    Count func(string, int64)
    Gauge func(string, int64)
)

func init() {
    user := env.StringDefault("LIBRATO_USER", "")
    token := env.StringDefault("LIBRATO_TOKEN", "")
    source := env.StringDefault("LIBRATO_SOURCE", "")

    if user == "" || token == "" || source == "" {
        Count = func(name string, value int64) {}
        Gauge = func(name string, value int64) {}
    } else {
        m := librato.NewSimpleMetrics(user, token, source)

        Count = func(name string, value int64) {
            m.GetCounter(name) <- value
        }

        Gauge = func(name string, value int64) {
            m.GetGauge(name) <- value
        }
    }
}

func Debug() {
    var ms runtime.MemStats
    runtime.ReadMemStats(&ms)
    Gauge(RuntimeMemory, int64(ms.Alloc))
    Gauge(RuntimeGoroutines, int64(runtime.NumGoroutine()))
}
