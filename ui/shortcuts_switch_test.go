package ui

import "testing"

func TestSwitchShortcuts(t *testing.T) {
	vnode := NewSwitchBuilder().
		Label("Auto-save").
		Checked(true).
		Build()
	if vnode == nil {
		t.Fatal("NewSwitchBuilder().Build() returned nil")
	}

	short := Switch("Sync", false)
	if short == nil {
		t.Fatal("Switch() returned nil")
	}
}
