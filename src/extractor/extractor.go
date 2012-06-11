package extractor

import (
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/go-html-transform/h5"
    "github.com/darkhelmet/go-html-transform/html/transform"
    "github.com/darkhelmet/readability"
    "github.com/trustmaster/goflow"
    "hashie"
    J "job"
    "log"
    "os"
    "stat"
    "sync"
    "time"
)

const (
    FriendlyMessage = "Sorry, extraction failed."
)

type JSON map[string]interface{}

var (
    timeout  = 5 * time.Second
    deadline = 10 * time.Second
    token    = env.String("READABILITY_TOKEN")
    rdb      = readability.New(token)
    logger   = log.New(os.Stdout, "[extractor] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
)

type Extractor struct {
    flow.Component
    Input  <-chan J.Job
    Output chan<- J.Job
    Error  chan<- J.Job
}

func New() *Extractor {
    return new(Extractor)
}

func (e *Extractor) error(job J.Job, format string, args ...interface{}) {
    logger.Printf(format, args...)
    job.Friendly = FriendlyMessage
    e.Error <- job
}

func (e *Extractor) OnInput(job J.Job) {
    resp, err := rdb.Extract(job.Url.String())
    if err != nil {
        e.error(job, "readability failed: %s", err)
        return
    }

    doc, err := transform.NewDoc(resp.Content)
    if err != nil {
        e.error(job, "HTML parsing failed: %s", err)
        return
    }

    job.Doc = rewriteAndDownloadImages(job.Root(), doc)
    job.Title = resp.Title
    job.Domain = resp.Domain
    if resp.Author != "" {
        job.Author = resp.Author
        stat.Count(stat.ExtractorAuthor, 1)
    }

    job.Progress("Extraction complete...")
    e.Output <- job
}

func rewriteAndDownloadImages(root string, doc *h5.Node) *h5.Node {
    var wg sync.WaitGroup
    imageDownloader := newDownloader(root, timeout, time.Now().Add(deadline))
    t := transform.NewTransform(doc)
    fix := transform.TransformAttrib("src", func(uri string) string {
        altered := fmt.Sprintf("%x.jpg", hashie.Sha1([]byte(uri)))
        wg.Add(1)
        go func() {
            defer wg.Done()
            if err := imageDownloader.downloadToFile(uri, altered); err != nil {
                logger.Printf("Downloading image failed: %s", err)
            }
            stat.Count(stat.ExtractorImage, 1)
        }()
        return altered
    })
    t.Apply(fix, "img")
    wg.Wait()
    return t.Doc()
}
