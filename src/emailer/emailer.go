package emailer

import (
    "blacklist"
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/postmark"
    J "job"
    "log"
    "net/http"
    "os"
    "stat"
)

type Any interface{}

const (
    MaxAttachmentSize = 10485760
    Subject           = "convert"
    Endpoint          = "https://api.postmarkapp.com/email"
    AuthHeader        = "X-Postmark-Server-Token"
    FriendlyMessage   = "Sorry, email sending failed."
)

var (
    from   = env.String("FROM")
    token  = env.String("POSTMARK_TOKEN")
    pm     = postmark.New(token)
    logger = log.New(os.Stdout, "[postmark] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
    client http.Client
)

type Emailer struct {
    Input  <-chan J.Job
    Output chan<- J.Job
    Error  chan<- J.Job
}

func New(input <-chan J.Job, output chan<- J.Job, error chan<- J.Job) *Emailer {
    return &Emailer{
        Input:  input,
        Output: output,
        Error:  error,
    }
}

func (e *Emailer) error(job J.Job, friendly, format string, args ...interface{}) {
    logger.Printf(format, args...)
    job.Friendly = friendly
    e.Error <- job
}

func (e *Emailer) Run() {
    for job := range e.Input {
        go e.Process(job)
    }
}

func (e *Emailer) Process(job J.Job) {
    if st, err := os.Stat(job.MobiFilePath()); err != nil {
        e.error(job, FriendlyMessage, "Something weird happened. Mobi is missing: %s", err)
        return
    } else {
        if st.Size() > MaxAttachmentSize {
            stat.Count(stat.PostmarkTooBig, 1)
            blacklist.Blacklist(job.Url)
            e.error(job, "Sorry, this article is too big to send!", "Attachment was too big (%d bytes)", st.Size())
            return
        }
    }

    m := &postmark.Message{
        From:     from,
        To:       job.Email,
        Subject:  Subject,
        TextBody: fmt.Sprintf("Straight to your Kindle! %s: %s", job.Title, job.Url),
    }

    if err := m.Attach(job.MobiFilePath()); err != nil {
        e.error(job, FriendlyMessage, "Failed attaching file: %s", err)
        return
    }

    resp, err := pm.Send(m)
    if err != nil {
        e.error(job, FriendlyMessage, "Failed sending emailer: %s", err)
        return
    }

    switch resp.ErrorCode {
    case 0:
        // All is well
    case 300, 406:
        blacklist.Blacklist(job.Email)
        stat.Count(stat.PostmarkBlacklist, 1)
        e.error(job, "Your email appears invalid or is inactive. Please try carefully remaking the bookmarklet.", "emailer: Email inactive or invalid")
        return
    default:
        e.error(job, FriendlyMessage, "Something bizarre happened with Postmark: %s", resp.Message)
        return
    }

    job.Progress("All done! Grab your Kindle and hang tight!")
    e.Output <- job
}
