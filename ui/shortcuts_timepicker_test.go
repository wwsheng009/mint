package ui

import "testing"

func TestNewTimePickerBuilder(t *testing.T) {
	vnode := NewTimePickerBuilder().
		Placeholder("HH:mm").
		Value("09:30").
		Build()
	if vnode == nil {
		t.Fatal("NewTimePickerBuilder().Build() returned nil")
	}
	if vnode.Tag() != "timepicker" {
		t.Fatalf("Tag = %q, want timepicker", vnode.Tag())
	}
}
