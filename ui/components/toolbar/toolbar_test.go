package toolbar

import (
	"testing"

	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	menucomp "github.com/wwsheng009/mint/ui/components/menu"
)

type testIntent struct {
	name string
}

func (i testIntent) IntentType() string { return i.name }

func TestBuilderAndProps(t *testing.T) {
	bar := NewBuilder().
		Key("ops-toolbar").
		Title("Load Balancer").
		TitleWidth(16).
		Width(96).
		Gap(2).
		Dense(true).
		Left(Text("scope", "group: default").WithWidth(18)).
		Center(Badge("state", "degraded").WithColors("black", "yellow")).
		Action(Button("refresh", "Refresh", testIntent{"refresh"}).Primary().WithHelp("Reload current page")).
		Action(Button("reset", "Reset", testIntent{"reset"}).Danger().WithDisabledReason("Select a provider before reset")).
		BuildVNode()

	if bar.Key() != "ops-toolbar" {
		t.Fatalf("key = %q, want ops-toolbar", bar.Key())
	}
	props := bar.Props()
	if got := props[propTitle]; got != "Load Balancer" {
		t.Fatalf("title = %v, want Load Balancer", got)
	}
	if got := props[propTitleWidth]; got != 16 {
		t.Fatalf("titleWidth = %v, want 16", got)
	}
	if got := props[propWidth]; got != 96 {
		t.Fatalf("width = %v, want 96", got)
	}
	if got := props[propDense]; got != true {
		t.Fatalf("dense = %v, want true", got)
	}
	left := props[propLeftItems].([]Item)
	if len(left) != 1 || left[0].Key != "scope" || left[0].Width != 18 {
		t.Fatalf("left items = %#v", left)
	}
	center := props[propCenterItems].([]Item)
	if len(center) != 1 || center[0].Kind != ItemBadge {
		t.Fatalf("center items = %#v", center)
	}
	right := props[propRightItems].([]Item)
	if len(right) != 2 || right[0].Variant != button.VariantPrimary || right[1].Variant != button.VariantDanger {
		t.Fatalf("right items = %#v", right)
	}
	if !right[1].Disabled || right[1].DisabledReason != "Select a provider before reset" {
		t.Fatalf("reset disabled reason = disabled:%v reason:%q", right[1].Disabled, right[1].DisabledReason)
	}
}

func TestItemWithDisabledReasonsJoinsNonEmptyReasons(t *testing.T) {
	item := Button("reset", "Reset", testIntent{"reset"}).WithDisabledReasons(
		"  Enter a reason.  ",
		"",
		"Select a provider key first.\n",
	)
	if !item.Disabled || item.DisabledReason != "Enter a reason. Select a provider key first." {
		t.Fatalf("item = %+v, want joined disabled reasons", item)
	}

	enabled := Button("reload", "Reload", testIntent{"reload"}).WithDisabledReasons("", " ")
	if enabled.Disabled || enabled.DisabledReason != "" {
		t.Fatalf("enabled item = %+v, want unchanged when reasons are empty", enabled)
	}

	if got := JoinDisabledReasons("Reason required.", "Select target."); got != "Reason required. Select target." {
		t.Fatalf("joined reasons = %q", got)
	}
}

func TestItemWithDisabledReasonIf(t *testing.T) {
	enabled := Button("refresh", "Refresh", testIntent{"refresh"}).WithDisabledReasonIf(false, "Loading.")
	if enabled.Disabled || enabled.DisabledReason != "" {
		t.Fatalf("enabled item = %+v, want unchanged", enabled)
	}

	disabled := Button("refresh", "Refresh", testIntent{"refresh"}).WithDisabledReasonIf(true, "Loading.")
	if !disabled.Disabled || disabled.DisabledReason != "Loading." {
		t.Fatalf("disabled item = %+v, want disabled reason", disabled)
	}

	chained := Button("refresh", "Refresh", testIntent{"refresh"}).
		WithDisabledReasonIf(true, "Current page cannot refresh.").
		WithDisabledReasonIf(false, "Ignored.").
		WithDisabledReasonIf(true, "A request is running.")
	if !chained.Disabled || chained.DisabledReason != "Current page cannot refresh. A request is running." {
		t.Fatalf("chained item = %+v, want appended disabled reasons", chained)
	}
}

func TestPaginationPresetsExposeBoundaryAndLoadingReasons(t *testing.T) {
	prev := PaginationPrev(testIntent{"prev"}, false, 1, "Logs are loading.")
	if prev.Key != "prev" || prev.Label != "Prev" {
		t.Fatalf("prev item = %+v", prev)
	}
	if !prev.Disabled || prev.DisabledReason != "Already at the first page." {
		t.Fatalf("prev disabled = %v reason = %q", prev.Disabled, prev.DisabledReason)
	}

	next := PaginationNext(testIntent{"next"}, false, 3, 3, "Logs are loading.")
	if next.Key != "next" || next.Label != "Next" {
		t.Fatalf("next item = %+v", next)
	}
	if !next.Disabled || next.DisabledReason != "Already at the last page." {
		t.Fatalf("next disabled = %v reason = %q", next.Disabled, next.DisabledReason)
	}

	loading := PaginationNext(testIntent{"next"}, true, 2, 3, "Logs are loading.")
	if !loading.Disabled || loading.DisabledReason != "Logs are loading." {
		t.Fatalf("loading disabled = %v reason = %q", loading.Disabled, loading.DisabledReason)
	}

	enabled := PaginationPrev(testIntent{"prev"}, false, 2, "Logs are loading.")
	if enabled.Disabled || enabled.DisabledReason != "" {
		t.Fatalf("enabled prev disabled = %v reason = %q", enabled.Disabled, enabled.DisabledReason)
	}
}

