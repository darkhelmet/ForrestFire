package transform

import (
    . "h5"
    "log"
    s "strings"
)

// SelectorQuery is the type of a CSS Selector Query.
// Each Selector in the slice is operated on order with
// subsequent selectors matching against the descendants
// of the previous selectors match.
type SelectorQuery []*Selector

// SelectorPart is the type of a single Selector's Class, ID or Pseudo part.
type SelectorPart struct {
    Type byte   // a bitmask of the selector types
    Val  string // the value we are matching against
}

// Selector is the type of a single selector in a selector query.
// A slice of Selectors makes up a SelectorQuery.
type Selector struct {
    TagName string // "*" for any tag otherwise the name of the tag
    Parts   []SelectorPart
    Attr    map[string]string
}

const (
    TAGNAME byte = iota // Tagname Selector Type
    CLASS   byte = '.'  // Class SelectorPart Type
    ID      byte = '#'  // Id  SelectorPart Type
    PSEUDO  byte = ':'  // Pseudo SelectoPart Type
    ANY     byte = '*'  // Any tag Selector Type
    ATTR    byte = '['  // Attr Selector Type
)

const (
    // Important characters in a Selector string
    SELECTOR_CHARS string = ".:#["
)

func matchAttrib(nodeAttr []*Attribute, matchAttr map[string]string) bool {
    attribResult := true
    for key, val := range matchAttr {
        exists := false
        matched := false
        for _, attr := range nodeAttr {
            if key == attr.Name {
                exists = true
                if val == attr.Value {
                    matched = true
                }
                attribResult = attribResult && exists && matched
            }
        }
        attribResult = attribResult && exists
    }
    return attribResult
}

func (part SelectorPart) match(node *Node) bool {
    switch part.Type {
    case CLASS:
        classAttr := make(map[string]string)
        classAttr["class"] = part.Val
        return matchAttrib(node.Attr, classAttr)
    case ID:
        idAttr := make(map[string]string)
        idAttr["id"] = part.Val
        return matchAttrib(node.Attr, idAttr)
    case PSEUDO:
    }
    return false
}

// The Match method tests if a Selector matches a Node.
// Returns true for a match false otherwise.
func (sel *Selector) Match(node *Node) bool {
    tagNameResult := true
    if sel.TagName != "" && sel.TagName != "*" && sel.TagName != node.Data() {
        tagNameResult = tagNameResult && false
    }
    attribResult := matchAttrib(node.Attr, sel.Attr)
    partsResult := true
    for _, part := range sel.Parts {
        partsResult = partsResult && part.match(node)
    }
    return partsResult && tagNameResult && attribResult
}

func newAnyTagClassOrIdSelector(str string) *Selector {
    return &Selector{
        Parts: []SelectorPart{
            SelectorPart{
                Type: str[0],
                Val:  str[1:],
            }},
        TagName: "*",
    }
}

func newAnyTagSelector(str string) *Selector {
    return &Selector{
        TagName: "*",
    }
}

func splitAttrs(str string) []string {
    attrs := s.FieldsFunc(str[1:len(str)-1], func(c rune) bool {
        if c == '=' {
            return true
        }
        return false
    })
    return attrs
}

func newAnyTagAttrSelector(str string) *Selector {
    parts := s.SplitAfter(str, "]")
    sel := Selector{
        TagName: "*",
        Attr:    map[string]string{},
    }
    for _, part := range parts[0 : len(parts)-1] {
        attrs := splitAttrs(part)
        sel.Attr[attrs[0]] = attrs[1]
    }
    return &sel
}

func newTagNameSelector(str string) *Selector {
    return &Selector{
        TagName: str,
    }
}

func newTagNameWithConstraints(str string, i int) *Selector {
    // TODO(jwall): indexAny use [CLASS,...]
    var selector = new(Selector)
    switch str[i] {
    case CLASS, ID, PSEUDO: // with class or id
        selector = newAnyTagClassOrIdSelector(str[i:])
    case ATTR: // with attribute
        selector = newAnyTagAttrSelector(str[i:])
    default:
        panic("Invalid constraint type for the tagname selector")
    }
    selector.TagName = str[0:i]
    //selector.Type = TAGNAME
    return selector
}

