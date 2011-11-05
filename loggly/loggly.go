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
    message string
    friendly string
    key fmt.Stringer
}

func NewError(message, friendly string, key fmt.Stringer) (* Err) {
    return &Err{message, friendly, key}
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
}

func SwallowError(f func()) {
    defer func() {
       if r := recover(); r != nil {
           err := r.(* Err)
           user.Notify(err.friendly, err.key.String())
           Error(err.message)
           fmt.Println("Error:", err.message)
       }
    }()
    f()
}
