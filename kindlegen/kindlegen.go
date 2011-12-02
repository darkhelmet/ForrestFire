package kindlegen

import (
    "fmt"
    "job"
    "loggly"
    "os"
    "os/exec"
    "path/filepath"
    "postmark"
    "runtime"
    "template"
    "util"
)

const Friendly = "Sorry, conversion failed."

var kindlegen string

func init() {
    kindlegen, _ = filepath.Abs(fmt.Sprintf("vendor/kindlegen-%s", runtime.GOOS))
}

func fail(format string, args ...interface{}) {
    panic(loggly.NewError(fmt.Sprintf(format, args...), Friendly))
}

func openFile(path string) *os.File {
    file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fail("Failed opening file: %s", err.Error())
    }
    return file
}

func writeHTML(j *job.Job) {
    tmpl := `
    <html>
        <head>
            <meta content="text/html, charset=utf-8" http-equiv="Content-Type" />
            <meta content="{{.Author}} ({{.Domain}})" name="author" />
            <title>{{.Title}}</title>
        </head>
        <body>
            <h1>{{.Title}}</h1>
            {{.HTML}}
            <hr />
            <p>Originally from <a href="{{.Url}}">{{.Url}}</a></p>
            <p>Sent with <a href="http://Tinderizer.com/">Tinderizer</a></p>
        </body>
    </html>
    `
    html := template.RenderToString(j.Title, tmpl, j)
    file := openFile(j.HTMLFilePath())
    defer file.Close()
    if _, err := file.WriteString(html); err != nil {
        fail("Failed writing HTML file: %s", err.Error())
    }
}

func Convert(j *job.Job) {
    go loggly.SwallowErrorAndNotify(j, func() {
        writeHTML(j)
        cmd := exec.Command(kindlegen, []string{j.HTMLFilename()}...)
        cmd.Dir = j.Root()
        out, err := cmd.CombinedOutput()
        if !util.FileExists(j.MobiFilePath()) {
            fail("Failed running kindlegen: %s {output=%s}", err.Error(), string(out))
        }
        j.Progress("Conversion complete...")
        postmark.Send(j)
    })
}
