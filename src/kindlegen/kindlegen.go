package kindlegen

import (
    "fmt"
    "github.com/darkhelmet/env"
    T "html/template"
    J "job"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
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
        <h1>{{.Title}}</h1>
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

type Kindlegen struct {
    Input  <-chan J.Job
    Output chan<- J.Job
    Error  chan<- J.Job
}

func New(input <-chan J.Job, output chan<- J.Job, error chan<- J.Job) *Kindlegen {
    return &Kindlegen{
        Input:  input,
        Output: output,
        Error:  error,
    }
}

func (k *Kindlegen) error(job J.Job, format string, args ...interface{}) {
    logger.Printf(format, args...)
    job.Friendly = FriendlyMessage
    k.Error <- job
}

func (k *Kindlegen) Run() {
    for job := range k.Input {
        go k.Process(job)
    }
}

func (k *Kindlegen) Process(job J.Job) {
    if err := writeHTML(job); err != nil {
        k.error(job, err.Error())
        return
    }

    cmd := exec.Command(kindlegen, []string{job.HTMLFilename()}...)
    cmd.Dir = job.Root()
    out, err := cmd.CombinedOutput()
    if !fileExists(job.MobiFilePath()) {
        k.error(job, "Failed running kindlegen: %s {output=%s}", err, out)
        return
    }

    job.Progress("Conversion complete...")
    k.Output <- job
}

func fileExists(path string) bool {
    stat, _ := os.Stat(path)
    return stat != nil
}

func writeHTML(job J.Job) error {
    file, err := os.OpenFile(job.HTMLFilePath(), os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("Failed opening file: %s", err)
    }
    defer file.Close()

    if err = template.Execute(file, &job); err != nil {
        return fmt.Errorf("Failed executing template: %s", err)
    }
    return nil
}
