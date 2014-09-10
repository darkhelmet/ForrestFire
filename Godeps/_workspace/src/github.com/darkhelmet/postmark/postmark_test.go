package postmark

import (
    "os"
    "testing"
)

// Probably should figure out a way to run these tests without 
// burning credits..
func setup(t *testing.T) map[string]string {

    cfg := make(map[string]string)
    cfg["key"] = os.Getenv("POSTMARK_API_KEY")
    cfg["to"] = os.Getenv("POSTMARK_TO")
    cfg["from"] = os.Getenv("POSTMARK_FROM")
    if len(cfg["key"]) <= 0 {
        t.Fatal("Need to setup POSTMARK_API_KEY")
    }
    return cfg
}

func TestSend(t *testing.T) {

    cfg := setup(t)
    println(cfg["key"])
    p := NewPostmark(cfg["key"])
    if p == nil {
        t.Fatal("NewPostmark (check api key)")
    }
    r, e := p.Send(&Message{
        From:     cfg["from"],
        To:       cfg["to"],
        Subject:  "TestSend",
        TextBody: "TestSend Body",
    })

    if e != nil {
        t.Fatal("Send failed! ", e.Error())
    }

    if r == nil || r.ErrorCode != 0 {
        t.Fatal("Response fail. ", r.String())
    }
}

func TestAttachment(t *testing.T) {
    cfg := setup(t)

    msg := &Message{
        From:     cfg["from"],
        To:       cfg["to"],
        Subject:  "TestAttach",
        TextBody: "Test attach body",
    }

    err := msg.Attach("testdata/attachment.txt")
    println(msg.String())
    if err != nil {
        t.Fatal("Attach failed. ", err.Error())
    }
    p := NewPostmark(cfg["key"])
    if p == nil {
        t.Fatal("NewPostmark (check api key)")
    }

    rsp, err := p.Send(msg)
    if err != nil {
        t.Fatal("Send failed.", err.Error())
    }

    if rsp.ErrorCode != 0 {
        t.Fatal("Response fail. ", rsp.String())
    }
}
