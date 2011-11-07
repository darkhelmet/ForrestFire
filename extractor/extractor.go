package extractor

import (
    "crypto/sha1"
    "env"
    "fmt"
    "h5"
    "html/transform"
    "http"
    "io"
    "job"
    "json"
    "kindlegen"
    "loggly"
    "os"
    "strings"
    "sync"
    "url"
)

const DefaultAuthor = "Tinderizer"
const Readability = "https://readability.com/api/content/v1/parser"
const Friendly = "Sorry, extraction failed."

type JSON map[string]interface{}

var token string

func init() {
    token = env.Get("READABILITY_TOKEN")
}

func buildReadabilityUrl(u string) string {
    return fmt.Sprintf("%s?url=%s&token=%s", Readability, url.QueryEscape(u), url.QueryEscape(token))
}

func downloadAndParse(j *job.Job) JSON {
    resp, err := http.Get(buildReadabilityUrl(j.Url.String()))
    defer resp.Body.Close()
    if err != nil {
        panic(loggly.NewError(
            fmt.Sprintf("Readability Error: %s", err.Error()),
            Friendly))
    }
    decoder := json.NewDecoder(resp.Body)
    var payload JSON
    err = decoder.Decode(&payload)
    if err != nil {
        panic(loggly.NewError(
            fmt.Sprintf("JSON Parsing Error: %s", err.Error()),
            Friendly))
    }
    return payload
}

func getFileExtension(uri string) string {
    url, err := url.Parse(uri)
    if err != nil {
        // Just pretend it's JPG
        return ".jpg"
    }
    path := url.Path
    dot := strings.LastIndex(path, ".")
    if dot == -1 {
        return ".jpg"
    }
    return path[dot:]
}

func openFile(path string) *os.File {
    file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(fmt.Sprintf("Failed opening file: %s", err.Error()))
    }
    return file
}

func getImage(url string) *http.Response {
    resp, err := http.Get(url)
    if err != nil {
        panic(fmt.Sprintf("Failed download image %s: %s", url, err.Error()))
    }
    return resp
}

func pipe(resp *http.Response, file *os.File) {
    written, err := io.Copy(file, resp.Body)
    if err != nil {
        panic(fmt.Sprintf("Error with io.Copy: %s", err.Error()))
    }
    if written != resp.ContentLength {
        loggly.Notice(fmt.Sprintf("written != resp.ContentLength: %d != %d", written, resp.ContentLength))
    }
}

func downloadToFile(url, name string) {
    resp := getImage(url)
    defer resp.Body.Close()
    file := openFile(name)
    defer file.Close()
    pipe(resp, file)
}

func rewriteAndDownloadImages(j *job.Job, doc *h5.Node) *h5.Node {
    var wg sync.WaitGroup
    root := j.Root()
    hash := sha1.New()
    t := transform.NewTransform(doc)
    fix := transform.TransformAttrib("src", func(uri string) string {
        hash.Reset()
        hash.Write([]byte(uri))
        altered := fmt.Sprintf("%x%s", hash.Sum(), getFileExtension(uri))
        wg.Add(1)
        go loggly.SwallowError(func() {
            defer wg.Done()
            downloadToFile(uri, fmt.Sprintf("%s/%s", root, altered))
        })
        return altered
    })
    t.Apply(fix, "img")
    wg.Wait()
    return t.Doc()
}

func parseHTML(content string) *h5.Node {
    doc, err := transform.NewDoc(content)
    if err != nil {
        panic(loggly.NewError(
            fmt.Sprintf("HTML Parsing Error: %s", err.Error()),
            Friendly))
    }
    return doc
}

func makeRoot(j *job.Job) {
    if err := os.MkdirAll(j.Root(), 0755); err != nil {
        panic(loggly.NewError(
            fmt.Sprintf("Failed to make working directory: %s", err.Error()),
            Friendly))
    }
}

func Extract(j *job.Job) {
    if j.Url == nil {
        j.Progress("This URL appears invalid. Sorry :(")
        return
    }

    go loggly.SwallowErrorAndNotify(j, func() {
        makeRoot(j)
        data := downloadAndParse(j)
        doc := parseHTML(data["content"].(string))
        j.Doc = rewriteAndDownloadImages(j, doc)
        j.Title = data["title"].(string)
        j.Domain = data["domain"].(string)
        author := data["author"]
        if author == nil {
            j.Author = DefaultAuthor
        } else {
            j.Author = author.(string)
        }
        j.Progress("Extraction complete...")
        kindlegen.Convert(j)
    })
}
