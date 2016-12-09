package webutil

import (
    "log"
    "net/http"
)

var (
    HerokuHeaders = []string{"X-Varnish", "X-Forwarded-For", "X-Heroku-Dynos-In-Use", "X-Request-Start", "X-Heroku-Queue-Wait-Time", "X-Heroku-Queue-Depth", "X-Real-Ip", "X-Forwarded-Proto", "X-Via", "X-Forwarded-Port"}
)

type HerokuHandler struct {
    H      http.Handler
    Logger *log.Logger
}

func (hh HerokuHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    info := make(map[string]string)
    for _, key := range HerokuHeaders {
        value := r.Header.Get(key)
        if value != "" {
            info[key] = value
        }
    }
    if len(info) > 0 {
        hh.Logger.Printf("%sn", info)
    }
    hh.H.ServeHTTP(w, r)
}
