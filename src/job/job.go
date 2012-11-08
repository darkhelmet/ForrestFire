package job

import (
    "blacklist"
    "errors"
    "fmt"
    "github.com/darkhelmet/go-html-transform/h5"
    "hashie"
    "html/template"
    "net/url"
    "os"
    "strings"
    "time"
    "user"
    "vendor/github.com/nu7hatch/gouuid"
)

const (
    DefaultAuthor = "Tinderizer"
    Tmp           = "tmp"
)

type Job struct {
    Url, Email, Title, Author, Domain, Friendly, Content string
    Key                                                  *uuid.UUID
    Doc                                                  *h5.Node
    urlError                                             error
}

func New(email, uri, content string) *Job {
    _, err := url.Parse(uri)
    key, _ := uuid.NewV4()
    return &Job{
        Content:  content,
        Title:    uri,
        Email:    email,
        Url:      uri,
        Key:      key,
        Doc:      nil,
        Author:   DefaultAuthor,
        urlError: err,
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
    return hashie.Sha1([]byte(j.Url), j.Key[:])
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
    if j.urlError != nil {
        blacklist.Blacklist(j.Url)
        return errors.New("Sorry, but this URL doesn't look like it'll work.")
    }

    // URL is already blacklisted
    if blacklist.IsBlacklisted(j.Url) {
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

func (j *Job) Now() string {
    return time.Now().Format(time.RFC822)
}
