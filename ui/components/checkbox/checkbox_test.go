package checkbox

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	cb := New("Test")
	if cb == nil {
		t.Fatal("New returned nil")
	}
	if cb.Label() != "Test" {
		t.Errorf("Label = %q, want %q", cb.Label(), "Test")
	}
	if cb.Tag() != "checkbox" {
		t.Errorf("Tag = %q, want %q", cb.Tag(), "checkbox")
	}
}

func TestVNode_Defaults(t *testing.T) {
	cb := New("Test")
	if cb.Checked() {
		t.Error("Default checked should be false")
	}
	if cb.Disabled() {
		t.Error("Default disabled should be false")
	}
	if cb.ToggleIntent() != nil {
		t.Error("Default toggleIntent should be nil")
	}
}

func TestVNode_Builder(t *testing.T) {
	cb := NewBuilder().
		Label("Accept Terms").
		Checked(true).
		Disabled(false).
		Key("terms").
		Build()

	vnode := cb.(*VNode)
	if vnode.Label() != "Accept Terms" {
		t.Errorf("Label = %q, want %q", vnode.Label(), "Accept Terms")
	}
	if !vnode.Checked() {
		t.Error("Checked should be true")
	}
	if vnode.Key() != "terms" {
		t.Errorf("Key = %q, want %q", vnode.Key(), "terms")
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	cb := New("Test").SetChecked(true)
	inst := cb.CreateInstance()

	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	ci, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}
	if ci.Label() != "Test" {
		t.Errorf("Label = %q, want %q", ci.Label(), "Test")
	}
	if !ci.IsChecked() {
		t.Error("Checked should be true")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label":   "Test",
		"checked": true,
	})

	if inst.Label() != "Test" {
		t.Errorf("Label = %q, want %q", inst.Label(), "Test")
	}
	if !inst.IsChecked() {
		t.Error("Checked should be true")
	}
}

func TestInstance_Measure(t *testing.T) {
	tests := []struct {
		name      string
		label     string
		wantWidth int
	}{
		{"Empty label", "", 4},           // "[X]" + " " = 4
		{"Short label", "OK", 6},         // 4 + 2 = 6
		{"Long label", "Accept Terms", 16}, // 4 + 12 = 16
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstance(rtui.Props{
				"label": tt.label,
			})

			size := inst.Measure(layout.UnboundedConstraints())

			if size.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", size.Width, tt.wantWidth)
			}
			if size.Height != 1 {
				t.Errorf("Height = %d, want 1", size.Height)
			}
		})
	}
}

func TestInstance_Measure_WithConstraints(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label": "Test Checkbox",
	})

	constraints := layout.Constraints{
		MinWidth: 10,
		MaxWidth: 20,
	}

	size := inst.Measure(constraints)

	// "Test Checkbox" = 13 chars + 4 = 17, within [10, 20]
	if size.Width < 10 || size.Width > 20 {
		t.Errorf("Width = %d, want between 10 and 20", size.Width)
	}
}

func TestInstance_Toggle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label":   "Test",
		"checked": false,
	})

	// Toggle should change state
	newState := inst.Toggle()
	if !newState {
		t.Error("Toggle should return true")
	}
	if !inst.IsChecked() {
		t.Error("Checked should be true after toggle")
	}

	// Toggle again
	newState = inst.Toggle()
	if newState {
		t.Error("Toggle should return false")
	}
	if inst.IsChecked() {
		t.Error("Checked should be false after second toggle")
	}
}

func TestInstance_SetChecked(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label":   "Test",
		"checked": false,
	})

	inst.SetChecked(true)
	if !inst.IsChecked() {
		t.Error("Checked should be true")
	}

	inst.SetChecked(false)
	if inst.IsChecked() {
		t.Error("Checked should be false")
	}

	// Setting same value should not mark dirty
	inst.ClearDirty()
	inst.SetChecked(false)
	if inst.IsDirty() {
		t.Error("Should not be dirty when setting same value")
	}
}

func TestInstance_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label":    "Test",
		"disabled": true,
	})

	if !inst.IsDisabled() {
		t.Error("Should be disabled")
	}

	// Disabled instance should not handle actions
	if inst.CanHandleAction("toggle") {
		t.Error("Disabled instance should not handle toggle action")
	}
}

func TestInstance_HandleAction(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label":   "Test",
		"checked": false,
	})

	// Toggle action
	handled := inst.HandleAction("toggle", nil)
	if !handled {
		t.Error("Toggle action should be handled")
	}
	if !inst.IsChecked() {
		t.Error("Should be checked after toggle")
	}

	// Click action
	handled = inst.HandleAction("click", nil)
	if !handled {
		t.Error("Click action should be handled")
	}
	if inst.IsChecked() {
		t.Error("Should be unchecked after click")
	}

	// Unknown action
	handled = inst.HandleAction("unknown", nil)
	if handled {
		t.Error("Unknown action should not be handled")
	}
}

func TestInstance_Focus(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label": "Test",
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
	tests := []struct {
		name    string
		checked bool
		want    string
	}{
		{"Unchecked", false, "[ ] Test"},
		{"Checked", true, "[X] Test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstance(rtui.Props{
				"label":   "Test",
				"checked": tt.checked,
			})

			cmds := inst.Paint(0, 0)
			if len(cmds) != 1 {
				t.Fatalf("Paint returned %d commands, want 1", len(cmds))
			}

			if cmds[0].Text != tt.want {
				t.Errorf("Text = %q, want %q", cmds[0].Text, tt.want)
			}
		})
	}
}

func TestInstance_Paint_EmptyLabel(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label":   "",
		"checked": true,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}

	if cmds[0].Text != "[X]" {
		t.Errorf("Text = %q, want %q", cmds[0].Text, "[X]")
	}
}