package splitpane

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/text"
)

func TestBuilderHorizontalRuntimeChildren(t *testing.T) {
	inst := NewBuilder().
		Key("ops").
		Width(80).
		Height(6).
		PrimarySize(24).
		SecondaryFlex(2).
		SeparatorColor(style.Color("bright-black")).
		Panes(text.New("Groups"), text.New("Details")).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("runtime children len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "hstack" {
		t.Fatalf("root tag = %q, want hstack", root.Tag())
	}
	if got := root.Props()["width"]; got != 80 {
		t.Fatalf("root width = %v, want 80", got)
	}
	if got := root.Props()["height"]; got != 6 {
		t.Fatalf("root height = %v, want 6", got)
	}

	rootChildren := root.Children()
	if len(rootChildren) != 3 {
		t.Fatalf("root children len = %d, want 3", len(rootChildren))
	}
	if got := rootChildren[0].Props()["width"]; got != 24 {
		t.Fatalf("primary width = %v, want 24", got)
	}
	if got := rootChildren[2].Props()["flex"]; got != 2 {
		t.Fatalf("secondary flex = %v, want 2", got)
	}
	if rootChildren[1].Tag() != "vstack" {
		t.Fatalf("separator tag = %q, want vstack", rootChildren[1].Tag())
	}
	if lines := rootChildren[1].Children(); len(lines) != 6 {
		t.Fatalf("separator lines len = %d, want 6", len(lines))
	}
}

func TestBuilderVerticalRuntimeChildren(t *testing.T) {
	inst := NewBuilder().
		Vertical().
		Key("runtime").
		Width(12).
		PrimarySize(3).
		Gap(0).
		Panes(text.New("Top"), text.New("Bottom")).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("runtime children len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "vstack" {
		t.Fatalf("root tag = %q, want vstack", root.Tag())
	}
	rootChildren := root.Children()
	if len(rootChildren) != 3 {
		t.Fatalf("root children len = %d, want 3", len(rootChildren))
	}
	if got := rootChildren[0].Props()["height"]; got != 3 {
		t.Fatalf("primary height = %v, want 3", got)
	}
	separatorProps := rootChildren[1].Props()
	if got := separatorProps["content"]; got != "────────────" {
		t.Fatalf("separator content = %q, want 12 horizontal glyphs", got)
	}
}

func TestSetPropsNormalizesInvalidValues(t *testing.T) {
	node := NewBuilder().
		Primary(text.New("A")).
		Secondary(text.New("B")).
		BuildVNode()

	node.SetProps(rtui.Props{
		propDirection:      Direction("bad"),
		propPrimarySize:    -1,
		propSecondarySize:  -2,
		propPrimaryFlex:    -3,
		propSecondaryFlex:  -4,
		propGap:            -5,
		propWidth:          -6,
		propHeight:         -7,
		propSeparatorGlyph: "",
	})

	props := node.Props()
	if got := props[propDirection]; got != DirectionHorizontal {
		t.Fatalf("direction = %v, want horizontal", got)
	}
	for _, key := range []string{propPrimarySize, propSecondarySize, propPrimaryFlex, propSecondaryFlex, propGap, propWidth, propHeight} {
		if got := props[key]; got != 0 {
			t.Fatalf("%s = %v, want 0", key, got)
		}
	}
	if got := props[propSeparatorGlyph]; got != "│" {
		t.Fatalf("separator glyph = %q, want vertical glyph", got)
	}
}
