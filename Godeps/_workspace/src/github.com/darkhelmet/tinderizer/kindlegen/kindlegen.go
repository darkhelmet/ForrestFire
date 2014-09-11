package kindlegen

import (
    "fmt"
    "github.com/darkhelmet/env"
    J "github.com/darkhelmet/tinderizer/job"
    T "html/template"
    "log"
    "os"
    "os/exec"
    "sync"
)

const (
    FriendlyMessage = "Sorry, conversion failed."
    Tmpl            = `
<html>
    <head>
        <meta content="text/html; charset=utf-8" http-equiv="Content-Type" />
        <meta content="{{.Author}} ({{.Domain}})" name="author" />
        <title>{{.Title}}</title>
        <style type="text/css">
            h1, h2, h3, h4, h5 {
                margin-bottom: 0.5em;
            }

            p, ol, ul {
                margin-bottom: 1em;
            }

            .meta {
                font-weight: bold;
                font-style: italic;
            }
        </style>
    </head>
    <body>
        <h1>{{.Title}}</h1>
        <hr />
        <p class="meta">{{if .Author}}By {{.Author}} on {{else}}On {{end}}<a href="{{.Url}}">{{.Domain}}</a></p>
        {{.HTML}}
        <hr />
        <p>Sent with <a href="https://Tinderizer.com/">Tinderizer</a> at {{.Now}} from <a href="{{.Url}}">{{.Url}}</a></p>
        <p>Please donate at <a href="https://Tinderizer.com/">https://Tinderizer.com/</a> if you find this application useful.</p>
    </body>
</html>
`
)

var (
    template *T.Template
    logger   = log.New(os.Stdout, "[kindlegen] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
)

func init() {
    template = T.Must(T.New("kindle").Parse(Tmpl))
}

type Kindlegen struct {
    wg     sync.WaitGroup
    binary string
    Input  <-chan J.Job
    Output chan<- J.Job
    Error  chan<- J.Job
}

func New(binary string, input <-chan J.Job, output chan<- J.Job, error chan<- J.Job) *Kindlegen {
    return &Kindlegen{
        binary: binary,
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

func (k *Kindlegen) Run(wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range k.Input {
        k.wg.Add(1)
        go k.Process(job)
    }
    k.wg.Wait()
    close(k.Output)
}

func (k *Kindlegen) Process(job J.Job) {
    job.Progress("Optimizing for Kindle...")

    defer k.wg.Done()
    if err := writeHTML(job); err != nil {
        k.error(job, err.Error())
        return
    }

    cmd := exec.Command(k.binary, []string{job.HTMLFilename()}...)
    cmd.Dir = job.Root()
    out, err := cmd.CombinedOutput()
    if !fileExists(job.MobiFilePath()) {
        k.error(job, "failed running kindlegen: %s {output=%s}", err, out)
        return
    }

    job.Progress("Optimization complete...")
    k.Output <- job
}

func fileExists(path string) bool {
    stat, _ := os.Stat(path)
    return stat != nil
}

func writeHTML(job J.Job) error {
    file, err := os.OpenFile(job.HTMLFilePath(), os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("failed opening file: %s", err)
    }
    defer file.Close()

    if err = template.Execute(file, &job); err != nil {
        return fmt.Errorf("failed executing template: %s", err)
    }
    return nil
}
