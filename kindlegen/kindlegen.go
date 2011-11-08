package kindlegen

import (
    "exec"
    "fmt"
    "job"
    "loggly"
    "os"
    "postmark"
    "runtime"
    "util"
)

const Friendly = "Sorry, conversion failed."

var kindlegen string

func init() {
    kindlegen = fmt.Sprintf("bin/kindlegen-%s", runtime.GOOS)
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
    template := `
    <html>
        <head>
            <meta content="text/html, charset=utf-8" http-equiv="Content-Type" />
            <meta content="%s (%s)" name="author" />
            <title>%s</title>
        </head>
        <body>
            <h1>%s</h1>
            %s
        </body>
    </html>
    `
    html := fmt.Sprintf(template, j.Author, j.Domain, j.Title, j.Title, j.HTML())
    file := openFile(j.HTMLFilePath())
    defer file.Close()
    if _, err := file.WriteString(html); err != nil {
        fail("Failed writing HTML file: %s", err.Error())
    }
}

func Convert(j *job.Job) {
    go loggly.SwallowErrorAndNotify(j, func() {
        writeHTML(j)
        cmd := exec.Command("kindlegen", []string{kindlegen, j.HTMLFilename()}...)
        cmd.Dir = j.Root()
        out, err := cmd.CombinedOutput()
        if !util.FileExists(j.MobiFilePath()) {
            fail("Failed running kindlegen: %s {output=%s}", err.Error(), string(out))
        }
        j.Progress("Conversion complete...")
        postmark.Send(j)
    })
}
