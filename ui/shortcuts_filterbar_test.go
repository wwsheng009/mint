package ui

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
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

func TestFilterBarSearchRefreshShortcut(t *testing.T) {
	action := FilterBarRefreshAction(filterBarShortcutIntent{"refresh"}, true, "Logs are loading.")
	if !action.Disabled || action.DisabledReason != "Logs are loading." {
		t.Fatalf("refresh action = %#v", action)
	}
	reset := FilterBarResetAction(filterBarShortcutIntent{"reset"}, true, "Logs are loading.")
	if !reset.Disabled || reset.DisabledReason != "Logs are loading." || reset.Key != "reset" {
		t.Fatalf("reset action = %#v", reset)
	}
	unchangedReset := FilterBarResetActionWhenChanged(filterBarShortcutIntent{"reset"}, false, false, "")
	if !unchangedReset.Disabled || unchangedReset.DisabledReason != "Nothing to reset." {
		t.Fatalf("unchanged reset action = %#v", unchangedReset)
	}
	clear := FilterBarClearFieldAction("logSearch", "trace-1", false, "")
	if clear.Disabled || clear.Key != "clear" {
		t.Fatalf("clear action = %#v", clear)
	}
	clearIntent, ok := clear.PressIntent.(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("clear intent = %T, want FieldChangeIntent", clear.PressIntent)
	}
	if clearIntent.Field != "logSearch" || clearIntent.Value != "" {
		t.Fatalf("clear intent = %#v, want logSearch empty value", clearIntent)
	}
	namedClear := FilterBarClearFieldActionWithLabel("clear-history", "Clear History", "alertHistorySearch", "provider", false, "")
	if namedClear.Disabled || namedClear.Key != "clear-history" || namedClear.Label != "Clear History" {
		t.Fatalf("named clear action = %#v", namedClear)
	}
	namedClearIntent, ok := namedClear.PressIntent.(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("named clear intent = %T, want FieldChangeIntent", namedClear.PressIntent)
	}
	if namedClearIntent.Field != "alertHistorySearch" || namedClearIntent.Value != "" {
		t.Fatalf("named clear intent = %#v, want alertHistorySearch empty value", namedClearIntent)
	}
	if got := FilterBarJoinDisabledReasons("Enter lookup.", "", "Data is loading."); got != "Enter lookup. Data is loading." {
		t.Fatalf("joined disabled reasons = %q, want normalized reason text", got)
	}
	joinedDisabled := FilterBarButton("load", "Load", filterBarShortcutIntent{"load"}).
		WithDisabledReasons("Enter lookup.", "Data is loading.")
	if !joinedDisabled.Disabled || joinedDisabled.DisabledReason != "Enter lookup. Data is loading." {
		t.Fatalf("joined disabled action = %#v", joinedDisabled)
	}

	bar := FilterBarSearchRefresh(
		"logs.filters",
		"page 1 total 25 search -",
		126,
		6,
		FilterBarSearch("query", "Search", "").WithPlaceholder("trace/request").WithWidth(42).ForField("logSearch"),
		filterBarShortcutIntent{"refresh"},
		true,
		"Logs are loading.",
	)
	node, ok := bar.(*filterbar.VNode)
	if !ok {
		t.Fatalf("FilterBarSearchRefresh returned %T, want *filterbar.VNode", bar)
	}
	props := node.Props()
	if got := props["width"]; got != 126 {
		t.Fatalf("width = %v, want 126", got)
	}
	fields := props["fields"].([]FilterBarField)
	if len(fields) != 1 || fields[0].FieldName != "logSearch" {
		t.Fatalf("fields = %#v", fields)
	}
	actions := props["actions"].([]FilterBarAction)
	if len(actions) != 1 || actions[0].Key != "refresh" || !actions[0].Disabled {
		t.Fatalf("actions = %#v", actions)
	}

	clearBar := FilterBarSearchRefreshClear(
		"logs.filters.clear",
		"page 1 total 25 search trace-1",
		126,
		6,
		FilterBarSearch("query", "Search", "trace-1").WithPlaceholder("trace/request").WithWidth(42).ForField("logSearch"),
		filterBarShortcutIntent{"refresh"},
		false,
		"Logs are loading.",
	)
	clearNode, ok := clearBar.(*filterbar.VNode)
	if !ok {
		t.Fatalf("FilterBarSearchRefreshClear returned %T, want *filterbar.VNode", clearBar)
	}
	clearActions := clearNode.Props()["actions"].([]FilterBarAction)
	if len(clearActions) != 2 || clearActions[0].Key != "refresh" || clearActions[1].Key != "clear" {
		t.Fatalf("clear actions = %#v", clearActions)
	}

	lookup := FilterBarSearchActions(
		"trace.filters",
		"lookup trace-1 resolved -",
		126,
		8,
		FilterBarSearch("trace", "Trace", "trace-1").WithWidth(48).ForField("traceLookupID"),
		FilterBarButton("refresh", "Load Trace", filterBarShortcutIntent{"refresh"}).Primary(),
		FilterBarButton("logs", "Back Logs", filterBarShortcutIntent{"logs"}),
	)
	lookupNode, ok := lookup.(*filterbar.VNode)
	if !ok {
		t.Fatalf("FilterBarSearchActions returned %T, want *filterbar.VNode", lookup)
	}
	lookupProps := lookupNode.Props()
	lookupFields := lookupProps["fields"].([]FilterBarField)
	if len(lookupFields) != 1 || lookupFields[0].FieldName != "traceLookupID" {
		t.Fatalf("lookup fields = %#v", lookupFields)
	}
	lookupActions := lookupProps["actions"].([]FilterBarAction)
	if len(lookupActions) != 2 || lookupActions[0].Key != "refresh" || lookupActions[1].Key != "logs" {
		t.Fatalf("lookup actions = %#v", lookupActions)
	}

	multi := FilterBarFieldsRefresh(
		"alerts.filters",
		"alerts 2 history 5",
		126,
		8,
		[]FilterBarField{
			FilterBarSearch("query", "Search", "latency").WithWidth(30).ForField("alertSearch"),
			FilterBarSelect("status", "Status", []FilterBarOption{
				{Value: "all", Label: "All"},
				{Value: "firing", Label: "Firing"},
			}).WithSelectedIndex(1).WithWidth(16).ForField("alertStatus"),
		},
		filterBarShortcutIntent{"refresh"},
		true,
		"Alerts are loading.",
	)
	multiNode, ok := multi.(*filterbar.VNode)
	if !ok {
		t.Fatalf("FilterBarFieldsRefresh returned %T, want *filterbar.VNode", multi)
	}
	multiProps := multiNode.Props()
	if got := multiProps["wrap"]; got != true {
		t.Fatalf("wrap = %v, want true", got)
	}
	multiFields := multiProps["fields"].([]FilterBarField)
	if len(multiFields) != 2 || multiFields[0].FieldName != "alertSearch" || multiFields[1].FieldName != "alertStatus" {
		t.Fatalf("multi fields = %#v", multiFields)
	}

	multiReset := FilterBarFieldsRefreshReset(
		"alerts.filters.reset",
		"alerts 2 history 5",
		126,
		8,
		[]FilterBarField{
			FilterBarSearch("query", "Search", "latency").WithWidth(30).ForField("alertSearch"),
			FilterBarSelect("status", "Status", []FilterBarOption{{Value: "actionable", Label: "Actionable"}}).WithSelectedIndex(0).ForField("alertStatus"),
		},
		filterBarShortcutIntent{"refresh"},
		filterBarShortcutIntent{"reset"},
		false,
		"Alerts are loading.",
	)
	multiResetNode, ok := multiReset.(*filterbar.VNode)
	if !ok {
		t.Fatalf("FilterBarFieldsRefreshReset returned %T, want *filterbar.VNode", multiReset)
	}
	multiResetActions := multiResetNode.Props()["actions"].([]FilterBarAction)
	if len(multiResetActions) != 2 || multiResetActions[0].Key != "refresh" || multiResetActions[1].Key != "reset" {
		t.Fatalf("multi reset actions = %#v", multiResetActions)
	}

	multiResetUnchanged := FilterBarFieldsRefreshResetWhenChanged(
		"alerts.filters.reset.unchanged",
		"alerts 2 history 5",
		126,
		8,
		[]FilterBarField{FilterBarSearch("query", "Search", "").ForField("alertSearch")},
		filterBarShortcutIntent{"refresh"},
		filterBarShortcutIntent{"reset"},
		false,
		false,
		"Alerts are loading.",
	)
	multiResetUnchangedNode, ok := multiResetUnchanged.(*filterbar.VNode)
	if !ok {
		t.Fatalf("FilterBarFieldsRefreshResetWhenChanged returned %T, want *filterbar.VNode", multiResetUnchanged)
	}
	multiResetUnchangedActions := multiResetUnchangedNode.Props()["actions"].([]FilterBarAction)
	if len(multiResetUnchangedActions) != 2 || !multiResetUnchangedActions[1].Disabled || multiResetUnchangedActions[1].DisabledReason != "Nothing to reset." {
		t.Fatalf("multi reset unchanged actions = %#v, want disabled reset", multiResetUnchangedActions)
	}
}

