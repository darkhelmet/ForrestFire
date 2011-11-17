/*
Package h5 implements an html5 parser for the go language.

    p := h5.NewParser(rdr)
    err := p.Parse()
    tree := p.Tree()

    tree.Walk(func(n *Node) {
       // do something with the node
    })

    tree2 := tree.Clone()
*/
package h5

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Represents an html5 parsing error. holds a message and the current html5 node
// when the error occured.
type ParseError struct {
	msg  string
	node *Node
}

// Constructor for an html5 parsing error
func NewParseError(n *Node, msg string, args ...interface{}) *ParseError {
	return &ParseError{node: n, msg: fmt.Sprintf(msg, args...)}
}

// Represent the parse error as a string
func (e ParseError) Error() string {
	return e.msg
}

// The html5 insertion mode for parsing
type InsertionMode int

const (
	IM_initial            InsertionMode = iota
	IM_beforeHtml         InsertionMode = iota
	IM_beforeHead         InsertionMode = iota
	IM_inHead             InsertionMode = iota
	IM_inHeadNoScript     InsertionMode = iota
	IM_afterHead          InsertionMode = iota
	IM_inBody             InsertionMode = iota
	IM_text               InsertionMode = iota
	IM_inTable            InsertionMode = iota
	IM_inTableText        InsertionMode = iota
	IM_inCaption          InsertionMode = iota
	IM_inColumnGroup      InsertionMode = iota
	IM_inTableBody        InsertionMode = iota
	IM_inRow              InsertionMode = iota
	IM_inCell             InsertionMode = iota
	IM_inSelect           InsertionMode = iota
	IM_inSelectInTable    InsertionMode = iota
	IM_afterBody          InsertionMode = iota
	IM_afterFrameset      InsertionMode = iota
	IM_afterAfterBody     InsertionMode = iota
	IM_afterAfterFrameset InsertionMode = iota
)

func insertionModeSwitch(p *Parser, n *Node) stateHandler {
	//fmt.Println("In insertionModeSwitch")
	currMode := p.Mode
	switch currMode {
	case IM_initial:
		fallthrough
	case IM_beforeHtml:
		//fmt.Println("starting doctypeStateHandler")
		p.Mode = IM_beforeHead
		return handleChar(startDoctypeStateHandler)
		//fallthrough
	case IM_beforeHead:
		switch n.Type {
		case DoctypeNode:
			// TODO(jwall): parse error
		case CommentNode:
		case ElementNode:
			switch n.Data() {
			case "head":
				p.Mode = IM_inHead
			case "body":
				p.Mode = IM_inBody
			default:
				// TODO(jwall): parse error
			}
		default:
			// TODO(jwall): parse error
		}
	case IM_inHead:
		switch n.Type {
		case DoctypeNode:
			// TODO(jwall): parse error
		case CommentNode:
		case ElementNode:
			switch n.Data() {
			case "script":
				//fmt.Println("In a script tag")
				p.Mode = IM_text
				return handleChar(startScriptDataState)
			case "body":
				p.Mode = IM_inBody
			default:
				// TODO(jwall): parse error
			}
		default:
			// TODO(jwall): parse error
		}
	case IM_inHeadNoScript:
	case IM_afterHead:
		switch n.Type {
		case DoctypeNode:
			// TODO(jwall): parse error
		case CommentNode:
		case ElementNode:
			switch n.Data() {
			case "body":
				p.Mode = IM_inBody
			default:
				// TODO(jwall): parse error
				// inject a body tag first?
			}
		default:
			// TODO(jwall): parse error
		}
	case IM_inTable:
		fallthrough
	case IM_inTableText:
		fallthrough
	case IM_inCaption:
		fallthrough
	case IM_inColumnGroup:
		fallthrough
	case IM_inTableBody:
		fallthrough
	case IM_inRow:
		fallthrough
	case IM_inCell:
		fallthrough
	case IM_inSelect:
		fallthrough
	case IM_inSelectInTable:
		fallthrough
	case IM_afterBody:
		fallthrough
	case IM_inBody:
		switch n.Type {
		case DoctypeNode:
			// TODO(jwall): parse error
		case CommentNode:
		case ElementNode:
			switch n.Data() {
			case "script":
				//fmt.Println("In a script tag")
				p.Mode = IM_text
				return handleChar(startScriptDataState)
			default:
				// TODO(jwall): parse error
			}
		}
	case IM_text:
		//fmt.Println("parsing script contents. data:", n.Data())
		if n.Data() == "script" {
			//fmt.Println("setting insertionMode to inBody")
			p.Mode = IM_inBody
			popNode(p)
			return handleChar(dataStateHandler)
		}
		return handleChar(scriptDataStateHandler)
	case IM_afterFrameset:
		fallthrough
	case IM_afterAfterFrameset:
		fallthrough
	case IM_afterAfterBody:
		fallthrough
		// TODO(jwall): parse error
	}
	return handleChar(dataStateHandler)
}

