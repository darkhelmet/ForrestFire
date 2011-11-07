package job

import (
    "crypto/sha1"
    "fmt"
    "h5"
    "time"
    "url"
    "user"
    "uuid"
)

var tmp string

func init() {
    tmp = "tmp"
}

type Job struct {
    Email  string
    Url    *url.URL
    Key    uuid.UUID
    Time   *time.Time
    Doc    *h5.Node
    Title  string
    Author string
    Domain string
}

func New(email, uri string) *Job {
    u, _ := url.ParseWithReference(uri)
    key := uuid.NewUUID()
    return &Job{email, u, key, time.UTC(), nil, "", "", ""}
}

func (j *Job) Hash() string {
    hash := sha1.New()
    hash.Write([]byte(j.Url.String()))
    hash.Write([]byte(j.Time.String()))
    return fmt.Sprintf("%x", hash.Sum())
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

func (j *Job) HTMLFilename() string {
    return fmt.Sprintf("%s.html", j.Title)
}

func (j *Job) MobiFilename() string {
    return fmt.Sprintf("%s.mobi", j.Title)
}

func (j *Job) HTMLFilePath() string {
    return fmt.Sprintf("%s/%s", j.Root(), j.HTMLFilename())
}

func (j *Job) MobiFilePath() string {
    return fmt.Sprintf("%s/%s", j.Root(), j.MobiFilename())
}
