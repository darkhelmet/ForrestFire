package job

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"time"

	"github.com/darkhelmet/tinderizer/blacklist"
	"github.com/darkhelmet/tinderizer/hashie"
	"github.com/darkhelmet/tinderizer/user"
	"github.com/nu7hatch/gouuid"
	"golang.org/x/net/html"
)

const DefaultAuthor = "Tinderizer"

var (
	Tmp                 = "tmp"
	BadUrlError         = errors.New("Sorry, but this URL doesn't look like it'll work.")
	BlacklistedUrlError = errors.New("Sorry, but this URL has proven to not work, and has been blacklisted.")
	NoKeyError          = errors.New("No key generated")
	NoDirectoryError    = errors.New("No working directory made")
	ParamsToClean       = []string{"utm_source", "utm_medium", "utm_campaign", "utm_content"}
)

type Job struct {
	Url, Email, Title, Author, Domain, Friendly string
	Key                                         *uuid.UUID
	Doc                                         *html.Node
	StartedAt                                   time.Time
}

func New(email, uri string) (*Job, error) {
	u, err := url.Parse(uri)
	if err != nil {
		blacklist.Blacklist(uri)
		return nil, BadUrlError
	}

	switch u.Scheme {
	case "http", "https":
		// Fine
	default:
		blacklist.Blacklist(uri)
		return nil, BadUrlError
	}

	query := u.Query()
	for _, param := range ParamsToClean {
		query.Del(param)
	}
	u.RawQuery = query.Encode()
	uri = u.String()

	if blacklist.IsBlacklisted(uri) {
		return nil, BlacklistedUrlError
	}

	key, err := uuid.NewV4()
	if err != nil {
		return nil, NoKeyError
	}

	j := &Job{
		Title:     uri,
		Email:     email,
		Url:       uri,
		Key:       key,
		Doc:       nil,
		Author:    DefaultAuthor,
		StartedAt: time.Now(),
	}

	err = os.MkdirAll(j.Root(), 0755)
	if err != nil {
		return nil, NoDirectoryError
	}

	return j, nil
}

func (j *Job) filename(extension string) string {
	return fmt.Sprintf("Tinderizer.%s", extension)
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
	var buffer bytes.Buffer
	html.Render(&buffer, j.Doc)
	return template.HTML(buffer.String())
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

func (j *Job) Now() string {
	return j.StartedAt.Format(time.RFC822)
}
