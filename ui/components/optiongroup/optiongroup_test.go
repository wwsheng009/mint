package optiongroup

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	opts := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
	}
	og := New(opts)

	if og == nil {
		t.Fatal("New returned nil")
	}
	if len(og.Options()) != 2 {
		t.Errorf("Options length = %d, want 2", len(og.Options()))
	}
	if og.Label() != "" {
		t.Errorf("Label = %q, want empty", og.Label())
	}
	if og.Tag() != "optiongroup" {
		t.Errorf("Tag = %q, want %q", og.Tag(), "optiongroup")
	}
}

func TestVNode_Defaults(t *testing.T) {
	opts := []Option{{Value: "opt1"}}
	og := New(opts)

	if og.Mode() != ModeSingle {
		t.Errorf("Default mode should be ModeSingle, got %v", og.Mode())
	}
	if og.Disabled() {
		t.Error("Default disabled should be false")
	}
	if og.Selected() != "" {
		t.Error("Default selected should be empty")
	}
	if og.SelectIntent() != nil {
		t.Error("Default selectIntent should be nil")
	}
	if og.Orientation() != OrientationVertical {
		t.Errorf("Default orientation should be Vertical, got %v", og.Orientation())
	}
}

func TestVNode_Builder(t *testing.T) {
	opts := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
	}
	og := NewBuilder(opts).
		Label("Select an option").
		Mode(ModeMultiple).
		Selecteds([]string{"opt1"}).
		Key("test-group").
		Horizontal().
		Spacing(2).
		Build()

	vnode := og.(*VNode)
	if vnode.Label() != "Select an option" {
		t.Errorf("Label = %q, want %q", vnode.Label(), "Select an option")
	}
	if vnode.Mode() != ModeMultiple {
		t.Errorf("Mode = %v, want ModeMultiple", vnode.Mode())
	}
	if vnode.Key() != "test-group" {
		t.Errorf("Key = %q, want %q", vnode.Key(), "test-group")
	}
	if vnode.Orientation() != OrientationHorizontal {
		t.Errorf("Orientation = %v, want Horizontal", vnode.Orientation())
	}
	if vnode.Spacing() != 2 {
		t.Errorf("Spacing = %d, want 2", vnode.Spacing())
	}
}

func TestVNode_Single_Multiple(t *testing.T) {
	opts := []Option{{Value: "opt1"}}

	og1 := New(opts).Single()
	if og1.Mode() != ModeSingle {
		t.Error("Single() did not set ModeSingle")
	}

	og2 := New(opts).Multiple()
	if og2.Mode() != ModeMultiple {
		t.Error("Multiple() did not set ModeMultiple")
	}
}

func TestVNode_Vertical_Horizontal(t *testing.T) {
	opts := []Option{{Value: "opt1"}}

	og1 := New(opts).Vertical()
	if og1.Orientation() != OrientationVertical {
		t.Error("Vertical() did not set OrientationVertical")
	}

	og2 := New(opts).Horizontal()
	if og2.Orientation() != OrientationHorizontal {
		t.Error("Horizontal() did not set OrientationHorizontal")
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	opts := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
	}
	og := New(opts).SetSelected("opt1")
	inst := og.CreateInstance()

	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	oi, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}
	if oi.GetSelected() != "opt1" {
		t.Errorf("Selected = %q, want %q", oi.GetSelected(), "opt1")
	}
	if len(oi.GetSelecteds()) != 1 {
		t.Errorf("Selecteds length = %d, want 1", len(oi.GetSelecteds()))
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	opts := []Option{{Value: "opt1", Label: "Option 1"}}
	inst := NewInstance(rtui.Props{
		"options":  opts,
		"selected": "opt1",
		"mode":     ModeSingle,
	})

	if inst.GetSelected() != "opt1" {
		t.Errorf("Selected = %q, want %q", inst.GetSelected(), "opt1")
	}
	if inst.Mode() != ModeSingle {
		t.Errorf("Mode = %v, want ModeSingle", inst.Mode())
	}
}