func TestPaginationControlsPresetBuildsStandardToolbar(t *testing.T) {
	node := PaginationControls(
		"logs.pagination",
		126,
		Text("page", "page 1/3 total 51 size 25").WithWidth(72),
		testIntent{"prev"},
		testIntent{"next"},
		false,
		1,
		3,
		"Logs are loading.",
	)
	props := node.Props()
	if got := props[propWidth]; got != 126 {
		t.Fatalf("width = %v, want 126", got)
	}
	if got := props[propDense]; got != true {
		t.Fatalf("dense = %v, want true", got)
	}
	left := props[propLeftItems].([]Item)
	if len(left) != 1 || left[0].Key != "page" || left[0].Width != 72 {
		t.Fatalf("left items = %+v", left)
	}
	right := props[propRightItems].([]Item)
	if len(right) != 2 || right[0].Key != "prev" || right[1].Key != "next" {
		t.Fatalf("right items = %+v", right)
	}
	if !right[0].Disabled || right[0].DisabledReason != "Already at the first page." {
		t.Fatalf("prev disabled = %v reason = %q", right[0].Disabled, right[0].DisabledReason)
	}
}

func TestPageSummaryBuildsMultipleScopeSegments(t *testing.T) {
	item := PageSummary(
		2,
		5,
		101,
		25,
		96,
		PageSummaryPart{Label: "provider", Value: "openai"},
		PageSummaryPart{Label: "search", Value: "azure"},
		PageSummaryPart{Label: "status", Value: ""},
		PageSummaryPart{Value: "server"},
		PageSummaryPart{},
	)
	if item.Key != "page" || item.Label != "page 2/5 total 101 size 25 provider openai search azure status - server" || item.Width != 96 {
		t.Fatalf("page summary item = %+v", item)
	}

	fallback := PageSummary(0, 0, -1, -1, 0)
	if fallback.Label != "page 1/1 total 0 size 0" || fallback.Width != 72 {
		t.Fatalf("fallback page summary = %+v", fallback)
	}

	compact := PageSummary(
		1,
		1,
		2,
		25,
		72,
		CompactPageSummaryPart("search", "very-long-provider-or-trace-lookup-value", 14),
	)
	if compact.Label != "page 1/1 total 2 size 25 search very-long-p..." {
		t.Fatalf("compact page summary = %+v", compact)
	}

	optional := PageSummary(
		1,
		1,
		2,
		25,
		72,
		PageSummaryPartIfValue("provider", ""),
		PageSummaryPartIfValue("status", "active"),
		CompactPageSummaryPartIfValue("search", "very-long-provider-or-trace-lookup-value", 14),
		CompactPageSummaryPartIfValue("group", " ", 14),
	)
	if optional.Label != "page 1/1 total 2 size 25 status active search very-long-p..." {
		t.Fatalf("optional page summary = %+v", optional)
	}

	changed := PageSummary(
		1,
		1,
		2,
		25,
		72,
		PageSummaryPartUnless("status", "all", "all"),
		PageSummaryPartUnless("last", "failed", "all"),
		PageSummaryPartUnless("source", " ", "all"),
		CompactPageSummaryPartUnless("search", "very-long-provider-or-trace-lookup-value", "", 14),
	)
	if changed.Label != "page 1/1 total 2 size 25 last failed search very-long-p..." {
		t.Fatalf("changed page summary = %+v", changed)
	}
}

func TestNormalizePaginationScopeAppliesDefaultsAndFallbackTotal(t *testing.T) {
	scope := NormalizePaginationScope(0, 0, 0, 0, 17)
	if scope.Page != 1 || scope.PageSize != 25 || scope.Total != 17 || scope.TotalPages != 1 {
		t.Fatalf("scope = %+v, want default page/size and fallback total", scope)
	}

	scope = NormalizePaginationScope(3, 50, 101, 5, 17)
	if scope.Page != 3 || scope.PageSize != 50 || scope.Total != 101 || scope.TotalPages != 5 {
		t.Fatalf("scope = %+v, want explicit values", scope)
	}
}

