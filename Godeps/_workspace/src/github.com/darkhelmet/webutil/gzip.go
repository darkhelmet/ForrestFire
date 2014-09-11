package webutil

import (
    "compress/gzip"
    "net/http"
    "strings"
)

const (
    HeaderAcceptEncoding  = "Accept-Encoding"
    HeaderContentEncoding = "Content-Encoding"
    HeaderContentLength   = "Content-Length"
    HeaderVary            = "Vary"
)

type GzipResponseWriter struct {
    gzr *gzip.Writer
    w   http.ResponseWriter
}

func (grw GzipResponseWriter) Header() http.Header {
    return grw.w.Header()
}

func (grw GzipResponseWriter) WriteHeader(code int) {
    grw.Header().Del(HeaderContentLength)
    grw.w.WriteHeader(code)
}

func (grw GzipResponseWriter) Write(b []byte) (int, error) {
    return grw.gzr.Write(b)
}

type GzipHandler struct {
    H http.Handler
}

func (gh GzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if strings.Contains(r.Header.Get(HeaderAcceptEncoding), "gzip") {
        headers := w.Header()
        headers.Set(HeaderContentEncoding, "gzip")
        headers.Set(HeaderVary, HeaderAcceptEncoding)
        gz := gzip.NewWriter(w)
        defer gz.Close()
        gzw := GzipResponseWriter{gz, w}
        gh.H.ServeHTTP(gzw, r)
        return
    }

    gh.H.ServeHTTP(w, r)
}
