package cleaner

import (
	J "github.com/darkhelmet/tinderizer/job"
	"os"
	"sync"
)

type Cleaner struct {
	wg    sync.WaitGroup
	Input <-chan J.Job
}

func New(input <-chan J.Job) *Cleaner {
	return &Cleaner{Input: input}
}

func (c *Cleaner) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range c.Input {
		c.wg.Add(1)
		go c.Process(job)
	}
	c.wg.Wait()
}

func (c *Cleaner) Process(job J.Job) {
	defer c.wg.Done()
	if job.Friendly != "" {
		job.Progress(job.Friendly)
	}

	os.RemoveAll(job.Root())
}
