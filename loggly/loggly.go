package loggly

import (
    "env"
    "fmt"
    "http"
    "strings"
    "time"
    "user"
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
    messages <- fmt.Sprintf("*** %s *** - %s - %s", level, time.UTC().Format("%Y-%m-%dT%H:%M:%S%Z"), message)
}

func Notice(message string) {
    send("NOTICE", message)
}

func Error(message string) {
    send("ERROR", message)
    fmt.Println("Error:", message)
}

func SwallowErrorAndNotify(key fmt.Stringer, f func()) {
    defer func() {
        if r := recover(); r != nil {
            err := r.(*Err)
            user.Notify(err.friendly, key.String())
            Error(err.message)
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
