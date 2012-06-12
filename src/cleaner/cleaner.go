package cleaner

import (
    J "job"
    "os"
)

type Cleaner struct {
    Input <-chan J.Job
}

func New(input <-chan J.Job) *Cleaner {
    return &Cleaner{Input: input}
}

func (c *Cleaner) Run() {
    for job := range c.Input {
        go c.Process(job)
    }
}

func (*Cleaner) Process(job J.Job) {
    if job.Friendly != "" {
        job.Progress(job.Friendly)
    }

    os.RemoveAll(job.Root())
}
