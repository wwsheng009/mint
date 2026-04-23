package breadcrumb

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestNew(t *testing.T) {
	v := New([]Item{Crumb("Home"), Current("Docs")})
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "breadcrumb" {
		t.Fatalf("Tag = %q, want breadcrumb", v.Tag())
	}
	if got := len(v.Items()); got != 2 {
		t.Fatalf("Items len = %d, want 2", got)
	}
	if v.Separator() != " / " {
		t.Fatalf("Separator = %q, want %q", v.Separator(), " / ")
	}
}

func TestVNodeSetPropsClonesItems(t *testing.T) {
	source := []Item{Crumb("Home")}
	v := New(nil)
	v.SetProps(rtui.Props{
		propItems:     source,
		propSeparator: " > ",
		propMaxWidth:  20,
	})
	source[0].Label = "Mutated"

	items := v.Items()
	if items[0].Label != "Home" {
		t.Fatalf("Items should be cloned, got %q", items[0].Label)
	}
	if v.Separator() != " > " {
		t.Fatalf("Separator = %q, want %q", v.Separator(), " > ")
	}
	if v.MaxWidth() != 20 {
		t.Fatalf("MaxWidth = %d, want 20", v.MaxWidth())
	}
}

func TestBuilderFluent(t *testing.T) {
	v := NewBuilder().
		Key("nav").
		Labels("Home", "Workspace").
		Item(Current("Breadcrumb")).
		Separator(" > ").
		MaxWidth(24).
		ItemStyle(style.NewStyle().Underline(true)).
		CurrentStyle(style.NewStyle().Bold(true)).
		Build()

	if v.Key() != "nav" {
		t.Fatalf("Key = %q, want nav", v.Key())
	}
	if got := len(v.Items()); got != 3 {
		t.Fatalf("Items len = %d, want 3", got)
	}
	if v.Separator() != " > " {
		t.Fatalf("Separator = %q, want %q", v.Separator(), " > ")
	}
	if v.MaxWidth() != 24 {
		t.Fatalf("MaxWidth = %d, want 24", v.MaxWidth())
	}
}

func TestInstanceMeasure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{Crumb("Home"), Crumb("Docs"), Current("API")},
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 3})
	wantWidth := paint.StringWidth("Home / Docs / API")
	if size.Width != wantWidth {
		t.Fatalf("Width = %d, want %d", size.Width, wantWidth)
	}
	if size.Height != 1 {
		t.Fatalf("Height = %d, want 1", size.Height)
	}
}

func TestInstancePaintRendersSegments(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Crumb("Home").WithIcon("H"),
			Crumb("Docs"),
			Current("API"),
		},
	})
	inst.SetBounds(0, 0, 80, 1)
	cmds := inst.Paint(0, 0)

	if len(cmds) != 5 {
		t.Fatalf("Paint returned %d commands, want 5", len(cmds))
	}
	if cmds[0].Text != "H Home" {
		t.Fatalf("first item = %q, want %q", cmds[0].Text, "H Home")
	}
	if cmds[1].Text != " / " || cmds[3].Text != " / " {
		t.Fatalf("separators = %q and %q, want %q", cmds[1].Text, cmds[3].Text, " / ")
	}
	if cmds[4].Text != "API" {
		t.Fatalf("last item = %q, want API", cmds[4].Text)
	}
	if !cmds[4].Style.IsBold() {
		t.Fatal("current breadcrumb item should be bold")
	}
}

func TestInstancePaintCollapsesLeftSide(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Crumb("Home"),
			Crumb("Workspace"),
			Crumb("Components"),
			Current("Breadcrumb"),
		},
	})
	inst.SetBounds(0, 0, 20, 1)
	cmds := inst.Paint(0, 0)
	if len(cmds) == 0 {
		t.Fatal("Paint returned no commands")
	}

	text := collectText(cmds)
	if !strings.HasPrefix(text, "… / ") {
		t.Fatalf("collapsed text = %q, want prefix %q", text, "… / ")
	}
	if !strings.Contains(text, "Breadcrumb") {
		t.Fatalf("collapsed text = %q, want to contain Breadcrumb", text)
	}
	if paint.StringWidth(text) > 20 {
		t.Fatalf("collapsed width = %d, want <= 20", paint.StringWidth(text))
	}
}

func TestInstancePaintTruncatesCurrentItemWhenNeeded(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{Current("Breadcrumb")},
	})
	inst.SetBounds(0, 0, 5, 1)
	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}
	if got := cmds[0].Text; got != "Brea…" {
		t.Fatalf("truncated text = %q, want %q", got, "Brea…")
	}
}

func TestInstanceSetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{Crumb("Home"), Current("Docs")},
	})
	inst.MarkClean()

	changed := inst.SetProps(rtui.Props{
		propItems:     []Item{Crumb("Home"), Current("API")},
		propSeparator: " > ",
	})
	if !changed {
		t.Fatal("SetProps should report changes")
	}
	if !inst.IsDirty() {
		t.Fatal("SetProps should mark the instance dirty")
	}
	if inst.separator != " > " {
		t.Fatalf("Separator = %q, want %q", inst.separator, " > ")
	}
}

func TestBreadcrumbItemHelpers(t *testing.T) {
	item := Crumb("Home").WithKey("home").WithIcon("H").AsCurrent()
	if item.Key != "home" || item.Icon != "H" || !item.Current {
		t.Fatalf("item helpers lost metadata: %+v", item)
	}
}

func TestVNodeImplementsInterfaces(t *testing.T) {
	var _ rtui.VNode = New(nil)
	var _ rtui.InstanceFactory = New(nil)
}

func collectText(cmds []paint.DrawCmd) string {
	var builder strings.Builder
	for _, cmd := range cmds {
		builder.WriteString(cmd.Text)
	}
	return builder.String()
}
