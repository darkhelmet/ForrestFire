package bookmarklet

import (
    "bytes"
    JSON "encoding/json"
    "fmt"
    "github.com/darkhelmet/env"
    "github.com/darkhelmet/webcompiler"
    "io/ioutil"
    "log"
    "os"
    "os/signal"
    "syscall"
    "text/template"
)

const (
    CoffeeScriptPath = "src/bookmarklet/bookmarklet.coffee"
    LessPath         = "src/bookmarklet/bookmarklet.less"
)

var (
    script   = make(chan []byte, 1)
    compress = env.StringDefault("BOOKMARKLET_PRECOMPILE", "") == "ugly"
    logger   = log.New(os.Stdout, "[bookmarklet] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
)

func init() {
    script <- compileCoffeeScript(compress)
    update := make(chan os.Signal, 1)
    signal.Notify(update, syscall.SIGUSR1)
    go handleSignals(update)
}

func handleSignals(c <-chan os.Signal) {
    for _ = range c {
        logger.Printf("Recompiling")
        s := compileCoffeeScript(compress)
        <-script
        script <- s
    }
}

func readFile(file string) []byte {
    data, err := ioutil.ReadFile(file)
    if err != nil {
        panic(fmt.Errorf("bookmarklet: Failed reading %s: %s", file, err))
    }
    return data
}

func Javascript() []byte {
    s := <-script
    script <- s
    return s
}

func compileCoffeeScript(compress bool) []byte {
    tmpl, err := template.ParseFiles(CoffeeScriptPath)
    if err != nil {
        panic(fmt.Errorf("bookmarklet: Failed building template: %s", err))
    }

    context := map[string]string{
        "Style": string(compileLessToJson(compress)),
    }

    var buffer bytes.Buffer
    tmpl.Execute(&buffer, context)

    js, err := webcompiler.CoffeeScript(&buffer, compress)
    if err != nil {
        panic(fmt.Errorf("bookmarklet: Failed compiling bookmarklet: %s", err))
    }
    return js
}

func compileLessToJson(compress bool) []byte {
    css, err := webcompiler.Less(bytes.NewReader(readFile(LessPath)), compress)
    if err != nil {
        panic(fmt.Errorf("bookmarklet: Failed compiling bookmarklet css: %s", err))
    }
    json, err := JSON.Marshal(string(css))
    if err != nil {
        panic(fmt.Errorf("bookmarklet: Failed dumping CSS to JSON: %s", err))
    }
    return json
}
