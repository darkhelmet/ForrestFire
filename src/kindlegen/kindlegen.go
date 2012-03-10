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

var kindlegen string
var logger *loggly.Logger
var tmpl = `
<html>
    <head>
        <meta content="text/html, charset=utf-8" http-equiv="Content-Type" />
        <meta content="{{.Author | html}} ({{.Domain | html}})" name="author" />
        <title>{{.Title | html}}</title>
    </head>
    <body>
        <h1>{{.Title | html}}</h1>
        {{.HTML}}
        <hr />
        <p>Originally from <a href="{{.Url}}">{{.Url}}</a></p>
        <p>Sent with <a href="http://Tinderizer.com/">Tinderizer</a></p>
    </body>
</html>
`

func init() {
    var err error
    kindlegen, err = filepath.Abs(fmt.Sprintf("vendor/kindlegen-%s", runtime.GOOS))
    if err != nil {
        panic(err)
    }
    logger = loggly.NewLogger("kindlegen", "Sorry, conversion failed.")
}

func openFile(path string) *os.File {
    file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        logger.Fail("Failed opening file: %s", err.Error())
    }
    return file
}

func writeHTML(j *job.Job) {
    html := template.RenderToString(j.Title, tmpl, j)
    file := openFile(j.HTMLFilePath())
    defer file.Close()
    if _, err := file.WriteString(html); err != nil {
        logger.Fail("Failed writing HTML file: %s", err.Error())
    }
}

func Convert(j *job.Job) {
    go logger.SwallowErrorAndNotify(j, func() {
        writeHTML(j)
        cmd := exec.Command(kindlegen, []string{j.HTMLFilename()}...)
        cmd.Dir = j.Root()
        out, err := cmd.CombinedOutput()
        if !util.FileExists(j.MobiFilePath()) {
            logger.Fail("Failed running kindlegen: %s {output=%s}", err.Error(), string(out))
        }
        j.Progress("Conversion complete...")
        postmark.Send(j)
    })
}
