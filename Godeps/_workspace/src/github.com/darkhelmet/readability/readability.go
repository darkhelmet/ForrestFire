package readability

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
)

var (
    ErrTransient = errors.New("readability: transient error, probably a 5xx, maybe try again")
    Parser       = "https://www.readability.com/api/content/v1/parser"
)

type Response struct {
    Domain        string  `json:"domain"`
    Author        *string `json:"author"`
    URL           string  `json:"url"`
    ShortURL      string  `json:"short_url"`
    Title         string  `json:"title"`
    TotalPages    int     `json:"total_pages"`
    WordCount     int     `json:"word_count"`
    Content       string  `json:"content"`
    DatePublished *Time   `json:"date_published"`
    NextPageId    *string `json:"next_page_id"`
    RenderedPages int     `json:"rendered_pages"`
}

type Endpoint struct {
    token  string
    logger *log.Logger
}

func New(token string, logger *log.Logger) *Endpoint {
    return &Endpoint{token, logger}
}

func parseResponse(uri string, r io.Reader) (*Response, error) {
    var rresp Response
    decoder := json.NewDecoder(r)
    err := decoder.Decode(&rresp)
    if err != nil {
        return nil, fmt.Errorf("readability: JSON error (%s): %s", uri, err)
    }
    return &rresp, nil
}

func (e *Endpoint) Extract(uri string) (*Response, error) {
    resp, err := http.PostForm(Parser, url.Values{"token": {e.token}, "url": {uri}})
    if err != nil {
        return nil, fmt.Errorf("readability: HTTP error (%s): %s", uri, err)
    }
    defer resp.Body.Close()
    return e.handleResponse(uri, resp)
}

func (e *Endpoint) ExtractWithContent(uri, content string) (*Response, error) {
    resp, err := http.PostForm(Parser, url.Values{"token": {e.token}, "url": {uri}, "content": {content}})
    if err != nil {
        return nil, fmt.Errorf("readability: HTTP error (%s): %s", uri, err)
    }
    defer resp.Body.Close()
    return e.handleResponse(uri, resp)
}

func (e *Endpoint) handleResponse(uri string, resp *http.Response) (*Response, error) {
    switch {
    case resp.StatusCode == 504:
        e.dumpResponse(resp)
        return nil, fmt.Errorf("readability: Failed to fetch %s", uri)
    case resp.StatusCode >= 500:
        e.dumpResponse(resp)
        return nil, ErrTransient
    case resp.StatusCode == 200:
        return parseResponse(uri, resp.Body)
    default:
        e.dumpResponse(resp)
        return nil, fmt.Errorf("readability: HTTP error (%s): %d", uri, resp.StatusCode)
    }
}

func (e *Endpoint) dumpResponse(resp *http.Response) {
    if e.logger == nil {
        return
    }

    dump, err := httputil.DumpResponse(resp, true)
    if err != nil {
        e.logger.Printf("readability: failed dumping response: %s", err)
    } else {
        e.logger.Printf("%s", dump)
    }
}
