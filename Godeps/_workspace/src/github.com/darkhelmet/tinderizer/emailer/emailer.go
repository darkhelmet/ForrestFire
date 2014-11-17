package emailer

import (
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/postmark"
    "github.com/darkhelmet/tinderizer/blacklist"
    "github.com/darkhelmet/tinderizer/cache"
    J "github.com/darkhelmet/tinderizer/job"
    "log"
    "net/http"
    "os"
    "sync"
    "time"
)

type Any interface{}

const (
    OneHour           = 60 * 60
    MaxAttachmentSize = 10485760
    Subject           = "convert"
    FriendlyMessage   = "Sorry, email sending failed."
)

var (
    logger = log.New(os.Stdout, "[emailer] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
    client http.Client
)

type Emailer struct {
    postmark *postmark.Postmark
    from     string
    wg       sync.WaitGroup
    Input    <-chan J.Job
    Output   chan<- J.Job
    Error    chan<- J.Job
}

func New(pm *postmark.Postmark, from string, input <-chan J.Job, output chan<- J.Job, error chan<- J.Job) *Emailer {
    return &Emailer{
        postmark: pm,
        from:     from,
        Input:    input,
        Output:   output,
        Error:    error,
    }
}

func (e *Emailer) error(job J.Job, friendly, format string, args ...interface{}) {
    logger.Printf(format, args...)
    job.Friendly = friendly
    e.Error <- job
}

func (e *Emailer) Run(wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range e.Input {
        e.wg.Add(1)
        go e.Process(job)
    }
    e.wg.Wait()
    close(e.Output)
}

func (e *Emailer) Process(job J.Job) {
    job.Progress("Sending to your Kindle...")

    defer e.wg.Done()
    if st, err := os.Stat(job.MobiFilePath()); err != nil {
        e.error(job, FriendlyMessage, "Something weird happened. Mobi is missing: %s", err)
        return
    } else {
        if st.Size() > MaxAttachmentSize {
            blacklist.Blacklist(job.Url)
            e.error(job, "Sorry, this article is too big to send!", "Attachment was too big (%d bytes)", st.Size())
            return
        }
    }

    m := &postmark.Message{
        From:     e.from,
        To:       job.Email,
        Subject:  Subject,
        TextBody: fmt.Sprintf("Straight to your Kindle! %s: %s", job.Title, job.Url),
    }

    if err := m.Attach(job.MobiFilePath()); err != nil {
        e.error(job, FriendlyMessage, "failed attaching file: %s", err)
        return
    }

    resp, err := e.postmark.Send(m)
    if resp == nil {
        e.error(job, FriendlyMessage, "failed sending email: %s", err)
        return
    }

    switch resp.ErrorCode {
    case 0:
        // All is well
    case 422:
        e.error(job, FriendlyMessage, "failed sending email: %s: %s", err, resp.Message)
        return
    case 300:
        e.error(job, "Your email appears invalid. Please try carefully remaking the bookmarklet.", "emailer: Email inactive or invalid")
        return
    case 406:
        e.error(job, "Your email appears to have bounced. Amazon likes to bounce emails sometimes, and my provider 'deactivates' the email. For now, try changing your Personal Documents Email. I'm trying to find a proper solution for this :(", "emailer: Email inactive or invalid")
        return
    default:
        e.error(job, FriendlyMessage, "Something bizarre happened with Postmark: %s", resp.Message)
        return
    }

    job.Progress("All done! Grab your Kindle and hang tight!")
    cache.Set(resp.MessageID, job.Url, OneHour)
    recordDurationStat(job)
    e.Output <- job
}

func recordDurationStat(job J.Job) {
    finishedAt := time.Now()
    duration := finishedAt.Sub(job.StartedAt)
    logger.Printf("job=%s duration=%s", job.Key, duration)
}
