package result

import (
	"testing"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "result" {
		t.Fatalf("Tag = %q, want result", v.Tag())
	}
	if v.status != StatusInfo {
		t.Fatalf("status = %v, want info", v.status)
	}
}

func TestBuilderFluent(t *testing.T) {
	extra := textcomp.New("Retry")
	v := NewBuilder().
		Key("save-result").
		Status(StatusSuccess).
		Icon("★").
		Title("Saved").
		Subtitle("Everything is synchronized.").
		Extra(extra).
		Bordered(true).
		Width(40).
		IconStyle(style.NewStyle().Bold(true)).
		BuildVNode()

	if v.Key() != "save-result" || v.status != StatusSuccess || v.icon != "★" {
		t.Fatalf("key/status/icon = (%q,%v,%q)", v.Key(), v.status, v.icon)
	}
	if v.title != "Saved" || v.subtitle != "Everything is synchronized." || v.extra != extra {
		t.Fatalf("title/subtitle/extra not set")
	}
	if !v.bordered || v.width != 40 {
		t.Fatalf("bordered/width = (%v,%d)", v.bordered, v.width)
	}
}

func TestRuntimeChildrenRenderPresetTexts(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propStatus:   Status404,
		propSubtitle: "The requested item was not found.",
	})
	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if !containsVNodeText(root, "404") {
		t.Fatal("expected preset icon")
	}
	if !containsVNodeText(root, "404 Not Found") {
		t.Fatal("expected preset title")
	}
	if !containsVNodeText(root, "The requested item was not found.") {
		t.Fatal("expected subtitle text")
	}
}

func TestResultBorderColorUsesStatusColor(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propStatus:   StatusError,
		propBordered: true,
	})
	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	props := children[0].Props()
	if props == nil {
		t.Fatal("expected props on bordered result root")
	}
	got, ok := props.Get("borderColor").(style.Color)
	if !ok || got != theme.Error() {
		t.Fatalf("borderColor = %q, want %q", got, theme.Error())
	}
}

func TestStatusColorAndIconStyle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propStatus: StatusSuccess,
	})
	if inst.statusColor() != theme.Success() {
		t.Fatalf("statusColor = %q, want %q", inst.statusColor(), theme.Success())
	}
	if inst.iconTextStyle().FG != theme.Success() {
		t.Fatalf("icon fg = %q, want %q", inst.iconTextStyle().FG, theme.Success())
	}
}

func containsVNodeText(node rtui.VNode, want string) bool {
	if node == nil {
		return false
	}
	if props := node.Props(); props != nil && props.GetString("content") == want {
		return true
	}
	if textNode, ok := node.(*textcomp.VNode); ok && textNode.Content() == want {
		return true
	}
	for _, child := range node.Children() {
		if containsVNodeText(child, want) {
			return true
		}
	}
	return false
}
