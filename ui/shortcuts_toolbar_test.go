package ui

import "testing"

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
}
