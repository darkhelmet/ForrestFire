/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
	"testing"
	. "h5"
)

func assertEqual(t *testing.T, val interface{}, expected interface{}) {
	if val != expected {
		t.Errorf("NotEqual Expected: [%s] Actual: [%s]",
			expected, val)
	}
}

func assertNotNil(t *testing.T, val interface{}) {
	if val == nil {
		t.Errorf("Value is Nil")
	}
}

func nodeTree() *Node {
	n, _ := NewDoc("<html><head /><body /></html>")
	return n
}

func TestNewDoc(t *testing.T) {
	docStr := "<a>foo</a>"
	node, _ := NewDoc(docStr)
	assertEqual(t, node.Children[0].Parent, node)
	assertEqual(t, node.Data(), "a")
	assertEqual(t, len(node.Children), 1)
	assertEqual(t, node.Type, ElementNode)
	assertEqual(t, node.Children[0].Data(), "foo")
	assertEqual(t, len(node.Children), 1)
	assertEqual(t, node.Children[0].Type, TextNode)
}
