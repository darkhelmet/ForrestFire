package extractor

import (
    "blacklist"
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/go-html-transform/h5"
    "github.com/darkhelmet/go-html-transform/html/transform"
    "hashie"
    "job"
    "kindlegen"
    "log"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "safely"
    "sync"
    "util"
)

const (
    Readability     = "https://readability.com/api/content/v1/parser"
    FriendlyMessage = "Sorry, extraction failed."
)

type JSON map[string]interface{}

var (
    token     = env.String("READABILITY_TOKEN")
    notParsed = regexp.MustCompile("(?i:Article Could not be Parsed)")
    logger    = log.New(os.Stdout, "[extractor] ", log.LstdFlags|log.Lmicroseconds)
)

func buildReadabilityUrl(u string) string {
    return fmt.Sprintf("%s?url=%s&token=%s", Readability, url.QueryEscape(u), url.QueryEscape(token))
}

func downloadAndParse(j *job.Job) JSON {
    resp, err := http.Get(buildReadabilityUrl(j.Url.String()))
    if err != nil {
        logger.Panicf("Readability Error: %s", err)
    }
    defer resp.Body.Close()
    return util.ParseJSON(resp.Body, func(err error) {
        logger.Panicf("JSON Parsing Error: %s", err)
    })
}

func getImage(url string) *http.Response {
    resp, err := http.Get(url)
    if err != nil {
        log.Panicf("Failed download image %s: %s", url, err)
    }
    return resp
}

func downloadToFile(url, name string) {
    resp := getImage(url)
    defer resp.Body.Close()
    file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        logger.Panicf("Failed opening file: %s", err)
    }
    defer file.Close()
    util.Pipe(file, resp.Body, resp.ContentLength, func(err error) {
        logger.Panicf("Error with io.Copy: %s", err)
    })
}

func rewriteAndDownloadImages(j *job.Job, doc *h5.Node) *h5.Node {
    var wg sync.WaitGroup
    root := j.Root()
    t := transform.NewTransform(doc)
    fix := transform.TransformAttrib("src", func(uri string) string {
        altered := fmt.Sprintf("%x.jpg", hashie.Sha1([]byte(uri)))
        wg.Add(1)
        go safely.Ignore(func() {
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
        logger.Panicf("HTML Parsing Error: %s", err)
    }
    return doc
}

func makeRoot(j *job.Job) {
    if err := os.MkdirAll(j.Root(), 0755); err != nil {
        logger.Panicf("Failed to make working directory: %s", err)
    }
}

func checkDoc(data JSON, j *job.Job) {
    if data["error"] != nil && data["error"].(bool) {
        blacklist.Blacklist(j.Url.String())
        logger.Panicf("Readability failed: %s", data["messages"].(string))
    }

    if notParsed.MatchString(data["title"].(string)) {
        blacklist.Blacklist(j.Url.String())
        logger.Panicf("Readability failed, article could not be parsed.")
    }
}

func Extract(j *job.Job) {
    go safely.Do(logger, j, FriendlyMessage, func() {
        makeRoot(j)
        data := downloadAndParse(j)
        checkDoc(data, j)
        doc := parseHTML(data["content"].(string))
        j.Doc = rewriteAndDownloadImages(j, doc)
        j.Title = data["title"].(string)
        j.Domain = data["domain"].(string)
        if author := data["author"]; author != nil {
            j.Author = author.(string)
        }
        j.Progress("Extraction complete...")
        kindlegen.Convert(j)
    })
}
