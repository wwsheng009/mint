package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
)

type testStatusBarIntent struct{}

func (testStatusBarIntent) IntentType() string { return "test.statusbar" }

func TestStatusBarShortcuts(t *testing.T) {
	bar := NewStatusBarBuilder().
		Theme(StatusBarThemeDefault()).
		Left(StatusBarBadge(" L ", "black", "yellow")).
		Center(StatusBarText(" Center ")).
		Right(StatusBarText("R").WithWidth(3)).
		Build()

	if bar == nil {
		t.Fatal("Build() returned nil")
	}
	if len(bar.Children()) != 3 {
		t.Fatalf("children len = %d, want 3", len(bar.Children()))
	}
}

func TestStatusBarDirectShortcut(t *testing.T) {
	bar := StatusBar(
		[]StatusBarSection{StatusBarBadge(" L ", "black", "yellow")},
		[]StatusBarSection{StatusBarText(" Center ")},
		[]StatusBarSection{StatusBarText("R")},
	)

	if bar == nil {
		t.Fatal("StatusBar() returned nil")
	}
	assertStatusBarShortcutGap(t, bar, 1)
}

func TestStatusBarOverflowAlias(t *testing.T) {
	bar := StatusBarWithTheme(
		StatusBarThemeMuted(),
		StatusBarSections(
			StatusBarText(" Native Select On ").WithWidth(10).WithOverflow(StatusBarOverflowEllipsis),
		),
		nil,
		StatusBarSections(
			StatusBarText("1234567890").WithWidth(4).WithOverflow(StatusBarOverflowClip),
		),
	)

	if bar == nil {
		t.Fatal("StatusBarWithTheme() returned nil")
	}
	assertStatusBarShortcutGap(t, bar, 1)
}

func TestStatusBarActionShortcuts(t *testing.T) {
	pressIntent := testStatusBarIntent{}
	bar := StatusBarWithTheme(
		StatusBarThemeDefault(),
		StatusBarSections(
			StatusBarActionBadge(" GO ", "black", "green", pressIntent),
		),
		nil,
		StatusBarSections(
			StatusBarActionText("Details", intent.Focus("statusbar.demo")),
		),
	)

	if bar == nil {
		t.Fatal("StatusBarWithTheme() returned nil")
	}
	assertStatusBarShortcutGap(t, bar, 1)
}

func TestStatusBarWithHelpShortcut(t *testing.T) {
	bar := StatusBarWithHelp(
		StatusBarThemeDefault(),
		"Ready",
		StatusBarSections(
			StatusBarActionBadge(" GO ", "black", "green", testStatusBarIntent{}).WithHelp("Run the current command"),
		),
		nil,
		StatusBarSections(
			StatusBarText("Hint").WithTooltip("Extra hint text"),
		),
	)

	if bar == nil {
		t.Fatal("StatusBarWithHelp() returned nil")
	}
	if len(bar.Children()) != 2 {
		t.Fatalf("children len = %d, want 2", len(bar.Children()))
	}
	assertStatusBarShortcutGap(t, bar.Children()[0], 1)
}

func TestStatusBarWithHelpModeShortcut(t *testing.T) {
	bar := StatusBarWithHelpMode(
		StatusBarThemeDefault(),
		"Ready",
		StatusBarHelpOverlay,
		StatusBarSections(
			StatusBarActionBadge(" GO ", "black", "green", testStatusBarIntent{}).WithHelp("Run the current command"),
		),
		nil,
		nil,
	)

	if bar == nil {
		t.Fatal("StatusBarWithHelpMode() returned nil")
	}
	if len(bar.Children()) != 2 {
		t.Fatalf("children len = %d, want 2", len(bar.Children()))
	}
	assertStatusBarShortcutGap(t, bar.Children()[0], 1)
}

func TestStatusBarTooltipArrowAliases(t *testing.T) {
	if StatusBarTooltipArrowDefault != 0 {
		t.Fatalf("default arrow alias = %v, want 0", StatusBarTooltipArrowDefault)
	}
	if StatusBarTooltipArrowSharp == StatusBarTooltipArrowRounded {
		t.Fatal("sharp and rounded arrow aliases should differ")
	}
	theme := StatusBarThemeMuted().WithTooltipArrowStyle(StatusBarTooltipArrowRounded)
	if theme.TooltipArrowStyle != StatusBarTooltipArrowRounded {
		t.Fatalf("theme arrow style = %v, want %v", theme.TooltipArrowStyle, StatusBarTooltipArrowRounded)
	}
}

