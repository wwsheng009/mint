package ui

import (
	"testing"
	"time"
)

type testToolbarIntent struct{}

func (testToolbarIntent) IntentType() string { return "test.toolbar" }

func TestToolbarShortcuts(t *testing.T) {
	bar := NewToolbarBuilder().
		Key("ops").
		Title("Ops").
		Left(ToolbarText("scope", "group: default")).
		Center(ToolbarBadge("state", "healthy")).
		Right(ToolbarButton("refresh", "Refresh", testToolbarIntent{}).Primary()).
		Right(ToolbarButton("reset", "Reset", testToolbarIntent{}).WithDisabledReason("Select a target first")).
		Build()
	if bar == nil {
		t.Fatal("NewToolbarBuilder().Build() returned nil")
	}
	if bar.Tag() != "toolbar" {
		t.Fatalf("toolbar tag = %q, want toolbar", bar.Tag())
	}
	right := bar.Props()["rightItems"].([]ToolbarItem)
	if len(right) != 2 || !right[1].Disabled || right[1].DisabledReason != "Select a target first" {
		t.Fatalf("toolbar disabled reason item = %#v", right)
	}
}

func TestToolbarDirectShortcut(t *testing.T) {
	bar := Toolbar(
		[]ToolbarItem{ToolbarText("scope", "default")},
		[]ToolbarItem{ToolbarBadge("state", "healthy")},
		[]ToolbarItem{ToolbarButton("refresh", "Refresh", testToolbarIntent{})},
	)
	if bar == nil {
		t.Fatal("Toolbar() returned nil")
	}
	if bar.Tag() != "toolbar" {
		t.Fatalf("Toolbar().Tag() = %q, want toolbar", bar.Tag())
	}
}

func TestToolbarOperationalShortcutPresets(t *testing.T) {
	kv := ToolbarKeyValue("endpoint", "endpoint", "http://localhost:8080")
	if kv.Label != "endpoint: http://localhost:8080" {
		t.Fatalf("key value label = %q", kv.Label)
	}
	muted := ToolbarMutedKeyValue("selection", "selection", "-")
	if muted.FgColor != "bright-black" {
		t.Fatalf("muted fg = %q", muted.FgColor)
	}
	state := ToolbarStateBadge("state", "healthy")
	if state.BgColor != "green" {
		t.Fatalf("state bg = %q", state.BgColor)
	}
	busy := ToolbarBusyBadge("busy", "")
	if busy.Label != "busy" {
		t.Fatalf("busy label = %q", busy.Label)
	}
	err := ToolbarErrorBadge("error", "failed")
	if err.BgColor != "red" {
		t.Fatalf("error bg = %q", err.BgColor)
	}
	for _, item := range []ToolbarItem{
		ToolbarEndpoint("local"),
		ToolbarScope("group: default"),
		ToolbarSelection("provider/openai"),
	} {
		if item.Label == "" {
			t.Fatalf("operational shortcut returned empty item: %+v", item)
		}
	}

	if got := ToolbarJoinDisabledReasons("Reason required.", "", "Select target."); got != "Reason required. Select target." {
		t.Fatalf("joined disabled reasons = %q", got)
	}
}

