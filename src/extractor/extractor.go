package extractor

import (
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/go-html-transform/h5"
    "github.com/darkhelmet/go-html-transform/html/transform"
    "github.com/darkhelmet/readability"
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
    Input  <-chan J.Job
    Output chan<- J.Job
    Error  chan<- J.Job
}

func New(input <-chan J.Job, output chan<- J.Job, error chan<- J.Job) *Extractor {
    return &Extractor{
        Input:  input,
        Output: output,
        Error:  error,
    }
}

func (e *Extractor) error(job J.Job, format string, args ...interface{}) {
    logger.Printf(format, args...)
    job.Friendly = FriendlyMessage
    e.Error <- job
}

func (e *Extractor) Run() {
    for job := range e.Input {
        go e.Process(job)
    }
}

func extract(url, content string) (*readability.Response, error) {
    if content == "" {
        return rdb.Extract(url)
    }
    return rdb.ExtractWithContent(url, content)
}

func (e *Extractor) Process(job J.Job) {
    resp, err := extract(job.Url, job.Content)
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
    if resp.Author != nil {
        job.Author = *resp.Author
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
                logger.Printf("downloading image failed: %s", err)
            }
            stat.Count(stat.ExtractorImage, 1)
        }()
        return altered
    })
    t.Apply(fix, "img")
    wg.Wait()
    return t.Doc()
}
