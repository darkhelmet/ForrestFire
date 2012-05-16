package safely

import (
    "cleanup"
    "index/suffixarray"
    "log"
    "runtime/debug"
    "sort"
)

const (
    DefaultProgress = "Something failed, sorry :("
)

type friendly interface {
    Friendly() string
}

type Jobber interface {
    Root() string
    Progress(string)
}

func Ignore(logger *log.Logger, f func()) {
    defer func() {
        if r := recover(); r != nil {
            debug.PrintStack()
        }
    }()
    f()
}

func pruneStack(stack []byte) []byte {
    index := suffixarray.New(stack)
    indexes := sort.IntSlice(index.Lookup([]byte{'\n'}, -1))
    sort.Sort(indexes)
    return stack[indexes[3]:]
}

func Do(logger *log.Logger, j Jobber, progress string, f func()) {
    defer func() {
        if r := recover(); r != nil {
            if err, ok := r.(friendly); ok {
                progress = err.Friendly()
                logger.Printf("%s: %#v", progress, j)
            } else {
                logger.Printf("%v: %#v", r, j)
            }
            logger.Printf("%s", pruneStack(debug.Stack()))
            j.Progress(progress)
            cleanup.Clean(j)
        }
    }()
    f()
}
