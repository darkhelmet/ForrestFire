package kindlegen

import (
    "fmt"
    "github.com/darkhelmet/env"
    T "html/template"
    "job"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "postmark"
    "runtime"
    "safely"
    "stat"
    "util"
)

const (
    FriendlyMessage = "Sorry, conversion failed."
    Tmpl            = `
<html>
    <head>
        <meta content="text/html, charset=utf-8" http-equiv="Content-Type" />
        <meta content="{{.Author}} ({{.Domain}})" name="author" />
        <title>{{.Title}}</title>
    </head>
    <body>
        <h1>{{.Title | html}}</h1>
        {{.HTML}}
        <hr />
        <p>Originally from <a href="{{.Url}}">{{.Url}}</a></p>
        <p>Sent with <a href="http://Tinderizer.com/">Tinderizer</a></p>
        <p>Generated at {{.Now}}</p>
    </body>
</html>
`
)

var (
    kindlegen string
    template  *T.Template
    logger    = log.New(os.Stdout, "[kindlegen] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
)

func init() {
    var err error
    kindlegen, err = filepath.Abs(fmt.Sprintf("vendor/kindlegen-%s", runtime.GOOS))
    if err != nil {
        panic(err)
    }
    template = T.Must(T.New("kindle").Parse(Tmpl))
}

func openFile(path string) *os.File {
    file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        logger.Panicf("Failed opening file: %s", err)
    }
    return file
}

func writeHTML(j *job.Job) {
    file := openFile(j.HTMLFilePath())
    defer file.Close()
    if err := template.Execute(file, j); err != nil {
        logger.Panicf("Failed rendering HTML to file: %s", err)
    }
}

func Convert(j *job.Job) {
    go safely.Do(logger, j, FriendlyMessage, stat.KindlegenUnhandled, func() {
        writeHTML(j)
        cmd := exec.Command(kindlegen, []string{j.HTMLFilename()}...)
        cmd.Dir = j.Root()
        out, err := cmd.CombinedOutput()
        if !util.FileExists(j.MobiFilePath()) {
            logger.Panicf("Failed running kindlegen: %s {output=%s}", err, out)
        }
        j.Progress("Conversion complete...")
        postmark.Mail(j)
    })
}
