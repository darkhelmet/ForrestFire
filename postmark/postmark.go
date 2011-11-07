package postmark

// TODO: Handle email invalid
// TODO: Handle large attachments
// TODO: Look through FogBugz and async.rb to see what I'm catching
// TODO: Cleanup

import (
    "env"
    "fmt"
    "http"
    "io"
    "io/ioutil"
    "job"
    "json"
    "loggly"
)

type Any interface{}

const Friendly = "Sorry, email sending failed."
const Subject = "convert"
const Endpoint = "https://api.postmarkapp.com/email"
const AuthHeader = "X-Postmark-Server-Token"

var from, token string
var client http.Client

func init() {
    from = env.Get("FROM")
    token = env.Get("POSTMARK_TOKEN")
}

func fail(format string, args ...interface{}) {
    panic(loggly.NewError(fmt.Sprintf(format, args...), Friendly))
}

func readFile(path string) []byte {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        fail("Failed reading file: %s", err.Error())
    }
    return data
}

func Send(j *job.Job) {
    go loggly.SwallowErrorAndNotify(j, func() {
        // TODO: Optimize this by using a JSON Encoder or something
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

        reader, writer := io.Pipe()
        encoder := json.NewEncoder(writer)
        go func() {
            encoder.Encode(payload)
            defer writer.Close()
        }()
        req, err := http.NewRequest("POST", Endpoint, reader)
        if err != nil {
            fail("Making HTTP Request failed: %s", err.Error())
        }
        req.Header.Add("Accept", "application/json")
        req.Header.Add("Content-Type", "application/json")
        req.Header.Add(AuthHeader, token)

        resp, err := client.Do(req)
        defer resp.Body.Close()
        if err != nil {
            fail("HTTP POST failed: %s", err.Error())
        }

        j.Progress("All done! Grab your Kindle and hang tight!")
    })
}
