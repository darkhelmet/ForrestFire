package job

import (
    "blacklist"
    "errors"
    "fmt"
    "github.com/darkhelmet/go-html-transform/h5"
    "github.com/nu7hatch/gouuid"
    "hashie"
    "html/template"
    "net/url"
    "os"
    "strings"
    "time"
    "user"
)

const (
    DefaultAuthor = "Tinderizer"
    Tmp           = "tmp"
)

type Job struct {
    Email    string
    Url      *url.URL
    Key      *uuid.UUID
    Doc      *h5.Node
    Title    string
    Author   string
    Domain   string
    Friendly string
}

func New(email, uri string) *Job {
    u, _ := url.Parse(uri)
    key, _ := uuid.NewV4()
    return &Job{
        Email:  email,
        Url:    u,
        Key:    key,
        Doc:    nil,
        Author: DefaultAuthor,
    }
}

func (j *Job) filename(extension string) string {
    safeName := strings.Replace(j.Title, string(os.PathSeparator), "", -1)
    return fmt.Sprintf("%s.%s", safeName, extension)
}

func (j *Job) GoString() string {
    return fmt.Sprintf("Job[Email: %s, URL: %s, Key: %s", j.Email, j.Url, j.Key)
}

func (j *Job) Hash() string {
    return hashie.Sha1([]byte(j.Url.String()), j.Key[:])
}

func (j *Job) Progress(message string) {
    user.Notify(j.Key.String(), message)
}

func (j *Job) Root() string {
    return fmt.Sprintf("%s/%s", Tmp, j.Hash())
}

func (j *Job) HTML() template.HTML {
    return template.HTML(j.Doc.String())
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

func (j *Job) Validate() error {
    // URL failed to parse
    if j.Url == nil {
        blacklist.Blacklist(j.Url.String())
        return errors.New("Sorry, but this URL doesn't look like it'll work.")
    }

    // URL is already blacklisted
    if blacklist.IsBlacklisted(j.Url.String()) {
        return errors.New("Sorry, but this URL has proven to not work, and has been blacklisted.")
    }

    // Email is blacklisted
    if blacklist.IsBlacklisted(j.Email) {
        return errors.New("Sorry, but this email has proven to not work. You might want to try carefully remaking your bookmarklet.")
    }

    if j.Key == nil {
        return errors.New("Submission failed, no key generated")
    }

    if err := os.MkdirAll(j.Root(), 0755); err != nil {
        return errors.New("Submission failed, no working directory made")
    }

    return nil
}

func (j *Job) Now() time.Time {
    return time.Now()
}