func TestToolbarShellShortcutPresets(t *testing.T) {
	header := ToolbarShellHeader(
		"manager.header",
		"ai-gateway-manager",
		"http://127.0.0.1:8080",
		"admin: ops",
		ToolbarButton("refresh", "Refresh", testToolbarIntent{}).Primary(),
		ToolbarButton("logout", "Logout", testToolbarIntent{}).WithDisabledReason("No local session."),
	)
	if header == nil || header.Tag() != "toolbar" {
		t.Fatalf("shell header = %v, want toolbar", header)
	}
	headerProps := header.Props()
	if headerProps["title"] != "ai-gateway-manager" || headerProps["titleWidth"] != 22 {
		t.Fatalf("header title props = %v/%v", headerProps["title"], headerProps["titleWidth"])
	}
	headerLeft := headerProps["leftItems"].([]ToolbarItem)
	if len(headerLeft) != 1 || headerLeft[0].Key != "base-url" || headerLeft[0].FgColor != "bright-black" {
		t.Fatalf("header left items = %+v", headerLeft)
	}
	headerCenter := headerProps["centerItems"].([]ToolbarItem)
	if len(headerCenter) != 1 || headerCenter[0].Key != "auth" || headerCenter[0].BgColor != "cyan" {
		t.Fatalf("header center items = %+v", headerCenter)
	}

	nav := ToolbarShellNav("manager.nav", []ToolbarItem{
		ToolbarButton("nav-overview", "F1 Overview", testToolbarIntent{}).Primary(),
		ToolbarButton("nav-actions", "F8 Actions", testToolbarIntent{}).WithDisabledReason("Login is required."),
	})
	if nav == nil || nav.Tag() != "toolbar" {
		t.Fatalf("shell nav = %v, want toolbar", nav)
	}
	navProps := nav.Props()
	if nav.Key() != "manager.nav" || navProps["dense"] != true {
		t.Fatalf("nav key/dense = %q/%v", nav.Key(), navProps["dense"])
	}
	navLeft := navProps["leftItems"].([]ToolbarItem)
	if len(navLeft) != 2 || !navLeft[1].Disabled || navLeft[1].DisabledReason != "Login is required." {
		t.Fatalf("nav left items = %+v", navLeft)
	}
}

func TestToolbarSyncShortcutPresets(t *testing.T) {
	syncAt := time.Date(2026, 5, 26, 9, 10, 11, 0, time.UTC)
	lastSync := ToolbarLastSync(syncAt)
	if lastSync.Key != "last-sync" || lastSync.Label != "last sync: 09:10:11" || lastSync.Width != 22 || lastSync.FgColor != "bright-black" {
		t.Fatalf("last sync item = %+v", lastSync)
	}
	never := ToolbarLastSync(time.Time{})
	if never.Label != "last sync: -" {
		t.Fatalf("empty last sync label = %q, want last sync: -", never.Label)
	}

	age := ToolbarSyncAge("page.runtime.sync-age", "age", syncAt, time.Time{}, 13)
	if age.Key != "page.runtime.sync-age" || age.Kind != ToolbarItemCustom {
		t.Fatalf("sync age item = %+v, want custom keyed item", age)
	}
	node, ok := age.Custom.(VNode)
	if !ok || node.Tag() != "timer" || node.Key() != "page.runtime.sync-age" {
		t.Fatalf("sync age custom = %T %v, want timer vnode with matching key", age.Custom, age.Custom)
	}
	if got := node.Props()["label"]; got != "age" {
		t.Fatalf("sync age label = %v, want age", got)
	}
	if got := node.Props()["width"]; got != 13 {
		t.Fatalf("sync age width = %v, want 13", got)
	}

	emptyAge := ToolbarSyncAge("", "", time.Time{}, time.Time{}, 0)
	if emptyAge.Key != "sync-age" || emptyAge.Label != "age: -" || emptyAge.Width != 13 || emptyAge.FgColor != "bright-black" {
		t.Fatalf("empty sync age = %+v", emptyAge)
	}
}