func TestInstance_Measure(t *testing.T) {
	tests := []struct {
		name       string
		label      string
		options    []Option
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "No label, one option",
			label:      "",
			options:    []Option{{Value: "opt1", Label: "Option 1"}},
			wantWidth:  12, // indicator(3) + space(1) + "Option 1"(9) = 13, but indicator is 3: "( )"
			wantHeight: 1,  // Option height is 1
		},
		{
			name:       "With label, one option",
			label:      "Select:",
			options:    []Option{{Value: "opt1", Label: "Option 1"}},
			wantWidth:  12, // Max of "Select:" (7) and option width (12)
			wantHeight: 2,  // Label (1) + option (1)
		},
		{
			name:  "Multiple options",
			label: "",
			options: []Option{
				{Value: "opt1", Label: "Option 1"},
				{Value: "opt2", Label: "Option 2"},
				{Value: "opt3", Label: "Option 3"},
			},
			wantWidth:  12, // All options have same width: indicator(3) + space(1) + text(8) = 12
			wantHeight: 3,  // 3 options, each height 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstance(rtui.Props{
				"label":   tt.label,
				"options": tt.options,
			})

			size := inst.Measure(layout.UnboundedConstraints())

			if size.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", size.Width, tt.wantWidth)
			}
			if size.Height != tt.wantHeight {
				t.Errorf("Height = %d, want %d", size.Height, tt.wantHeight)
			}
		})
	}
}

func TestInstance_Measure_WithConstraints(t *testing.T) {
	// After refactoring: Options are child nodes, so group Measure only considers label
	inst := NewInstance(rtui.Props{
		"label":    "Test Label",
		"options":  []Option{{Value: "opt1", Label: "Long Option Name"}},
	})

	constraints := layout.Constraints{
		MinWidth: 10,
		MaxWidth: 15,
	}

	size := inst.Measure(constraints)

	// "Test Label" = 10 chars, "Long Option Name" = 15 chars
	// Max width is 15, but constraint limits to [10, 15], so width = 15
	// Height = label (1) + option (1) = 2
	if size.Width != 15 {
		t.Errorf("Width = %d, want 15", size.Width)
	}
	if size.Height != 2 {
		t.Errorf("Height = %d, want 2", size.Height)
	}
}

func TestInstance_SelectOption_SingleMode(t *testing.T) {
	opts := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
	}
	inst := NewInstance(rtui.Props{
		"mode":    ModeSingle,
		"options": opts,
	})

	// Select first option
	inst.SelectOption("opt1")
	if inst.GetSelected() != "opt1" {
		t.Errorf("Selected = %q, want %q", inst.GetSelected(), "opt1")
	}
	if len(inst.GetSelecteds()) != 1 {
		t.Errorf("Selecteds length = %d, want 1", len(inst.GetSelecteds()))
	}
	if inst.GetSelecteds()[0] != "opt1" {
		t.Errorf("Selecteds[0] = %q, want %q", inst.GetSelecteds()[0], "opt1")
	}

	// Select second option
	inst.SelectOption("opt2")
	if inst.GetSelected() != "opt2" {
		t.Errorf("Selected = %q, want %q", inst.GetSelected(), "opt2")
	}
}

func TestInstance_SelectOption_MultipleMode(t *testing.T) {
	opts := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
		{Value: "opt3", Label: "Option 3"},
	}
	inst := NewInstance(rtui.Props{
		"mode":    ModeMultiple,
		"options": opts,
	})

	// Select first option
	inst.SelectOption("opt1")
	if len(inst.GetSelecteds()) != 1 {
		t.Errorf("Selecteds length = %d, want 1", len(inst.GetSelecteds()))
	}

	// Select second option
	inst.SelectOption("opt2")
	if len(inst.GetSelecteds()) != 2 {
		t.Errorf("Selecteds length = %d, want 2", len(inst.GetSelecteds()))
	}

	// Toggle first option (deselect)
	inst.SelectOption("opt1")
	if len(inst.GetSelecteds()) != 1 {
		t.Errorf("Selecteds length = %d, want 1", len(inst.GetSelecteds()))
	}
	if inst.GetSelecteds()[0] != "opt2" {
		t.Errorf("Selecteds[0] = %q, want %q", inst.GetSelecteds()[0], "opt2")
	}
}

