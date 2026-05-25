package ui

import (
	"testing"

	"github.com/wwsheng009/mint/framework"
)

type menuShortcutTestIntent struct{}

func (menuShortcutTestIntent) IntentType() string { return "menu-shortcut-test" }

func TestMenuShortcuts(t *testing.T) {
	items := MenuItems(
		MenuAction("open", "Open", menuShortcutTestIntent{}).WithShortcut("ctrl+o"),
		MenuSeparator(),
		MenuSubmenu("file", "File", MenuAction("save", "Save", menuShortcutTestIntent{}).WithShortcut("ctrl+s")),
	)
	builder := NewMenuPopupBuilder(items).Theme(MenuThemeDefault())
	vnode := builder.Build()
	// Build() wraps popup in a portal "box" element; check the first child for the popup tag.
	tag := vnode.Tag()
	if tag == "box" {
		if children := vnode.Children(); len(children) > 0 {
			tag = children[0].Tag()
		}
	}
	if tag != "menu-popup" {
		t.Fatalf("Tag() = %q, want menu-popup", tag)
	}
	bindings := builder.Shortcuts()
	if len(bindings) != 2 {
		t.Fatalf("Shortcuts() len = %d, want 2", len(bindings))
	}
	if MenuPortalFixed != 2 {
		t.Fatalf("MenuPortalFixed = %d, want 2", MenuPortalFixed)
	}
}

func TestMenuOperationalShortcutPresets(t *testing.T) {
	items := MenuItems()
	items = MenuAppendGroup(items, "Operations",
		MenuRefreshAction(menuShortcutTestIntent{}),
		MenuReloadRuntimeAction(menuShortcutTestIntent{}),
		MenuResetRuntimeAction(menuShortcutTestIntent{}),
		MenuClearCircuitBreakersAction(menuShortcutTestIntent{}),
		MenuDangerAction("disable-key", "Disable Key", "Disable selected provider key", menuShortcutTestIntent{}),
		MenuDisabledAction("reset-key", "Reset Key", "Select a key first", menuShortcutTestIntent{}),
	)

	if len(items) != 7 {
		t.Fatalf("items len = %d, want 7", len(items))
	}
	if !items[0].IsLabel() {
		t.Fatalf("first item should be group label: %+v", items[0])
	}
	if !items[2].Danger {
		t.Fatalf("reload action should be dangerous: %+v", items[2])
	}
	if !items[6].Disabled || items[6].Metadata["disabledReason"] != "Select a key first" {
		t.Fatalf("disabled action = %+v", items[6])
	}

	if group := MenuGroup("", MenuActionWithDescription("inspect", "Inspect", "Open details", menuShortcutTestIntent{})); len(group) != 1 || group[0].Description == "" {
		t.Fatalf("empty-label group = %+v", group)
	}
}

func TestBindMenuGlobalShortcuts(t *testing.T) {
	app := framework.NewApp()
	builder := NewMenuPopupBuilder(MenuItems(
		MenuAction("open", "Open", menuShortcutTestIntent{}).WithShortcut("ctrl+o"),
	)).RegisterShortcuts(true)

	if count := BindMenuGlobalShortcuts(app, builder); count != 1 {
		t.Fatalf("BindMenuGlobalShortcuts() = %d, want 1", count)
	}
}

func TestInstallMenu(t *testing.T) {
	app := framework.NewApp()
	builder := NewMenuPopupBuilder(MenuItems(
		MenuAction("open", "Open", menuShortcutTestIntent{}).WithShortcut("ctrl+o"),
	)).RegisterShortcuts(true)

	if count := InstallMenu(app, builder); count != 1 {
		t.Fatalf("InstallMenu() = %d, want 1", count)
	}
}
