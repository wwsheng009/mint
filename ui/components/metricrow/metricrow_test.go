package metricrow

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestNew(t *testing.T) {
	row := New(
		[]Item{
			{Title: "Runtime", Value: "healthy"},
			{Title: "Requests", Value: 42},
			{Title: "Missing", Value: nil},
		},
		ItemWidth(24),
		Gap(2),
		Border(layout.BorderDouble, style.Color("green")),
	)

	if row.Tag() != "hstack" {
		t.Fatalf("MetricRow tag = %q, want hstack", row.Tag())
	}
	props := row.Props()
	if got := props["gap"]; got != 2 {
		t.Fatalf("gap = %v, want 2", got)
	}
	children := row.Children()
	if len(children) != 3 {
		t.Fatalf("children len = %d, want 3", len(children))
	}
	firstProps := children[0].Props()
	if got := firstProps["title"]; got != "Runtime" {
		t.Fatalf("first title = %v, want Runtime", got)
	}
	if got := firstProps["width"]; got != 24 {
		t.Fatalf("first width = %v, want 24", got)
	}
	if got := firstProps["borderStyle"]; got != layout.BorderDouble {
		t.Fatalf("first borderStyle = %v, want %v", got, layout.BorderDouble)
	}
	if got := firstProps["borderColor"]; got != style.Color("green") {
		t.Fatalf("first borderColor = %v, want green", got)
	}
}

func TestOperational(t *testing.T) {
	row := Operational([]Item{
		{Title: "Runtime", Value: "healthy"},
		{Title: "Alerts", Value: 2},
	})
	if row.Tag() != "hstack" {
		t.Fatalf("Operational tag = %q, want hstack", row.Tag())
	}
	props := row.Props()
	if got := props["gap"]; got != 1 {
		t.Fatalf("gap = %v, want 1", got)
	}
	children := row.Children()
	if len(children) != 2 {
		t.Fatalf("children len = %d, want 2", len(children))
	}
	if got := children[0].Props()["width"]; got != 20 {
		t.Fatalf("item width = %v, want 20", got)
	}
	if got := children[0].Props()["title"]; got != "Runtime" {
		t.Fatalf("title = %v, want Runtime", got)
	}
}

func TestItemWithWidthOverridesRowDefault(t *testing.T) {
	row := Operational([]Item{
		Value("Jobs", 12),
		CompactValue("Filters", "search=cleanup / status=active / last=failed", 42).WithWidth(46),
	})
	children := row.Children()
	if len(children) != 2 {
		t.Fatalf("children len = %d, want 2", len(children))
	}
	if got := children[0].Props()["width"]; got != 20 {
		t.Fatalf("default item width = %v, want 20", got)
	}
	if got := children[1].Props()["width"]; got != 46 {
		t.Fatalf("custom item width = %v, want 46", got)
	}
}

func TestMetricItemHelpersNormalizeOperationalValues(t *testing.T) {
	items := []Item{
		Value("Status", ""),
		FallbackValue("Hint", nil, "open from logs"),
		CompactValue("Trace", "trace-1234567890-extra", 12),
		Count("Errors", -3),
	}
	if items[0].Value != "-" {
		t.Fatalf("blank value = %v, want -", items[0].Value)
	}
	if items[1].Value != "open from logs" {
		t.Fatalf("fallback value = %v, want open from logs", items[1].Value)
	}
	if items[2].Value != "trace-123..." {
		t.Fatalf("compact value = %v, want trace-123...", items[2].Value)
	}
	if items[3].Value != 0 {
		t.Fatalf("count value = %v, want 0", items[3].Value)
	}
}

func TestFormatter(t *testing.T) {
	row := New(
		[]Item{{Title: "Rate", Value: 0.95}},
		Formatter(func(value interface{}) string {
			return "95%"
		}),
	)
	children := row.Children()
	if len(children) != 1 {
		t.Fatalf("children len = %d, want 1", len(children))
	}
	content := children[0].Props()["content"]
	contentNode, ok := content.(rtui.VNode)
	if !ok {
		t.Fatalf("metric content = %T, want ui.VNode", content)
	}
	if got := contentNode.Props()["content"]; got != "95%" {
		t.Fatalf("content text = %q, want 95%%", got)
	}
}
