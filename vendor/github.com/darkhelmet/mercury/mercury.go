package mercury

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	Parser = "https://mercury.postlight.com/parser"
)

type Response struct {
	Title         string  `json:"title"`
	Content       string  `json:"content"`
	URL           string  `json:"url"`
	Domain        string  `json:"domain"`
	WordCount     int     `json:"word_count"`
	TotalPages    int     `json:"total_pages"`
	RenderedPages int     `json:"rendered_pages"`
	NextPageUrl   *string `json:"next_page_url"`
}

type Endpoint struct {
	apiKey string
	logger *log.Logger
}

func New(apiKey string, logger *log.Logger) *Endpoint {
	return &Endpoint{apiKey, logger}
}

func parseResponse(uri string, r io.Reader) (*Response, error) {
	var rresp Response
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&rresp)
	if err != nil {
		return nil, fmt.Errorf("mercury: JSON error (%s): %s", uri, err)
	}
	return &rresp, nil
}

func (e *Endpoint) Extract(uri string) (*Response, error) {
	query := url.Values{"url": {uri}}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", Parser, query.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("mercury: failed creating request: %s", err)
	}
	req.Header.Add("x-api-key", e.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mercury: HTTP error (%s): %s", uri, err)
	}
	defer resp.Body.Close()
	return e.handleResponse(uri, resp)
}

func (e *Endpoint) handleResponse(uri string, resp *http.Response) (*Response, error) {
	switch {
	case resp.StatusCode == 200:
		return parseResponse(uri, resp.Body)
	default:
		e.dumpResponse(resp)
		return nil, fmt.Errorf("mercury: HTTP error (%s): %d", uri, resp.StatusCode)
	}
}

func (e *Endpoint) dumpResponse(resp *http.Response) {
	if e.logger == nil {
		return
	}

	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		e.logger.Printf("mercury: failed dumping response: %s", err)
	} else {
		e.logger.Printf("%s", dump)
	}
}
