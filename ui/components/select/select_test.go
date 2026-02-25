package selectcomp

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New returned nil")
	}
	if s.Tag() != "select" {
		t.Errorf("Tag = %q, want %q", s.Tag(), "select")
	}
}

func TestVNode_Defaults(t *testing.T) {
	s := New()
	if len(s.Options()) != 0 {
		t.Error("Default options should be empty")
	}
	if s.SelectedIndex() != -1 {
		t.Errorf("Default selectedIndex = %d, want -1", s.SelectedIndex())
	}
	if s.Disabled() {
		t.Error("Default disabled should be false")
	}
}

func TestVNode_Builder(t *testing.T) {
	opts := []Option{
		{Value: "a", Label: "Option A"},
		{Value: "b", Label: "Option B"},
	}

	s := NewBuilder().
		Options(opts).
		Selected(1).
		Key("myselect").
		Build()

	vnode := s.(*VNode)
	if len(vnode.Options()) != 2 {
		t.Errorf("Options count = %d, want 2", len(vnode.Options()))
	}
	if vnode.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", vnode.SelectedIndex())
	}
	if vnode.Key() != "myselect" {
		t.Errorf("Key = %q, want %q", vnode.Key(), "myselect")
	}
}

func TestVNode_AddOption(t *testing.T) {
	s := New().
		AddOption("a", "Option A").
		AddOption("b", "Option B")

	if len(s.Options()) != 2 {
		t.Errorf("Options count = %d, want 2", len(s.Options()))
	}
	if s.Options()[0].Value != "a" {
		t.Errorf("First option value = %q, want %q", s.Options()[0].Value, "a")
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	s := New().
		AddOption("a", "Option A").
		SetSelectedIndex(0)

	inst := s.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	ci, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}
	if ci.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0", ci.SelectedIndex())
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
		"selectedIndex": 1,
	})

	if len(inst.options) != 2 {
		t.Errorf("Options count = %d, want 2", len(inst.options))
	}
	if inst.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", inst.SelectedIndex())
	}
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "Short"},
			{Value: "b", Label: "Longer Option"},
		},
	})

	size := inst.Measure(layout.UnboundedConstraints())

	// "Longer Option" = 13 chars + 4 for "< >" = 17
	if size.Width < 17 {
		t.Errorf("Width = %d, want >= 17", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Height = %d, want 1", size.Height)
	}
}

func TestInstance_SelectNext(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
			{Value: "c", Label: "C"},
		},
		"selectedIndex": 0,
	})

	inst.SelectNext()
	if inst.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", inst.SelectedIndex())
	}

	inst.SelectNext()
	if inst.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex = %d, want 2", inst.SelectedIndex())
	}

	// Wrap around
	inst.SelectNext()
	if inst.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0 (wrap)", inst.SelectedIndex())
	}
}

func TestInstance_SelectPrev(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
			{Value: "c", Label: "C"},
		},
		"selectedIndex": 0,
	})

	inst.SelectPrev()
	if inst.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex = %d, want 2 (wrap)", inst.SelectedIndex())
	}

	inst.SelectPrev()
	if inst.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", inst.SelectedIndex())
	}
}

func TestInstance_SelectedValue(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "val_a", Label: "Label A"},
			{Value: "val_b", Label: "Label B"},
		},
		"selectedIndex": 1,
	})

	if inst.SelectedValue() != "val_b" {
		t.Errorf("SelectedValue = %q, want %q", inst.SelectedValue(), "val_b")
	}
	if inst.SelectedLabel() != "Label B" {
		t.Errorf("SelectedLabel = %q, want %q", inst.SelectedLabel(), "Label B")
	}
}

func TestInstance_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
		},
		"disabled": true,
	})

	if !inst.IsDisabled() {
		t.Error("Should be disabled")
	}

	// Disabled instance should not handle actions
	if inst.CanHandleAction("select") {
		t.Error("Disabled instance should not handle select action")
	}
}

func TestInstance_HandleAction(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
		},
		"selectedIndex": 0,
	})

	// Select/Click/Down action
	handled := inst.HandleAction("select", nil)
	if !handled {
		t.Error("Select action should be handled")
	}
	if inst.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", inst.SelectedIndex())
	}

	// Up action
	handled = inst.HandleAction("up", nil)
	if !handled {
		t.Error("Up action should be handled")
	}
	if inst.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0", inst.SelectedIndex())
	}

	// Unknown action
	handled = inst.HandleAction("unknown", nil)
	if handled {
		t.Error("Unknown action should not be handled")
	}
}

func TestInstance_Focus(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
		},
	})

	if inst.HasFocus() {
		t.Error("Should not have focus initially")
	}

	inst.SetFocus(true)
	if !inst.HasFocus() {
		t.Error("Should have focus after SetFocus(true)")
	}

	inst.SetFocus(false)
	if inst.HasFocus() {
		t.Error("Should not have focus after SetFocus(false)")
	}
}

func TestInstance_Paint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
		"selectedIndex": 1,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}

	expected := "< Option B >"
	if cmds[0].Text != expected {
		t.Errorf("Text = %q, want %q", cmds[0].Text, expected)
	}
}

func TestInstance_Paint_NoSelection(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "Option A"},
		},
		"selectedIndex": -1,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}

	// No selection shows first option
	expected := "< Option A >"
	if cmds[0].Text != expected {
		t.Errorf("Text = %q, want %q", cmds[0].Text, expected)
	}
}