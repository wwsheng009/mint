package timeline

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestNew(t *testing.T) {
	v := New([]Item{Event("Created")})
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "timeline" {
		t.Fatalf("Tag = %q, want timeline", v.Tag())
	}
	if len(v.items) != 1 {
		t.Fatalf("items len = %d, want 1", len(v.items))
	}
}

func TestBuilderFluent(t *testing.T) {
	v := NewBuilder().
		Key("history").
		Item(Event("Created").WithLabel("09:00").WithDescription("ticket opened")).
		Pending("waiting").
		Reverse(true).
		Width(48).
		LabelStyle(style.NewStyle().Bold(true)).
		ContentStyle(style.NewStyle().Foreground(style.Color("cyan"))).
		BuildVNode()

	if v.Key() != "history" {
		t.Fatalf("Key = %q, want history", v.Key())
	}
	if v.pending != "waiting" || !v.reverse || v.width != 48 {
		t.Fatalf("pending/reverse/width = (%q,%v,%d)", v.pending, v.reverse, v.width)
	}
}

func TestNormalizeItemsAssignsPendingDot(t *testing.T) {
	items := normalizeItems([]Item{Event("Pending").WithStatus(StatusPending)})
	if items[0].Dot != "○" {
		t.Fatalf("pending dot = %q, want ○", items[0].Dot)
	}
}

func TestMeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Event("Build completed").WithLabel("09:30").WithDescription("CI finished").WithStatus(StatusSuccess),
			Event("Deploy started").WithLabel("09:45"),
		},
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 20})
	if size.Height == 0 || size.Width == 0 {
		t.Fatalf("size = %#v, want non-zero", size)
	}
	inst.SetBounds(0, 0, 80, 20)
	lines := collectLines(inst.Paint(0, 0))
	if len(lines) < 4 {
		t.Fatalf("line count = %d, want >= 4", len(lines))
	}
	if !strings.Contains(lines[0], "09:30") || !strings.Contains(lines[0], "Build completed") {
		t.Fatalf("line 0 = %q", lines[0])
	}
	if !strings.Contains(strings.Join(lines, "\n"), "CI finished") {
		t.Fatal("expected description text")
	}
}

func TestReverseAndPending(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Event("First").WithLabel("1"),
			Event("Second").WithLabel("2"),
		},
		propPending: "Queued",
		propReverse: true,
	})
	inst.SetBounds(0, 0, 60, 20)
	lines := collectLines(inst.Paint(0, 0))
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Second") || !strings.Contains(joined, "Queued") {
		t.Fatalf("unexpected timeline output:\n%s", joined)
	}
	firstContentLine := ""
	for _, line := range lines {
		if strings.Contains(line, "Second") || strings.Contains(line, "First") {
			firstContentLine = line
			break
		}
	}
	if !strings.Contains(firstContentLine, "Second") {
		t.Fatalf("first event line = %q, want Second", firstContentLine)
	}
}

func TestMarkerStyleUsesStatusColor(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Event("Warn").WithStatus(StatusWarning),
			Event("Error").WithStatus(StatusError),
		},
	})
	inst.SetBounds(0, 0, 40, 10)
	cmds := inst.Paint(0, 0)
	if styleForMarker(cmds, "▲").FG != theme.Warning() {
		t.Fatalf("warning marker fg = %q, want %q", styleForMarker(cmds, "▲").FG, theme.Warning())
	}
	if styleForMarker(cmds, "✖").FG != theme.Error() {
		t.Fatalf("error marker fg = %q, want %q", styleForMarker(cmds, "✖").FG, theme.Error())
	}
}

func collectLines(cmds []paint.DrawCmd) []string {
	maxY := -1
	for _, cmd := range cmds {
		if cmd.Y > maxY {
			maxY = cmd.Y
		}
	}
	lines := make([]string, maxY+1)
	for _, cmd := range cmds {
		lines[cmd.Y] += cmd.Text
	}
	return lines
}

func styleForMarker(cmds []paint.DrawCmd, marker string) style.Style {
	for _, cmd := range cmds {
		if cmd.Text == marker {
			return cmd.Style
		}
	}
	return style.Style{}
}
