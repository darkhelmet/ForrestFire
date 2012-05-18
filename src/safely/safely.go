package safely

import (
    "cleanup"
    "index/suffixarray"
    "runtime/debug"
    "sort"
    "stat"
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

type Logger interface {
    Printf(string, ...interface{})
}

func Ignore(f func()) {
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

func Do(logger Logger, j Jobber, progress, statName string, f func()) {
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
            stat.Count(statName, 1)
        }
    }()
    f()
}