func dataStateHandlerSwitch(p *Parser) stateHandler {
	n := p.curr
	/*fmt.Printf(
	"insertionMode: %v in dataStateHandlerSwitch with node: %v\n",
	p.Mode, n)*/
	return insertionModeSwitch(p, n)
}

// An html5 parsing struct. It holds the parsing state for the html5 parsing
// state machine.
type Parser struct {
	In   *bufio.Reader
	Top  *Node
	curr *Node
	c    *rune
	Mode InsertionMode
	buf  []rune // temporary buffer
}

type stateHandler func(p *Parser) (stateHandler, error)

// Construct a new h5 parser from a string
func NewParserFromString(s string) *Parser {
	return NewParser(strings.NewReader(s))
}

// Construct a new h5 parser from a io.Reader
func NewParser(r io.Reader) *Parser {
	return &Parser{In: bufio.NewReader(r)}
}

func (p *Parser) nextInput() (rune, error) {
	if p.c != nil {
		c := p.c
		p.c = nil
		//fmt.Printf("reread rune: %c\n", *c)
		return *c, nil
	}
	r, _, err := p.In.ReadRune()
	//fmt.Printf("rune: %c\n", r)
	return r, err
}

func (p *Parser) pushBack(c rune) {
	p.c = &c
}

// Parse an html stream.
// Returns an error if there was problem parsing the stream.
// The result of parsing can be retrieved with p.Tree()
func (p *Parser) Parse() error {
	// we start in the Doctype state
	// and in the Initial InsertionMode
	// start with a docType
	h := dataStateHandlerSwitch(p)
	for h != nil {
		//if p.curr != nil && p.curr.data != nil {
		//fmt.Printf("YYY: %v\n", p.curr.Data())
		//}
		h2, err := h(p)
		if err == io.EOF {
			//fmt.Println("End of file:")
			return nil
		}
		if err != nil {
			//fmt.Println("End of file: ", err)
			// TODO parse error
			return errors.New(fmt.Sprintf("Parse error: %s", err))
		}
		h = h2
	}
	return nil
}

// Return the parsed html5 tree or nil if parsing hasn't occured yet
func (p *Parser) Tree() *Node {
	return p.Top
}

// TODO(jwall): UNITTESTS!!!!
func textConsumer(p *Parser, chars ...rune) {
	if p.curr == nil {
		pushNode(p)
	}
	p.curr.data = append(p.curr.data, chars...) // ugly but safer
}

var memoized = make(map[func(*Parser, rune) stateHandler]stateHandler)

// TODO(jwall): UNITTESTS!!!!
func handleChar(h func(*Parser, rune) stateHandler) stateHandler {
	if f, ok := memoized[h]; ok {
		return f
	}
	memoized[h] = func(p *Parser) (stateHandler, error) {
		c, err := p.nextInput()
		if err != nil {
			return nil, err
		}
		//fmt.Printf("YYY: char %c\n", c)
		return h(p, c), nil
	}
	return memoized[h]
}

func startDoctypeStateHandler(p *Parser, c rune) stateHandler {
	//fmt.Printf("Starting Doctype handler c:%c\n", c)
	switch c {
	case '<':
		c2, err := p.nextInput()
		if err != nil {
			// correctly handle EOF
			return nil
		}
		if c2 == '!' { // its a doc type tag yay
			return handleChar(doctypeStateHandler)
		} else { // whoops not a doctype tag
			// TODO setup a default doctype
			// TODO we need a way to reconsume two characters :-(
			p.pushBack(c2)
			return dataStateHandler(p, c)
		}
	default:
		// TODO setup a default doctype
		return dataStateHandler(p, c)
	}
	panic("unreachable")
}