func TestInstance_Focus(t *testing.T) {
	opts := []Option{{Value: "opt1"}}
	inst := NewInstance(rtui.Props{
		"options": opts,
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

func TestInstance_Disabled(t *testing.T) {
	opts := []Option{{Value: "opt1"}}
	inst := NewInstance(rtui.Props{
		"options":  opts,
		"disabled": true,
	})

	if !inst.IsDisabled() {
		t.Error("Should be disabled")
	}
}

func TestInstance_HandleAction(t *testing.T) {
	opts := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
		{Value: "opt3", Label: "Option 3"},
	}
	inst := NewInstance(rtui.Props{
		"mode":    ModeSingle,
		"options": opts,
	})

	// After refactoring: NavigateDown/Up are handled by FocusManager at option level
	// Group-level HandleAction should return false for navigation actions
	handled := inst.HandleAction(action.NewAction(action.ActionNavigateDown))
	if handled {
		t.Error("ActionNavigateDown should not be handled at group level")
	}

	handled = inst.HandleAction(action.NewAction(action.ActionNavigateUp))
	if handled {
		t.Error("ActionNavigateUp should not be handled at group level")
	}

	// Enter/Click/Space are also handled at option level now
	// Group-level HandleAction should return false
	handled = inst.HandleAction(action.NewAction(action.ActionEnter))
	if handled {
		t.Error("ActionEnter should not be handled at group level")
	}
}

func TestInstance_Paint(t *testing.T) {
	tests := []struct {
		name    string
		mode    SelectMode
		opts    []Option
		wantLen int
	}{
		{
			name: "Single mode with label",
			mode: ModeSingle,
			opts: []Option{
				{Value: "opt1", Label: "Option 1"},
				{Value: "opt2", Label: "Option 2"},
			},
			wantLen: 0, // Paint returns empty, all content rendered by child nodes
		},
		{
			name: "Multiple mode no label",
			mode: ModeMultiple,
			opts: []Option{
				{Value: "opt1", Label: "Option 1"},
			},
			wantLen: 0, // Paint returns empty, all content rendered by child nodes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := rtui.Props{
				"mode":    tt.mode,
				"options": tt.opts,
			}
			if tt.name == "Single mode with label" {
				props["label"] = "Select an option"
			}
			inst := NewInstance(props)

			cmds := inst.Paint(0, 0)
			if len(cmds) != tt.wantLen {
				t.Fatalf("Paint returned %d commands, want %d", len(cmds), tt.wantLen)
			}
		})
	}
}

func TestInstance_Paint_WithSelection(t *testing.T) {
	opts := []Option{
		{Value: "opt1", Label: "Option 1"},
		{Value: "opt2", Label: "Option 2"},
	}
	inst := NewInstance(rtui.Props{
		"mode":     ModeSingle,
		"options":  opts,
		"selected": "opt1",
		"label":    "Select an option",
	})

	cmds := inst.Paint(0, 0)
	// After refactoring: OptionGroup Paint returns empty
	// All content (label and options) is rendered by child nodes
	if len(cmds) != 0 {
		t.Fatalf("Paint returned %d commands, want 0 (all content rendered by child nodes)", len(cmds))
	}

	// Verify that child instances were created (options + label)
	if len(inst.childInstances) != 2 {
		t.Errorf("Expected 2 child instances (for 2 options), got %d", len(inst.childInstances))
	}
}

func TestInstance_SetSelected(t *testing.T) {
	opts := []Option{{Value: "opt1"}}
	inst := NewInstance(rtui.Props{
		"options": opts,
	})

	inst.SetSelected("opt1")
	if inst.GetSelected() != "opt1" {
		t.Errorf("Selected = %q, want %q", inst.GetSelected(), "opt1")
	}

	inst.SetSelecteds([]string{"opt1", "opt2"})
	if len(inst.GetSelecteds()) != 2 {
		t.Errorf("Selecteds length = %d, want 2", len(inst.GetSelecteds()))
	}
}

func TestInstance_SetProps(t *testing.T) {
	opts := []Option{{Value: "opt1"}}
	inst := NewInstance(rtui.Props{
		"options": opts,
	})

	// Update props
	changed := inst.SetProps(rtui.Props{
		"label":     "New Label",
		"selected":  "opt1",
		"selecteds": []string{"opt1"},
		"mode":      ModeMultiple,
	})

	if !changed {
		t.Error("SetProps returned false, want true")
	}

	if inst.label != "New Label" {
		t.Errorf("label = %q, want %q", inst.label, "New Label")
	}

	if inst.mode != ModeMultiple {
		t.Errorf("mode = %v, want ModeMultiple", inst.mode)
	}
}

// =============================================================================
// Helpers
// =============================================================================

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr))
}
