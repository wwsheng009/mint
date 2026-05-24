package ui

import (
	"testing"

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
}
