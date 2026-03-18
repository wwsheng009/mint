package skeleton

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
	if v.Tag() != "skeleton" {
		t.Fatalf("Tag = %q, want skeleton", v.Tag())
	}
	if !v.loading || !v.showTitle || !v.showParagraph {
		t.Fatalf("unexpected defaults: loading=%v title=%v paragraph=%v", v.loading, v.showTitle, v.showParagraph)
	}
}

func TestBuilderFluent(t *testing.T) {
	content := textcomp.New("Loaded")
	v := NewBuilder().
		Key("profile-skeleton").
		Loading(false).
		Active(true).
		Avatar(true).
		AvatarShape(ShapeRound).
		AvatarSize(5).
		TitleWidth(18).
		ParagraphRows(2).
		ParagraphWidths(20, 12).
		Width(32).
		Gap(2).
		Content(content).
		PlaceholderStyle(style.NewStyle().Foreground(style.Blue)).
		BuildVNode()

	if v.Key() != "profile-skeleton" || v.content != content {
		t.Fatalf("key/content not set")
	}
	if !v.active || !v.showAvatar || v.avatarShape != ShapeRound || v.avatarSize != 5 {
		t.Fatalf("avatar props not set: %+v", v)
	}
	if v.titleWidth != 18 || v.paragraphRows != 2 || v.width != 32 || v.gap != 2 {
		t.Fatalf("layout props not set: title=%d rows=%d width=%d gap=%d", v.titleWidth, v.paragraphRows, v.width, v.gap)
	}
	if len(v.paragraphWidths) != 2 || v.paragraphWidths[1] != 12 {
		t.Fatalf("paragraph widths = %v", v.paragraphWidths)
	}
}

func TestRuntimeChildrenWhenLoadingFalseReturnsContent(t *testing.T) {
	content := textcomp.New("Loaded profile content")
	inst := NewInstance(rtui.Props{
		propLoading: false,
		propContent: content,
	})
	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	if children[0] != content {
		t.Fatal("expected content vnode when not loading")
	}
}

func TestRuntimeChildrenBuildsSkeletonLines(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propAvatar:        true,
		propAvatarSize:    3,
		propTitleWidth:    12,
		propParagraphRows: 2,
		propParagraphWidths: []int{
			18,
			10,
		},
	})
	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if !containsVNodeText(root, "▓▓▓") {
		t.Fatal("expected avatar placeholder line")
	}
	if !containsVNodeText(root, "▓▓▓▓▓▓▓▓▓▓▓▓") {
		t.Fatal("expected title placeholder line")
	}
	if !containsVNodeText(root, "▓▓▓▓▓▓▓▓▓▓") {
		t.Fatal("expected paragraph placeholder line")
	}
}

func TestRoundAvatarLines(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propAvatar:      true,
		propAvatarShape: ShapeRound,
		propAvatarSize:  4,
	})
	lines := inst.avatarLines()
	if len(lines) != 4 {
		t.Fatalf("avatarLines len = %d, want 4", len(lines))
	}
	if lines[0] != " ▓▓ " || lines[3] != " ▓▓ " {
		t.Fatalf("round avatar top/bottom = %q / %q", lines[0], lines[3])
	}
	if lines[1] != "▓▓▓▓" || lines[2] != "▓▓▓▓" {
		t.Fatalf("round avatar middle = %q / %q", lines[1], lines[2])
	}
}

func TestResolvedPlaceholderStyleUsesThemeDefaults(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	resolved := inst.resolvedPlaceholderStyle()
	if resolved.FG != theme.Muted() {
		t.Fatalf("placeholder fg = %q, want %q", resolved.FG, theme.Muted())
	}
	if resolved.BG != theme.Surface() {
		t.Fatalf("placeholder bg = %q, want %q", resolved.BG, theme.Surface())
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
