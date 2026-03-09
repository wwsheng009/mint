package ui

import (
	"testing"
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
	if vnode.Tag() != "menu-popup" {
		t.Fatalf("Tag() = %q, want menu-popup", vnode.Tag())
	}
	bindings := builder.Shortcuts()
	if len(bindings) != 2 {
		t.Fatalf("Shortcuts() len = %d, want 2", len(bindings))
	}
	if MenuPortalFixed != 2 {
		t.Fatalf("MenuPortalFixed = %d, want 2", MenuPortalFixed)
	}
}