func TestPaginationControlsWithScopeBuildsScopedSummary(t *testing.T) {
	scope := NormalizePaginationScope(2, 25, 76, 4, 0)
	node := PaginationControlsWithScope(
		"logs.pagination",
		126,
		scope,
		testIntent{"prev"},
		testIntent{"next"},
		false,
		"Logs are loading.",
		84,
		PageSummaryPart{Label: "search", Value: "trace-1"},
	)
	props := node.Props()
	if got := props[propKey]; got != "logs.pagination" {
		t.Fatalf("key = %v, want logs.pagination", got)
	}
	left := props[propLeftItems].([]Item)
	if len(left) != 1 {
		t.Fatalf("left items len = %d, want 1", len(left))
	}
	if left[0].Label != "page 2/4 total 76 size 25 search trace-1" {
		t.Fatalf("summary label = %q", left[0].Label)
	}
	right := props[propRightItems].([]Item)
	if len(right) != 2 || right[0].Key != "prev" || right[1].Key != "next" {
		t.Fatalf("right items = %+v, want prev/next", right)
	}
}

func TestPaginationConfigBuildsItemsAndControls(t *testing.T) {
	config := PaginationConfig{
		Key:           "jobs.pagination",
		Width:         118,
		Summary:       Text("page", "page 2/5 total 101 size 25").WithWidth(72),
		PrevIntent:    testIntent{"prev"},
		NextIntent:    testIntent{"next"},
		Busy:          false,
		Page:          2,
		TotalPages:    5,
		LoadingReason: "Background jobs are loading.",
	}

	prev := config.Prev()
	if prev.Disabled || prev.Key != "prev" {
		t.Fatalf("prev = %+v, want enabled prev", prev)
	}
	next := config.Next()
	if next.Disabled || next.Key != "next" {
		t.Fatalf("next = %+v, want enabled next", next)
	}

	controls := config.Controls()
	props := controls.Props()
	if got := props[propWidth]; got != 118 {
		t.Fatalf("controls width = %v, want 118", got)
	}
	if got := props[propDense]; got != true {
		t.Fatalf("controls dense = %v, want true", got)
	}
	left := props[propLeftItems].([]Item)
	if len(left) != 1 || left[0].Key != "page" || left[0].Label != "page 2/5 total 101 size 25" {
		t.Fatalf("left items = %+v", left)
	}
	right := props[propRightItems].([]Item)
	if len(right) != 2 || right[0].Key != "prev" || right[1].Key != "next" {
		t.Fatalf("right items = %+v", right)
	}
}

func TestPaginationConfigDefaultsAndDisabledReasons(t *testing.T) {
	config := PaginationConfig{
		Key:           "alerts.history.pagination",
		PrevIntent:    testIntent{"prev"},
		NextIntent:    testIntent{"next"},
		Busy:          true,
		Page:          1,
		TotalPages:    3,
		LoadingReason: "Alert history is loading.",
	}

	if got := config.Controls().Props()[propWidth]; got != 96 {
		t.Fatalf("controls default width = %v, want 96", got)
	}
	if prev := config.Prev(); !prev.Disabled || prev.DisabledReason != "Alert history is loading." {
		t.Fatalf("busy prev disabled = %v reason = %q", prev.Disabled, prev.DisabledReason)
	}
	if next := (PaginationConfig{Page: 3, TotalPages: 3}).Next(); !next.Disabled || next.DisabledReason != "Already at the last page." {
		t.Fatalf("last-page next disabled = %v reason = %q", next.Disabled, next.DisabledReason)
	}
	left := config.Controls().Props()[propLeftItems].([]Item)
	if len(left) != 1 || left[0].Key != "page" || left[0].Label != "-" || left[0].Width != 72 {
		t.Fatalf("default summary = %+v", left)
	}
}

func TestSelectionPresetsExposeBoundaryAndLoadingReasons(t *testing.T) {
	prev := SelectionPrev(testIntent{"prev"}, false, 0, 3, "log", "Logs are loading.")
	if prev.Key != "select-prev" || prev.Label != "Up" {
		t.Fatalf("prev item = %+v", prev)
	}
	if !prev.Disabled || prev.DisabledReason != "Already at the first log." {
		t.Fatalf("prev disabled = %v reason = %q", prev.Disabled, prev.DisabledReason)
	}

	next := SelectionNext(testIntent{"next"}, false, 2, 3, "log", "Logs are loading.")
	if next.Key != "select-next" || next.Label != "Down" {
		t.Fatalf("next item = %+v", next)
	}
	if !next.Disabled || next.DisabledReason != "Already at the last log." {
		t.Fatalf("next disabled = %v reason = %q", next.Disabled, next.DisabledReason)
	}

	empty := SelectionNext(testIntent{"next"}, false, 0, 0, "log", "Logs are loading.")
	if !empty.Disabled || empty.DisabledReason != "No log available." {
		t.Fatalf("empty disabled = %v reason = %q", empty.Disabled, empty.DisabledReason)
	}

	loading := SelectionNext(testIntent{"next"}, true, 1, 3, "log", "Logs are loading.")
	if !loading.Disabled || loading.DisabledReason != "Logs are loading." {
		t.Fatalf("loading disabled = %v reason = %q", loading.Disabled, loading.DisabledReason)
	}

	enabled := SelectionPrev(testIntent{"prev"}, false, 1, 3, "log", "Logs are loading.")
	if enabled.Disabled || enabled.DisabledReason != "" {
		t.Fatalf("enabled prev disabled = %v reason = %q", enabled.Disabled, enabled.DisabledReason)
	}
}

