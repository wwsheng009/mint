package ui

import (
	"testing"

	"github.com/wwsheng009/mint/ui/components/text"
)

func TestNewSpaceBuilder(t *testing.T) {
	vnode := NewSpaceBuilder().
		Vertical().
		Children(text.New("A"), text.New("B")).
		Build()
	if vnode == nil {
		t.Fatal("NewSpaceBuilder().Build() returned nil")
	}
	if vnode.Tag() != "space" {
		t.Fatalf("Tag = %q, want space", vnode.Tag())
	}
}

func TestSpaceShortcut(t *testing.T) {
	vnode := Space(text.New("A"), text.New("B"))
	if vnode == nil {
		t.Fatal("Space() returned nil")
	}
	if vnode.Tag() != "space" {
		t.Fatalf("Tag = %q, want space", vnode.Tag())
	}
}
