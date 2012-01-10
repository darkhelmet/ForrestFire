package h5

import (
    "fmt"
    "os"
    "testing"
    "testing/util"
)

func TestPushNode(t *testing.T) {
    p := new(Parser)
    util.AssertTrue(t, p.Top == nil, "Top is not nil")
    util.AssertTrue(t, p.curr == nil, "curr is not nil")
    top := pushNode(p)
    top.data = append(top.data, []rune("foo")...)
    util.AssertTrue(t, p.Top != nil, "Top is still nil")
    util.AssertTrue(t, p.curr != nil, "curr is stil nil")
    util.AssertEqual(t, p.Top, top)
    util.AssertEqual(t, p.curr, top)
    next := pushNode(p)
    next.data = append(next.data, []rune("bar")...)
    util.AssertEqual(t, len(top.Children), 1)
    util.AssertEqual(t, p.Top, top)
    util.AssertEqual(t, p.curr, next)
    util.AssertEqual(t, p.curr.Parent, p.Top)
}

func TestPopNode(t *testing.T) {
    p := new(Parser)
    top := pushNode(p)
    top.data = append(top.data, []rune("foo")...)
    next := pushNode(p)
    next.data = append(next.data, []rune("bar")...)
    popped := popNode(p)
    util.AssertEqual(t, popped, top)
    util.AssertEqual(t, p.Top, p.curr)
}

func TestAddSibling(t *testing.T) {
    p := new(Parser)
    top := pushNode(p)
    top.data = append(top.data, []rune("foo")...)
    next := pushNode(p)
    next.data = append(next.data, []rune("bar")...)
    sib := addSibling(p)
    sib.data = append(sib.data, []rune("baz")...)
    util.AssertEqual(t, len(top.Children), 2)
    util.AssertEqual(t, top.Children[0], next)
    util.AssertEqual(t, top.Children[1], sib)
}

func TestBogusCommentHandlerNoEOF(t *testing.T) {
    p := NewParserFromString("foo comment >")
    top := pushNode(p)
    pushNode(p)
    st, err := bogusCommentHandler(p)
    util.AssertEqual(t, len(top.Children), 2)
    util.AssertEqual(t, string(top.Children[1].data), "foo comment ")
    util.AssertTrue(t, st != nil, "next state handler is nil")
    util.AssertTrue(t, err == nil, "err is not nil")
}

// TODO error cases
func TestBogusCommentHandlerEOF(t *testing.T) {
    p := NewParserFromString("foo comment")
    top := pushNode(p)
    pushNode(p)
    st, err := bogusCommentHandler(p)
    util.AssertEqual(t, len(top.Children), 2)
    util.AssertEqual(t, string(top.Children[1].data), "foo comment")
    util.AssertTrue(t, st == nil, "next state handler is not nil")
    util.AssertTrue(t, err != nil, "err is nil")
}

func TestEndTagOpenHandlerOk(t *testing.T) {
    p := NewParserFromString("FoO>")
    curr := pushNode(p)
    curr.data = []rune("foo")
    util.AssertTrue(t, p.curr != nil, "curr is not nil")
    st, err := endTagOpenHandler(p)
    util.AssertTrue(t, st != nil, "next state handler is nil")
    util.AssertEqual(t, err, nil)
    util.AssertTrue(t, err == nil, "err is not nil")
    //util.AssertTrue(t, p.curr == nil, "did not pop node")
}