func TestSelectionControlsPresetBuildsStandardToolbar(t *testing.T) {
	node := SelectionControls("logs.selection", "Log", testIntent{"prev"}, testIntent{"next"}, false, 0, 2, "log", "Logs are loading.")
	props := node.Props()
	if got := props[propTitle]; got != "Log" {
		t.Fatalf("title = %v, want Log", got)
	}
	if got := props[propTitleWidth]; got != 10 {
		t.Fatalf("titleWidth = %v, want 10", got)
	}
	if got := props[propWidth]; got != 54 {
		t.Fatalf("width = %v, want 54", got)
	}
	if got := props[propDense]; got != true {
		t.Fatalf("dense = %v, want true", got)
	}
	items := props[propLeftItems].([]Item)
	if len(items) != 2 || items[0].Key != "select-prev" || items[1].Key != "select-next" {
		t.Fatalf("items = %+v", items)
	}
	if !items[0].Disabled || items[0].DisabledReason != "Already at the first log." {
		t.Fatalf("prev disabled = %v reason = %q", items[0].Disabled, items[0].DisabledReason)
	}
}

func TestSelectionActionControlsPresetBuildsStandardToolbar(t *testing.T) {
	node := SelectionActionControls(
		"logs.selection.actions",
		"Log",
		56,
		testIntent{"prev"},
		testIntent{"next"},
		false,
		0,
		2,
		"log",
		"Logs are loading.",
		Button("open-trace", "Open Trace", testIntent{"open"}).Primary(),
	)
	props := node.Props()
	if got := props[propTitle]; got != "Log" {
		t.Fatalf("title = %v, want Log", got)
	}
	if got := props[propTitleWidth]; got != 10 {
		t.Fatalf("titleWidth = %v, want 10", got)
	}
	if got := props[propWidth]; got != 56 {
		t.Fatalf("width = %v, want 56", got)
	}
	if got := props[propDense]; got != true {
		t.Fatalf("dense = %v, want true", got)
	}
	items := props[propLeftItems].([]Item)
	if len(items) != 3 || items[0].Key != "select-prev" || items[1].Key != "select-next" || items[2].Key != "open-trace" {
		t.Fatalf("items = %+v", items)
	}
	if !items[0].Disabled || items[0].DisabledReason != "Already at the first log." {
		t.Fatalf("prev disabled = %v reason = %q", items[0].Disabled, items[0].DisabledReason)
	}
	if items[2].Variant != button.VariantPrimary {
		t.Fatalf("action variant = %v, want primary", items[2].Variant)
	}

	defaultWidth := SelectionActionControls("trace.actions", "Span", 0, testIntent{"prev"}, testIntent{"next"}, false, 1, 2, "span", "")
	if got := defaultWidth.Props()[propWidth]; got != 64 {
		t.Fatalf("default width = %v, want 64", got)
	}
}

func TestSelectionConfigBuildsItemsControlsAndActions(t *testing.T) {
	config := SelectionConfig{
		Key:           "jobs.selection.actions",
		Title:         "Job",
		Width:         62,
		PrevIntent:    testIntent{"prev"},
		NextIntent:    testIntent{"next"},
		Busy:          false,
		Index:         1,
		Total:         3,
		ItemLabel:     "job",
		LoadingReason: "Background jobs are loading.",
	}

	prev := config.Prev()
	if prev.Disabled || prev.Key != "select-prev" {
		t.Fatalf("prev = %+v, want enabled select-prev", prev)
	}
	next := config.Next()
	if next.Disabled || next.Key != "select-next" {
		t.Fatalf("next = %+v, want enabled select-next", next)
	}

	controls := config.Controls()
	props := controls.Props()
	if got := props[propWidth]; got != 62 {
		t.Fatalf("controls width = %v, want 62", got)
	}
	if got := props[propTitle]; got != "Job" {
		t.Fatalf("controls title = %v, want Job", got)
	}
	items := props[propLeftItems].([]Item)
	if len(items) != 2 || items[0].Key != "select-prev" || items[1].Key != "select-next" {
		t.Fatalf("controls items = %+v", items)
	}

	actions := config.ActionControls(Button("open", "Open", testIntent{"open"}).Primary())
	actionItems := actions.Props()[propLeftItems].([]Item)
	if len(actionItems) != 3 || actionItems[2].Key != "open" {
		t.Fatalf("action items = %+v", actionItems)
	}
	if actionItems[2].Variant != button.VariantPrimary {
		t.Fatalf("action variant = %v, want primary", actionItems[2].Variant)
	}
}

func TestSelectionConfigDefaultsWidthsAndDisabledReasons(t *testing.T) {
	config := SelectionConfig{
		Key:           "alerts.selection.actions",
		Title:         "Alert",
		PrevIntent:    testIntent{"prev"},
		NextIntent:    testIntent{"next"},
		Busy:          true,
		Index:         0,
		Total:         2,
		ItemLabel:     "alert",
		LoadingReason: "Alerts data is loading.",
	}

	if got := config.Controls().Props()[propWidth]; got != 54 {
		t.Fatalf("controls default width = %v, want 54", got)
	}
	if got := config.ActionControls().Props()[propWidth]; got != 64 {
		t.Fatalf("action controls default width = %v, want 64", got)
	}
	if next := config.Next(); !next.Disabled || next.DisabledReason != "Alerts data is loading." {
		t.Fatalf("busy next disabled = %v reason = %q", next.Disabled, next.DisabledReason)
	}
}

