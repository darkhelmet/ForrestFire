package job

import (
    "crypto/sha1"
    "fmt"
    "h5"
    "time"
    "url"
    "uuid"
)

type Job struct {
    Email string
    Url   *url.URL
    Key   uuid.UUID
    Time  *time.Time
    Doc   *h5.Node
    Title string
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

func (j *Job) Root() string {
    return fmt.Sprintf("tmp/%s", j.Hash())
}
