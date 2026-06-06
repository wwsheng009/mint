package filterbar

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	rtlayout "github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
)

type testIntent struct {
	name string
}

func (i testIntent) IntentType() string { return i.name }

func TestBuilderAndProps(t *testing.T) {
	bar := NewBuilder().
		Key("providers").
		Title("Provider Filters").
		Summary("scope: prod\nstatus: degraded").
		Width(80).
		LabelWidth(8).
		Gap(2).
		RowGap(1).
		RootGap(2).
		Field(Search("query", "Query", "openai").WithPlaceholder("provider").WithWidth(20).ForField("query")).
		Field(Select("status", "Status", []Option{
			{Value: "all", Label: "All"},
			{Value: "degraded", Label: "Degraded"},
		}).WithSelectedIndex(1).ForField("status")).
		Action(Button("refresh", "Refresh", testIntent{"refresh"}).Primary()).
		Action(Button("export", "Export", testIntent{"export"}).WithDisabledReason("Select a provider first.\n")).
		BuildVNode()

	if bar.Key() != "providers" {
		t.Fatalf("key = %q, want providers", bar.Key())
	}
	props := bar.Props()
	if got := props[propTitle]; got != "Provider Filters" {
		t.Fatalf("title = %v, want Provider Filters", got)
	}
	if got := props[propSummary]; got != "scope: prod status: degraded" {
		t.Fatalf("summary = %v, want normalized summary", got)
	}
	if got := props[propWidth]; got != 80 {
		t.Fatalf("width = %v, want 80", got)
	}
	if got := props[propLabelWidth]; got != 8 {
		t.Fatalf("label width = %v, want 8", got)
	}
	if got := props[propRootGap]; got != 2 {
		t.Fatalf("root gap = %v, want 2", got)
	}
	fields := props[propFields].([]Field)
	if len(fields) != 2 {
		t.Fatalf("fields len = %d, want 2", len(fields))
	}
	if fields[0].Kind != FieldSearch || fields[0].FieldName != "query" {
		t.Fatalf("search field = %#v", fields[0])
	}
	if fields[1].Kind != FieldSelect || fields[1].SelectedIndex != 1 || !fields[1].HasSelectedIndex {
		t.Fatalf("select field = %#v", fields[1])
	}
	actions := props[propActions].([]Action)
	if len(actions) != 2 || actions[0].Variant != button.VariantPrimary {
		t.Fatalf("actions = %#v", actions)
	}
	if !actions[1].Disabled || actions[1].DisabledReason != "Select a provider first." {
		t.Fatalf("disabled action = %#v", actions[1])
	}
}

func TestActionWithDisabledReasonIf(t *testing.T) {
	enabled := Button("load", "Load", testIntent{"load"}).WithDisabledReasonIf(false, "Loading.")
	if enabled.Disabled || enabled.DisabledReason != "" {
		t.Fatalf("enabled action = %+v, want unchanged", enabled)
	}

	disabled := Button("load", "Load", testIntent{"load"}).WithDisabledReasonIf(true, "Loading.")
	if !disabled.Disabled || disabled.DisabledReason != "Loading." {
		t.Fatalf("disabled action = %+v, want disabled reason", disabled)
	}

	chained := Button("load", "Load", testIntent{"load"}).
		WithDisabledReasonIf(true, "Enter lookup.").
		WithDisabledReasonIf(false, "Ignored.").
		WithDisabledReasonIf(true, "Data is loading.")
	if !chained.Disabled || chained.DisabledReason != "Enter lookup. Data is loading." {
		t.Fatalf("chained action = %+v, want appended disabled reasons", chained)
	}
}

func TestActionWithDisabledReasonsJoinsNonEmptyReasons(t *testing.T) {
	enabled := Button("load", "Load", testIntent{"load"}).WithDisabledReasons("", " \n\t ")
	if enabled.Disabled || enabled.DisabledReason != "" {
		t.Fatalf("enabled action = %+v, want unchanged", enabled)
	}

	disabled := Button("load", "Load", testIntent{"load"}).WithDisabledReasons("Enter lookup.", "", "Data is loading.\n")
	if !disabled.Disabled || disabled.DisabledReason != "Enter lookup. Data is loading." {
		t.Fatalf("disabled action = %+v, want joined disabled reason", disabled)
	}

	if got := JoinDisabledReasons(" Enter lookup. ", "", "Data\tis loading."); got != "Enter lookup. Data is loading." {
		t.Fatalf("joined reasons = %q, want normalized reasons", got)
	}
}