func TestToolbarPageHeaderShortcutPreset(t *testing.T) {
	syncAt := time.Date(2026, 5, 26, 9, 10, 11, 0, time.UTC)
	bar := ToolbarPageHeader("page.runtime.toolbar", "Runtime", "runtime status", syncAt, time.Time{})
	if bar == nil || bar.Tag() != "toolbar" {
		t.Fatalf("page toolbar = %v, want toolbar", bar)
	}
	if bar.Key() != "page.runtime.toolbar" {
		t.Fatalf("page toolbar key = %q, want page.runtime.toolbar", bar.Key())
	}
	props := bar.Props()
	if got := props["title"]; got != "Runtime" {
		t.Fatalf("title = %v, want Runtime", got)
	}
	if got := props["titleWidth"]; got != 18 {
		t.Fatalf("title width = %v, want 18", got)
	}
	left := props["leftItems"].([]ToolbarItem)
	if len(left) != 1 || left[0].Key != "subtitle" || left[0].Label != "runtime status" || left[0].Width != 46 || left[0].FgColor != "bright-black" {
		t.Fatalf("left items = %+v, want muted subtitle", left)
	}
	right := props["rightItems"].([]ToolbarItem)
	if len(right) != 2 {
		t.Fatalf("right items = %+v, want last sync and sync age", right)
	}
	if right[0].Key != "last-sync" || right[0].Label != "last sync: 09:10:11" {
		t.Fatalf("last sync item = %+v", right[0])
	}
	if right[1].Key != "page.runtime.sync-age" || right[1].Kind != ToolbarItemCustom {
		t.Fatalf("sync age item = %+v, want derived page sync-age key", right[1])
	}
	node, ok := right[1].Custom.(VNode)
	if !ok || node.Tag() != "timer" || node.Key() != "page.runtime.sync-age" {
		t.Fatalf("sync age custom = %T %v, want timer with page.runtime.sync-age key", right[1].Custom, right[1].Custom)
	}

	fallback := ToolbarPageHeader("", "", "", time.Time{}, time.Time{})
	if fallback.Key() != "page.toolbar" {
		t.Fatalf("fallback key = %q, want page.toolbar", fallback.Key())
	}
	fallbackProps := fallback.Props()
	fallbackLeft := fallbackProps["leftItems"].([]ToolbarItem)
	if len(fallbackLeft) != 1 || fallbackLeft[0].Label != "-" {
		t.Fatalf("fallback left items = %+v, want subtitle fallback", fallbackLeft)
	}
	fallbackRight := fallbackProps["rightItems"].([]ToolbarItem)
	if len(fallbackRight) != 2 || fallbackRight[1].Key != "page.sync-age" || fallbackRight[1].Label != "age: -" {
		t.Fatalf("fallback right items = %+v, want page.sync-age fallback", fallbackRight)
	}
}

func TestToolbarPageSummaryShortcutPreset(t *testing.T) {
	item := ToolbarPageSummary(2, 5, 101, 25, "search", "openai", 78)
	if item.Key != "page" || item.Label != "page 2/5 total 101 size 25 search openai" || item.Width != 78 {
		t.Fatalf("page summary item = %+v", item)
	}

	fallback := ToolbarPageSummary(0, 0, -1, -1, "provider", "", 0)
	if fallback.Label != "page 1/1 total 0 size 0 provider -" || fallback.Width != 72 {
		t.Fatalf("fallback page summary = %+v", fallback)
	}

	plain := ToolbarPageSummary(1, 3, 50, 25, "", "ignored", 36)
	if plain.Label != "page 1/3 total 50 size 25" || plain.Width != 36 {
		t.Fatalf("plain page summary = %+v", plain)
	}

	scoped := ToolbarPageSummaryWithParts(
		2,
		5,
		101,
		25,
		96,
		ToolbarPageSummaryPart{Label: "provider", Value: "openai"},
		ToolbarPageSummaryPart{Label: "search", Value: "azure"},
	)
	if scoped.Label != "page 2/5 total 101 size 25 provider openai search azure" || scoped.Width != 96 {
		t.Fatalf("scoped page summary = %+v", scoped)
	}

	compact := ToolbarPageSummaryWithParts(
		1,
		1,
		2,
		25,
		72,
		ToolbarCompactPageSummaryPart("search", "very-long-provider-or-trace-lookup-value", 14),
	)
	if compact.Label != "page 1/1 total 2 size 25 search very-long-p..." {
		t.Fatalf("compact page summary = %+v", compact)
	}

	optional := ToolbarPageSummaryWithParts(
		1,
		1,
		2,
		25,
		72,
		ToolbarPageSummaryPartIfValue("provider", ""),
		ToolbarPageSummaryPartIfValue("status", "active"),
		ToolbarCompactPageSummaryPartIfValue("search", "very-long-provider-or-trace-lookup-value", 14),
	)
	if optional.Label != "page 1/1 total 2 size 25 status active search very-long-p..." {
		t.Fatalf("optional page summary = %+v", optional)
	}

	changed := ToolbarPageSummaryWithParts(
		1,
		1,
		2,
		25,
		72,
		ToolbarPageSummaryPartUnless("status", "all", "all"),
		ToolbarPageSummaryPartUnless("last", "failed", "all"),
		ToolbarCompactPageSummaryPartUnless("search", "very-long-provider-or-trace-lookup-value", "", 14),
	)
	if changed.Label != "page 1/1 total 2 size 25 last failed search very-long-p..." {
		t.Fatalf("changed page summary = %+v", changed)
	}
}

