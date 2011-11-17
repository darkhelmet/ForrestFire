// Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
// Use of this source code is governed by the Artistic License 2.0.
// That License is included in the LICENSE file.
/*

The html transform package implements a html css selector and transformer.

An html doc can be inspected and queried using css selectors as well as
transformed.

 	doc := NewDoc(str)
 	sel1 := NewSelector("li.menuitem")
 	sel2 := NewSelector("a")
	t := NewTransform(doc)
 	t.Apply(CopyAnd(myModifiers...), sel1)
  t..Apply(Replace(Text("my new text"), sel2)
  newDoc := t.Doc()
*/
package transform

// TODO(jwall): Documentation...
import (
	. "h5"
	"log"
)

// The TransformFunc type is the type of a Node transformation function.
type TransformFunc func(*Node)

// Transformer encapsulates a document under transformation.
type Transformer struct {
	doc *Node
}

// Constructor for a Transformer. It makes a copy of the document
// and transforms that instead of the original.
func NewTransform(d *Node) *Transformer {
	return &Transformer{doc: d.Clone()}
}

// The Doc method returns the document under transformation.
func (t *Transformer) Doc() *Node {
	return t.doc
}

func (t *Transformer) String() string {
	return t.doc.String()
}

// The Apply method applies a TransformFunc to the nodes returned from
// the Selector query
func (t *Transformer) Apply(f TransformFunc, sel ...string) *Transformer {
	// TODO come up with a way to walk tree once?
	sq := NewSelectorQuery(sel...)
	nodes := sq.Apply(t.doc)
	for _, n := range nodes {
		f(n)
	}
	return t
}

// Compose a set of TransformFuncs into a single TransformFunc
func Compose(fs ...TransformFunc) TransformFunc {
	return func(n *Node) {
		for _, f := range fs {
			f(n)
		}
	}
}

// AppendChildren creates a TransformFunc that appends the Children passed in.
func AppendChildren(cs ...*Node) TransformFunc {
	return func(n *Node) {
		sz := len(n.Children)
		newChild := make([]*Node, sz+len(cs))
		copy(newChild, n.Children)
		copy(newChild[sz:], cs)
		n.Children = newChild
	}
}

// PrependChildren creates a TransformFunc that prepends the Children passed in.
func PrependChildren(cs ...*Node) TransformFunc {
	return func(n *Node) {
		sz := len(n.Children)
		sz2 := len(cs)
		newChild := make([]*Node, sz+len(cs))
		copy(newChild[sz2:], n.Children)
		copy(newChild[0:sz2], cs)
		n.Children = newChild
	}
}

// RemoveChildren creates a TransformFunc that removes the Children of the node
// it operates on.
func RemoveChildren() TransformFunc {
	return func(n *Node) {
		n.Children = make([]*Node, 0)
	}
}

// ReplaceChildren creates a TransformFunc that replaces the Children of the
// node it operates on with the Children passed in.
func ReplaceChildren(ns ...*Node) TransformFunc {
	return func(n *Node) {
		n.Children = ns
	}
}

func Replace(ns ...*Node) TransformFunc {
	return func(n *Node) {
		p := n.Parent
		// TODO(jwall): splice the new nodes into the spot the current node is
		for i, c := range p.Children {
			if c == n {
				n := i - 1
				if n < 0 {
					n = 0
				}
				var newChild []*Node
				pre := p.Children[:n]
				post := p.Children[i+1:]
				newChild = append(pre, ns...)
				p.Children = append(newChild, post...)
			}
		}
	}
}

// ModifyAttrb creates a TransformFunc that modifies the attributes
// of the node it operates on.
func ModifyAttrib(key string, val string) TransformFunc {
	return func(n *Node) {
		found := false
		for i, attr := range n.Attr {
			if attr.Name == key {
				n.Attr[i].Value = val
				found = true
			}
		}
		if !found {
			newAttr := make([]*Attribute, len(n.Attr)+1)
			newAttr[len(n.Attr)] = &Attribute{Name: key, Value: val}
			n.Attr = newAttr
		}
	}
}

func TransformAttrib(key string, f func(string) string) TransformFunc {
	return func(n *Node) {
		for i, attr := range n.Attr {
			if attr.Name == key {
				n.Attr[i].Value = f(n.Attr[i].Value)
			}
		}
	}
}

func DoAll(fs ...TransformFunc) TransformFunc {
	return func(n *Node) {
		for _, f := range fs {
			f(n)
		}
	}
}

// ForEach takes a function and a list of Nodes and performs that
// function for each node in the list.
// The function should be of a type either func(...*Node) TransformFunc
// or func(*Node) TransformFunc. Any other type will panic.
// Returns a TransformFunc.
func ForEach(f interface{}, ns ...*Node) TransformFunc {
	switch t := f.(type) {
	case func(...*Node) TransformFunc:
		return func(n *Node) {
			for _, n2 := range ns {
				f1 := f.(func(...*Node) TransformFunc)
				f2 := f1(n2)
				f2(n)
			}
		}
	case func(*Node) TransformFunc:
		return func(n *Node) {
			for _, n2 := range ns {
				f1 := f.(func(*Node) TransformFunc)
				f2 := f1(n2)
				f2(n)
			}
		}
	default:
		log.Panicf("Wrong function type passed to ForEach %s", t)
	}
	return nil
}

// CopyAnd will construct a TransformFunc that will
// make a copy of the node for each passed in TransformFunc
// And replace the passed in node with the resulting transformed
// Nodes.
// Returns a TransformFunc
func CopyAnd(fns ...TransformFunc) TransformFunc {
	return func(n *Node) {
		newNodes := make([]*Node, len(fns))
		for i, fn := range fns {
			node := n.Clone()
			fn(node)
			newNodes[i] = node
		}
		replaceFn := Replace(newNodes...)
		replaceFn(n)
	}
}
