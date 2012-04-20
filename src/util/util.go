package util

import (
    "encoding/json"
    "io"
    "log"
    "os"
)

var logger = log.New(os.Stdout, "[util] ", log.LstdFlags|log.Lmicroseconds)

func Must(err error) {
    if err != nil {
        panic(err)
    }
}

func Pipe(w io.Writer, r io.Reader, expected int64, f func(error)) {
    written, err := io.Copy(w, r)
    if err != nil {
        f(err)
    }
    if expected > 0 && written != expected {
        logger.Printf("written != expected: %d != %d", written, expected)
    }
}

func FileExists(path string) bool {
    stat, _ := os.Stat(path)
    return stat != nil
}

func ParseJSON(r io.Reader, f func(error)) map[string]interface{} {
    decoder := json.NewDecoder(r)
    var payload map[string]interface{}
    if err := decoder.Decode(&payload); err != nil {
        f(err)
    }
    return payload
}
