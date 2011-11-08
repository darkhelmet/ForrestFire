package util

import (
    "fmt"
    "io"
    "json"
    "loggly"
    "os"
    "strings"
    "url"
)

type ErrorFunc func(error)

func GetUrlFileExtension(uri, def string) string {
    url, err := url.Parse(uri)
    if err != nil {
        return def
    }
    path := url.Path
    dot := strings.LastIndex(path, ".")
    if dot == -1 {
        return def
    }
    return path[dot:]
}

func Pipe(w io.Writer, r io.Reader, expected int64, f ErrorFunc) {
    written, err := io.Copy(w, r)
    if err != nil {
        f(err)
    }
    if written != expected {
        loggly.Notice(fmt.Sprintf("written != expected: %d != %d", written, expected))
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