// Section 11.2.4.52
func doctypeStateHandler(p *Parser, c rune) stateHandler {
	//fmt.Printf("Parsing Doctype c:%c\n", c)
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(beforeDoctypeHandler)
	default:
		// TODO parse error
		// reconsume in BeforeDoctypeState
		return beforeDoctypeHandler(p, c)
	}
	panic("unreachable")
}

// Section 11.2.4.53
func beforeDoctypeHandler(p *Parser, c rune) stateHandler {
	curr := pushNode(p)
	curr.Type = DoctypeNode
	switch {
	case c == '\t', c == '\n', c == '\f', c == ' ':
		// ignore
		return handleChar(beforeDoctypeHandler)
	case c == '>':
		// TODO parse error, quirks mode
		return dataStateHandlerSwitch(p)
	case 'A' <= c && c <= 'Z':
		lc := c + 0x0020 // lowercase it
		curr.data = append(curr.data, lc)
		return handleChar(doctypeNameState)
	default:
		curr.data = append(curr.data, c)
		return handleChar(doctypeNameState)
	}
	panic("unreachable")
}

// Section 11.2.4.54
func doctypeNameState(p *Parser, c rune) stateHandler {
	n := p.curr
	switch {
	case c == '\t', c == '\n', c == '\f', c == ' ':
		// ignore
		return afterDoctypeNameHandler
	case c == '>':
		return afterDoctypeNameHandler
	case 'A' <= c && c <= 'Z':
		lc := c + 0x0020 // lowercase it
		n.data = append(n.data, lc)
		return handleChar(doctypeNameState)
	default:
		n.data = append(n.data, c)
		return handleChar(doctypeNameState)
	}
	panic("unreachable")
}

var (
	// The public doctype keyword constant
	public = "public"
	// The system doctype keyword constant
	system = "system"
)

// Section 11.2.4.55
func afterDoctypeNameHandler(p *Parser) (stateHandler, error) {
	firstSix := make([]rune, 0, 6)
	//n := p.curr
	for {
		c, err := p.nextInput()
		if err == io.EOF {
			// TODO parse error
			return dataStateHandlerSwitch(p), nil
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			// ignore
			return afterDoctypeNameHandler, nil
		case '>':
			return dataStateHandlerSwitch(p), nil
		default:
			if len(firstSix) == cap(firstSix) {
				switch string(firstSix) {
				case public:
					p.curr.Public = true
					return handleChar(afterDoctypeHandler), nil
				case system:
					p.curr.System = true
					return handleChar(afterDoctypeHandler), nil
				}
			} else {
				lc := c + 0x0020 // lowercase it
				firstSix = append(firstSix, lc)
			}
		}
	}
	panic("unreachable")
}

// Section 11.2.4.56
func afterDoctypeHandler(p *Parser, c rune) stateHandler {
	switch c {
	case '\t', '\n', '\f', ' ':
		// ignore
		return handleChar(beforeDoctypeIdentHandler)
	case '"', '\'':
		// TODO parse error
		return handleChar(makeIdentQuotedHandler(c))
	case '>':
		// TODO parse error
		return dataStateHandlerSwitch(p)
	default:
		// TODO parse error
		// TODO bogusDoctypeState
	}
	panic("unreachable")
}

// Section 11.2.4.57
func beforeDoctypeIdentHandler(p *Parser, c rune) stateHandler {
	switch c {
	case '\t', '\n', '\f', ' ':
		// ignore
		return handleChar(beforeDoctypeIdentHandler)
	case '"', '\'':
		return handleChar(makeIdentQuotedHandler(c))
	case '>':
		// TODO parse error
		return dataStateHandlerSwitch(p)
	default:
		// TODO parse error
		// TODO bogusDoctypeState
	}
	panic("unreachable")
}

// Section 11.2.4.58
func makeIdentQuotedHandler(q rune) func(*Parser, rune) stateHandler {
	return func(p *Parser, c rune) stateHandler {
		c2 := c
		for {
			if q == c2 {
				return handleChar(afterDoctypeIdentifierHandler)
			}
			if c2 == '>' {
				// TODO parse error
				return dataStateHandlerSwitch(p)
			}
			p.curr.Identifier = append(p.curr.Identifier, c2)
			next, err := p.nextInput()
			if err != nil {
				// TODO parse error
				return nil
			}
			c2 = next
		}
		panic("unreachable")
	}
}

