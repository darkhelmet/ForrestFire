package extractor

import (
    "env"
    "fmt"
    "http"
    "json"
    "loggly"
    "url"
    "user"
    "uuid"
)

const Readability = "https://readability.com/api/content/v1/parser"

type JSON map[string]interface{}

type Job struct {
    Email string
    Url   *url.URL
    Key   uuid.UUID
}

var token string

func init() {
    token = env.Get("READABILITY_TOKEN")
}

func buildReadabilityUrl(u string) string {
    return fmt.Sprintf("%s?url=%s&token=%s", Readability, url.QueryEscape(u), url.QueryEscape(token))
}

func downloadAndParse(job * Job) JSON {
    println("Downloading")
    resp, err := http.Get(buildReadabilityUrl(job.Url.String()))
    if err != nil {
        panic(loggly.NewError(
            fmt.Sprintf("Readability Error: %s", err.Error()),
            "Sorry, extraction failed.",
            job.Key))
    }
    decoder := json.NewDecoder(resp.Body)
    var payload JSON
    err = decoder.Decode(&payload)
    resp.Body.Close()
    if err != nil {
        panic(loggly.NewError(
            fmt.Sprintf("JSON Parsing Error: %s", err.Error()),
            "Sorry, extraction failed.",
            job.Key))
    }
    return payload
}

func Extract(job *Job) {
    if job.Url == nil {
        user.Notify(job.Key.String(), "This URL appears invalid. Sorry :(")
        return
    }

    go loggly.SwallowError(func() {
        data := downloadAndParse(job)
        fmt.Println(data)
        user.Notify(job.Key.String(), "Done")
    })
}
