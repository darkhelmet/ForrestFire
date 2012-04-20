package cleanup

import (
    "os"
)

type rooter interface {
    Root() string
}

func Clean(r rooter) {
    go os.RemoveAll(r.Root())
}