func TestEndTagOpenHandlerTrunc(t *testing.T) {
    p := NewParserFromString("fo>")
    curr := pushNode(p)
    curr.data = []rune("foo")
    util.AssertTrue(t, p.curr != nil, "curr is not nil")
    st, err := endTagOpenHandler(p)
    util.AssertTrue(t, st == nil, "next state handler is not nil")
    util.AssertTrue(t, err != nil, "err is nil")
    util.AssertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerLong(t *testing.T) {
    p := NewParserFromString("fooo>")
    curr := pushNode(p)
    curr.data = []rune("foo")
    util.AssertTrue(t, p.curr != nil, "curr is not nil")
    st, err := endTagOpenHandler(p)
    util.AssertTrue(t, st == nil, "next state handler is not nil")
    util.AssertTrue(t, err != nil, "err is nil")
    util.AssertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerWrong(t *testing.T) {
    p := NewParserFromString("bar>")
    curr := pushNode(p)
    curr.data = []rune("foo")
    util.AssertTrue(t, p.curr != nil, "curr is not nil")
    st, err := endTagOpenHandler(p)
    util.AssertTrue(t, st == nil, "next state handler is not nil")
    util.AssertTrue(t, err != nil, "err is nil")
    util.AssertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerBogusComment(t *testing.T) {
    p := NewParserFromString("f\no>")
    curr := pushNode(p)
    curr.data = []rune("foo")
    util.AssertTrue(t, p.curr != nil, "curr is not nil")
    st, err := endTagOpenHandler(p)
    util.AssertTrue(t, st != nil, "next state handler is not nil")
    util.AssertTrue(t, err != nil, "err is nil")
    util.AssertEqual(t, p.curr, curr)
}

func TestEndTagOpenHandlerEOF(t *testing.T) {
    p := NewParserFromString("foo")
    curr := pushNode(p)
    curr.data = []rune("foo")
    util.AssertTrue(t, p.curr != nil, "curr is not nil")
    st, err := endTagOpenHandler(p)
    util.AssertTrue(t, st == nil, "next state handler is nil")
    util.AssertTrue(t, err != nil, "err is nil")
    util.AssertEqual(t, p.curr, curr)
}

func TestTagNameHandler(t *testing.T) {
    p := NewParserFromString("f> ")
    curr := pushNode(p)
    st, err := handleChar(tagNameHandler)(p)
    util.AssertTrue(t, st != nil, "next state handler is nil")
    util.AssertTrue(t, err == nil, "err is not nil")
    util.AssertEqual(t, curr.data[0], 'f')
    st, err = handleChar(tagNameHandler)(p)
    util.AssertTrue(t, st != nil, "next state handler is nil")
    util.AssertTrue(t, err == nil, "err is not nil")
    util.AssertEqual(t, curr.data[0], 'f')
    p = NewParserFromString("F")
    curr = pushNode(p)
    st, err = handleChar(tagNameHandler)(p)
    util.AssertTrue(t, st != nil, "next state handler is nil")
    util.AssertTrue(t, err == nil, "err is not nil")
    util.AssertEqual(t, curr.data[0], 'f')
}

func TestTagOpenHandler(t *testing.T) {
    p := NewParserFromString("")
    st := tagOpenHandler(p, 'f')
    util.AssertTrue(t, st != nil, "next state handler is nil")
    util.AssertEqual(t, st, handleChar(tagNameHandler))
    util.AssertEqual(t, p.curr.data[0], 'f')
    util.AssertEqual(t, p.curr.Type, ElementNode)
}

func TestTagOpenHandlerEndTag(t *testing.T) {
    p := NewParserFromString("")
    st := tagOpenHandler(p, '/')
    util.AssertTrue(t, st != nil, "next state handler is nil")
    //util.AssertEqual(t, st, endTagOpenHandler)
}

func TestDataStateHandler(t *testing.T) {
    p := NewParserFromString("")
    st := dataStateHandler(p, '<')
    util.AssertTrue(t, st != nil, "next state handler is nil")
    util.AssertEqual(t, st, handleChar(tagOpenHandler))
    util.AssertTrue(t, p.curr == nil, "curr is currently nil")
    util.AssertTrue(t, p.Top == nil, "Top is currently nil")
    p = NewParserFromString("oo<")
    st = dataStateHandler(p, 'f')
    util.AssertTrue(t, st != nil, "next state handler is nil")
    util.AssertTrue(t, p.curr != nil, "curr is currently nil")
    util.AssertTrue(t, p.Top != nil, "Top is currently nil")
    util.AssertEqual(t, p.curr.data, []rune("foo"))
}

func TestSimpledoc(t *testing.T) {
    p := NewParserFromString("<html><body>foo</body></html>")
    err := p.Parse()
    util.AssertTrue(t, err == nil, "err is not nil: %v", err)
    //fmt.Printf("XXX doc: %s\n", p.Top)
    util.AssertEqual(t, p.Top.Data(), "html")
    util.AssertEqual(t, len(p.Top.Children), 1)
    util.AssertEqual(t, len(p.Top.Children[0].Children), 1)
    util.AssertEqual(t, p.Top.Children[0].Data(), "body")
    util.AssertEqual(t, p.Top.Children[0].Children[0].Data(), "foo")
}

func TestScriptDoc(t *testing.T) {
    p := NewParserFromString(
        "<html><body><script> if (foo < 10) { }</script></body></html>")
    err := p.Parse()
    util.AssertTrue(t, err == nil, "err is not nil: %v", err)
    //fmt.Printf("XXX doc: %s\n", p.Top)
    util.AssertEqual(t, p.Top.Data(), "html")
    util.AssertEqual(t, len(p.Top.Children), 1)
    util.AssertEqual(t, p.Top.Children[0].Data(), "body")
    util.AssertEqual(t, len(p.Top.Children[0].Children), 1)
    util.AssertEqual(t, p.Top.Children[0].Children[0].Data(), "script")
    util.AssertEqual(t, p.Top.Children[0].Children[0].Children[0].Data(),
        " if (foo < 10) { }")
}

func TestSimpledocSiblings(t *testing.T) {
    p := NewParserFromString(
        "<html><body><a>foo</a><div>bar</div></body></html>")
    err := p.Parse()
    util.AssertTrue(t, err == nil, "err is not nil: %v", err)
    //fmt.Printf("XXX doc: %s\n", p.Top)
    util.AssertEqual(t, p.Top.Data(), "html")
    util.AssertEqual(t, len(p.Top.Children), 1)
    util.AssertEqual(t, len(p.Top.Children[0].Children), 2)
    util.AssertEqual(t, p.Top.Children[0].Data(), "body")
    util.AssertEqual(t, p.Top.Children[0].Children[0].Data(), "a")
}

func TestParseFromReader(t *testing.T) {
    rdr, err := os.Open("test_data/page.html")
    if err != nil {
        fmt.Println("Error: ", err)
        os.Exit(1)
    }
    p := NewParser(rdr)
    err = p.Parse()
    if err != nil {
        util.AssertTrue(t, false, "Failed to parse")
    }
    util.AssertTrue(t, p.Top != nil, "We got a parse tree back")
    //fmt.Println("Doc: ", p.Top.String())
}

func TestNodeClone(t *testing.T) {
    p := NewParserFromString(
        "<html><body><a>foo</a><div>bar</div></body></html>")
    p.Parse()
    n := p.Top.Clone()
    util.AssertTrue(t, n != nil, "n is nil")
    util.AssertEqual(t, n.Data(), "html")
    util.AssertEqual(t, len(n.Children), 1)
    util.AssertEqual(t, len(n.Children[0].Children), 2)
    util.AssertEqual(t, n.Children[0].Data(), "body")
    util.AssertEqual(t, n.Children[0].Children[0].Data(), "a")
}

func TestNodeWalk(t *testing.T) {
    p := NewParserFromString(
        "<html><body><a>foo</a><div>bar</div></body></html>")
    p.Parse()
    i := 0
    ns := make([]string, 6)
    f := func(n *Node) {
        ns[i] = n.Data()
        i++
    }
    p.Top.Walk(f)
    util.AssertEqual(t, i, 6)
    util.AssertEqual(t, ns, []string{"html", "body", "a", "foo", "div", "bar"})
}

func TestSnippet(t *testing.T) {
    p := NewParserFromString("<a></a>")
    err := p.Parse()
    util.AssertTrue(t, err == nil, "we errored while parsing snippet %s", err)
    util.AssertTrue(
        t, p.Top != nil, "We didn't get a node tree back while parsing snippet")
    util.AssertEqual(t, p.Top.Data(), "a")
}

func TestMeta(t *testing.T) {
    p := NewParserFromString(
        "<html><head><meta><link href='foo'></head><body><div>foo</div></body></html>")
    err := p.Parse()
    util.AssertTrue(t, err == nil, "err was not nil, %s", err)
    n := p.Top
    fmt.Println(p.Top)
    util.AssertTrue(t, n != nil, "n is nil")
    util.AssertEqual(t, n.Data(), "html")
    util.AssertEqual(t, len(n.Children), 2)
    util.AssertEqual(t, len(n.Children[0].Children), 2)
    util.AssertEqual(t, n.Children[0].Data(), "head")
    util.AssertEqual(t, n.Children[0].Children[0].Data(), "meta")
    util.AssertEqual(t, n.Children[0].Children[1].Data(), "link")
    util.AssertEqual(t, n.Children[1].Data(), "body")
    util.AssertEqual(t, n.Children[1].Children[0].Data(), "div")
    util.AssertEqual(t, n.Children[1].Children[0].Children[0].Data(), "foo")
}

func TestComment(t *testing.T) {
    p := NewParserFromString(
        "<html><head><!-- comment --></head><body><div>foo</div></body></html>")
    err := p.Parse()
    util.AssertTrue(t, err == nil, "err was not nil, %s", err)
    n := p.Top
    fmt.Println(p.Top)
    util.AssertTrue(t, n != nil, "n is nil")
    util.AssertEqual(t, n.Data(), "html")
    util.AssertEqual(t, len(n.Children), 2)
    util.AssertEqual(t, len(n.Children[0].Children), 1)
    util.AssertEqual(t, n.Children[0].Data(), "head")
    util.AssertEqual(t, n.Children[0].Children[0].Data(), " comment ")
    util.AssertEqual(t, n.Children[0].Children[0].Type, CommentNode)
    util.AssertEqual(t, n.Children[1].Data(), "body")
    util.AssertEqual(t, n.Children[1].Children[0].Data(), "div")
    util.AssertEqual(t, n.Children[1].Children[0].Children[0].Data(), "foo")
}
