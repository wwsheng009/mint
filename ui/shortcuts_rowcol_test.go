package ui

import (
	"testing"

	"github.com/wwsheng009/mint/ui/components/text"
)

func TestNewRowBuilder(t *testing.T) {
	vnode := NewRowBuilder().
		Children(NewColBuilder().Span(24).Children(text.New("Body")).Build()).
		Build()
	if vnode == nil {
		t.Fatal("NewRowBuilder().Build() returned nil")
	}
	if vnode.Tag() != "row" {
		t.Fatalf("Tag = %q, want row", vnode.Tag())
	}
}

func TestNewColBuilder(t *testing.T) {
	vnode := NewColBuilder().
		Span(12).
		Children(text.New("Body")).
		Build()
	if vnode == nil {
		t.Fatal("NewColBuilder().Build() returned nil")
	}
	if vnode.Tag() != "col" {
		t.Fatalf("Tag = %q, want col", vnode.Tag())
	}
}
