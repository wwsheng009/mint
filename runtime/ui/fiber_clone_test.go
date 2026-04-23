package ui

import "testing"

func TestCloneFiber_PreservesBusinessID(t *testing.T) {
	original := &Fiber{
		Tag:     "select",
		Key:     "select-1",
		DiffKey: "select-1",
		ID:      "profile.country",
		NodeID:  42,
	}

	clone := CloneFiber(original)
	if clone == nil {
		t.Fatal("CloneFiber returned nil")
	}
	if clone.ID != original.ID {
		t.Fatalf("clone.ID = %q, want %q", clone.ID, original.ID)
	}
}
