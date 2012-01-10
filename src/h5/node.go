package h5

import (
    "fmt"
)

// The type of a html5 nodes attributes
type Attribute struct {
    Name  string
    Value string
    // TODO for gob this should be public field
    quote rune
}

// Serialize an html5 attribute to a string
func (a *Attribute) String() string {
    // TODO handle different quoting styles.
    c := '"'
    if a.quote != 0 {
        c = a.quote
    }
    return fmt.Sprintf(
        "%s=%c%s%c", a.Name, c, a.Value, c)
}

// Clone an html5 attribute
func (a *Attribute) Clone() *Attribute {
    return &Attribute{Name: a.Name, Value: a.Value}
}

// Represents the type of an html5 node
type NodeType int

const (
    TextNode    NodeType = iota // zero value so the default
    ElementNode NodeType = iota
    DoctypeNode NodeType = iota
    CommentNode NodeType = iota
)

// The type of an html5 node
type Node struct {
    Type NodeType // The type of node this is.
    // TODO for gob this should be a public field
    data       []rune
    Attr       []*Attribute // The attributes of the html5 node
    Parent     *Node        // The parent of the html5 node if it has one, nil otherwise
    Children   []*Node      // The children of the html5 node if it has any.
    Public     bool         // True if this is a PUBLIC doctype node
    System     bool         // True if this is a SYSTEM doctype node
    Identifier []rune       // The identifier if this is a doctype node
}

// Sets a Nodes data. (eg: The Tagname for ElementNodes or text for TextNodes)
func (n *Node) SetData(rs []rune) {
    n.data = rs
}

func attrString(attrs []*Attribute) string {
    if attrs == nil {
        return ""
    }
    s := ""
    for _, a := range attrs {
        s += fmt.Sprintf(" %s", a)
    }
    return s
}

func doctypeString(n *Node) string {
    keyword := ""
    identifier := string(n.Identifier)
    switch {
    case n.Public:
        keyword = "PUBLIC"
    case n.System:
        keyword = "SYSTEM"
    default:
        return "<!DOCTYPE html>"
    }
    return fmt.Sprintf("<!DOCTYPE %s=\"%s\">", keyword, identifier)
}

func commentString(n *Node) string {
    return fmt.Sprintf("<!--%s-->", n.Data())
}

// Serialize an html5 node to a string.
func (n *Node) String() string {
    switch n.Type {
    case TextNode:
        return n.Data()
    case ElementNode:
        // TODO handle the strange self close tags
        if n.Children == nil || len(n.Children) == 0 {
            name := n.Data()
            switch name {
            case "textarea":
                return "<textarea" + attrString(n.Attr) + "></textarea>"
            }
            return "<" + n.Data() + attrString(n.Attr) + "/>"
        } else {
            s := "<" + n.Data() + attrString(n.Attr) + ">"
            for _, n := range n.Children {
                s += n.String()
            }
            s += "</" + n.Data() + ">"
            return s
        }
    case DoctypeNode:
        // TODO Doctype stringification
        s := doctypeString(n)
        for _, n := range n.Children {
            s += n.String()
        }
        return s
    case CommentNode:
        return commentString(n)
    }
    return ""
}

// Walk a Node tree with a given function.
func (n *Node) Walk(f func(*Node)) {
    f(n)
    if len(n.Children) > 0 {
        for _, c := range n.Children {
            c.Walk(f)
        }
    }
}

func cloneNode(n, p *Node) *Node {
    clone := new(Node)
    clone.data = make([]rune, len(n.data))
    clone.Attr = make([]*Attribute, len(n.Attr))
    clone.Children = make([]*Node, len(n.Children))
    clone.Parent = p
    clone.Type = n.Type
    clone.Public = n.Public
    clone.System = n.System
    clone.Identifier = n.Identifier
    copy(clone.data, n.data)
    for i, a := range n.Attr {
        clone.Attr[i] = a.Clone()
    }
    if len(n.Children) > 0 {
        for i, c := range n.Children {
            clone.Children[i] = cloneNode(c, n)
        }
    }
    return clone
}

// Clone an html5 nodetree to get a copy.
func (n *Node) Clone() *Node {
    return cloneNode(n, n.Parent)
}

// String form of an html nodes data.
// (eg: The Tagname for ElementNodes or text for TextNodes)
func (n *Node) Data() string {
    if n.data != nil {
        return string(n.data)
    }
    return ""
}

// Construct a TextNode
func Text(str string) *Node {
    return &Node{data: []rune(str)}
}

// TODO Constructors for other html node types.

// Copyright 2011 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.
