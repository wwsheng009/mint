package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
)

func TestMetricRow(t *testing.T) {
	row := MetricRow(
		[]MetricItem{
			{Title: "Runtime", Value: "healthy"},
			{Title: "Requests", Value: 42},
			{Title: "Missing", Value: nil},
		},
		MetricRowItemWidth(24),
		MetricRowGap(2),
		MetricRowBorder(layout.BorderDouble, style.Color("green")),
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

func TestMetricRowFormatter(t *testing.T) {
	row := MetricRow(
		[]MetricItem{{Title: "Rate", Value: 0.95}},
		MetricRowFormatter(func(value interface{}) string {
			return "95%"
		}),
	)
	children := row.Children()
	if len(children) != 1 {
		t.Fatalf("children len = %d, want 1", len(children))
	}
	content := children[0].Props()["content"]
	contentNode, ok := content.(VNode)
	if !ok {
		t.Fatalf("metric content = %T, want ui.VNode", content)
	}
	if got := contentNode.Props()["content"]; got != "95%" {
		t.Fatalf("content text = %q, want 95%%", got)
	}
}
