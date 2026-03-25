package space

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/text"
)

func TestVNodeNewDefaults(t *testing.T) {
	node := New()
	if node.Tag() != "space" {
		t.Fatalf("Tag = %q, want space", node.Tag())
	}
	if node.direction != DirectionHorizontal {
		t.Fatalf("direction = %q, want horizontal", node.direction)
	}
	if node.size != SizeSmall {
		t.Fatalf("size = %d, want %d", node.size, SizeSmall)
	}
}

func TestBuilderFluentAPI(t *testing.T) {
	child := text.New("A")
	node := NewBuilder().
		Key("toolbar-space").
		Vertical().
		Large().
		Wrap(true).
		Width(42).
		Align(AlignCenter).
		Split("|").
		Children(child).
		BuildVNode()

	if node.Key() != "toolbar-space" {
		t.Fatalf("Key = %q, want toolbar-space", node.Key())
	}
	if node.direction != DirectionVertical {
		t.Fatalf("direction = %q, want vertical", node.direction)
	}
	if node.size != SizeLarge {
		t.Fatalf("size = %d, want %d", node.size, SizeLarge)
	}
	if !node.wrap || node.width != 42 || node.align != AlignCenter || node.split != "|" {
		t.Fatalf("unexpected builder state: wrap=%v width=%d align=%v split=%q", node.wrap, node.width, node.align, node.split)
	}
	if children := node.Children(); len(children) != 0 {
		t.Fatalf("VNode.Children len = %d, want 0 so runtime children drive rendering", len(children))
	}
	if propsChildren, ok := node.Props()[propChildren].([]rtui.VNode); !ok || len(propsChildren) != 1 {
		t.Fatalf("props children = %v, want 1 child preserved in props", node.Props()[propChildren])
	}
}

func TestInstanceRuntimeChildrenHorizontal(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propKey:      "toolbar",
		propChildren: []rtui.VNode{text.New("A"), text.New("B")},
		propSize:     2,
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	if children[0].Tag() != "hstack" {
		t.Fatalf("root tag = %q, want hstack", children[0].Tag())
	}
	if gap, ok := children[0].Props()["gap"].(int); !ok || gap != 2 {
		t.Fatalf("gap = %v, want 2", children[0].Props()["gap"])
	}
	if len(children[0].Children()) != 2 {
		t.Fatalf("root children len = %d, want 2", len(children[0].Children()))
	}
}

func TestInstanceRuntimeChildrenVerticalWithSplit(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propDirection: DirectionVertical,
		propSplit:     "|",
		propChildren:  []rtui.VNode{text.New("A"), text.New("B"), text.New("C")},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "vstack" {
		t.Fatalf("root tag = %q, want vstack", root.Tag())
	}
	if len(root.Children()) != 5 {
		t.Fatalf("root children len = %d, want 5", len(root.Children()))
	}
}

func TestInstanceRuntimeChildrenWrapUsesWrapComponent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propWrap:     true,
		propWidth:    24,
		propChildren: []rtui.VNode{text.New("A"), text.New("B")},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "wrap" {
		t.Fatalf("root tag = %q, want wrap", root.Tag())
	}
	if width, ok := root.Props()["width"].(int); !ok || width != 24 {
		t.Fatalf("width = %v, want 24", root.Props()["width"])
	}
}

func TestInstanceSetPropsTracksChanges(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propChildren: []rtui.VNode{text.New("A")},
	})
	inst.dirty = false

	if !inst.SetProps(rtui.Props{
		propChildren: []rtui.VNode{text.New("A"), text.New("B")},
	}) {
		t.Fatal("SetProps should report change")
	}
	if !inst.dirty {
		t.Fatal("dirty should be true after prop change")
	}
}