func TestInstanceBuildsCompactInputFields(t *testing.T) {
	inst := NewInstance(rtui.Props{propKey: "filters"})

	tests := []struct {
		name       string
		field      Field
		wantSearch bool
	}{
		{
			name:       "search",
			field:      Search("query", "Search", "").WithWidth(20).ForField("query"),
			wantSearch: true,
		},
		{
			name:       "text",
			field:      Text("provider", "Provider", "").WithWidth(20).ForField("provider"),
			wantSearch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control := inst.buildFieldControl(tt.field)
			if control == nil || control.Tag() != "input" {
				t.Fatalf("control = %v, want input", control)
			}

			props := control.Props()
			if got := props["borderStyle"]; got != rtlayout.BorderNone {
				t.Fatalf("borderStyle = %v, want BorderNone for compact filter input", got)
			}
			if got := props["searchVariant"]; got != tt.wantSearch {
				t.Fatalf("searchVariant = %v, want %v", got, tt.wantSearch)
			}

			factory, ok := control.(rtui.InstanceFactory)
			if !ok {
				t.Fatalf("control %T does not create an instance", control)
			}
			measurable, ok := factory.CreateInstance().(interface {
				Measure(rtlayout.Constraints) rtlayout.Size
			})
			if !ok {
				t.Fatalf("control instance does not implement Measure")
			}
			if size := measurable.Measure(rtlayout.UnboundedConstraints()); size.Height != 1 {
				t.Fatalf("compact filter input height = %d, want 1", size.Height)
			}
		})
	}
}

func TestRefreshActionAndSearchRefreshPreset(t *testing.T) {
	action := RefreshAction(testIntent{"refresh"}, true, "Logs are loading.\n")
	if action.Key != "refresh" || action.Label != "Refresh" || action.Variant != button.VariantPrimary {
		t.Fatalf("refresh action = %#v", action)
	}
	if !action.Disabled || action.DisabledReason != "Logs are loading." {
		t.Fatalf("refresh disabled = %v reason = %q", action.Disabled, action.DisabledReason)
	}

	enabled := RefreshAction(testIntent{"refresh"}, false, "Logs are loading.")
	if enabled.Disabled || enabled.DisabledReason != "" {
		t.Fatalf("enabled refresh = %#v", enabled)
	}

	bar := SearchRefresh(
		"logs.filters",
		"page 1 total 25 search -",
		126,
		6,
		Search("query", "Search", "").WithPlaceholder("trace/request").WithWidth(42).ForField("logSearch"),
		testIntent{"refresh"},
		true,
		"Logs are loading.",
	)
	props := bar.Props()
	if got := props[propWidth]; got != 126 {
		t.Fatalf("width = %v, want 126", got)
	}
	if got := props[propLabelWidth]; got != 6 {
		t.Fatalf("label width = %v, want 6", got)
	}
	fields := props[propFields].([]Field)
	if len(fields) != 1 || fields[0].Key != "query" || fields[0].FieldName != "logSearch" {
		t.Fatalf("fields = %#v", fields)
	}
	actions := props[propActions].([]Action)
	if len(actions) != 1 || actions[0].Key != "refresh" || !actions[0].Disabled {
		t.Fatalf("actions = %#v", actions)
	}
	if actions[0].DisabledReason != "Logs are loading." {
		t.Fatalf("disabled reason = %q", actions[0].DisabledReason)
	}
}

