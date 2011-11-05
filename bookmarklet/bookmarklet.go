package bookmarklet

import (
    "env"
    "exec"
    "fmt"
)

type Marker struct {
    f func() []byte
}

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
    script := "CoffeeScript.compile(File.read('bookmarklet/bookmarklet.coffee'))"
    if uglifier {
        script = fmt.Sprintf("Uglifier.compile(%s)", script)
    }
    bundle, err := exec.LookPath("bundle")
    if err != nil {
        panic("ruby/bundle not found")
    }
    args := []string{bundle, "exec", "ruby", "-rcoffee-script", "-ruglifier", "-Eutf-8:utf-8", "-e", fmt.Sprintf("print %s", script)}
    cmd := exec.Command("ruby", args...)
    out, err := cmd.Output()
    if err != nil {
        panic(fmt.Sprintf("Error running coffee-script: %s", err.Error()))
    }
    return out
}
