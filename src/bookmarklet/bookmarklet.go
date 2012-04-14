package bookmarklet

import (
    "compiler"
    "env"
    "fmt"
    "io/ioutil"
)

type Marker struct {
    f func() []byte
}

const CoffeeScriptPath = "src/bookmarklet/bookmarklet.coffee"

var script []byte
var bm Marker

func init() {
    cs, err := ioutil.ReadFile(CoffeeScriptPath)
    if err != nil {
        panic(fmt.Sprintf("Failed reading bookmarklet: %s", err))
    }
    script = cs

    precompile := env.GetDefault("BOOKMARKLET_PRECOMPILE", "")
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
    js, err := compiler.CoffeeScript(script, uglifier)
    if err != nil {
        panic(fmt.Sprintf("Failed compiling bookmarklet: %s", err))
    }
    return js
}
