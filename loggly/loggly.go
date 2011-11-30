package loggly

import (
    "cleanup"
    "env"
    "fmt"
    "job"
    "net/http"
    "reflect"
    "strings"
    "time"
)

type Stringer interface {
    String() string
}

type Err struct {
    message  string
    friendly string
}

var messages chan string

func NewError(message, friendly string) *Err {
    return &Err{message, friendly}
}

func init() {
    endpoint := env.Get("LOGGLY_URL")
    messages = make(chan string, 10)
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

func unhandled(message string) {
    Error(fmt.Sprintf("Unhandled/run-time panic: %s", message))
}

func formatError(j *job.Job, message string) string {
    return fmt.Sprintf("%s {url=%s, email=%s}", message, j.Url, j.Email)
}

func handleErrors(r interface{}) {
    // Handle the error interface
    if err, ok := r.(error); ok {
        unhandled(err.Error())
        return
    }

    // Handler the Stringer interface
    if str, ok := r.(Stringer); ok {
        unhandled(str.String())
        return
    }

    // Fallback and just use reflect to get a string of it
    value := reflect.ValueOf(r)
    unhandled(value.String())
}

func SwallowErrorAndNotify(j *job.Job, f func()) {
    defer func() {
        if r := recover(); r != nil {
            if err, ok := r.(*Err); ok {
                j.Progress(err.friendly)
                Error(formatError(j, err.message))
                cleanup.Clean(j)
            } else {
                handleErrors(r)
            }
        }
    }()
    f()
}

func SwallowError(f func()) {
    defer func() {
        if r := recover(); r != nil {
            handleErrors(r)
        }
    }()
    f()
}
