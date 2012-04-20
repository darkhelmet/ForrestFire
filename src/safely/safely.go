package safely

import (
    "runtime/debug"
    "log"
    "job"
    "cleanup"
)

const (
    DefaultProgress = "Something failed, sorry :("
)

type friendly interface {
    Friendly() string
}

func Ignore(logger *log.Logger, f func()) {
    defer func() {
        if r := recover(); r != nil {
            debug.PrintStack()
        }
    }()
    f()
}

func Do(logger *log.Logger, j *job.Job, progress string, f func()) {
    defer func() {
        if r := recover(); r != nil {
            if err, ok := r.(friendly); ok {
                progress = err.Friendly()
                logger.Printf("%s: %#v", progress, j)
            } else {
                logger.Printf("%v: %#v", r, j)
            }
            debug.PrintStack()
            j.Progress(progress)
            cleanup.Clean(j)
        }
    }()
    f()
}
