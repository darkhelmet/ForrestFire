package cleanup

import (
    "job"
    "os"
)

func Clean(j *job.Job) {
    go os.RemoveAll(j.Root())
}