func TestActionControlsPresetBuildsCompactToolbar(t *testing.T) {
	node := ActionControls(
		"login.actions",
		88,
		Button("login", "Login", testIntent{"login"}).Primary(),
		Button("refresh", "Refresh", testIntent{"refresh"}).WithDisabledReason("Captcha is loading."),
	)
	props := node.Props()
	if got := props[propKey]; got != "login.actions" {
		t.Fatalf("key = %v, want login.actions", got)
	}
	if got := props[propWidth]; got != 88 {
		t.Fatalf("width = %v, want 88", got)
	}
	if got := props[propDense]; got != true {
		t.Fatalf("dense = %v, want true", got)
	}
	items := props[propLeftItems].([]Item)
	if len(items) != 2 || items[0].Key != "login" || items[1].Key != "refresh" {
		t.Fatalf("items = %+v", items)
	}
	if !items[1].Disabled || items[1].DisabledReason != "Captcha is loading." {
		t.Fatalf("disabled item = %+v", items[1])
	}

	defaultWidth := ActionControls("login.actions", 0, Button("login", "Login", testIntent{"login"}))
	if got := defaultWidth.Props()[propWidth]; got != 64 {
		t.Fatalf("default width = %v, want 64", got)
	}
}

func TestActionGroupPresetBuildsStandardToolbar(t *testing.T) {
	node := ActionGroup("actions.global", "Global", []Item{
		Button("reload", "Reload", testIntent{"reload"}).Primary(),
		Button("reset", "Reset", testIntent{"reset"}).Danger().WithDisabledReason("Reason required."),
	})
	props := node.Props()
	if got := props[propTitle]; got != "Global" {
		t.Fatalf("title = %v, want Global", got)
	}
	if got := props[propTitleWidth]; got != 16 {
		t.Fatalf("titleWidth = %v, want 16", got)
	}
	if got := props[propWidth]; got != 118 {
		t.Fatalf("width = %v, want 118", got)
	}
	if got := props[propDense]; got != true {
		t.Fatalf("dense = %v, want true", got)
	}
	items := props[propLeftItems].([]Item)
	if len(items) != 2 || items[0].Key != "reload" || items[1].Key != "reset" {
		t.Fatalf("items = %+v", items)
	}
	if !items[1].Disabled || items[1].DisabledReason != "Reason required." {
		t.Fatalf("disabled action = %+v", items[1])
	}

	custom := ActionGroupWithLayout("actions.key", "Key", 126, 10, []Item{
		Button("disable", "Disable", testIntent{"disable"}).Danger(),
	})
	customProps := custom.Props()
	if customProps[propWidth] != 126 || customProps[propTitleWidth] != 10 {
		t.Fatalf("custom action group layout = width %v titleWidth %v, want 126/10", customProps[propWidth], customProps[propTitleWidth])
	}
}

func TestActionGroupsPresetBuildsGroupedOperationSurface(t *testing.T) {
	node := ActionGroups("actions.groups", []ActionGroupConfig{
		{
			Key:   "actions.global",
			Title: "Global",
			Items: []Item{
				Button("reload", "Reload", testIntent{"reload"}).Primary(),
			},
		},
		{
			Key:   "actions.empty",
			Title: "Empty",
		},
		{
			Key:        "actions.target",
			Title:      "Selected Target",
			Width:      126,
			TitleWidth: 18,
			Items: []Item{
				Button("disable", "Disable", testIntent{"disable"}).Danger().WithDisabledReason("Reason required."),
			},
		},
	}, "selected key: provider/default/key-1", "No actions matched.")
	if node.Tag() != "vstack" {
		t.Fatalf("tag = %q, want vstack", node.Tag())
	}
	if node.Key() != "actions.groups" {
		t.Fatalf("key = %q, want actions.groups", node.Key())
	}
	children := node.Children()
	if len(children) != 3 {
		t.Fatalf("children = %d, want global/target/summary", len(children))
	}
	if got := children[0].Props()[propTitle]; got != "Global" {
		t.Fatalf("first title = %v, want Global", got)
	}
	if got := children[1].Props()[propTitle]; got != "Selected Target" {
		t.Fatalf("second title = %v, want Selected Target", got)
	}
	if got := children[1].Props()[propWidth]; got != 126 {
		t.Fatalf("second width = %v, want configured width 126", got)
	}
	if got := children[1].Props()[propTitleWidth]; got != 18 {
		t.Fatalf("second title width = %v, want configured title width 18", got)
	}
	items := children[1].Props()[propLeftItems].([]Item)
	if len(items) != 1 || items[0].Key != "disable" || items[0].DisabledReason != "Reason required." {
		t.Fatalf("target items = %+v", items)
	}
	summary := children[2]
	if summary.Tag() != "text" || summary.Props()["content"] != "selected key: provider/default/key-1" {
		t.Fatalf("summary node = %s %+v", summary.Tag(), summary.Props())
	}
}