func TestFilterBarPageSummaryShortcut(t *testing.T) {
	if got := FilterBarPageSummary(2, 34, "openai"); got != "page 2 · total 34 · search openai" {
		t.Fatalf("page summary = %q", got)
	}
	if got := FilterBarPageSummary(0, -1, ""); got != "page 1 · total 0 · search -" {
		t.Fatalf("page summary fallback = %q", got)
	}
	if got := FilterBarCompactPageSummary(1, 2, "very-long-provider-or-trace-lookup-value", 14); got != "page 1 · total 2 · search very-long-p..." {
		t.Fatalf("compact page summary = %q", got)
	}
	lookup := FilterBarLookupSummary(FilterBarLookupSummaryConfig{
		Lookup:      "trace-1",
		Source:      "input",
		Resolved:    "trace_id",
		ItemsLabel:  "spans",
		Items:       2,
		ErrorsLabel: "errors",
		Errors:      1,
	})
	if lookup != "lookup trace-1 · source input · resolved trace_id · spans 2 · errors 1" {
		t.Fatalf("lookup summary = %q", lookup)
	}
	actionable := FilterBarLookupSummary(FilterBarLookupSummaryConfig{
		LookupFallback:   "required",
		SourceFallback:   "none",
		ResolvedFallback: "pending",
		ItemsLabel:       "spans",
	})
	if actionable != "lookup required · source none · resolved pending · spans 0 · errors 0" {
		t.Fatalf("actionable lookup summary = %q", actionable)
	}
}

