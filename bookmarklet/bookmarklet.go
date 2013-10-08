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
    CoffeeScriptPath = "bookmarklet/bookmarklet.coffee"
    LessPath         = "bookmarklet/bookmarklet.less"
)

var (
    script   = make(chan []byte, 1)
    protocol = env.StringDefault("PROTOCOL", "http")
    port     = env.IntDefault("PORT", 8080)
    host     = env.StringDefaultF("CANONICAL_HOST", func() string { return fmt.Sprintf("tinderizer.dev:%d", port) })
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
    tmpl := template.Must(template.ParseFiles(CoffeeScriptPath))

    context := map[string]string{
        "Style":    string(compileLessToJson(compress)),
        "Protocol": protocol,
        "Host":     host,
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
