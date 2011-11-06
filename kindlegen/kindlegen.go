package kindlegen

import (
    "job"
    "loggly"
    "postmark"
    "user"
)

func Convert(j *job.Job) {
    go loggly.SwallowErrorAndNotify(j.Key, func() {
        user.Notify(j.KeyString(), "Conversion complete...")
        postmark.Send(j)
    })
}
