package extractor

import (
    "net/http"
    "time"
    "fmt"
    "os"
    "util"
)

type downloader struct {
    root string
    client *http.Client
}

func newDownloader(root string, timeout time.Duration, deadline time.Time) downloader {
    return downloader{
        root: root,
        client: newTimeoutDeadlineDialer(timeout, deadline),
    }
}

func (d *downloader) output(path string) string {
    return fmt.Sprintf("%s/%s", d.root, path)
}

func (d *downloader) get(url string) *http.Response {
    resp, err := d.client.Get(url)
    if err != nil {
        logger.Panicf("Failed downloading %s: %s", url, err)
    }
    return resp
}

func (d *downloader) downloadToFile(url, path string) {
    resp := d.get(url)
    defer resp.Body.Close()
    file, err := os.OpenFile(d.output(path), os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        logger.Panicf("Failed opening file: %s", err)
    }
    defer file.Close()
    util.Pipe(file, resp.Body, resp.ContentLength, func(err error) {
        logger.Panicf("Error with io.Copy: %s", err)
    })
}
