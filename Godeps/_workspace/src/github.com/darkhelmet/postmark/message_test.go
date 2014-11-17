package postmark

import (
    "testing"
)

// From API demo
var jsonMsg string = `{
    "From" : "sender@example.com",
    "To" : "receiver@example.com",
    "Cc" : "copied@example.com",
    "Bcc": "blank-copied@example.com",
    "Subject" : "Test",
    "Tag" : "Invitation",
    "HtmlBody" : "<b>Hello</b>",
    "TextBody" : "Hello",
    "ReplyTo" : "reply@example.com",
    "Headers" : [{ "Name" : "CUSTOM-HEADER", "Value" : "value" }]
}`

var jsonRsp string = `{ 
  "ErrorCode" : 0,
  "Message" : "OK",
  "MessageID" : "b7bc2f4a-e38e-4336-af7d-e6c392c2f817",
  "SubmittedAt" : "2010-11-26T12:01:05.1794748-05:00",
  "To" : "receiver@example.com"
}`

func TestMarshal(t *testing.T) {
    email := &Message{
        From:     "sender@example.com",
        To:       "receiver@example.com",
        Cc:       "copied@example.com",
        Bcc:      "blank-copied@example.com",
        Subject:  "Test",
        Tag:      "Invitation",
        HtmlBody: "<b>Hello</b>",
        TextBody: "Hello",
        ReplyTo:  "reply@example.com",
        Headers:  []Header{Header{Name: "CUSTOM-HEADER", Value: "value"}}}
    msg, err := email.Marshal()
    if err != nil {
        t.Errorf("Can't marshal object to json: %s\n", err)
    }
    println(string(msg))
}

func TestUnmarshal(t *testing.T) {

    email, err := UnmarshalMessage([]byte(jsonMsg))
    if err != nil {
        t.Errorf("Can't unmarshal message: %s\n", err.Error())
    }
    println(email.String())

    rsp, err := UnmarshalResponse([]byte(jsonRsp))
    if err != nil {
        t.Errorf("Can't unmarshal response: %s\n", err.Error())
    }
    println(rsp.String())
}

func TestAttach(t *testing.T) {

    // Load object from template json
    email, err := UnmarshalMessage([]byte(jsonMsg))
    if err != nil {
        t.Errorf("Can't unmarshal mesage: %s\n", err.Error())
    }

    err = email.Attach("testdata/attachment.txt")
    if err != nil {
        t.Errorf("Failed to attach file: %s\n", err.Error())
    }
    println(email.String())

}