func TestClearFieldActionAndSearchRefreshClearPreset(t *testing.T) {
	action := ClearFieldAction(" logSearch ", "trace-1", false, "")
	if action.Key != "clear" || action.Label != "Clear" || action.Variant != button.VariantSecondary {
		t.Fatalf("clear action = %#v", action)
	}
	event, ok := action.PressIntent.(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("clear intent = %T, want FieldChangeIntent", action.PressIntent)
	}
	if event.Field != "logSearch" || event.Value != "" {
		t.Fatalf("clear intent = %#v, want logSearch empty value", event)
	}

	if empty := ClearFieldAction("logSearch", "", false, ""); !empty.Disabled || empty.DisabledReason != "Nothing to clear." {
		t.Fatalf("empty clear action = %#v", empty)
	}
	if unbound := ClearFieldAction("", "trace-1", false, ""); !unbound.Disabled || unbound.DisabledReason != "No field is bound." {
		t.Fatalf("unbound clear action = %#v", unbound)
	}
	if busy := ClearFieldAction("logSearch", "trace-1", true, "Logs are loading."); !busy.Disabled || busy.DisabledReason != "Logs are loading." {
		t.Fatalf("busy clear action = %#v", busy)
	}

	named := ClearFieldActionWithLabel("clear-history", "Clear History", " alertHistorySearch ", "provider", false, "")
	if named.Key != "clear-history" || named.Label != "Clear History" || named.Variant != button.VariantSecondary {
		t.Fatalf("named clear action = %#v", named)
	}
	namedEvent, ok := named.PressIntent.(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("named clear intent = %T, want FieldChangeIntent", named.PressIntent)
	}
	if namedEvent.Field != "alertHistorySearch" || namedEvent.Value != "" {
		t.Fatalf("named clear intent = %#v, want alertHistorySearch empty value", namedEvent)
	}
	fallbackNamed := ClearFieldActionWithLabel("", "", "alertSearch", "quota", false, "")
	if fallbackNamed.Key != "clear" || fallbackNamed.Label != "Clear" {
		t.Fatalf("fallback named clear action = %#v, want default key/label", fallbackNamed)
	}

	bar := SearchRefreshClear(
		"logs.filters",
		"page 1 total 25 search trace-1",
		126,
		6,
		Search("query", "Search", "trace-1").WithPlaceholder("trace/request").WithWidth(42).ForField("logSearch"),
		testIntent{"refresh"},
		false,
		"Logs are loading.",
	)
	props := bar.Props()
	actions := props[propActions].([]Action)
	if len(actions) != 2 || actions[0].Key != "refresh" || actions[1].Key != "clear" {
		t.Fatalf("actions = %#v, want refresh and clear", actions)
	}
	if actions[1].Disabled {
		t.Fatalf("clear action disabled = %#v, want enabled", actions[1])
	}
}

func TestResetActionAndFieldsRefreshResetPreset(t *testing.T) {
	action := ResetAction(testIntent{"reset"}, false, "")
	if action.Key != "reset" || action.Label != "Reset" || action.Variant != button.VariantSecondary {
		t.Fatalf("reset action = %#v", action)
	}
	if got, ok := action.PressIntent.(testIntent); !ok || got.name != "reset" {
		t.Fatalf("reset intent = %#v, want reset intent", action.PressIntent)
	}

	busy := ResetAction(testIntent{"reset"}, true, "Alerts are loading.")
	if !busy.Disabled || busy.DisabledReason != "Alerts are loading." {
		t.Fatalf("busy reset action = %#v", busy)
	}
	unchanged := ResetActionWhenChanged(testIntent{"reset"}, false, false, "")
	if !unchanged.Disabled || unchanged.DisabledReason != "Nothing to reset." {
		t.Fatalf("unchanged reset action = %#v", unchanged)
	}
	changed := ResetActionWhenChanged(testIntent{"reset"}, true, false, "")
	if changed.Disabled || changed.DisabledReason != "" {
		t.Fatalf("changed reset action = %#v", changed)
	}
	changedBusy := ResetActionWhenChanged(testIntent{"reset"}, true, true, "Alerts are loading.")
	if !changedBusy.Disabled || changedBusy.DisabledReason != "Alerts are loading." {
		t.Fatalf("changed busy reset action = %#v", changedBusy)
	}

	bar := FieldsRefreshReset(
		"alerts.filters",
		"alerts 2 history 5",
		126,
		8,
		[]Field{
			Search("query", "Search", "quota").WithWidth(30).ForField("alertSearch"),
			Select("status", "Status", []Option{{Value: "actionable", Label: "Actionable"}}).WithSelectedIndex(0).ForField("alertStatus"),
		},
		testIntent{"refresh"},
		testIntent{"reset"},
		false,
		"Alerts are loading.",
	)
	props := bar.Props()
	if got := props[propWrap]; got != true {
		t.Fatalf("wrap = %v, want true", got)
	}
	actions := props[propActions].([]Action)
	if len(actions) != 2 || actions[0].Key != "refresh" || actions[1].Key != "reset" {
		t.Fatalf("actions = %#v, want refresh and reset", actions)
	}

	unchangedBar := FieldsRefreshResetWhenChanged(
		"alerts.filters.unchanged",
		"alerts 2 history 5",
		126,
		8,
		[]Field{Search("query", "Search", "").ForField("alertSearch")},
		testIntent{"refresh"},
		testIntent{"reset"},
		false,
		false,
		"Alerts are loading.",
	)
	unchangedActions := unchangedBar.Props()[propActions].([]Action)
	if len(unchangedActions) != 2 || !unchangedActions[1].Disabled || unchangedActions[1].DisabledReason != "Nothing to reset." {
		t.Fatalf("unchanged bar actions = %#v, want disabled reset", unchangedActions)
	}
}

