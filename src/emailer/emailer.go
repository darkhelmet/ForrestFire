package emailer

import (
    "blacklist"
    "counter"
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/postmark"
    J "job"
    "log"
    "looper"
    "net/http"
    "os"
    "retry"
    "stat"
    "time"
)

type Any interface{}

const (
    MaxAttachmentSize = 10485760
    Subject           = "convert"
    FriendlyMessage   = "Sorry, email sending failed."
)

var (
    from   = env.String("FROM")
    token  = env.String("POSTMARK_TOKEN")
    Pm     = postmark.New(token)
    logger = log.New(os.Stdout, "[emailer] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
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
        e.error(job, FriendlyMessage, "failed attaching file: %s", err)
        return
    }

    resp, err := Pm.Send(m)
    if resp == nil {
        e.error(job, FriendlyMessage, "failed sending email: %s", err)
        return
    }

    switch resp.ErrorCode {
    case 0:
        // All is well
        key := time.Now().Format("2006:01")
        retry.Times(3, func() error {
            return counter.Inc(key, 1)
        })
        stat.Count(stat.PostmarkSuccess, 1)
    case 422:
        e.error(job, FriendlyMessage, "failed sending email: %s: %s", err, resp.Message)
        stat.Count(stat.PostmarkError, 1)
        return
    case 300:
        e.error(job, "Your email appears invalid. Please try carefully remaking the bookmarklet.", "emailer: Email inactive or invalid")
        stat.Count(stat.PostmarkInvalidEmail, 1)
        return
    case 406:
        e.error(job, "Your email appears to have bounced. Amazon likes to bounce emails sometimes, and my provider 'deactivates' the email. For now, try changing your Personal Documents Email. I'm trying to find a proper solution for this :(", "emailer: Email inactive or invalid")
        stat.Count(stat.PostmarkDeactivated, 1)
        return
    default:
        e.error(job, FriendlyMessage, "Something bizarre happened with Postmark: %s", resp.Message)
        return
    }

    job.Progress("All done! Grab your Kindle and hang tight!")
    looper.MapUrl(resp.MessageID, job.Url)
    e.Output <- job
}