func TestStatusBarOperationalShortcutPresets(t *testing.T) {
	if StatusBarDefaultTone("healthy") != StatusBarToneNormal {
		t.Fatalf("healthy tone = %q", StatusBarDefaultTone("healthy"))
	}
	if StatusBarDefaultTone("pending_restart") != StatusBarToneWarn {
		t.Fatalf("pending tone = %q", StatusBarDefaultTone("pending_restart"))
	}
	if StatusBarDefaultTone("failed") != StatusBarToneError {
		t.Fatalf("failed tone = %q", StatusBarDefaultTone("failed"))
	}
	if StatusBarDefaultTone("syncing") != StatusBarToneInfo {
		t.Fatalf("syncing tone = %q", StatusBarDefaultTone("syncing"))
	}
	if StatusBarDefaultTone("custom") != StatusBarToneNeutral {
		t.Fatalf("custom tone = %q", StatusBarDefaultTone("custom"))
	}

	kv := StatusBarKeyValue("endpoint", "http://localhost:8080")
	if kv.Text != "endpoint: http://localhost:8080" {
		t.Fatalf("key value text = %q", kv.Text)
	}
	muted := StatusBarMutedKeyValue("selection", "-")
	if muted.FgColor != "bright-black" {
		t.Fatalf("muted fg = %q", muted.FgColor)
	}
	state := StatusBarStateBadge("degraded")
	if state.BgColor != "yellow" {
		t.Fatalf("state bg = %q", state.BgColor)
	}
	busy := StatusBarBusyBadge("")
	if busy.Text != " busy " {
		t.Fatalf("busy text = %q", busy.Text)
	}
	err := StatusBarErrorBadge("failed")
	if err.BgColor != "red" {
		t.Fatalf("error bg = %q", err.BgColor)
	}

	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	for _, section := range StatusBarSections(
		StatusBarEndpoint("http://localhost:8080"),
		StatusBarProfile("local"),
		StatusBarUser("admin"),
		StatusBarRole("ops"),
		StatusBarPage("jobs"),
		StatusBarScope("provider"),
		StatusBarTarget("openai/key-1"),
		StatusBarSelection("job-1"),
		StatusBarFilter("failed"),
		StatusBarCount("keys", 12),
		StatusBarLatency(250*time.Millisecond),
		StatusBarUptime(3*time.Hour),
		StatusBarHotkey("r", "reload"),
		StatusBarSeparator(),
		StatusBarLastSync(now.Add(-2*time.Minute), now),
		StatusBarAutoRefresh(30*time.Second),
	) {
		if section.Text == "" {
			t.Fatalf("operational status shortcut returned empty section: %+v", section)
		}
	}

	if got := StatusBarSelectionTarget("job", "Sync", 12); got != "job Sync" {
		t.Fatalf("selection target = %q, want job Sync", got)
	}
}

func TestOperationalStatusBarShortcutPreset(t *testing.T) {
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	bar := OperationalStatusBar(
		StatusBarBusyBadge("BUSY").WithHelp("Request running"),
		"http://localhost:8080",
		"admin: ops",
		"Runtime",
		now.Add(-2*time.Minute),
		now,
		"key openai/default/key-op...ai-1",
		StatusBarText("Up/Down select").WithHelp("Move selection"),
		StatusBarText("HTTP Admin API only").WithHelp("No internal imports"),
	)
	if bar == nil {
		t.Fatal("OperationalStatusBar() returned nil")
	}
	if len(bar.Children()) != 3 {
		t.Fatalf("children len = %d, want 3", len(bar.Children()))
	}
	texts := statusBarShortcutSectionTexts(bar)
	joined := strings.Join(texts, "")
	for _, want := range []string{
		" | endpoint: http://localhost:8080",
		" | user: admin: ops",
		" | page: Runtime",
		" | last sync: 2m ago",
		" | selection: key openai/default/key-op...ai-1",
		" | Up/Down select",
		" | HTTP Admin API only",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("operational status bar text = %q, want segment %q", joined, want)
		}
	}

	fallback := OperationalStatusBar(StatusBarSection{}, "", "", "", time.Time{}, time.Time{}, "")
	if fallback == nil {
		t.Fatal("OperationalStatusBar() fallback returned nil")
	}
	if len(fallback.Children()) != 3 {
		t.Fatalf("fallback children len = %d, want 3", len(fallback.Children()))
	}
	blankEndpoint := OperationalStatusBar(
		StatusBarStateBadge("ready"),
		"",
		"admin: ops",
		"Runtime",
		now,
		now,
		"-",
	)
	blankEndpointText := strings.Join(statusBarShortcutSectionTexts(blankEndpoint), "")
	if strings.Contains(blankEndpointText, "endpoint:") {
		t.Fatalf("blank endpoint status bar text = %q, should omit endpoint section", blankEndpointText)
	}
	for _, want := range []string{"user: admin: ops", "page: Runtime"} {
		if !strings.Contains(blankEndpointText, want) {
			t.Fatalf("blank endpoint status bar text = %q, want %q", blankEndpointText, want)
		}
	}
}

func statusBarShortcutSectionTexts(node VNode) []string {
	if node == nil {
		return nil
	}
	texts := []string{}
	if node.Tag() == "statusbar-section" {
		if text, ok := node.Props()["text"].(string); ok {
			texts = append(texts, text)
		}
	}
	for _, child := range node.Children() {
		texts = append(texts, statusBarShortcutSectionTexts(child)...)
	}
	return texts
}

func assertStatusBarShortcutGap(t *testing.T, node VNode, want int) {
	t.Helper()
	if node == nil {
		t.Fatal("status bar node is nil")
	}
	children := node.Children()
	if len(children) != 3 {
		t.Fatalf("status bar children len = %d, want 3", len(children))
	}
	for index, child := range children {
		gapNode, ok := child.(interface{ Gap() int })
		if !ok {
			t.Fatalf("status bar slot %d does not expose gap", index)
		}
		if got := gapNode.Gap(); got != want {
			t.Fatalf("status bar slot %d gap = %d, want %d", index, got, want)
		}
	}
}
