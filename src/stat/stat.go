package stat

import (
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/stathat/go"
    "log"
    "os"
    "runtime"
)

const (
    Prefix = "[Tinderizer]"

    SubmitOld     = "submit.old"
    SubmitSuccess = "submit.success"
    SubmitError   = "submit.error"
    SubmitBounce  = "submit.bounce"
    SubmitEmail   = "submit.email"

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

    JobDuration = "job.duration"

    OneMillion = 1000000
)

var (
    Count func(string, int)
    Gauge func(string, float64)
)

func init() {
    token := env.StringDefault("STAT_HAT_KEY", "")
    logger := log.New(os.Stdout, "[stat] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))

    if token == "" {
        Count = func(name string, value int) {
            logger.Printf("count: %s: %d", name, value)
        }

        Gauge = func(name string, value float64) {
            logger.Printf("gauge: %s: %f", name, value)
        }
    } else {

        Count = func(name string, value int) {
            logger.Printf("count: %s: %d", name, value)
            stathat.PostEZCount(fmt.Sprintf("%s %s", Prefix, name), token, value)
        }

        Gauge = func(name string, value float64) {
            logger.Printf("gauge: %s: %f", name, value)
            stathat.PostEZValue(fmt.Sprintf("%s %s", Prefix, name), token, value)
        }
    }
}

func Debug() {
    var ms runtime.MemStats
    runtime.ReadMemStats(&ms)
    kilobytes := ms.Alloc / 1024
    megabytes := float64(kilobytes) / 1024
    Gauge(RuntimeMemory, megabytes)
    Gauge(RuntimeGoroutines, float64(runtime.NumGoroutine()))
}