func TestActionTargetSummaryBuildsOperationalTargetText(t *testing.T) {
	summary := ActionTargetSummary(
		CompactActionTargetPart("group", "default", 24),
		CompactActionTargetPart("provider", "very-long-provider-name-for-runtime", 16),
		CompactActionTargetPart("key", "", 12),
	)
	if summary != "group default · provider very-long-pro... · key -" {
		t.Fatalf("summary = %q, want compact target summary", summary)
	}

	if got := ActionTargetSummary(ActionTargetPart{Value: "endpoint"}); got != "endpoint" {
		t.Fatalf("unlabeled target summary = %q, want endpoint", got)
	}
	if got := ActionTargetSummary(ActionTargetPart{}); got != "" {
		t.Fatalf("blank target summary = %q, want empty", got)
	}
}

func TestActionTargetSummaryWithScopePrefixesActiveOperationScope(t *testing.T) {
	summary := ActionTargetSummaryWithScope(
		"provider",
		CompactActionTargetPart("group", "default", 24),
		CompactActionTargetPart("provider", "openai", 16),
		CompactActionTargetPart("key", "key-op...ai-1", 24),
	)
	if summary != "scope provider · group default · provider openai · key key-op...ai-1" {
		t.Fatalf("scoped summary = %q, want operation scope prefix", summary)
	}

	if got := ActionTargetSummaryWithScope(" ", ActionTargetPart{Value: "endpoint"}); got != "endpoint" {
		t.Fatalf("blank scope summary = %q, want unscoped target summary", got)
	}
}

func TestActionTargetSummaryWithScopesPrefixesVisibleOperationScopes(t *testing.T) {
	summary := ActionTargetSummaryWithScopes(
		[]string{"global", "provider", "key"},
		CompactActionTargetPart("endpoint", "http://127.0.0.1:8080", 40),
		CompactActionTargetPart("group", "default", 24),
		CompactActionTargetPart("provider", "openai", 16),
		CompactActionTargetPart("key", "key-op...ai-1", 24),
	)
	if summary != "scopes global/provider/key · endpoint http://127.0.0.1:8080 · group default · provider openai · key key-op...ai-1" {
		t.Fatalf("scoped summary = %q, want visible scopes and target chain", summary)
	}

	if got := ActionTargetSummaryWithScopes([]string{"key", "key", " "}, ActionTargetPart{Value: "selected"}); got != "scope key · selected" {
		t.Fatalf("single normalized scope summary = %q, want single-scope format", got)
	}

	if got := ActionTargetSummaryWithScopes(nil, ActionTargetPart{Value: "endpoint"}); got != "endpoint" {
		t.Fatalf("empty scopes summary = %q, want unscoped target summary", got)
	}
}

func TestActionGroupsPresetShowsEmptyTextWhenAllGroupsAreEmpty(t *testing.T) {
	node := ActionGroups("actions.groups", []ActionGroupConfig{
		{Key: "actions.global", Title: "Global"},
	}, "", "No actions matched.")
	children := node.Children()
	if len(children) != 1 {
		t.Fatalf("children = %d, want empty text only", len(children))
	}
	if children[0].Tag() != "text" || children[0].Props()["content"] != "No actions matched." {
		t.Fatalf("empty node = %s %+v", children[0].Tag(), children[0].Props())
	}
}

func TestRuntimeChildrenBuildToolbarControls(t *testing.T) {
	inst := NewBuilder().
		Key("ops-toolbar").
		Title("Runtime").
		Left(Text("focus", "F2 Load Balancer")).
		Right(Button("refresh", "Refresh", testIntent{"refresh"}).Primary()).
		Right(Button("reload", "Reload", testIntent{"reload"}).Danger().WithDisabledReason("Runtime is already reloading")).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "hstack" {
		t.Fatalf("root tag = %q, want hstack", root.Tag())
	}
	if findVNodeByKey(root, "ops-toolbar-title") == nil {
		t.Fatal("title node not found")
	}
	refresh := findVNodeByKey(root, "ops-toolbar-right-refresh")
	if refresh == nil {
		t.Fatal("refresh action not found")
	}
	if got := refresh.Props()["variant"]; got != button.VariantPrimary {
		t.Fatalf("refresh variant = %v, want primary", got)
	}
	reload := findVNodeByKey(root, "ops-toolbar-right-reload")
	if reload == nil {
		t.Fatal("reload action not found")
	}
	if got := reload.Props()["disabled"]; got != true {
		t.Fatalf("reload disabled = %v, want true", got)
	}
	tooltip := findVNodeByKey(root, "ops-toolbar-right-reload-tooltip")
	if tooltip == nil {
		t.Fatal("reload disabled reason tooltip not found")
	}
	if got := tooltip.GetLayer(); got != rtui.LayerBase {
		t.Fatalf("tooltip wrapper layer = %v, want %v so the button stays below overlays", got, rtui.LayerBase)
	}
	if got := tooltip.Props()["text"]; got != "Runtime is already reloading" {
		t.Fatalf("tooltip text = %v, want disabled reason", got)
	}
}

