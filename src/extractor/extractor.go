package extractor

import (
    "blacklist"
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/go-html-transform/h5"
    "github.com/darkhelmet/go-html-transform/html/transform"
    "hashie"
    "io/ioutil"
    "job"
    "kindlegen"
    "log"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "safely"
    "stat"
    "sync"
    "time"
    "util"
)

const (
    Readability     = "https://readability.com/api/content/v1/parser"
    FriendlyMessage = "Sorry, extraction failed."
)

type JSON map[string]interface{}

var (
    timeout   = 5 * time.Second
    deadline  = 10 * time.Second
    token     = env.String("READABILITY_TOKEN")
    notParsed = regexp.MustCompile("(?i:Article Could not be Parsed)")
    logger    = log.New(os.Stdout, "[extractor] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
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
    if resp.StatusCode != 200 {
        body, err := ioutil.ReadAll(resp.Body)
        logger.Panicf("Readability returned wrong error code (%d/%s): %s", resp.StatusCode, err, body)
    }
    return util.ParseJSON(resp.Body, func(err error) {
        logger.Panicf("JSON Parsing Error: %s", err)
    })
}

func rewriteAndDownloadImages(j *job.Job, doc *h5.Node) *h5.Node {
    var wg sync.WaitGroup
    imageDownloader := newDownloader(j.Root(), timeout, time.Now().Add(deadline))
    t := transform.NewTransform(doc)
    fix := transform.TransformAttrib("src", func(uri string) string {
        altered := fmt.Sprintf("%x.jpg", hashie.Sha1([]byte(uri)))
        wg.Add(1)
        go safely.Ignore(func() {
            defer wg.Done()
            imageDownloader.downloadToFile(uri, altered)
            stat.Count(stat.ExtractorImage, 1)
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
    go safely.Do(logger, j, FriendlyMessage, stat.ExtractorUnhandled, func() {
        makeRoot(j)
        data := downloadAndParse(j)
        checkDoc(data, j)
        doc := parseHTML(data["content"].(string))
        j.Doc = rewriteAndDownloadImages(j, doc)
        j.Title = data["title"].(string)
        j.Domain = data["domain"].(string)
        if author := data["author"]; author != nil {
            j.Author = author.(string)
            stat.Count(stat.ExtractorAuthor, 1)
        }
        j.Progress("Extraction complete...")
        kindlegen.Convert(j)
    })
}
