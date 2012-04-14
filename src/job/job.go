package job

import (
    "blacklist"
    "fmt"
    "github.com/darkhelmet/go-html-transform/h5"
    "hashie"
    "net/url"
    "os"
    "strings"
    "time"
    "user"
    "uuid"
)

const DefaultAuthor = "Tinderizer"

var tmp string

func init() {
    tmp = "tmp"
}

type Job struct {
    Email        string
    Url          *url.URL
    Key          uuid.UUID
    Time         time.Time
    Doc          *h5.Node
    Title        string
    Author       string
    Domain       string
    ErrorMessage string
}

func New(email, uri string) *Job {
    u, _ := url.Parse(uri)
    key := uuid.NewUUID()
    return &Job{email, u, key, time.Now().UTC(), nil, "", DefaultAuthor, "", ""}
}

func (j *Job) Hash() string {
    return hashie.Sha1([]byte(j.Url.String()), []byte(j.Time.String()))
}

func (j *Job) KeyString() string {
    return j.Key.String()
}

func (j *Job) Progress(message string) {
    user.Notify(j.KeyString(), message)
}

func (j *Job) Root() string {
    return fmt.Sprintf("%s/%s", tmp, j.Hash())
}

func (j *Job) HTML() string {
    return j.Doc.String()
}

func (j *Job) filename(extension string) string {
    safeName := strings.Replace(j.Title, string(os.PathSeparator), "-", -1)
    return fmt.Sprintf("%s.%s", safeName, extension)
}

func (j *Job) HTMLFilename() string {
    return j.filename("html")
}

func (j *Job) MobiFilename() string {
    return j.filename("mobi")
}

func (j *Job) HTMLFilePath() string {
    return fmt.Sprintf("%s/%s", j.Root(), j.HTMLFilename())
}

func (j *Job) MobiFilePath() string {
    return fmt.Sprintf("%s/%s", j.Root(), j.MobiFilename())
}

func (j *Job) IsValid() bool {
    // URL failed to parse
    if j.Url == nil {
        blacklist.Blacklist(j.Url.String())
        j.ErrorMessage = "Sorry, but this URL doesn't look like it'll work."
        return false
    }

    // URL is already blacklisted
    if blacklist.IsBlacklisted(j.Url.String()) {
        j.ErrorMessage = "Sorry, but this URL has proven to not work, and has been blacklisted."
        return false
    }

    // Email is blacklisted
    if blacklist.IsBlacklisted(j.Email) {
        j.ErrorMessage = "Sorry, but this email has proven to not work. You might want to try carefully remaking your bookmarklet."
        return false
    }

    return true
}