func TestFilterBarSummaryShortcuts(t *testing.T) {
	got := FilterBarSummary(
		FilterBarSummaryCount("alerts", 4),
		FilterBarSummaryCount("history", 12),
		FilterBarSummaryValue("status", "actionable"),
		FilterBarSummaryValue("source", "all"),
	)
	if got != "alerts 4 · history 12 · status actionable · source all" {
		t.Fatalf("summary = %q", got)
	}

	got = FilterBarSummary(
		FilterBarSummaryRatio("actions", 5, 8),
		FilterBarSummarySearch(""),
	)
	if got != "actions 5/8 · search -" {
		t.Fatalf("ratio summary = %q", got)
	}

	got = FilterBarSummary(
		FilterBarSummaryValue("page", "1"),
		FilterBarSummaryValueUnless("status", "all", "all"),
		FilterBarSummaryValueUnless("last", "failed", "all"),
		FilterBarSummaryPresence("reason", "maintenance", "ready", "missing"),
		FilterBarSummaryCompactUnless("search", "very-long-provider-or-trace-lookup-value", "", 14),
	)
	if got != "page 1 · last failed · reason ready · search very-long-p..." {
		t.Fatalf("unless summary = %q", got)
	}

	if got := FilterBarSummary(FilterBarSummaryPresence("reason", "", "", "")); got != "reason missing" {
		t.Fatalf("presence summary = %q", got)
	}

	wide := strings.Repeat("界", 20)
	value := strings.TrimPrefix(FilterBarSummary(FilterBarSummaryCompact("lookup", wide, 10)), "lookup ")
	if paint.StringWidth(value) > 10 {
		t.Fatalf("compact shortcut width = %d, want <= 10 (%q)", paint.StringWidth(value), value)
	}

	value = strings.TrimPrefix(FilterBarSummary(FilterBarSummaryCompactSearch(wide, 10)), "search ")
	if paint.StringWidth(value) > 10 {
		t.Fatalf("compact search shortcut width = %d, want <= 10 (%q)", paint.StringWidth(value), value)
	}
}
