package postmark

// TODO: Handle email invalid
// TODO: Handle large attachments
// TODO: Look through FogBugz and async.rb to see what I'm catching
// TODO: Cleanup

import (
    "job"
    "loggly"
    "user"
)

func Send(j *job.Job) {
    go loggly.SwallowErrorAndNotify(j.Key, func() {
        user.Notify(j.KeyString(), "All done! Grab your Kindle and hang tight!")
    })
}
