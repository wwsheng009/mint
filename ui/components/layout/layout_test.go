package layout

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/text"
)

func TestVNodeNewDefaults(t *testing.T) {
	node := New()
	if node.Tag() != "layout" {
		t.Fatalf("Tag = %q, want layout", node.Tag())
	}
	if node.sectionGap != 0 {
		t.Fatalf("sectionGap = %d, want 0", node.sectionGap)
	}
	if node.bodyGap != 0 {
		t.Fatalf("bodyGap = %d, want 0", node.bodyGap)
	}
}

func TestBuilderFluentAPI(t *testing.T) {
	rootStyle := style.NewStyle().Foreground(style.Cyan).Bold(true)
	node := NewBuilder().
		Key("app-shell").
		Header(text.New("Header")).
		LeftSider(text.New("Left")).
		Content(text.New("Content")).
		RightSider(text.New("Right")).
		Footer(text.New("Footer")).
		Gap(1).
		BodyGap(2).
		Width(80).
		Height(24).
		Style(rootStyle).
		BuildVNode()

	if node.Key() != "app-shell" {
		t.Fatalf("Key = %q, want app-shell", node.Key())
	}
	if node.header == nil || node.leftSider == nil || node.content == nil || node.rightSider == nil || node.footer == nil {
		t.Fatal("expected all layout slots to be populated")
	}
	if node.sectionGap != 1 || node.bodyGap != 2 {
		t.Fatalf("unexpected gaps: section=%d body=%d", node.sectionGap, node.bodyGap)
	}
	if node.width != 80 || node.height != 24 {
		t.Fatalf("unexpected size: width=%d height=%d", node.width, node.height)
	}
	if node.rootStyle != rootStyle {
		t.Fatalf("rootStyle = %#v, want %#v", node.rootStyle, rootStyle)
	}
}

func TestInstanceRuntimeChildrenBuildsShell(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propKey:        "shell",
		propHeader:     text.New("Header"),
		propLeftSider:  text.New("Left"),
		propContent:    text.New("Content"),
		propRightSider: text.New("Right"),
		propFooter:     text.New("Footer"),
		propSectionGap: 1,
		propBodyGap:    2,
		propWidth:      80,
		propHeight:     24,
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}

	root := children[0]
	if root.Tag() != "vstack" {
		t.Fatalf("root tag = %q, want vstack", root.Tag())
	}
	if gap, ok := root.Props()["gap"].(int); !ok || gap != 1 {
		t.Fatalf("root gap = %v, want 1", root.Props()["gap"])
	}
	if width, ok := root.Props()["width"].(int); !ok || width != 80 {
		t.Fatalf("root width = %v, want 80", root.Props()["width"])
	}
	if height, ok := root.Props()["height"].(int); !ok || height != 24 {
		t.Fatalf("root height = %v, want 24", root.Props()["height"])
	}
	if len(root.Children()) != 3 {
		t.Fatalf("root children len = %d, want 3", len(root.Children()))
	}

	body := root.Children()[1]
	if body.Tag() != "hstack" {
		t.Fatalf("body tag = %q, want hstack", body.Tag())
	}
	if flex, ok := body.Props()["flex"].(int); !ok || flex != 1 {
		t.Fatalf("body flex = %v, want 1", body.Props()["flex"])
	}
	if gap, ok := body.Props()["gap"].(int); !ok || gap != 2 {
		t.Fatalf("body gap = %v, want 2", body.Props()["gap"])
	}
	if len(body.Children()) != 3 {
		t.Fatalf("body children len = %d, want 3", len(body.Children()))
	}
	content := body.Children()[1]
	if content.Tag() != "vstack" {
		t.Fatalf("content wrapper tag = %q, want vstack", content.Tag())
	}
	if flex, ok := content.Props()["flex"].(int); !ok || flex != 1 {
		t.Fatalf("content flex = %v, want 1", content.Props()["flex"])
	}
	if len(content.Children()) != 1 {
		t.Fatalf("content wrapper children len = %d, want 1", len(content.Children()))
	}
}

func TestInstanceRuntimeChildrenWithoutSections(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	if children := inst.RuntimeChildren(); len(children) != 0 {
		t.Fatalf("RuntimeChildren len = %d, want 0", len(children))
	}
}

func TestInstanceSetPropsTracksChanges(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propContent: text.New("A"),
	})
	inst.dirty = false

	if !inst.SetProps(rtui.Props{
		propContent: text.New("B"),
	}) {
		t.Fatal("SetProps should report change")
	}
	if !inst.dirty {
		t.Fatal("dirty should be true after prop change")
	}
}