func TestToolbarActionTargetSummaryShortcutPreset(t *testing.T) {
	summary := ToolbarActionTargetSummary(
		ToolbarCompactActionTargetPart("group", "default", 24),
		ToolbarCompactActionTargetPart("provider", "openai", 24),
		ToolbarCompactActionTargetPart("key", "", 24),
	)
	if summary != "group default · provider openai · key -" {
		t.Fatalf("target summary = %q, want action target summary", summary)
	}

	scoped := ToolbarActionTargetSummaryWithScope(
		"key",
		ToolbarCompactActionTargetPart("group", "default", 24),
		ToolbarCompactActionTargetPart("provider", "openai", 24),
		ToolbarCompactActionTargetPart("key", "key-op...ai-1", 24),
	)
	if scoped != "scope key · group default · provider openai · key key-op...ai-1" {
		t.Fatalf("scoped target summary = %q, want scope-prefixed target summary", scoped)
	}

	visibleScopes := ToolbarActionTargetSummaryWithScopes(
		[]string{"global", "group", "provider", "key"},
		ToolbarCompactActionTargetPart("endpoint", "http://127.0.0.1:8080", 40),
		ToolbarCompactActionTargetPart("group", "default", 24),
		ToolbarCompactActionTargetPart("provider", "openai", 24),
		ToolbarCompactActionTargetPart("key", "key-op...ai-1", 24),
	)
	if visibleScopes != "scopes global/group/provider/key · endpoint http://127.0.0.1:8080 · group default · provider openai · key key-op...ai-1" {
		t.Fatalf("multi-scope target summary = %q, want visible scopes summary", visibleScopes)
	}
}

func TestToolbarPaginationShortcutPresets(t *testing.T) {
	prev := ToolbarPaginationPrev(testToolbarIntent{}, false, 1, "Logs are loading.")
	if !prev.Disabled || prev.DisabledReason != "Already at the first page." {
		t.Fatalf("prev disabled = %v reason = %q", prev.Disabled, prev.DisabledReason)
	}
	next := ToolbarPaginationNext(testToolbarIntent{}, true, 2, 3, "Logs are loading.")
	if !next.Disabled || next.DisabledReason != "Logs are loading." {
		t.Fatalf("next disabled = %v reason = %q", next.Disabled, next.DisabledReason)
	}

	controls := ToolbarPaginationControls(
		"logs.pagination",
		126,
		ToolbarText("page", "page 1/3 total 51 size 25").WithWidth(72),
		testToolbarIntent{},
		testToolbarIntent{},
		false,
		1,
		3,
		"Logs are loading.",
	)
	if controls == nil || controls.Tag() != "toolbar" {
		t.Fatalf("pagination controls = %v", controls)
	}
	props := controls.Props()
	if got := props["width"]; got != 126 {
		t.Fatalf("pagination controls width = %v, want 126", got)
	}
	left := props["leftItems"].([]ToolbarItem)
	if len(left) != 1 || left[0].Key != "page" || left[0].Width != 72 {
		t.Fatalf("pagination controls left items = %+v", left)
	}
	right := props["rightItems"].([]ToolbarItem)
	if len(right) != 2 || right[0].Key != "prev" || right[1].Key != "next" {
		t.Fatalf("pagination controls right items = %+v", right)
	}
	if !right[0].Disabled || right[0].DisabledReason != "Already at the first page." {
		t.Fatalf("pagination controls prev disabled = %v reason = %q", right[0].Disabled, right[0].DisabledReason)
	}

	config := ToolbarPaginationConfig{
		Key:           "jobs.pagination",
		Width:         118,
		Summary:       ToolbarPageSummary(2, 5, 101, 25, "", "", 72),
		PrevIntent:    testToolbarIntent{},
		NextIntent:    testToolbarIntent{},
		Busy:          false,
		Page:          2,
		TotalPages:    5,
		LoadingReason: "Background jobs are loading.",
	}
	configControls := config.Controls()
	if configControls == nil || configControls.Tag() != "toolbar" {
		t.Fatalf("pagination config controls = %v", configControls)
	}
	if got := configControls.Props()["width"]; got != 118 {
		t.Fatalf("pagination config controls width = %v, want 118", got)
	}
	configRight := configControls.Props()["rightItems"].([]ToolbarItem)
	if len(configRight) != 2 || configRight[0].Disabled || configRight[1].Disabled {
		t.Fatalf("pagination config right items = %+v", configRight)
	}

	scope := ToolbarPaginationScopeOf(0, 0, 0, 0, 12)
	if scope.Page != 1 || scope.PageSize != 25 || scope.Total != 12 || scope.TotalPages != 1 {
		t.Fatalf("scope = %+v, want normalized defaults", scope)
	}
	scoped := ToolbarPaginationControlsWithScope(
		"logs.pagination",
		126,
		scope,
		testToolbarIntent{},
		testToolbarIntent{},
		false,
		"Logs are loading.",
		84,
		ToolbarPageSummaryPart{Label: "search", Value: "trace-1"},
	)
	scopedLeft := scoped.Props()["leftItems"].([]ToolbarItem)
	if len(scopedLeft) != 1 || scopedLeft[0].Label != "page 1/1 total 12 size 25 search trace-1" {
		t.Fatalf("scoped pagination left = %+v", scopedLeft)
	}
}

