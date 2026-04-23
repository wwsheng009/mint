package ui

import (
	"testing"

	"github.com/wwsheng009/mint/ui/components/text"
)

func TestNewLayoutBuilder(t *testing.T) {
	vnode := NewLayoutBuilder().
		Content(text.New("Body")).
		Build()
	if vnode == nil {
		t.Fatal("NewLayoutBuilder().Build() returned nil")
	}
	if vnode.Tag() != "layout" {
		t.Fatalf("Tag = %q, want layout", vnode.Tag())
	}
}

func TestLayoutShortcut(t *testing.T) {
	vnode := Layout(text.New("Body"))
	if vnode == nil {
		t.Fatal("Layout() returned nil")
	}
	if vnode.Tag() != "layout" {
		t.Fatalf("Tag = %q, want layout", vnode.Tag())
	}
}