// section 11.2.4.59
func afterDoctypeIdentifierHandler(p *Parser, c rune) stateHandler {
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(afterDoctypeIdentifierHandler)
	case '>':
		p.Mode = IM_beforeHtml
		return dataStateHandlerSwitch(p)
	default:
		// TODO parse error
		// TODO switch to bogus Doctype state
	}
	panic("unreachable")
}

func startScriptDataState(p *Parser, c rune) stateHandler {
	//fmt.Println("Adding TextNode")
	pushNode(p) // push a text node onto the stack
	return scriptDataStateHandler(p, c)
}

func scriptDataStateHandler(p *Parser, c rune) stateHandler {
	switch c {
	case '<':
		return handleChar(scriptDataLessThanHandler)
	default:
		textConsumer(p, c)
		for {
			c2, err := p.nextInput()
			if err != nil {
				// TODO parseError
				return nil
			}
			if c2 == '<' {
				return handleChar(scriptDataLessThanHandler)
			}
			textConsumer(p, c2)
		}
	}
	panic("unreachable")
}

func scriptDataLessThanHandler(p *Parser, c rune) stateHandler {
	//fmt.Printf("handling a '<' in script data c: %c\n", c)
	switch c {
	case '/':
		p.buf = make([]rune, 0, 1)
		return handleChar(scriptDataEndTagOpenHandler)
	default:
		textConsumer(p, '<', c)
		return handleChar(scriptDataStateHandler)
	}
	panic("unreachable")
}

func scriptDataEndTagOpenHandler(p *Parser, c rune) stateHandler {
	//fmt.Printf("trying to close script tag c: %c\n", c)
	switch {
	case 'A' <= c && c <= 'Z':
		lc := c + 0x0020 // lowercase it
		p.buf = append(p.buf, lc)
		return handleChar(scriptDataEndTagNameHandler)
	case 'a' <= c && c <= 'z':
		p.buf = append(p.buf, c)
		return handleChar(scriptDataEndTagNameHandler)
	default:
		textConsumer(p, '<', '/')
		return handleChar(scriptDataStateHandler)
	}
	panic("unreachable")
}

func scriptDataEndTagNameHandler(p *Parser, c rune) stateHandler {
	//fmt.Printf("script tag name handler c:%c\n", c)
	n := p.curr
	switch {
	case c == '\t', c == '\f', c == '\n', c == ' ':
		if n.Data() == string(p.buf) {
			return handleChar(beforeAttributeNameHandler)
		} else {
			p.buf = append(p.buf, c)
			return handleChar(scriptDataStateHandler)
		}
	case c == '/':
		if n.Parent.Data() == string(p.buf) {
			return handleChar(selfClosingTagStartHandler)
		} else {
			//fmt.Println("we don't match :-( keep going")
			return handleChar(scriptDataStateHandler)
		}
	case c == '>':
		if n.Parent.Data() == string(p.buf) {
			//fmt.Printf("time to see about closing it :-)")
			popNode(p)
			return dataStateHandlerSwitch(p)
		} else {
			//fmt.Println("we don't match :-( keep going")
			return handleChar(scriptDataStateHandler)
		}
	case 'A' <= c && c <= 'Z':
		lc := c + 0x0020 // lowercase it
		p.buf = append(p.buf, lc)
		return handleChar(scriptDataEndTagNameHandler)
	case 'a' <= c && c <= 'z':
		p.buf = append(p.buf, c)
		return handleChar(scriptDataEndTagNameHandler)
	default:
		textConsumer(p, '<', '/')
		textConsumer(p, p.buf...)
		return handleChar(scriptDataStateHandler)
	}
	panic("unreachable")
}

// Section 11.2.4.1
func dataStateHandler(p *Parser, c rune) stateHandler {
	//fmt.Printf("In dataStateHandler c:%c\n", c)
	//if p.curr != nil { fmt.Println("curr node: ", p.curr.Data()) }
	//fmt.Println("curr node textNode?",
	//	(p.curr != nil) && (p.curr.Type == TextNode))
	// consume the token
	if p.curr != nil {
		switch p.curr.Data() {
		case "base", "bgsound", "command", "link", "meta",
			"area", "br", "embed", "img", "keygen", "wbr",
			"param", "source", "track", "hr", "input", "image":
			popNode(p)
		}
		// this is the end of the textNode so pop it from stack
		//fmt.Println("TTT: popping textNode from stack")
		if p.curr.Type == TextNode {
			popNode(p)
		}
	}
	switch c {
	case '<':
		return handleChar(tagOpenHandler)
	default:
		pushNode(p)
		textConsumer(p, c)
		for {
			c2, err := p.nextInput()
			if err != nil {
				// TODO parseError
				return nil
			}
			if c2 == '<' { // for loop end condition
				return dataStateHandler(p, c2)
			}
			textConsumer(p, c2)
		}
	}
	panic("Unreachable")
}

