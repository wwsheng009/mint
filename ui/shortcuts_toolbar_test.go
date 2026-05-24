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
