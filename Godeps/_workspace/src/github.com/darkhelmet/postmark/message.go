package postmark

import (
    "bytes"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "mime"
    "os"
    "path"
)

const (
    MaxAttachmentSize = 10485760
)

var (
    BadOrMissingAPIToken        = fmt.Errorf("postmark: Bad or missing API token")
    InvalidEmail                = fmt.Errorf("postmark: Invalid email request")
    SenderSignatureNotFound     = fmt.Errorf("postmark: Sender signature not found")
    SenderSignatureNotConfirmed = fmt.Errorf("postmark: Sender signature not confirmed")
    InvalidJSON                 = fmt.Errorf("postmark: Invalid JSON")
    IncompatibleJSON            = fmt.Errorf("postmark: Incompatible JSON")
    NotAllowedToSend            = fmt.Errorf("postmark: Not allowed to send")
    InactiveRecipient           = fmt.Errorf("postmark: Inactive recipient")
    JSONRequired                = fmt.Errorf("postmark: JSON required")

    ErrorCodes = map[int]error{
        0:   BadOrMissingAPIToken,
        300: InvalidEmail,
        400: SenderSignatureNotFound,
        401: SenderSignatureNotConfirmed,
        402: InvalidJSON,
        403: IncompatibleJSON,
        405: NotAllowedToSend,
        406: InactiveRecipient,
        409: JSONRequired,
    }
)

type Header struct {
    Name  string
    Value string
}

type Attachment struct {
    Name        string
    Content     string // Base 64 encoded string
    ContentType string
}

type Response struct {
    ErrorCode   int
    Message     string
    MessageID   string
    SubmittedAt string //Date
    To          string
}

type BatchResponse []Response

type Message struct {
    From        string
    To          string
    Cc          string
    Bcc         string
    Subject     string
    Tag         string
    HtmlBody    string
    TextBody    string
    ReplyTo     string
    Headers     []Header
    Attachments []Attachment
}

type BatchMessage []Message

func (p *Message) String() string {
    js, e := json.MarshalIndent(p, "", "")
    if e != nil {
        return ""
    }
    return string(js)
}

// Attach file to message (base64 encoded)
func (p *Message) Attach(file string) error {
    finfo, err := os.Stat(file)
    if err != nil {
        return err
    }

    if finfo.Size() > MaxAttachmentSize {
        return fmt.Errorf("File size %d exceeds 10MB limit.", finfo.Size())
    }

    fh, err := os.Open(file)
    if err != nil {
        return err
    }

    var buffer bytes.Buffer
    encoder := base64.NewEncoder(base64.StdEncoding, &buffer)
    io.Copy(encoder, fh)
    encoder.Close()
    cnt, err := ioutil.ReadAll(&buffer)
    if err != nil {
        return err
    }
    fh.Close()

    mimeType := mime.TypeByExtension(path.Ext(file))
    if len(mimeType) == 0 {
        mimeType = "application/octet-stream"
    }

    attachment := Attachment{
        Name:        finfo.Name(),
        Content:     string(cnt),
        ContentType: mimeType,
    }
    p.Attachments = append(p.Attachments, attachment)
    return nil
}

func unmarshal(msg []byte, i interface{}) error {
    return json.Unmarshal(msg, i)
}

func (m *Message) Marshal() ([]byte, error) {
    return json.Marshal(*m)
}

func UnmarshalMessage(msg []byte) (*Message, error) {
    var m Message
    err := unmarshal(msg, &m)
    return &m, err
}

func (r *Response) Marshal() ([]byte, error) {
    return json.Marshal(*r)
}

func (r *Response) Error() error {
    return ErrorCodes[r.ErrorCode]
}

func UnmarshalResponse(rsp []byte) (*Response, error) {
    var r Response
    err := unmarshal(rsp, &r)
    return &r, err
}

func (r *Response) String() string {
    js, err := json.MarshalIndent(r, "", "")
    if err != nil {
        return ""
    }
    return string(js)
}
