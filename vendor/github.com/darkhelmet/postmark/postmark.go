package postmark

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
)

const (
    Endpoint   = "https://api.postmarkapp.com"
    AuthHeader = "X-Postmark-Server-Token"
)

var (
    MissingOrIncorrectAPIKey = fmt.Errorf("postmark: Missing or incorrect API key header")
    InvalidRequest           = fmt.Errorf("postmark: Unprocessable Entity")
    ServerError              = fmt.Errorf("postmark: Server error")
    client                   http.Client
)

type Postmark struct {
    key string
}

func New(apikey string) *Postmark {
    return &Postmark{key: apikey}
}

func (p *Postmark) Send(m *Message) (*Response, error) {
    data, err := m.Marshal()
    if err != nil {
        return nil, err
    }
    postData := bytes.NewBuffer(data)
    resp, err := p.request("POST", Endpoint+"/email", postData)
    if err != nil {
        return nil, err
    }

    switch {
    case resp.StatusCode == 401:
        return nil, MissingOrIncorrectAPIKey
    case resp.StatusCode == 500:
        return nil, ServerError
    }

    var body bytes.Buffer
    _, err = io.Copy(&body, resp.Body)
    resp.Body.Close()
    if err != nil {
        return nil, err
    }

    prsp, err := UnmarshalResponse([]byte(body.String()))
    if err != nil {
        return nil, err
    }

    if resp.StatusCode == 422 {
        return prsp, InvalidRequest
    }

    return prsp, nil
}

func (p *Postmark) Reactivate(b Bounce) error {
    if b.CanActivate {
        resp, err := p.request("PUT", fmt.Sprintf("%s/bounces/%d/activate", Endpoint, b.ID), nil)
        if err != nil {
            return err
        }
        switch {
        case resp.StatusCode == 401:
            return MissingOrIncorrectAPIKey
        case resp.StatusCode == 500:
            return ServerError
        }
        return nil
    }
    return nil
}

func (p *Postmark) request(method, urlStr string, body io.Reader) (*http.Response, error) {
    req, err := http.NewRequest(method, urlStr, body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set(AuthHeader, p.key)
    return client.Do(req)
}
