package ui

import "testing"

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
