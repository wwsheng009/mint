package ui

import "testing"

func TestNewTransferBuilder(t *testing.T) {
	vnode := NewTransferBuilder().
		Items([]TransferItem{
			NewTransferItem("a", "Alpha"),
		}).
		Build()
	if vnode == nil {
		t.Fatal("NewTransferBuilder().Build() returned nil")
	}
	if vnode.Tag() != "transfer" {
		t.Fatalf("Tag = %q, want transfer", vnode.Tag())
	}
}

func TestTransferShortcut(t *testing.T) {
	vnode := Transfer([]TransferItem{
		NewTransferItem("a", "Alpha"),
	})
	if vnode == nil {
		t.Fatal("Transfer() returned nil")
	}
	if vnode.Tag() != "transfer" {
		t.Fatalf("Tag = %q, want transfer", vnode.Tag())
	}
}
