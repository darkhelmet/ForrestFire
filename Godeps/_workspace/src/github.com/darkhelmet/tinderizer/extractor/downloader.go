package extractor

import (
    "fmt"
    "github.com/pkulak/simpletransport/simpletransport"
    "io"
    "net/http"
    "os"
    "time"
)

type downloader struct {
    root   string
    client *http.Client
}

func newDownloader(root string, timeout time.Duration) downloader {
    return downloader{
        root: root,
        client: &http.Client{
            Transport: &simpletransport.SimpleTransport{
                ReadTimeout:    timeout,
                RequestTimeout: timeout,
            },
        },
    }
}

func (d *downloader) output(path string) string {
    return fmt.Sprintf("%s/%s", d.root, path)
}

func (d *downloader) downloadToFile(url, path string) error {
    resp, err := d.client.Get(url)
    if err != nil {
        return fmt.Errorf("downloader: HTTP request failed: %s", err)
    }
    defer resp.Body.Close()
    file, err := os.OpenFile(d.output(path), os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("downloader: file open failed: %s", err)
    }
    defer file.Close()

    written, err := io.Copy(file, resp.Body)
    if err != nil {
        return fmt.Errorf("downloader: failed copying to file; %s", err)
    }

    if resp.ContentLength > 0 && written != resp.ContentLength {
        return fmt.Errorf("downloader: written != expected: %d != %d", written, resp.ContentLength)
    }

    return nil
}