func TestRuntimeChildrenBuildStatusBarMode(t *testing.T) {
	inst := NewBuilder().
		Key("ops-status").
		Title("Manager").
		UseStatusBar(true).
		Left(Badge("mode", "ADMIN").WithHelp("Admin mode")).
		Right(Button("reload", "Reload", testIntent{"reload"}).WithDisabledReason("Select a target first")).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "fragment" && root.Tag() != "hstack" && root.Tag() != "vstack" {
		t.Fatalf("unexpected statusbar root tag = %q", root.Tag())
	}
	if findVNodeByKey(root, "ops-status-left-mode") == nil {
		t.Fatal("statusbar badge section not found")
	}
	if findVNodeByKey(root, "ops-status-right-reload") == nil {
		t.Fatal("statusbar disabled action not found")
	}
}

func TestRuntimeChildrenBuildControlledDropdownMenu(t *testing.T) {
	items := []menucomp.MenuItem{
		menucomp.Action("reload", "Reload", testIntent{"reload"}),
		menucomp.Action("inspect", "Inspect", testIntent{"inspect"}).WithDescription("Open diagnostics"),
	}
	inst := NewBuilder().
		Key("ops-toolbar").
		Right(Dropdown("more", "More", items, true).
			WithMenuID("ops-actions").
			WithMenuPlacement(menucomp.PlacementBottomEnd).
			WithMenuMinWidth(24).
			WithMenuMaxHeight(8).
			WithMenuDescriptions(true)).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "fragment" {
		t.Fatalf("root tag = %q, want fragment when dropdown is open", root.Tag())
	}
	buttonNode := findVNodeByKey(root, "ops-toolbar-right-more")
	if buttonNode == nil {
		t.Fatal("dropdown button not found")
	}
	if buttonNode.ID() != "ops-toolbar-right-more" {
		t.Fatalf("dropdown button ID = %q, want anchorable ID", buttonNode.ID())
	}
	pressIntent, ok := buttonNode.Props()["pressIntent"].(menucomp.OpenMenuIntent)
	if !ok {
		t.Fatalf("press intent = %T, want menu.OpenMenuIntent", buttonNode.Props()["pressIntent"])
	}
	if pressIntent.MenuID != "ops-actions" {
		t.Fatalf("OpenMenuIntent.MenuID = %q, want ops-actions", pressIntent.MenuID)
	}

	portal := findVNodeByKey(root, "ops-actions-portal")
	if portal == nil {
		t.Fatal("dropdown menu portal not found")
	}
	if got, _ := portal.Props()["anchorId"].(string); got != "ops-toolbar-right-more" {
		t.Fatalf("portal anchorId = %q, want toolbar button ID", got)
	}
	if got, _ := portal.Props()["anchor"].(rttypes.Anchor); got != rttypes.AnchorBottomRight {
		t.Fatalf("portal anchor = %v, want AnchorBottomRight", got)
	}
	if got, _ := portal.Props()["popupPlacement"].(string); got != string(menucomp.PlacementBottomEnd) {
		t.Fatalf("popupPlacement = %q, want bottom-end", got)
	}
}

func TestDropdownClosedOnlyRendersButton(t *testing.T) {
	inst := NewBuilder().
		Key("ops-toolbar").
		Right(Dropdown("more", "More", []menucomp.MenuItem{
			menucomp.Action("reload", "Reload", testIntent{"reload"}),
		}, false)).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "hstack" {
		t.Fatalf("root tag = %q, want hstack when dropdown is closed", root.Tag())
	}
	buttonNode := findVNodeByKey(root, "ops-toolbar-right-more")
	if buttonNode == nil {
		t.Fatal("dropdown button not found")
	}
	pressIntent, ok := buttonNode.Props()["pressIntent"].(menucomp.OpenMenuIntent)
	if !ok {
		t.Fatalf("press intent = %T, want menu.OpenMenuIntent", buttonNode.Props()["pressIntent"])
	}
	if pressIntent.MenuID != "ops-toolbar-right-more-menu" {
		t.Fatalf("OpenMenuIntent.MenuID = %q, want generated menu id", pressIntent.MenuID)
	}
}

func TestNormalizeItems(t *testing.T) {
	items := normalizeItems([]Item{
		{Key: "dup", Kind: ItemKind("bad"), Width: -1},
		{Key: "dup", Kind: ItemButton},
		{Kind: ItemSeparator},
	})
	if items[0].Key != "dup" || items[1].Key != "dup-1" || items[2].Key != "item-2" {
		t.Fatalf("item keys = %q, %q, %q", items[0].Key, items[1].Key, items[2].Key)
	}
	if items[0].Kind != ItemText {
		t.Fatalf("first item kind = %v, want text", items[0].Kind)
	}
	if items[0].Width != 0 {
		t.Fatalf("width = %d, want 0", items[0].Width)
	}
}

