package compiler

import (
    "bytes"
    "io/ioutil"
    "mime/multipart"
    "net/http"
)

const Endpoint = "http://compiler.herokuapp.com/"

func call(f func(*multipart.Writer)) ([]byte, error) {
    var buffer bytes.Buffer
    w := multipart.NewWriter(&buffer)
    f(w)
    w.Close()

    resp, err := http.Post(Endpoint, w.FormDataContentType(), &buffer)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return ioutil.ReadAll(resp.Body)
}

func Less(data []byte, compress bool) ([]byte, error) {
    return call(func(w *multipart.Writer) {
        if compress {
            w.WriteField("compress", "1")
        }
        file, _ := w.CreateFormFile("less", "style.less")
        file.Write(data)
    })
}

func CoffeeScript(data []byte, compress bool) ([]byte, error) {
    return call(func(w *multipart.Writer) {
        if compress {
            w.WriteField("uglify", "1")
        }
        file, _ := w.CreateFormFile("coffee", "script.coffee")
        file.Write(data)
    })
}
