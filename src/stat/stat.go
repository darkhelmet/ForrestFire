package stat

import (
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/go-librato"
    "log"
    "os"
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
    logger := log.New(os.Stdout, "[stat] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))

    if user == "" || token == "" || source == "" {
        Count = func(name string, value int64) {
            logger.Printf("count: %d", value)
        }

        Gauge = func(name string, value int64) {
            logger.Printf("gauge: %d", value)
        }
    } else {

        m := librato.NewSimpleMetrics(user, token, source)

        Count = func(name string, value int64) {
            logger.Printf("count: %d", value)
            m.GetCounter(name) <- value
        }

        Gauge = func(name string, value int64) {
            logger.Printf("gauge: %d", value)
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
