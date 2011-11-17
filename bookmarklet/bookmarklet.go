package bookmarklet

import (
    "env"
    "fmt"
    "os/exec"
)

type Marker struct {
    f func() []byte
}

const CoffeeScriptCompile = "coffee -p -c bookmarklet/bookmarklet.coffee"

var bm Marker
var bash string

func init() {
    var err error
    bash, err = exec.LookPath("bash")
    if err != nil {
        panic("bash not found")
    }

    if _, e := exec.LookPath("coffee"); e != nil {
        panic("coffee-script not found")
    }

    if _, e := exec.LookPath("uglifyjs"); e != nil {
        panic("uglify-js not found")
    }

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
    script := CoffeeScriptCompile
    if uglifier {
        script = fmt.Sprintf("%s | uglifyjs", script)
    }
    args := []string{bash, "-c", script}
    cmd := exec.Command("bash", args...)
    out, err := cmd.Output()
    if err != nil {
        panic(fmt.Sprintf("Failed compiling: %s", err.Error()))
    }
    return out
}
