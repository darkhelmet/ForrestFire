package extractor

import (
    "boots"
    "code.google.com/p/go.net/html"
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/readability"
    "hashie"
    J "job"
    "log"
    "os"
    "stat"
    "strings"
    "sync"
    "time"
)

const (
    FriendlyMessage = "Sorry, extraction failed."
    RetryTimes      = 3
    RetryPause      = 3 * time.Second
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
    stat.Count(stat.ExtractorError, 1)
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

func extractRetry(url, content string) (resp *readability.Response, err error) {
    for i := 0; i < RetryTimes; i++ {
        resp, err = extract(url, content)
        switch err {
        case readability.ErrTransient:
            // Hmm, let's try that again. Sleep and let the loop repeat
            logger.Println("readability transient error, retrying")
            time.Sleep(RetryPause)
        default:
            // Either it worked, or we don't want to retry
            return
        }
    }
    // If we get here, just return what we had last
    return
}

func (e *Extractor) Process(job J.Job) {
    resp, err := extractRetry(job.Url, job.Content)
    if err != nil {
        e.error(job, "readability failed: %s", err)
        return
    }

    doc, err := rewriteAndDownloadImages(job.Root(), resp.Content)
    if err != nil {
        e.error(job, "HTML parsing failed: %s", err)
        return
    }

    job.Doc = doc
    if resp.Title != "" {
        job.Title = resp.Title
    }
    job.Domain = resp.Domain
    if resp.Author != nil {
        job.Author = *resp.Author
        stat.Count(stat.ExtractorAuthor, 1)
    }

    job.Progress("Extraction complete...")
    e.Output <- job
}

func rewriteAndDownloadImages(root string, content string) (*html.Node, error) {
    var wg sync.WaitGroup
    imageDownloader := newDownloader(root, timeout, time.Now().Add(deadline))
    doc, err := boots.Walk(strings.NewReader(content), "img", func(node *html.Node) {
        for index, attr := range node.Attr {
            if attr.Key == "src" {
                uri := attr.Val
                altered := fmt.Sprintf("%x.jpg", hashie.Sha1([]byte(uri)))
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    if err := imageDownloader.downloadToFile(uri, altered); err != nil {
                        logger.Printf("downloading image failed: %s", err)
                        stat.Count(stat.ExtractorImageError, 1)
                    } else {
                        stat.Count(stat.ExtractorImage, 1)
                    }
                }()
                node.Attr[index].Val = altered
                break
            }
        }
    })
    wg.Wait()
    return doc, err
}
