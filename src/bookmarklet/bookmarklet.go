package bookmarklet

import (
    "bytes"
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/webcompiler"
    "io/ioutil"
)

type Marker struct {
    f func() []byte
}

const CoffeeScriptPath = "src/bookmarklet/bookmarklet.coffee"

var (
    script []byte
    bm     Marker
)

func init() {
    cs, err := ioutil.ReadFile(CoffeeScriptPath)
    if err != nil {
        panic(fmt.Errorf("Failed reading bookmarklet: %s", err))
    }
    script = cs

    precompile := env.StringDefault("BOOKMARKLET_PRECOMPILE", "")
    if precompile != "" {
        js := Compile(precompile == "ugly")
        bm = Marker{func() []byte {
            return js
        }}
    } else {
        bm = Marker{func() []byte {
            return Compile(false)
        }}
    }
}

func Javascript() []byte {
    return bm.f()
}

func Compile(uglifier bool) []byte {
    js, err := webcompiler.CoffeeScript(bytes.NewReader(script), uglifier)
    if err != nil {
        panic(fmt.Errorf("Failed compiling bookmarklet: %s", err))
    }
    return js
}
