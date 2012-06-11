package cleaner

import (
    "github.com/trustmaster/goflow"
    J "job"
    "os"
)

type Cleaner struct {
    flow.Component
    Input <-chan J.Job
}

func New() *Cleaner {
    return new(Cleaner)
}

func (e *Cleaner) OnInput(job J.Job) {
    if job.Friendly != "" {
        job.Progress(job.Friendly)
    }

    os.RemoveAll(job.Root())
}