func TestSearchActionsPresetBuildsOneSearchWithContextActions(t *testing.T) {
	bar := SearchActions(
		"trace.filters",
		"lookup trace-1 resolved -",
		126,
		8,
		Search("trace", "Trace", "trace-1").WithPlaceholder("trace id or request id").WithWidth(48).ForField("traceLookupID"),
		Button("refresh", "Load Trace", testIntent{"refresh"}).Primary(),
		Button("logs", "Back Logs", testIntent{"logs"}).WithDisabledReason("Trace is loading."),
	)
	props := bar.Props()
	if got := props[propWidth]; got != 126 {
		t.Fatalf("width = %v, want 126", got)
	}
	if got := props[propLabelWidth]; got != 8 {
		t.Fatalf("label width = %v, want 8", got)
	}
	fields := props[propFields].([]Field)
	if len(fields) != 1 || fields[0].Key != "trace" || fields[0].FieldName != "traceLookupID" {
		t.Fatalf("fields = %#v", fields)
	}
	actions := props[propActions].([]Action)
	if len(actions) != 2 || actions[0].Key != "refresh" || actions[1].Key != "logs" {
		t.Fatalf("actions = %#v", actions)
	}
	if actions[0].Variant != button.VariantPrimary {
		t.Fatalf("load action variant = %v, want primary", actions[0].Variant)
	}
	if !actions[1].Disabled || actions[1].DisabledReason != "Trace is loading." {
		t.Fatalf("back action disabled = %v reason = %q", actions[1].Disabled, actions[1].DisabledReason)
	}
}

func TestFieldsRefreshPresetBuildsWrappedMultiFieldBar(t *testing.T) {
	fields := []Field{
		Search("query", "Search", "latency").WithPlaceholder("alert keyword").WithWidth(30).ForField("alertSearch"),
		Select("status", "Status", []Option{
			{Value: "all", Label: "All"},
			{Value: "firing", Label: "Firing"},
		}).WithSelectedIndex(1).WithWidth(16).ForField("alertStatus"),
	}
	bar := FieldsRefresh(
		"alerts.filters",
		"alerts 2 history 5",
		126,
		8,
		fields,
		testIntent{"refresh"},
		true,
		"Alerts are loading.",
	)
	props := bar.Props()
	if got := props[propWidth]; got != 126 {
		t.Fatalf("width = %v, want 126", got)
	}
	if got := props[propLabelWidth]; got != 8 {
		t.Fatalf("label width = %v, want 8", got)
	}
	if got := props[propWrap]; got != true {
		t.Fatalf("wrap = %v, want true", got)
	}
	gotFields := props[propFields].([]Field)
	if len(gotFields) != 2 || gotFields[0].FieldName != "alertSearch" || gotFields[1].SelectedIndex != 1 {
		t.Fatalf("fields = %#v", gotFields)
	}
	actions := props[propActions].([]Action)
	if len(actions) != 1 || actions[0].Key != "refresh" || !actions[0].Disabled {
		t.Fatalf("actions = %#v", actions)
	}
	if actions[0].DisabledReason != "Alerts are loading." {
		t.Fatalf("disabled reason = %q", actions[0].DisabledReason)
	}
}