func TestToolbarSelectionShortcutPresets(t *testing.T) {
	prev := ToolbarSelectionPrev(testToolbarIntent{}, false, 0, 2, "log", "Logs are loading.")
	if !prev.Disabled || prev.DisabledReason != "Already at the first log." {
		t.Fatalf("prev disabled = %v reason = %q", prev.Disabled, prev.DisabledReason)
	}
	next := ToolbarSelectionNext(testToolbarIntent{}, true, 0, 2, "log", "Logs are loading.")
	if !next.Disabled || next.DisabledReason != "Logs are loading." {
		t.Fatalf("next disabled = %v reason = %q", next.Disabled, next.DisabledReason)
	}

	controls := ToolbarSelectionControls("logs.selection", "Log", testToolbarIntent{}, testToolbarIntent{}, false, 0, 2, "log", "Logs are loading.")
	if controls == nil || controls.Tag() != "toolbar" {
		t.Fatalf("selection controls = %v", controls)
	}
	items := controls.Props()["leftItems"].([]ToolbarItem)
	if len(items) != 2 || items[0].Key != "select-prev" || items[1].Key != "select-next" {
		t.Fatalf("selection controls items = %+v", items)
	}

	actions := ToolbarSelectionActionControls(
		"logs.selection.actions",
		"Log",
		56,
		testToolbarIntent{},
		testToolbarIntent{},
		false,
		0,
		2,
		"log",
		"Logs are loading.",
		ToolbarButton("open-trace", "Open Trace", testToolbarIntent{}).Primary(),
	)
	if actions == nil || actions.Tag() != "toolbar" {
		t.Fatalf("selection action controls = %v", actions)
	}
	actionItems := actions.Props()["leftItems"].([]ToolbarItem)
	if len(actionItems) != 3 || actionItems[0].Key != "select-prev" || actionItems[1].Key != "select-next" || actionItems[2].Key != "open-trace" {
		t.Fatalf("selection action controls items = %+v", actionItems)
	}
	if got := actions.Props()["width"]; got != 56 {
		t.Fatalf("selection action controls width = %v, want 56", got)
	}

	config := ToolbarSelectionConfig{
		Key:           "jobs.selection.actions",
		Title:         "Job",
		Width:         62,
		PrevIntent:    testToolbarIntent{},
		NextIntent:    testToolbarIntent{},
		Busy:          false,
		Index:         1,
		Total:         3,
		ItemLabel:     "job",
		LoadingReason: "Background jobs are loading.",
	}
	configControls := config.Controls()
	if configControls == nil || configControls.Tag() != "toolbar" {
		t.Fatalf("selection config controls = %v", configControls)
	}
	if got := configControls.Props()["width"]; got != 62 {
		t.Fatalf("selection config controls width = %v, want 62", got)
	}
	configItems := config.ActionControls(ToolbarButton("open", "Open", testToolbarIntent{}).Primary()).Props()["leftItems"].([]ToolbarItem)
	if len(configItems) != 3 || configItems[0].Key != "select-prev" || configItems[1].Key != "select-next" || configItems[2].Key != "open" {
		t.Fatalf("selection config action items = %+v", configItems)
	}
}