func TestOperationalPresets(t *testing.T) {
	tests := []struct {
		name string
		got  Item
		want Item
	}{
		{
			name: "key value",
			got:  KeyValue("endpoint", "endpoint", "http://localhost:8080"),
			want: Item{Key: "endpoint", Label: "endpoint: http://localhost:8080", Kind: ItemText},
		},
		{
			name: "key value empty",
			got:  KeyValue("scope", "scope", ""),
			want: Item{Key: "scope", Label: "scope: -", Kind: ItemText},
		},
		{
			name: "scope",
			got:  Scope("group: default"),
			want: Item{Key: "scope", Label: "scope: group: default", Kind: ItemText},
		},
		{
			name: "selection",
			got:  Selection("provider/openai"),
			want: Item{Key: "selection", Label: "selection: provider/openai", Kind: ItemText, FgColor: "bright-black"},
		},
	}

	for _, tt := range tests {
		if tt.got.Key != tt.want.Key || tt.got.Label != tt.want.Label || tt.got.Kind != tt.want.Kind || tt.got.FgColor != tt.want.FgColor {
			t.Fatalf("%s item = %+v, want %+v", tt.name, tt.got, tt.want)
		}
	}

	state := StateBadge("state", "degraded")
	if state.Kind != ItemBadge || state.Label != "degraded" || state.FgColor != "black" || state.BgColor != "yellow" {
		t.Fatalf("state badge = %+v", state)
	}
	busy := BusyBadge("busy", "")
	if busy.Label != "busy" || busy.BgColor != "yellow" {
		t.Fatalf("busy badge = %+v", busy)
	}
	err := ErrorBadge("error", "failed")
	if err.Label != "failed" || err.BgColor != "red" {
		t.Fatalf("error badge = %+v", err)
	}
}

func TestShellHeaderPresetBuildsStandardOperationsHeader(t *testing.T) {
	header := ShellHeader(
		"manager.header",
		"ai-gateway-manager",
		"http://127.0.0.1:8080",
		"admin: ops",
		Button("refresh", "Refresh", testIntent{"refresh"}).Primary(),
		Button("logout", "Logout", testIntent{"logout"}).WithDisabledReason("No local session."),
	)
	if header == nil || header.Tag() != "toolbar" {
		t.Fatalf("shell header = %v, want toolbar", header)
	}
	if header.Key() != "manager.header" {
		t.Fatalf("shell header key = %q, want manager.header", header.Key())
	}
	props := header.Props()
	if props[propTitle] != "ai-gateway-manager" || props[propTitleWidth] != 22 {
		t.Fatalf("title props = %v/%v, want app title with width 22", props[propTitle], props[propTitleWidth])
	}
	left := props[propLeftItems].([]Item)
	if len(left) != 1 || left[0].Key != "base-url" || left[0].Label != "http://127.0.0.1:8080" || left[0].FgColor != "bright-black" {
		t.Fatalf("left items = %+v", left)
	}
	center := props[propCenterItems].([]Item)
	if len(center) != 1 || center[0].Key != "auth" || center[0].Label != "admin: ops" || center[0].BgColor != "cyan" {
		t.Fatalf("center items = %+v", center)
	}
	right := props[propRightItems].([]Item)
	if len(right) != 2 || right[0].Key != "refresh" || right[1].DisabledReason != "No local session." {
		t.Fatalf("right items = %+v", right)
	}

	fallback := ShellHeader("", "", "", "")
	fallbackProps := fallback.Props()
	if fallback.Key() != "shell.header" || fallbackProps[propTitle] != "-" {
		t.Fatalf("fallback shell header key/title = %q/%v", fallback.Key(), fallbackProps[propTitle])
	}
	fallbackLeft := fallbackProps[propLeftItems].([]Item)
	fallbackCenter := fallbackProps[propCenterItems].([]Item)
	if fallbackLeft[0].Label != "-" || fallbackCenter[0].Label != "-" {
		t.Fatalf("fallback context items = %+v %+v", fallbackLeft, fallbackCenter)
	}
}

func TestShellNavPresetBuildsDenseNavigationToolbar(t *testing.T) {
	nav := ShellNav("manager.nav", []Item{
		Button("nav-overview", "F1 Overview", testIntent{"overview"}).Primary(),
		Button("nav-actions", "F8 Actions", testIntent{"actions"}).Secondary().WithDisabledReason("Login is required."),
	})
	if nav == nil || nav.Tag() != "toolbar" {
		t.Fatalf("shell nav = %v, want toolbar", nav)
	}
	props := nav.Props()
	if nav.Key() != "manager.nav" || props[propDense] != true {
		t.Fatalf("nav key/dense = %q/%v, want manager.nav/dense", nav.Key(), props[propDense])
	}
	left := props[propLeftItems].([]Item)
	if len(left) != 2 || left[0].Key != "nav-overview" || left[1].DisabledReason != "Login is required." {
		t.Fatalf("nav left items = %+v", left)
	}

	fallback := ShellNav("", nil)
	if fallback.Key() != "shell.nav" {
		t.Fatalf("fallback shell nav key = %q, want shell.nav", fallback.Key())
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
