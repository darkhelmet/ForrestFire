package stat

import (
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/stathatgo"
    "runtime"
)

const (
    Prefix            = "[Tinderizer]"
    SubmitSuccess     = "submit.success"
    SubmitBlacklist   = "submit.blacklist"
    HttpRedirect      = "http.redirect"
    RuntimeGoroutines = "runtime.goroutines"
    RuntimeMemory     = "runtime.memory"
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
    var ms runtime.MemStats
    runtime.ReadMemStats(&ms)
    // Yes, this could overflow the float,
    // but this app is fairly low on usage, so it's fine.
    Value(RuntimeMemory, float64(ms.Alloc))
    Value(RuntimeGoroutines, float64(runtime.NumGoroutine()))
}
