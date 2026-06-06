package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui/components/tabs"
)

func TestTabsShortcutHelpers(t *testing.T) {
	item := NewTabItem("home", "Home").WithIcon("H").WithBadge("3").WithHotkey('h')
	if item.ID != "home" || item.Label != "Home" {
		t.Fatal("NewTabItem should set ID and label")
	}
	if item.Icon != "H" || item.Badge != "3" || item.Hotkey != 'h' {
		t.Fatal("tab item helper chaining should preserve metadata")
	}

	next := TabsNext("workspace")
	if next.ComponentID != "workspace" {
		t.Fatalf("TabsNext should preserve component ID, got %q", next.ComponentID)
	}

	prev := TabsPrevious("workspace")
	if prev.ComponentID != "workspace" {
		t.Fatalf("TabsPrevious should preserve component ID, got %q", prev.ComponentID)
	}

	selectByID := TabsSelect("workspace", "settings")
	if selectByID.TabID != "settings" || selectByID.TabIndex != -1 {
		t.Fatalf("TabsSelect should target tab ID, got %+v", selectByID)
	}

	selectByIndex := TabsSelectIndex("workspace", 2)
	if selectByIndex.TabIndex != 2 || selectByIndex.TabID != "" {
		t.Fatalf("TabsSelectIndex should target tab index, got %+v", selectByIndex)
	}

	change := TabsChange("workspace", 1, "settings", "Settings")
	if change.ComponentID != "workspace" || change.ActiveTab != 1 || change.TabID != "settings" {
		t.Fatalf("TabsChange should preserve payload, got %+v", change)
	}
}

func TestWorkspaceTabsPresetUsesDistinctCardStyleAndFieldBinding(t *testing.T) {
	node := WorkspaceTabs("ops.tabs", "ops.tabs", "opsWorkspace", "requests", []TabItem{
		NewTabItem("requests", "Requests"),
		NewTabItem("detail", "Detail"),
	}, 126)

	props := node.Props()
	if node.Tag() != "tabs" {
		t.Fatalf("WorkspaceTabs tag = %q, want tabs", node.Tag())
	}
	if got := props["key"]; got != "ops.tabs" {
		t.Fatalf("key = %v, want ops.tabs", got)
	}
	if got := props["componentID"]; got != "ops.tabs" {
		t.Fatalf("componentID = %v, want ops.tabs", got)
	}
	if got := props["activeTabID"]; got != "requests" {
		t.Fatalf("activeTabID = %v, want requests", got)
	}
	if got := props["tabVariant"]; got != tabs.TabVariantCard {
		t.Fatalf("tabVariant = %v, want card", got)
	}
	if got := props["wrapTabs"]; got != true {
		t.Fatalf("wrapTabs = %v, want true", got)
	}
	if got := props["activeTabMarker"]; got != "" {
		t.Fatalf("activeTabMarker = %v, want empty", got)
	}
	if got := props["divider"]; got != " " {
		t.Fatalf("divider = %v, want single space", got)
	}
	if got := props["width"]; got != 126 {
		t.Fatalf("width = %v, want 126", got)
	}
	if _, ok := props["changeIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("changeIntentField = %T, want FieldIntent", props["changeIntentField"])
	}
	inactive, ok := props["tabStyle"].(style.Style)
	if !ok {
		t.Fatalf("tabStyle = %T, want style.Style", props["tabStyle"])
	}
	if inactive.FG == style.NoColor {
		t.Fatalf("tabStyle = %+v, want muted foreground", inactive)
	}
	active, ok := props["activeTabStyle"].(style.Style)
	if !ok {
		t.Fatalf("activeTabStyle = %T, want style.Style", props["activeTabStyle"])
	}
	if !active.IsReverse() || !active.IsBold() || active.IsUnderline() {
		t.Fatalf("activeTabStyle = %+v, want reverse+bold without underline", active)
	}
}
