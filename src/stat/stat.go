package stat

import (
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/stathatgo"
    "runtime"
)

const (
    Prefix             = "[Tinderizer]"
    SubmitSuccess      = "submit.success"
    EmailSuccess       = "email.success"
    SubmitBlacklist    = "submit.blacklist"
    EmailBlacklist     = "email.blacklist"
    HttpRedirect       = "http.redirect"
    RuntimeGoroutines  = "runtime.goroutines"
    RuntimeMemory      = "runtime.memory"
    RuntimeBoot        = "runtime.boot"
    KindlegenUnhandled = "kindlegen.unhandled"
    PostmarkTooBig     = "postmark.too-big"
    PostmarkInactive   = "postmark.inactive"
    PostmarkBlacklist  = "postmark.blacklist"
    PostmarkUnhandled  = "postmark.unhandled"
    ExtractorAuthor    = "extractor.author"
    ExtractorImage     = "extractor.image"
    ExtractorUnhandled = "extractor.unhandled"
    OneMillion         = 1000000
)

var (
    key   = env.StringDefault("STAT_HAT_KEY", "")
    Count func(string, int)
    Value func(string, float64)
)

func init() {
    if key == "" {
        Count = func(name string, value int) {}
        Value = func(name string, value float64) {}
    } else {
        Count = count
        Value = value
    }
}

func count(name string, value int) {
    stathat.PostEZCount(fmt.Sprintf("%s %s", Prefix, name), key, value)
}

func value(name string, value float64) {
    stathat.PostEZValue(fmt.Sprintf("%s %s", Prefix, name), key, value)
}

func Debug() {
    Value(RuntimeMemory, allocInBaseTen())
    Value(RuntimeGoroutines, float64(runtime.NumGoroutine()))
}

// Read amount of memory, but converted to base ten
// so when stathat does math it's actually accurate.
// They show 3.91M or something to mean 3.91 million,
// but doing this math, it will end up being 3.whatever,
// megabytes
func allocInBaseTen() float64 {
    var ms runtime.MemStats
    runtime.ReadMemStats(&ms)
    // Yes, this could overflow the float,
    // but this app is fairly low on usage, so it's fine.
    alloc := float64(ms.Alloc)
    return alloc / 1024 / 1024 * OneMillion
}
