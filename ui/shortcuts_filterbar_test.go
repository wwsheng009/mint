package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui/components/filterbar"
)

type filterBarShortcutIntent struct {
	name string
}

func (i filterBarShortcutIntent) IntentType() string { return i.name }

func TestFilterBarShortcuts(t *testing.T) {
	bar := NewFilterBarBuilder().
		Key("logs.filterbar").
		Title("Log Filters").
		Summary("range: 15m | status: failed").
		Field(FilterBarSearch("query", "Query", "trace-1").ForField("logQuery")).
		Field(FilterBarSelect("status", "Status", []FilterBarOption{
			{Value: "all", Label: "All"},
			{Value: "failed", Label: "Failed"},
		}).WithSelectedIndex(1).ForField("logStatus")).
		Action(FilterBarButton("refresh", "Refresh", filterBarShortcutIntent{"refresh"}).Primary()).
		Action(FilterBarButton("export", "Export", filterBarShortcutIntent{"export"}).WithDisabledReason("Select rows first.")).
		Build()

	node, ok := bar.(*filterbar.VNode)
	if !ok {
		t.Fatalf("FilterBar builder returned %T, want *filterbar.VNode", bar)
	}
	props := node.Props()
	if got := props["summary"]; got != "range: 15m | status: failed" {
		t.Fatalf("summary = %v, want range summary", got)
	}
	actions := props["actions"].([]FilterBarAction)
	if len(actions) != 2 {
		t.Fatalf("actions len = %d, want 2", len(actions))
	}
	if !actions[1].Disabled || actions[1].DisabledReason != "Select rows first." {
		t.Fatalf("disabled action = %#v", actions[1])
	}
	if _, ok := actions[0].PressIntent.(intent.Intent); !ok {
		t.Fatalf("press intent = %T, want intent.Intent", actions[0].PressIntent)
	}
}
