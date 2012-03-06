package loggly

import (
    "bytes"
    "cleanup"
    "encoding/json"
    "env"
    "fmt"
    "job"
    "net/http"
    "reflect"
    "time"
)

type Stringer interface {
    String() string
}

type Err struct {
    message  string
    friendly string
}

type Logger struct {
    area     string
    friendly string
}

var messages chan map[string]interface{}

func NewLogger(area, friendly string) *Logger {
    return &Logger{area, friendly}
}

func init() {
    endpoint := env.GetDefault("LOGGLY_URL", "")
    messages = make(chan map[string]interface{}, 10)
    reader := func(fn func(buffer *bytes.Buffer)) {
        for message := range messages {
            buffer := new(bytes.Buffer)
            json.NewEncoder(buffer).Encode(message)
            fn(buffer)
        }
    }
    if endpoint == "" {
        // No endpoint, just write to stdout
        go reader(func(buffer *bytes.Buffer) {
            println(buffer.String())
        })
    } else {
        // We have an endpoint so actually post messages
        go reader(func(buffer *bytes.Buffer) {
            http.Post(endpoint, "application/json", buffer)
        })
    }
}

func send(payload map[string]interface{}) {
    payload["timestamp"] = time.Now().UTC().Format(time.RFC3339)
    messages <- payload
}

func sendMessage(severity, message string) {
    send(map[string]interface{}{
        "severity": severity,
        "message":  message,
    })
}

func (l *Logger) Info(message string) {
    sendMessage("info", message)
}

func (l *Logger) Notice(message string) {
    sendMessage("notice", message)
}

func (l *Logger) Error(message string) {
    sendMessage("error", message)
}

func (l *Logger) Unhandled(message string) {
    sendMessage("unhandled", message)
}

func (l *Logger) JobError(j *job.Job, message string) {
    send(map[string]interface{}{
        "area":     l.area,
        "severity": "error",
        "message":  message,
        "url":      j.Url,
        "email":    j.Email,
    })
}

func (l *Logger) SwallowErrorAndNotify(j *job.Job, f func()) {
    defer func() {
        if r := recover(); r != nil {
            progress := "Something failed, sorry :("
            if err, ok := r.(*Err); ok {
                progress = err.friendly
                l.JobError(j, err.message)
            } else {
                l.handleErrors(r)
            }
            j.Progress(progress)
            cleanup.Clean(j)
        }
    }()
    f()
}

func (l *Logger) SwallowError(f func()) {
    defer func() {
        if r := recover(); r != nil {
            l.handleErrors(r)
        }
    }()
    f()
}

func (l *Logger) NewFriendlyError(message, friendly string) *Err {
    return &Err{message, friendly}
}

func (l *Logger) NewError(message string) *Err {
    return l.NewFriendlyError(message, l.friendly)
}

func (l *Logger) Fail(format string, args ...interface{}) {
    panic(l.NewError(fmt.Sprintf(format, args...)))
}

func (l *Logger) FailFriendly(friendly, format string, args ...interface{}) {
    panic(l.NewFriendlyError(fmt.Sprintf(format, args...), friendly))
}

func (l *Logger) unhandled(message string) {
    l.Unhandled(fmt.Sprintf("Unhandled/run-time panic: %s", message))
}

func (l *Logger) handleErrors(r interface{}) {
    // Handle the error interface
    if err, ok := r.(error); ok {
        l.unhandled(err.Error())
        return
    }

    // Handler the Stringer interface
    if str, ok := r.(Stringer); ok {
        l.unhandled(str.String())
        return
    }

    // Fallback and just use reflect to get a gross string of it
    value := reflect.ValueOf(r)
    l.unhandled(value.String())
}