func startHtmlCommentHandler(p *Parser) (stateHandler, error) {
	//fmt.Println("handling an html comment")
	d1, err := p.nextInput()
	if err != nil {
		// TODO parse error
		return nil, err
	}
	d2, err := p.nextInput()
	if err != nil {
		// TODO parse error
		return nil, err
	}
	if d1 == '-' && d2 == '-' {
		n := pushNode(p)
		n.Type = CommentNode
		return htmlCommentHandler, nil
	}
	return dataStateHandlerSwitch(p), nil
}

func htmlCommentHandler(p *Parser) (stateHandler, error) {
	n := p.curr
	for {
		c, err := p.nextInput()
		if err != nil {
			return nil, err
		}
		if c == '-' {
			return endHtmlCommentHandler, nil
		} else {
			n.data = append(n.data, c)
		}
	}
	return dataStateHandlerSwitch(p), nil
}

func endHtmlCommentHandler(p *Parser) (stateHandler, error) {
	c, err := p.nextInput()
	if err != nil {
		return nil, err
	}
	if c == '-' {
		c2, err := p.nextInput()
		if err != nil {
			return nil, err
		}
		if c2 == '>' { // close the comment
			popNode(p)
			return dataStateHandlerSwitch(p), nil
		} else { // still in comment
			return htmlCommentHandler, nil
		}
	}
	// still in a comment
	return htmlCommentHandler, nil
}

// Section 11.2.4.8
func tagOpenHandler(p *Parser, c rune) stateHandler {
	//fmt.Printf("opening a tag\n")
	switch {
	case c == '!': // markup declaration state
		return startHtmlCommentHandler
	case c == '/': // end tag open state
		return endTagOpenHandler
	case c == '?': // TODO parse error // bogus comment state
		return bogusCommentHandler
	case 'A' <= c && c <= 'Z':
		//fmt.Printf("ZZZ: opening a new tag\n")
		curr := pushNode(p)
		curr.Type = ElementNode
		lc := c + 0x0020 // lowercase it
		curr.data = []rune{lc}
		return handleChar(tagNameHandler)
	case 'a' <= c && c <= 'z':
		//fmt.Printf("ZZZ: opening a new tag\n")
		curr := pushNode(p)
		curr.Type = ElementNode
		curr.data = []rune{c}
		return handleChar(tagNameHandler)
	default: // parse error // recover using Section 11.2.4.8 rules
		// TODO
	}
	return nil
}

