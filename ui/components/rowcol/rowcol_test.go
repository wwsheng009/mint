package rowcol

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/text"
)

func TestNewRowDefaults(t *testing.T) {
	row := NewRow()
	if row.Tag() != "row" {
		t.Fatalf("Tag = %q, want row", row.Tag())
	}
	if !row.wrap {
		t.Fatal("wrap should default to true")
	}
	if row.justify != rtui.AlignStart || row.align != rtui.AlignStart {
		t.Fatalf("unexpected defaults: justify=%v align=%v", row.justify, row.align)
	}
}

func TestNewColDefaults(t *testing.T) {
	col := NewCol()
	if col.Tag() != "col" {
		t.Fatalf("Tag = %q, want col", col.Tag())
	}
	if col.span != 0 || col.offset != 0 {
		t.Fatalf("unexpected defaults: span=%d offset=%d", col.span, col.offset)
	}
}

func TestRowBuilderFluentAPI(t *testing.T) {
	node := NewRowBuilder().
		Key("toolbar-row").
		Justify(rtui.AlignCenter).
		Align(rtui.AlignEnd).
		Gutter(2, 1).
		Wrap(false).
		Width(80).
		Children(NewColBuilder().Span(12).Children(text.New("A")).Build()).
		BuildVNode()

	if node.Key() != "toolbar-row" {
		t.Fatalf("Key = %q, want toolbar-row", node.Key())
	}
	if node.justify != rtui.AlignCenter || node.align != rtui.AlignEnd {
		t.Fatalf("unexpected alignment: justify=%v align=%v", node.justify, node.align)
	}
	if node.gutter != 2 || node.verticalGutter != 1 {
		t.Fatalf("unexpected gutter: horizontal=%d vertical=%d", node.gutter, node.verticalGutter)
	}
	if node.wrap || node.width != 80 {
		t.Fatalf("unexpected row state: wrap=%v width=%d", node.wrap, node.width)
	}
}

func TestColBuilderFluentAPI(t *testing.T) {
	node := NewColBuilder().
		Key("sidebar-col").
		Span(8).
		Offset(2).
		Children(text.New("Sidebar")).
		BuildVNode()

	if node.Key() != "sidebar-col" {
		t.Fatalf("Key = %q, want sidebar-col", node.Key())
	}
	if node.span != 8 || node.offset != 2 {
		t.Fatalf("unexpected col state: span=%d offset=%d", node.span, node.offset)
	}
	if len(node.children) != 1 {
		t.Fatalf("children len = %d, want 1", len(node.children))
	}
}

func TestRowInstanceRuntimeChildrenSingleLine(t *testing.T) {
	inst := NewRowInstance(rtui.Props{
		propChildren: []rtui.VNode{
			NewColBuilder().Span(12).Children(text.New("A")).Build(),
			NewColBuilder().Span(12).Children(text.New("B")).Build(),
		},
		propGutter: 2,
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "hstack" {
		t.Fatalf("root tag = %q, want hstack", root.Tag())
	}
	if len(root.Children()) != 3 {
		t.Fatalf("root children len = %d, want 3", len(root.Children()))
	}
	if flex, ok := root.Children()[0].Props()["flex"].(int); !ok || flex != 12 {
		t.Fatalf("first col flex = %v, want 12", root.Children()[0].Props()["flex"])
	}
	if width, ok := root.Children()[1].Props()["width"].(int); !ok || width != 2 {
		t.Fatalf("gap width = %v, want 2", root.Children()[1].Props()["width"])
	}
	if flex, ok := root.Children()[2].Props()["flex"].(int); !ok || flex != 12 {
		t.Fatalf("second col flex = %v, want 12", root.Children()[2].Props()["flex"])
	}
}

func TestRowInstanceRuntimeChildrenWrapsLines(t *testing.T) {
	inst := NewRowInstance(rtui.Props{
		propChildren: []rtui.VNode{
			NewColBuilder().Span(12).Children(text.New("A")).Build(),
			NewColBuilder().Span(12).Children(text.New("B")).Build(),
			NewColBuilder().Span(12).Children(text.New("C")).Build(),
		},
		propGutter:         1,
		propVerticalGutter: 2,
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "vstack" {
		t.Fatalf("root tag = %q, want vstack", root.Tag())
	}
	if gap, ok := root.Props()["gap"].(int); !ok || gap != 2 {
		t.Fatalf("vertical gutter = %v, want 2", root.Props()["gap"])
	}
	if len(root.Children()) != 2 {
		t.Fatalf("wrapped lines len = %d, want 2", len(root.Children()))
	}
}

func TestRowInstanceJustifyAndOffset(t *testing.T) {
	inst := NewRowInstance(rtui.Props{
		propChildren: []rtui.VNode{
			NewColBuilder().Span(6).Offset(6).Children(text.New("Main")).Build(),
		},
		propJustify: rtui.AlignEnd,
	})

	children := inst.RuntimeChildren()
	root := children[0]
	if len(root.Children()) != 3 {
		t.Fatalf("root children len = %d, want 3", len(root.Children()))
	}
	if flex, ok := root.Children()[0].Props()["flex"].(int); !ok || flex != 12 {
		t.Fatalf("leading spacer flex = %v, want 12", root.Children()[0].Props()["flex"])
	}
	if flex, ok := root.Children()[1].Props()["flex"].(int); !ok || flex != 6 {
		t.Fatalf("offset spacer flex = %v, want 6", root.Children()[1].Props()["flex"])
	}
	if flex, ok := root.Children()[2].Props()["flex"].(int); !ok || flex != 6 {
		t.Fatalf("col flex = %v, want 6", root.Children()[2].Props()["flex"])
	}
}

func TestColInstanceRuntimeChildren(t *testing.T) {
	inst := NewColInstance(rtui.Props{
		propChildren: []rtui.VNode{text.New("A"), text.New("B")},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	if children[0].Tag() != "vstack" {
		t.Fatalf("root tag = %q, want vstack", children[0].Tag())
	}
	if len(children[0].Children()) != 2 {
		t.Fatalf("col children len = %d, want 2", len(children[0].Children()))
	}
}

func TestRowInstanceSetPropsTracksChanges(t *testing.T) {
	inst := NewRowInstance(rtui.Props{
		propChildren: []rtui.VNode{
			NewColBuilder().Span(12).Children(text.New("A")).Build(),
		},
	})
	inst.dirty = false

	if !inst.SetProps(rtui.Props{
		propChildren: []rtui.VNode{
			NewColBuilder().Span(8).Children(text.New("B")).Build(),
		},
	}) {
		t.Fatal("SetProps should report change")
	}
	if !inst.dirty {
		t.Fatal("dirty should be true after prop change")
	}
}
