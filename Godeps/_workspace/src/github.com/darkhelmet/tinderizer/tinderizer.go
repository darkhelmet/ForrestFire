package tinderizer

import (
    "github.com/darkhelmet/postmark"
    "github.com/darkhelmet/readability"
    "github.com/darkhelmet/tinderizer/cache"
    "github.com/darkhelmet/tinderizer/cleaner"
    "github.com/darkhelmet/tinderizer/emailer"
    "github.com/darkhelmet/tinderizer/extractor"
    J "github.com/darkhelmet/tinderizer/job"
    "github.com/darkhelmet/tinderizer/kindlegen"
    "log"
    "sync"
)

type App struct {
    postmark    *postmark.Postmark
    readability *readability.Endpoint
    kindlegen   string
    from        string
    input       chan J.Job
    wg          sync.WaitGroup
}

func (a *App) Run(size int) {
    a.input = make(chan J.Job, size)
    conversion := make(chan J.Job, size)
    emailing := make(chan J.Job, size)
    cleaning := make(chan J.Job, size)

    a.wg.Add(4)
    go extractor.New(a.readability, a.input, conversion, cleaning).Run(&a.wg)
    go kindlegen.New(a.kindlegen, conversion, emailing, cleaning).Run(&a.wg)
    go emailer.New(a.postmark, a.from, emailing, cleaning, cleaning).Run(&a.wg)
    go cleaner.New(cleaning).Run(&a.wg)
}

func (a *App) Shutdown() {
    close(a.input)
    a.wg.Wait()
}

func (a *App) Queue(job J.Job) {
    a.input <- job
}

func (a *App) Status(id string) (string, error) {
    return cache.Get(id)
}

func (a *App) Reactivate(b postmark.Bounce) error {
    return a.postmark.Reactivate(b)
}

func New(readabilityToken, postmarkToken, fromEmailAddress string, kindlegenBinary string, logger *log.Logger) *App {
    return &App{
        kindlegen:   kindlegenBinary,
        postmark:    postmark.New(postmarkToken),
        readability: readability.New(readabilityToken, nil),
        from:        fromEmailAddress,
    }
}