// Section 11.2.4.10
func tagNameHandler(p *Parser, c rune) stateHandler {
	n := p.curr
	// TODO(jwall): make this more efficient with a for loop
	switch {
	case c == '\t', c == '\n', c == '\f', c == ' ':
		return handleChar(beforeAttributeNameHandler)
	case c == '/':
		return handleChar(selfClosingTagStartHandler)
	case c == '>':
		return dataStateHandlerSwitch(p)
	case 'A' <= c && c <= 'Z':
		lc := c + 0x0020 // lowercase it
		n.data = append(n.data, lc)
		return handleChar(tagNameHandler)
	default:
		n.data = append(n.data, c)
		return handleChar(tagNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.34
func beforeAttributeNameHandler(p *Parser, c rune) stateHandler {
	n := p.curr
	switch {
	case c == '\t' || c == '\n' || c == '\f', c == ' ':
		// ignore
		return handleChar(beforeAttributeNameHandler)
	case c == '/':
		return handleChar(selfClosingTagStartHandler)
	case c == '>':
		return dataStateHandlerSwitch(p)
	case 'A' <= c && c <= 'Z':
		lc := c + 0x0020 // lowercase it
		newAttr := new(Attribute)
		newAttr.Name = string(lc)
		n.Attr = append(n.Attr, newAttr)
		return handleChar(attributeNameHandler)
	case c == '=', c == '"', c == '\'', c == '<':
		// TODO parse error
		fallthrough
	default:
		newAttr := new(Attribute)
		newAttr.Name = string(c)
		n.Attr = append(n.Attr, newAttr)
		return handleChar(attributeNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.35
func attributeNameHandler(p *Parser, c rune) stateHandler {
	n := p.curr
	switch {
	case c == '\t', c == '\n', c == '\f', c == ' ':
		return handleChar(afterAttributeNameHandler)
	case c == '/':
		return handleChar(selfClosingTagStartHandler)
	case c == '>':
		return dataStateHandlerSwitch(p)
	case c == '=':
		return handleChar(beforeAttributeValueHandler)
	case 'A' <= c && c <= 'Z':
		lc := c + 0x0020 // lowercase it
		currAttr := n.Attr[len(n.Attr)-1]
		currAttr.Name += string(lc)
		return handleChar(attributeNameHandler)
	case c == '"', c == '\'', c == '<':
		// TODO parse error
		fallthrough
	default:
		currAttr := n.Attr[len(n.Attr)-1]
		currAttr.Name += string(c)
		return handleChar(attributeNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.37
func beforeAttributeValueHandler(p *Parser, c rune) stateHandler {
	n := p.curr
	currAttr := n.Attr[len(n.Attr)-1]
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(beforeAttributeValueHandler)
	case '"', '\'':
		currAttr.quote = c
		return handleChar(makeAttributeValueQuotedHandler(c))
	case '>':
		return dataStateHandlerSwitch(p)
	case '<', '=', '`':
		// todo parse error
		fallthrough
	default:
		currAttr.Value += string(c)
		return handleChar(attributeValueUnquotedHandler)
	}
	panic("Unreachable")
}

var memoizedQuotedAttributeHandlers = make(map[rune]func(p *Parser, c rune) stateHandler)
// Section 11.2.4.3{8,9}
func makeAttributeValueQuotedHandler(c rune) func(p *Parser, c rune) stateHandler {
	if memoizedQuotedAttributeHandlers[c] != nil {
		return memoizedQuotedAttributeHandlers[c]
	}
	f := func(p *Parser, c2 rune) stateHandler {
		n := p.curr
		switch c2 {
		case '"', '\'':
			if c2 == c {
				return handleChar(afterAttributeValueQuotedHandler)
			}
			fallthrough
		default:
			currAttr := n.Attr[len(n.Attr)-1]
			currAttr.Value += string(c2)
			return handleChar(makeAttributeValueQuotedHandler(c))
		}
		panic("Unreachable")
	}
	memoizedQuotedAttributeHandlers[c] = f
	return f
}

// Section 11.2.4.40
func attributeValueUnquotedHandler(p *Parser, c rune) stateHandler {
	n := p.curr
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(beforeAttributeNameHandler)
	case '>':
		return dataStateHandlerSwitch(p)
	case '"', '\'', '<', '=', '`':
		// TODO parse error
		fallthrough
	default:
		currAttr := n.Attr[len(n.Attr)-1]
		currAttr.Value += string(c)
		return handleChar(attributeValueUnquotedHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.42
func afterAttributeValueQuotedHandler(p *Parser, c rune) stateHandler {
	switch c {
	case '\t', '\n', '\f', ' ':
		return handleChar(beforeAttributeNameHandler)
	case '/':
		return handleChar(selfClosingTagStartHandler)
	case '>':
		return dataStateHandlerSwitch(p)
	default:
		// TODO Parse error Reconsume the Character in the before attribute name state
		return handleChar(beforeAttributeNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.36
func afterAttributeNameHandler(p *Parser, c rune) stateHandler {
	n := p.curr
	switch {
	case c == '\t', c == '\n', c == '\f', c == ' ':
		return handleChar(afterAttributeNameHandler)
	case c == '/':
		return handleChar(selfClosingTagStartHandler)
	case c == '>':
		return dataStateHandlerSwitch(p)
	case c == '=':
		return handleChar(beforeAttributeValueHandler)
	case 'A' <= c && c <= 'Z':
		lc := c + 0x0020 // lowercase it
		newAttr := new(Attribute)
		newAttr.Name = string(lc)
		n.Attr = append(n.Attr, newAttr)
		return handleChar(attributeNameHandler)
	case c == '"', c == '\'', c == '<':
		// TODO parse error
		fallthrough
	default:
		newAttr := new(Attribute)
		newAttr.Name = string(c)
		n.Attr = append(n.Attr, newAttr)
		return handleChar(attributeNameHandler)
	}
	panic("Unreachable")
}

// Section 11.2.4.43
func selfClosingTagStartHandler(p *Parser, c rune) stateHandler {
	//fmt.Println("starting self closing tag handler")
	switch c {
	case '>':
		popNode(p)
		return dataStateHandlerSwitch(p)
	default:
		// TODO parse error reconsume as before attribute handler
		return handleChar(beforeAttributeNameHandler)
	}
	panic("Unreachable")
}

func newEndTagError(problem string, n *Node, tag []rune) error {
	msg := fmt.Sprintf(
		"%s: End Tag does not match Start Tag start:[%s] end:[%s]",
		problem, n.Data(), string(tag))
	//fmt.Println(msg)
	return NewParseError(n, msg)
}

func genImpliedEndTags(p *Parser) {
	for {
		switch p.curr.Data() {
		case "dd", "dt", "li", "option", "optgroup", "p", "rp", "rt":
			//fmt.Println("Found an implied end tag", p.curr.Data())
			popNode(p)
		default:
			return
		}
	}
	return
}

// Section 11.2.4.9
func endTagOpenHandler(p *Parser) (stateHandler, error) {
	// compare to current tags name
	//fmt.Println("YYY: attempting to close a node")
	n := p.curr
	tag := make([]rune, 0, len(n.data))
	for {
		c, err := p.nextInput()
		if err == io.EOF { // Parse Error
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		switch {
		case c == '>':
			// TODO tests for this
			switch string(tag) { // quirks mode case
			case "base", "bgsound", "command", "link", "meta",
				"area", "br", "embed", "img", "keygen", "wbr",
				"param", "source", "track", "hr", "input", "image":
				return dataStateHandlerSwitch(p), nil
			case "address", "article", "aside", "blockquote", "button",
				"center", "details", "dir", "div", "dl", "fieldset",
				"figcaption", "figure", "footer", "header", "hgroup",
				"listing", "menu", "nav", "ol", "pre", "section", "summary",
				"ul", "td", "th", "font":
				// generate implied end tags
				genImpliedEndTags(p)
				// reset the current node
				n = p.curr
			}
			if string(n.data) != string(tag) {
				return nil, newEndTagError("NotSameTag", n, tag)
			}
			//fmt.Println("YYY: closing a tag")
			popNode(p)
			return dataStateHandlerSwitch(p), nil
		case 'A' <= c && c <= 'Z':
			lc := c + 0x0020 // lowercase it
			tag = append(tag, lc)
		case 'a' <= c && c <= 'z', '0' <= c && c <= '9':
			tag = append(tag, c)
		default: // Bogus Comment state
			tag = append(tag, c)
			return bogusCommentHandler, NewParseError(n,
				"Strange characters in end tag: [%c] switching to BogusCommentState", c)
		}
	}
	panic("Unreachable")
}

// Section 11.2.4.44
func bogusCommentHandler(p *Parser) (stateHandler, error) {
	n := addSibling(p)
	for {
		c, err := p.nextInput()
		if err != nil {
			return nil, err
		}
		switch c {
		case '>':
			return dataStateHandlerSwitch(p), nil
		default:
			n.data = append(n.data, c)
		}
	}
	panic("Unreachable")
}

func addSibling(p *Parser) *Node {
	//fmt.Printf("adding sibling to: %s\n", p.curr.Data())
	n := new(Node)
	p.curr.Parent.Children = append(p.curr.Parent.Children, n)
	p.curr = n
	return n
}

func pushNode(p *Parser) *Node {
	n := new(Node)
	if p.Top == nil {
		p.Top = n
	}
	if p.curr == nil {
		p.curr = n
	} else {
		//fmt.Printf("pushing child onto curr node: %s\n", p.curr.Data())
		n.Parent = p.curr
		n.Parent.Children = append(n.Parent.Children, n)
		p.curr = n
	}
	return n
}

func popNode(p *Parser) *Node {
	if p.curr != nil && p.curr.Parent != nil {
		//fmt.Printf("popping node: %s\n", p.curr.Data())
		p.curr = p.curr.Parent
		//fmt.Printf("curr node: %s\n", p.curr.Data())
	}
	return p.curr
}

// Copyright 2011 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.