func partition(s string, f func(c rune) bool) []string {
    parts := []string{}
    start := 0
    for i, char := range s {
        if i < 1 { // we don't want empty first partitions
            continue
        }
        if f(char) {
            parts = append(parts, s[start:i])
            start = i
        }
    }
    parts = append(parts, s[start:])
    return parts
}

// MergeSelectors merges two *Selector types into one
// *Selector. It merges the second selector into the first
// modifying the first selector.
func MergeSelectors(sel1 *Selector, sel2 *Selector) {
    if sel2.TagName != "" && sel2.TagName != "*" {
        if sel1.TagName == "" || sel1.TagName == "*" {
            sel1.TagName = sel2.TagName
        } else {
            log.Panicf("Can't merge multiple tagnames in"+
                " selectors: First [%s] Second [%s]",
                sel1.TagName, sel2.TagName)
        }
    }
    if sel1.Parts == nil {
        sel1.Parts = make([]SelectorPart, 0)
    }
    newParts := make([]SelectorPart, len(sel1.Parts)+len(sel2.Parts))
    copy(newParts, sel1.Parts)
    copy(newParts[len(sel1.Parts):], sel2.Parts)
    sel1.Parts = newParts
    if sel1.Attr == nil {
        sel1.Attr = make(map[string]string)
    }
    for key, val := range sel2.Attr {
        sel1.Attr[key] = val
    }
}

// NewSelector is a constructor for a *Selector type.
// It creates a Selector by parsing the string passed in.
func NewSelector(str string) *Selector {
    str = s.TrimSpace(str) // trim whitespace
    // TODO(jwall): support combinators > + \S
    parts := partition(str, func(c rune) bool {
        for _, c2 := range SELECTOR_CHARS {
            if c == c2 {
                return true
            }
        }
        return false
    })
    result := new(Selector)
    result.TagName = "*"
    for _, p := range parts {
        selector := new(Selector)
        switch p[0] {
        case CLASS, ID: // Any tagname with class or id
            selector = newAnyTagClassOrIdSelector(p)
        case ANY: // Any tagname
            selector = newAnyTagSelector(p)
        case ATTR: // any tagname with attribute
            selector = newAnyTagAttrSelector(p)
        default: // TAGNAME
            if i := s.IndexAny(p, SELECTOR_CHARS); i != -1 {
                selector = newTagNameWithConstraints(p, i)
            } else { // just a tagname
                selector = newTagNameSelector(p)
            }
        }
        MergeSelectors(result, selector)
    }
    return result
}

// NewSelectorQuery is a constructor for a SelectorQuery
// It creates the query using the strings passed in.
func NewSelectorQuery(sel ...string) SelectorQuery {
    q := make([]*Selector, len(sel))
    for i, str := range sel {
        selPart := NewSelector(str)
        if selPart == nil {
            log.Panic("Invalid Selector in query")
        }
        q[i] = selPart
    }
    return q
}

func applyToNode(sel []*Selector, n *Node) []*Node {
    var nodes []*Node
    if sel[0].Match(n) {
        if len(sel) == 1 {
            nodes = []*Node{n}
        } else {
            for _, c := range n.Children {
                if len(sel) > 1 {
                    ns := applyToNode(sel[1:], c)
                    if len(ns) > 0 {
                        nodes = append(nodes, ns...)
                    }
                } else {
                    nodes = []*Node{n}
                }
            }
        }
    }
    return nodes
}

// Apply the css selector to a document.
// Returns a Vector of nodes that the selector matched.
func (sel SelectorQuery) Apply(doc *Node) []*Node {
    interesting := []*Node{}
    f := func(n *Node) {
        found := applyToNode(sel, n)
        interesting = append(interesting, found...)
    }
    doc.Walk(f)
    return interesting
}

// Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.