func TestToolbarActionControlsShortcutPreset(t *testing.T) {
	controls := ToolbarActionControls(
		"login.captcha.actions",
		82,
		ToolbarButton("captcha-refresh", "Refresh Captcha", testToolbarIntent{}),
		ToolbarButton("captcha-open", "Open Captcha Image", testToolbarIntent{}).WithDisabledReason("No captcha image file is available."),
	)
	if controls == nil || controls.Tag() != "toolbar" {
		t.Fatalf("action controls = %v", controls)
	}
	props := controls.Props()
	if got := props["width"]; got != 82 {
		t.Fatalf("width = %v, want 82", got)
	}
	items := props["leftItems"].([]ToolbarItem)
	if len(items) != 2 || items[0].Key != "captcha-refresh" || items[1].Key != "captcha-open" {
		t.Fatalf("items = %+v", items)
	}
	if !items[1].Disabled || items[1].DisabledReason != "No captcha image file is available." {
		t.Fatalf("disabled item = %+v", items[1])
	}
}

func TestToolbarActionGroupShortcutPreset(t *testing.T) {
	group := ToolbarActionGroup("actions.global", "Global", []ToolbarItem{
		ToolbarButton("reload", "Reload", testToolbarIntent{}).Primary(),
		ToolbarButton("reset", "Reset", testToolbarIntent{}).WithDisabledReason("Reason required."),
	})
	if group == nil || group.Tag() != "toolbar" {
		t.Fatalf("action group = %v", group)
	}
	items := group.Props()["leftItems"].([]ToolbarItem)
	if len(items) != 2 || items[0].Key != "reload" || items[1].Key != "reset" {
		t.Fatalf("action group items = %+v", items)
	}
	if !items[1].Disabled || items[1].DisabledReason != "Reason required." {
		t.Fatalf("disabled action = %+v", items[1])
	}

	custom := ToolbarActionGroupWithLayout("actions.key", "Key", 126, 10, []ToolbarItem{
		ToolbarButton("disable", "Disable", testToolbarIntent{}),
	})
	if custom.Props()["width"] != 126 || custom.Props()["titleWidth"] != 10 {
		t.Fatalf("custom action group props = %+v, want explicit layout", custom.Props())
	}
}

func TestToolbarActionGroupsShortcutPreset(t *testing.T) {
	groups := ToolbarActionGroups("actions.groups", []ToolbarActionGroupConfig{
		{
			Key:   "actions.global",
			Title: "Global",
			Items: []ToolbarItem{
				ToolbarButton("reload", "Reload", testToolbarIntent{}).Primary(),
			},
		},
		{
			Key:        "actions.target",
			Title:      "Selected Target",
			Width:      126,
			TitleWidth: 18,
			Items: []ToolbarItem{
				ToolbarButton("disable", "Disable", testToolbarIntent{}).WithDisabledReason("Reason required."),
			},
		},
	}, "selected key: provider/default/key-1", "No actions matched.")
	if groups == nil || groups.Tag() != "vstack" {
		t.Fatalf("action groups = %v", groups)
	}
	children := groups.Children()
	if len(children) != 3 {
		t.Fatalf("action group children = %d, want 3", len(children))
	}
	if got := children[0].Props()["title"]; got != "Global" {
		t.Fatalf("first title = %v, want Global", got)
	}
	if got := children[1].Props()["title"]; got != "Selected Target" {
		t.Fatalf("second title = %v, want Selected Target", got)
	}
	if got := children[1].Props()["width"]; got != 126 {
		t.Fatalf("second width = %v, want configured width 126", got)
	}
	if got := children[1].Props()["titleWidth"]; got != 18 {
		t.Fatalf("second title width = %v, want configured title width 18", got)
	}
	if got := children[2].Props()["content"]; got != "selected key: provider/default/key-1" {
		t.Fatalf("summary = %v", got)
	}
}
