package ui

import "testing"

func TestContainerAndDisplayShortcuts(t *testing.T) {
	if vnode := Text("Hello"); vnode.Tag() != "text" {
		t.Fatalf("Text().Tag() = %q, want text", vnode.Tag())
	}
	if vnode := Divider(); vnode.Tag() != "divider" {
		t.Fatalf("Divider().Tag() = %q, want divider", vnode.Tag())
	}
	if vnode := HDivider(); vnode.Tag() != "divider" {
		t.Fatalf("HDivider().Tag() = %q, want divider", vnode.Tag())
	}
	if vnode := VDivider(); vnode.Tag() != "divider" {
		t.Fatalf("VDivider().Tag() = %q, want divider", vnode.Tag())
	}
	if vnode := Panel(Text("Body")); vnode.Tag() != "panel" {
		t.Fatalf("Panel().Tag() = %q, want panel", vnode.Tag())
	}
	if vnode := ScrollView(Text("Body")); vnode.Tag() != "scrollview" {
		t.Fatalf("ScrollView().Tag() = %q, want scrollview", vnode.Tag())
	}
	if vnode := Empty("No records"); vnode.Tag() != "empty" {
		t.Fatalf("Empty().Tag() = %q, want empty", vnode.Tag())
	}
}

func TestNewEmptyBuilder(t *testing.T) {
	vnode := NewEmptyBuilder().
		Description("Nothing here").
		Image("[ ]").
		Build()
	if vnode == nil {
		t.Fatal("NewEmptyBuilder().Build() returned nil")
	}
	if vnode.Tag() != "empty" {
		t.Fatalf("Tag() = %q, want empty", vnode.Tag())
	}
}
