/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.
*/
package transform

import (
    . "h5"
    "testing"
)

func TestNewTransform(t *testing.T) {
    doc, _ := NewDoc("<html><body><div id=\"foo\"></div></body></html")
    tf := NewTransform(doc)
    // hacky way of comparing an uncomparable type
    assertEqual(t, (*tf.doc).Type, (*doc).Type)
}

func TestTransformApply(t *testing.T) {
    doc, _ := NewDoc("<html><body><div id=\"foo\"></div></body></html")
    tf := NewTransform(doc)
    newDoc := tf.Apply(AppendChildren(new(Node)), "body").Doc()
    assertEqual(t, len(newDoc.Children[0].Children), 2)
}

func TestTransformApplyMulti(t *testing.T) {
    doc, _ := NewDoc("<html><body><div id=\"foo\"></div></body></html")
    tf := NewTransform(doc)
    tf.Apply(AppendChildren(new(Node)), "body")
    newDoc := tf.Apply(TransformAttrib("id", func(val string) string {
        t.Logf("Rewriting Url")
        return "bar"
    }),
        "div").Doc()
    assertEqual(t, len(newDoc.Children[0].Children), 2)
    assertEqual(t, newDoc.Children[0].Children[0].Attr[0].Value,
        "bar")
}

func TestAppendChildren(t *testing.T) {
    node, _ := NewDoc("<div id=\"foo\"></div><")
    child := new(Node)
    child2 := new(Node)
    f := AppendChildren(child, child2)
    f(node)
    assertEqual(t, len(node.Children), 2)
    assertEqual(t, node.Children[0], child)
    assertEqual(t, node.Children[1], child2)
}

func TestRemoveChildren(t *testing.T) {
    doc, _ := NewDoc("<div id=\"foo\">foo</div>")
    node := doc.Children[0]
    f := RemoveChildren()
    f(node)
    assertEqual(t, len(node.Children), 0)
}

func TestReplaceChildren(t *testing.T) {
    doc, _ := NewDoc("<div id=\"foo\">foo</div>")
    node := doc.Children[0]
    child := new(Node)
    child2 := new(Node)
    f := ReplaceChildren(child, child2)
    f(node)
    assertEqual(t, len(node.Children), 2)
    assertEqual(t, node.Children[0], child)
    assertEqual(t, node.Children[1], child2)
}

func TestReplace(t *testing.T) {
    defer func() {
        if err := recover(); err != nil {
            t.Error("TestReplace paniced")
        }
    }()
    doc, _ := NewDoc("<div id=\"foo\">foo</div><")
    node := doc.Children[0]
    ns, _ := NewDoc("<span>foo</span>")
    f := Replace(ns)
    f(node)
    assertEqual(t, len(doc.Children), 1)
    assertEqual(t, doc.Children[0].Data(), "span")
}

func TestReplaceSplice(t *testing.T) {
    defer func() {
        if err := recover(); err != nil {
            t.Error("TestReplaceSplice paniced")
        }
    }()
    doc, _ := NewDoc("<div id=\"foo\">foo<span>bar</span></div><")
    node := doc.Children[0]
    ns, _ := NewDoc("<span>foo</span>")
    f := Replace(ns)
    f(node)
    assertEqual(t, len(doc.Children), 2)
    assertEqual(t, doc.Children[0].Data(), "span")
    assertEqual(t, doc.Children[0].Children[0].Data(), "foo")
    assertEqual(t, doc.Children[1].Children[0].Data(), "bar")
}

func TestModifyAttrib(t *testing.T) {
    node, _ := NewDoc("<div id=\"foo\">foo</div><")
    assertEqual(t, node.Attr[0].Value, "foo")
    f := ModifyAttrib("id", "bar")
    f(node)
    assertEqual(t, node.Attr[0].Value, "bar")
    f = ModifyAttrib("class", "baz")
    f(node)
    assertEqual(t, node.Attr[1].Name, "class")
    assertEqual(t, node.Attr[1].Value, "baz")
}

func TestTransformAttrib(t *testing.T) {
    node, _ := NewDoc("<div id=\"foo\">foo</div><")
    assertEqual(t, node.Attr[0].Value, "foo")
    f := TransformAttrib("id", func(s string) string { return "bar" })
    f(node)
    assertEqual(t, node.Attr[0].Value, "bar")
}

func TestDoAll(t *testing.T) {
    node, _ := NewDoc("<div id=\"foo\">foo</div><")
    preNode := Text("pre node")
    postNode := Text("post node")
    f := DoAll(AppendChildren(postNode),
        PrependChildren(preNode))
    f(node)
    assertEqual(t, len(node.Children), 3)
    assertEqual(t, node.Children[0].Data(), preNode.Data())
    assertEqual(t, node.Children[len(node.Children)-1].Data(), postNode.Data())
}

func TestForEach(t *testing.T) {
    node, _ := NewDoc("<div id=\"foo\">foo</div><")
    txtNode1 := Text(" bar")
    txtNode2 := Text(" baz")
    f := ForEach(AppendChildren, txtNode1, txtNode2)
    f(node)
    assertEqual(t, len(node.Children), 3)
    assertEqual(t, node.Children[1].Data(), txtNode1.Data())
    assertEqual(t, node.Children[2].Data(), txtNode2.Data())
}

func TestForEachSingleArgFuncs(t *testing.T) {
    node, _ := NewDoc("<div id=\"foo\">foo</div><")
    txtNode1 := Text(" bar")
    txtNode2 := Text(" baz")
    singleArgFun := func(n *Node) TransformFunc {
        return AppendChildren(n)
    }
    f := ForEach(singleArgFun, txtNode1, txtNode2)
    f(node)
    assertEqual(t, len(node.Children), 3)
    assertEqual(t, node.Children[1].Data(), txtNode1.Data())
    assertEqual(t, node.Children[2].Data(), txtNode2.Data())
}

func TestForEachPanic(t *testing.T) {
    defer func() {
        if err := recover(); err == nil {
            t.Error("ForEach Failed to panic")
        }
    }()
    txtNode1 := Text(" bar")
    txtNode2 := Text(" baz")
    ForEach("foo", txtNode1, txtNode2)
}

func TestCopyAnd(t *testing.T) {
    defer func() {
        if err := recover(); err != nil {
            t.Error("TestCopyAnd paniced %s", err)
        }
    }()
    ul, _ := NewDoc("<ul><li class=\"item\">item1</li></ul>")
    node := ul.Children[0]
    fn1 := func(n *Node) {
        n.Children[0].SetData([]int("foo"))
    }
    fn2 := func(n *Node) {
        n.Children[0].SetData([]int("bar"))
    }
    f := CopyAnd(fn1, fn2)

    assertEqual(t, len(ul.Children), 1)
    f(node)
    assertEqual(t, len(ul.Children), 2)
    assertEqual(t, ul.Children[0].Data(), "li")
    assertEqual(t, ul.Children[0].Attr[0].Name, "class")
    assertEqual(t, ul.Children[0].Attr[0].Value, "item")
    assertEqual(t, ul.Children[0].Children[0].Data(), "foo")

    assertEqual(t, ul.Children[1].Data(), "li")
    assertEqual(t, ul.Children[1].Attr[0].Name, "class")
    assertEqual(t, ul.Children[1].Attr[0].Value, "item")
    assertEqual(t, ul.Children[1].Children[0].Data(), "bar")
}

// TODO(jwall): benchmarking tests
