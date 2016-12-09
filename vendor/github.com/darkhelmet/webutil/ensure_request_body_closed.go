package webutil

import "net/http"

type EnsureRequestBodyClosedHandler struct {
    H http.Handler
}

func (erbch EnsureRequestBodyClosedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()
    erbch.H.ServeHTTP(w, r)
}
