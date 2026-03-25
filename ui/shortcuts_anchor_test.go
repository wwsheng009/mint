package ui

import "testing"

func TestNewAnchorBuilder(t *testing.T) {
	vnode := NewAnchorBuilder().
		Items([]AnchorItem{
			NewAnchorItem("intro", "Introduction"),
		}).
		Build()
	if vnode == nil {
		t.Fatal("NewAnchorBuilder().Build() returned nil")
	}
	if vnode.Tag() != "anchor" {
		t.Fatalf("Tag = %q, want anchor", vnode.Tag())
	}
}

func TestAnchorNavShortcut(t *testing.T) {
	vnode := AnchorNav([]AnchorItem{
		NewAnchorItem("intro", "Introduction"),
	})
	if vnode == nil {
		t.Fatal("AnchorNav() returned nil")
	}
	if vnode.Tag() != "anchor" {
		t.Fatalf("Tag = %q, want anchor", vnode.Tag())
	}
}
