package bookmarklet

import (
    "env"
    "ruby"
    "fmt"
)

type Marker struct {
    f func() []byte
}

const CoffeeScriptCompile = "CoffeeScript.compile(File.read('bookmarklet/bookmarklet.coffee'))"

var bm Marker

func init() {
    precompile := env.GetDefault("BOOKMARKLET_PRECOMPILE", "")
    if precompile == "ugly" {
        js := Compile(true)
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
    script := CoffeeScriptCompile
    if uglifier {
        script = fmt.Sprintf("Uglifier.compile(%s)", script)
    }
    script = fmt.Sprintf("STDOUT.write(%s)", script)
    out, err := ruby.Run(script, []string{"coffee-script", "uglifier"})
    if err != nil {
        panic(fmt.Sprintf("Error running coffee-script: %s: %s", err.Error()))
    }
    return out
}