func TestSummaryHelpers(t *testing.T) {
	got := Summary(
		SummaryCount("groups", 3),
		SummaryRatio("providers", 2, 5),
		SummaryRatio("keys", 10, 12),
		SummarySearch("openai"),
	)
	if got != "groups 3 · providers 2/5 · keys 10/12 · search openai" {
		t.Fatalf("summary = %q", got)
	}

	fallback := Summary(
		SummaryValue("status", ""),
		SummarySearch(""),
		SummaryCount("negative", -3),
		SummaryRatio("ratio", -1, -2),
	)
	if fallback != "status - · search - · negative 0 · ratio 0/0" {
		t.Fatalf("fallback summary = %q", fallback)
	}

	changed := Summary(
		SummaryValue("page", "1"),
		SummaryValueUnless("status", "all", "all"),
		SummaryValueUnless("last", "failed", "all"),
		SummaryValueUnless("source", " ", "all"),
		SummaryPresence("reason", "maintenance", "ready", "missing"),
		SummaryPresence("ticket", "", "set", "required"),
		SummaryCompactUnless("search", "very-long-provider-or-trace-lookup-value", "", 14),
	)
	if changed != "page 1 · last failed · reason ready · ticket required · search very-long-p..." {
		t.Fatalf("changed summary = %q", changed)
	}

	if got := Summary(SummaryPresence("reason", "maintenance", "", "")); got != "reason ready" {
		t.Fatalf("presence default present = %q", got)
	}
	if got := Summary(SummaryPresence("reason", " \t\n", "", "")); got != "reason missing" {
		t.Fatalf("presence default missing = %q", got)
	}
}

func TestSummaryCompactUsesDisplayWidth(t *testing.T) {
	wide := strings.Repeat("界", 20)
	got := Summary(SummaryCompact("lookup", wide, 10))
	if !strings.HasPrefix(got, "lookup ") {
		t.Fatalf("compact summary = %q", got)
	}
	value := strings.TrimPrefix(got, "lookup ")
	if paint.StringWidth(value) > 10 {
		t.Fatalf("compact value width = %d, want <= 10 (%q)", paint.StringWidth(value), value)
	}
	if !strings.HasSuffix(value, "...") {
		t.Fatalf("compact value = %q, want ellipsis suffix", value)
	}
}

func TestPageSummaryHelper(t *testing.T) {
	if got := PageSummary(2, 34, "openai"); got != "page 2 · total 34 · search openai" {
		t.Fatalf("page summary = %q", got)
	}
	if got := PageSummary(0, -1, ""); got != "page 1 · total 0 · search -" {
		t.Fatalf("page summary fallback = %q", got)
	}
	if got := CompactPageSummary(1, 2, "very-long-provider-or-trace-lookup-value", 14); got != "page 1 · total 2 · search very-long-p..." {
		t.Fatalf("compact page summary = %q", got)
	}
}

func TestLookupSummaryHelper(t *testing.T) {
	got := LookupSummary(LookupSummaryConfig{
		Lookup:        "trace-1234567890-extra",
		LookupWidth:   14,
		Source:        "selected log",
		Resolved:      "request_id",
		ResolvedWidth: 16,
		ItemsLabel:    "spans",
		Items:         3,
		Errors:        1,
	})
	if got != "lookup trace-12345... · source selected log · resolved request_id · spans 3 · errors 1" {
		t.Fatalf("lookup summary = %q", got)
	}

	fallback := LookupSummary(LookupSummaryConfig{Items: -1, Errors: -2})
	if fallback != "lookup - · source - · resolved - · items 0 · errors 0" {
		t.Fatalf("fallback lookup summary = %q", fallback)
	}

	actionable := LookupSummary(LookupSummaryConfig{
		LookupFallback:   "required",
		SourceFallback:   "none",
		ResolvedFallback: "pending",
		ItemsLabel:       "spans",
		Items:            -1,
		Errors:           -2,
	})
	if actionable != "lookup required · source none · resolved pending · spans 0 · errors 0" {
		t.Fatalf("actionable lookup summary = %q", actionable)
	}
}

