package kindlegen

import (
    "exec"
    "fmt"
    "job"
    "loggly"
    "os"
    "postmark"
    "runtime"
)

const Friendly = "Sorry, conversion failed."

var kindlegen string

func init() {
    kindlegen = fmt.Sprintf("bin/kindlegen-%s", runtime.GOOS)
}

func openFile(path string) *os.File {
    file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(loggly.NewError(
            fmt.Sprintf("Failed opening file: %s", err.Error()),
            Friendly))
    }
    return file
}

func fileExists(path string) bool {
    stat, _ := os.Stat(path)
    return stat != nil
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
    if _, err := file.WriteString(html); err != nil {
        panic(loggly.NewError(
            fmt.Sprintf("Failed writing HTML file: %s", err.Error()),
            Friendly))
    }
}

func Convert(j *job.Job) {
    go loggly.SwallowErrorAndNotify(j, func() {
        writeHTML(j)
        cmd := exec.Command("kindlegen", []string{kindlegen, j.HTMLFilename()}...)
        cmd.Dir = j.Root()
        out, err := cmd.CombinedOutput()
        if !fileExists(j.MobiFilePath()) {
            panic(loggly.NewError(
                fmt.Sprintf("Failed running kindlegen: %s {output=%s}", err.Error(), string(out)),
                Friendly))
        }
        j.Progress("Conversion complete...")
        postmark.Send(j)
    })
}
