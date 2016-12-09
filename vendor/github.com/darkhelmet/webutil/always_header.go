package webutil

import (
    "net/http"
)

type AlwaysHeaderHandler struct {
    H       http.Handler
    Headers http.Header
}

func (ahh AlwaysHeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    headers := w.Header()
    for key, values := range ahh.Headers {
        for _, value := range values {
            headers.Add(key, value)
        }
    }
    ahh.H.ServeHTTP(w, r)
}
