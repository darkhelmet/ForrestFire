package postmark

import (
    "blacklist"
    "bytes"
    "cleanup"
    "encoding/json"
    "env"
    "fmt"
    "io/ioutil"
    "job"
    "loggly"
    "net/http"
    "os"
    "util"
)

type Any interface{}

const (
    MaxAttachmentSize = 10485760
    Subject           = "convert"
    Endpoint          = "https://api.postmarkapp.com/email"
    AuthHeader        = "X-Postmark-Server-Token"
)

var from, token string
var client http.Client
var logger *loggly.Logger

func init() {
    from = env.Get("FROM")
    token = env.Get("POSTMARK_TOKEN")
    logger = loggly.NewLogger("postmark", "Sorry, email sending failed.")
}

func readFile(path string) []byte {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        logger.Fail("Failed reading file: %s", err.Error())
    }
    return data
}

func setupHeaders(req *http.Request) {
    req.Header.Add("Accept", "application/json")
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add(AuthHeader, token)
}

func Send(j *job.Job) {
    go logger.SwallowErrorAndNotify(j, func() {
        if stat, err := os.Stat(j.MobiFilePath()); err != nil {
            logger.Fail("Something weird happen. Mobi is missing in postmark.go: %s", err.Error())
        } else {
            if stat.Size() > MaxAttachmentSize {
                blacklist.Blacklist(j.Url.String())
                logger.FailFriendly("Sorry, this article is too big to send!", "URL %s is too big", j.Url.String())
            }
        }

        payload := map[string]Any{
            "From":     from,
            "To":       j.Email,
            "Subject":  Subject,
            "TextBody": fmt.Sprintf("Straight to your Kindle! %s: %s", j.Title, j.Url),
            "Attachments": []Any{
                map[string]Any{
                    "Name":        j.MobiFilename(),
                    "ContentType": "application/octet-stream",
                    "Content":     readFile(j.MobiFilePath()),
                },
            },
        }

        var buffer bytes.Buffer
        json.NewEncoder(&buffer).Encode(payload)

        req, err := http.NewRequest("POST", Endpoint, &buffer)
        if err != nil {
            logger.Fail("Making HTTP Request failed: %s", err.Error())
        }

        setupHeaders(req)
        resp, err := client.Do(req)
        if err != nil {
            logger.Fail("Postmark failed: %s", err.Error())
        }

        defer resp.Body.Close()
        answer := util.ParseJSON(resp.Body, func(err error) {
            logger.Fail("Something bad happened with Postmark: %s", err.Error())
        })

        if answer["ErrorCode"] != nil {
            code := int(answer["ErrorCode"].(float64))
            switch code {
            case 0:
                // All is well
            case 300:
                blacklist.Blacklist(j.Email)
                logger.FailFriendly("Your email appears invalid. Please try carefully remaking the bookmarklet.",
                    "Invalid email given: %s", j.Email)
            default:
                logger.Fail("Unknown error code from Postmark: %d, %s", code, answer)
            }
        }

        j.Progress("All done! Grab your Kindle and hang tight!")
        cleanup.Clean(j)
    })
}
