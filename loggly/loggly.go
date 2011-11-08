package loggly

import (
    "cleanup"
    "env"
    "fmt"
    "http"
    "job"
    "strings"
    "time"
)

var messages chan string

type Err struct {
    message  string
    friendly string
}

func NewError(message, friendly string) *Err {
    return &Err{message, friendly}
}

func init() {
    endpoint := env.Get("LOGGLY_URL")
    messages = make(chan string, 25)
    go func() {
        for message := range messages {
            http.Post(endpoint, "text/plain", strings.NewReader(message))
        }
    }()
}

func send(level, message string) {
    messages <- fmt.Sprintf("*** %s *** - %s - %s", level, time.UTC().Format(time.RFC3339), message)
}

func Notice(message string) {
    send("NOTICE", message)
}

func Error(message string) {
    send("ERROR", message)
    fmt.Println("Error:", message)
}

func formatError(j *job.Job, message string) string {
    return fmt.Sprintf("%s {url=%s, email=%s}", message, j.Url, j.Email)
}

func SwallowErrorAndNotify(j *job.Job, f func()) {
    defer func() {
        if r := recover(); r != nil {
            err := r.(*Err)
            j.Progress(err.friendly)
            Error(formatError(j, err.message))
            cleanup.Clean(j)
        }
    }()
    f()
}

func SwallowError(f func()) {
    defer func() {
        if r := recover(); r != nil {
            Error(r.(string))
        }
    }()
    f()
}
