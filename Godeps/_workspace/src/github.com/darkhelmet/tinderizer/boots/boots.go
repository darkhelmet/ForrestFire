package boots

import (
    "code.google.com/p/go.net/html"
    "io"
)

func Walk(r io.Reader, tag string, f func(*html.Node)) (*html.Node, error) {
    doc, err := html.Parse(r)
    if err != nil {
        return nil, err
    }

    var walker func(*html.Node)
    walker = func(node *html.Node) {
        if node.Type == html.ElementNode && node.Data == tag {
            f(node)
        }

        for c := node.FirstChild; c != nil; c = c.NextSibling {
            walker(c)
        }
    }
    walker(doc)

    return doc, nil
}
