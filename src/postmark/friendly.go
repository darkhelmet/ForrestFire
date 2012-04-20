package postmark

import (
    "fmt"
)

type friendly struct {
    message string
}

func (f friendly) Friendly() string {
    return f.message
}

func failFriendly(format string, v ...interface{}) {
    panic(friendly{fmt.Sprintf(format, v...)})
}