func TestRuntimeChildrenBuildControlsWithBindings(t *testing.T) {
	inst := NewBuilder().
		Key("providers").
		Title("Provider Filters").
		Summary("group: default").
		Width(72).
		LabelWidth(10).
		Field(Search("query", "Query", "openai").ForField("query")).
		Field(Select("status", "Status", []Option{
			{Value: "all", Label: "All"},
			{Value: "failed", Label: "Failed"},
		}).WithSelectedIndex(1).ForField("status")).
		Action(Button("reset", "Reset", testIntent{"reset"}).Secondary()).
		Action(Button("export", "Export", testIntent{"export"}).WithDisabledReason("Select one provider first.")).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "vstack" {
		t.Fatalf("root tag = %q, want vstack", root.Tag())
	}
	if got := root.Props()["gap"]; got != 0 {
		t.Fatalf("default root gap = %v, want compact summary/control layout", got)
	}

	search := findVNodeByKey(root, "providers-control-query")
	if search == nil {
		t.Fatal("search control not found")
	}
	if got := search.Props()["value"]; got != "openai" {
		t.Fatalf("search value = %v, want openai", got)
	}
	if _, ok := search.Props()["changeIntent"].(intent.FieldIntent); !ok {
		t.Fatalf("search changeIntent = %T, want FieldIntent", search.Props()["changeIntent"])
	}

	status := findVNodeByKey(root, "providers-control-status")
	if status == nil {
		t.Fatal("status control not found")
	}
	if got := status.Props()["selectedIndex"]; got != 1 {
		t.Fatalf("selectedIndex = %v, want 1", got)
	}
	if _, ok := status.Props()["changeIntent"].(intent.FieldIntent); !ok {
		t.Fatalf("select changeIntent = %T, want FieldIntent", status.Props()["changeIntent"])
	}

	reset := findVNodeByKey(root, "providers-action-reset")
	if reset == nil {
		t.Fatal("reset action not found")
	}
	if got := reset.Props()["variant"]; got != button.VariantSecondary {
		t.Fatalf("reset variant = %v, want secondary", got)
	}
	summary := findVNodeByKey(root, "providers-summary")
	if summary == nil {
		t.Fatal("summary node not found")
	}
	reasons := findVNodeByKey(root, "providers-disabled-reasons")
	if reasons == nil {
		t.Fatal("disabled reasons node not found")
	}
	export := findVNodeByKey(root, "providers-action-export")
	if export == nil {
		t.Fatal("export action not found")
	}
	if got := export.Props()["disabled"]; got != true {
		t.Fatalf("export disabled = %v, want true", got)
	}
}

func TestNormalizeFieldsAndActions(t *testing.T) {
	fields := normalizeFields([]Field{
		{Key: "dup", Width: -1, LabelWidth: -1, Kind: FieldKind("bad")},
		{Key: "dup", Kind: FieldSelect},
	})
	if fields[0].Key != "dup" || fields[1].Key != "dup-1" {
		t.Fatalf("field keys = %q, %q", fields[0].Key, fields[1].Key)
	}
	if fields[0].Width != 0 || fields[0].LabelWidth != 0 {
		t.Fatalf("field widths = %d/%d, want 0/0", fields[0].Width, fields[0].LabelWidth)
	}
	if fields[0].Kind != FieldText {
		t.Fatalf("field kind = %v, want text", fields[0].Kind)
	}
	if fields[1].SelectedIndex != -1 {
		t.Fatalf("select selected index = %d, want -1", fields[1].SelectedIndex)
	}

	actions := normalizeActions([]Action{{Key: "dup", Width: -1}, {Key: "dup"}})
	if actions[0].Key != "dup" || actions[1].Key != "dup-1" {
		t.Fatalf("action keys = %q, %q", actions[0].Key, actions[1].Key)
	}
	if actions[0].Width != 0 {
		t.Fatalf("action width = %d, want 0", actions[0].Width)
	}
}

func findVNodeByKey(node rtui.VNode, key string) rtui.VNode {
	if node == nil {
		return nil
	}
	if node.Key() == key {
		return node
	}
	for _, child := range node.Children() {
		if found := findVNodeByKey(child, key); found != nil {
			return found
		}
	}
	return nil
}
