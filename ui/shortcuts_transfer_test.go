package ui

import "testing"

func TestNewTransferBuilder(t *testing.T) {
	vnode := NewTransferBuilder().
		BulkOperations(true).
		BulkOperationLabels("All Send", "All Return").
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
	if bulkOperations, _ := vnode.Props()["bulkOperations"].(bool); !bulkOperations {
		t.Fatal("bulkOperations = false, want true")
	}
	if labels, _ := vnode.Props()["bulkOperationLabels"].([2]string); labels != [2]string{"All Send", "All Return"} {
		t.Fatalf("bulkOperationLabels = %#v, want custom labels", labels)
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
