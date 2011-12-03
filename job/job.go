package job

import (
    "fmt"
    "h5"
    "hashie"
    "net/url"
    "time"
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
    Time   time.Time
    Doc    *h5.Node
    Title  string
    Author string
    Domain string
}

func New(email, uri string) *Job {
    u, _ := url.ParseWithReference(uri)
    key := uuid.NewUUID()
    return &Job{email, u, key, time.Now().UTC(), nil, "", "", ""}
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
