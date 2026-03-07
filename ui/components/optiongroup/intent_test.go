package optiongroup

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/layout"
)

// =============================================================================
// Phase 5: Intent Bubble Integration Tests
// =============================================================================

// TestOptionGroup_HandleIntent_SingleSelect tests IntentHandler interface
// for single-select mode.
func TestOptionGroup_HandleIntent_SingleSelect(t *testing.T) {
	// Create an OptionGroup instance
	props := rtui.Props{
		"key":   "test-group",
		"label": "Test Group",
		"mode":  ModeSingle,
	}
	inst := NewInstance(props)

	// Test initial state
	if inst.GetSelected() != "" {
		t.Errorf("Expected empty selection, got %s", inst.GetSelected())
	}

	// Emit OptionSelectIntent
	intent := OptionSelectIntent{
		GroupKey:   "test-group",
		Value:      "option1",
		IsSelected: true,
		Mode:       ModeSingle,
	}

	handled := inst.HandleIntent(intent)
	if !handled {
		t.Error("Intent should be handled by OptionGroup")
	}

	// Verify selection
	if inst.GetSelected() != "option1" {
		t.Errorf("Expected selection 'option1', got '%s'", inst.GetSelected())
	}
}

// TestOptionGroup_HandleIntent_MultiSelect tests IntentHandler interface
// for multi-select mode.
func TestOptionGroup_HandleIntent_MultiSelect(t *testing.T) {
	// Create an OptionGroup instance in multi-select mode
	props := rtui.Props{
		"key":   "test-group",
		"label": "Test Group",
		"mode":  ModeMultiple,
	}
	inst := NewInstance(props)

	// Select option1
	intent1 := OptionSelectIntent{
		GroupKey:   "test-group",
		Value:      "option1",
		IsSelected: true,
		Mode:       ModeMultiple,
	}

	handled := inst.HandleIntent(intent1)
	if !handled {
		t.Error("Intent should be handled")
	}

	if !inst.isOptionSelected("option1") {
		t.Error("Option1 should be selected")
	}

	// Select option2
	intent2 := OptionSelectIntent{
		GroupKey:   "test-group",
		Value:      "option2",
		IsSelected: true,
		Mode:       ModeMultiple,
	}

	inst.HandleIntent(intent2)

	if !inst.isOptionSelected("option2") {
		t.Error("Option2 should be selected")
	}

	// Verify both are selected
	if inst.isOptionSelected("option1") && inst.isOptionSelected("option2") {
		t.Log("Both option1 and option2 are selected correctly")
	} else {
		t.Error("Both options should be selected")
	}
}

// TestOptionGroup_HandleIntent_Deselect tests deselecting in multi-select mode.
func TestOptionGroup_HandleIntent_Deselect(t *testing.T) {
	props := rtui.Props{
		"key":   "test-group",
		"label": "Test Group",
		"mode":  ModeMultiple,
	}
	inst := NewInstance(props)

	// Select both options
	intent1 := OptionSelectIntent{
		GroupKey:   "test-group",
		Value:      "option1",
		IsSelected: true,
		Mode:       ModeMultiple,
	}
	intent2 := OptionSelectIntent{
		GroupKey:   "test-group",
		Value:      "option2",
		IsSelected: true,
		Mode:       ModeMultiple,
	}

	inst.HandleIntent(intent1)
	inst.HandleIntent(intent2)

	// Deselect option1
	deselectIntent := OptionSelectIntent{
		GroupKey:   "test-group",
		Value:      "option1",
		IsSelected: false,
		Mode:       ModeMultiple,
	}

	handled := inst.HandleIntent(deselectIntent)
	if !handled {
		t.Error("Deselect intent should be handled")
	}

	if inst.isOptionSelected("option1") {
		t.Error("Option1 should be deselected")
	}

	if !inst.isOptionSelected("option2") {
		t.Error("Option2 should still be selected")
	}
}

// TestOptionGroup_HandleIntent_WrongGroup tests that intent with wrong group key is ignored.
func TestOptionGroup_HandleIntent_WrongGroup(t *testing.T) {
	props := rtui.Props{
		"key":   "test-group",
		"label": "Test Group",
		"mode":  ModeSingle,
	}
	inst := NewInstance(props)

	// Emit intent for different group
	intent := OptionSelectIntent{
		GroupKey:   "other-group",
		Value:      "option1",
		IsSelected: true,
		Mode:       ModeSingle,
	}

	handled := inst.HandleIntent(intent)
	if handled {
		t.Error("Intent with wrong group key should not be handled")
	}

	if inst.GetSelected() != "" {
		t.Error("Selection should be empty")
	}
}

// TestOptionHandleAction_EmitIntent tests that Option.HandleAction emits OptionSelectIntent.
func TestOptionHandleAction_EmitIntent(t *testing.T) {
	// Create parent OptionGroup
	groupProps := rtui.Props{
		"key":   "parent-group",
		"label": "Parent Group",
		"mode":  ModeSingle,
	}
	groupInst := NewInstance(groupProps)

	// Create Option instance with parent reference
	optionProps := rtui.Props{
		"key":   "opt1",
		"value": "option1",
		"label": "Option 1",
		"mode":  ModeSingle,
	}
	optionInst := NewOptionInstance(optionProps)
	optionInst.parent = groupInst // Set parent reference (Instance Tree)

	// Handle action (simulate click/click) - this will emit intent and bubble to parent
	act := &action.Action{Type: action.ActionClick}
	_ = optionInst.HandleAction(act)

	// The intent should have bubbled up and been handled by groupInst.HandleIntent()
	// Check that the group handled it correctly
	if groupInst.GetSelected() != "option1" {
		t.Errorf("Expected group selection 'option1', got '%s'", groupInst.GetSelected())
	}
}

// TestOption_isOptionSelected tests the selection query method.
func TestOption_isOptionSelected(t *testing.T) {
	groupProps := rtui.Props{
		"key":   "test-group",
		"label": "Test Group",
		"mode":  ModeSingle,
	}
	groupInst := NewInstance(groupProps)

	// Create Option instance with parent reference
	optionProps := rtui.Props{
		"key":   "opt1",
		"value": "option1",
		"label": "Option 1",
		"mode":  ModeSingle,
	}
	optionInst := NewOptionInstance(optionProps)
	optionInst.parent = groupInst

	// Initially not selected
	if optionInst.isOptionSelected() {
		t.Error("Option should not be selected initially")
	}

	// Select the option
	groupInst.SelectOption("option1")

	// Now should be selected
	if !optionInst.isOptionSelected() {
		t.Error("Option should be selected after SelectOption")
	}
}

// TestOption_InstanceTree_Methods tests Parent() and Children() methods.
func TestOption_InstanceTree_Methods(t *testing.T) {
	groupProps := rtui.Props{
		"key":   "test-group",
		"label": "Test Group",
		"mode":  ModeSingle,
	}
	groupInst := NewInstance(groupProps)

	optionProps := rtui.Props{
		"key":   "opt1",
		"value": "option1",
		"label": "Option 1",
		"mode":  ModeSingle,
	}
	optionInst := NewOptionInstance(optionProps)
	optionInst.parent = groupInst

	// Test Parent() returns interface{} (for TreeComponent)
	parent := optionInst.Parent()
	if parent == nil {
		t.Error("Parent should not be nil")
	}

	if parentGroup, ok := parent.(*Instance); ok {
		if parentGroup.key != "test-group" {
			t.Errorf("Expected parent key 'test-group', got '%s'", parentGroup.key)
		}
	} else {
		t.Error("Parent should be of type *Instance")
	}

	// Test Children() returns nil (Option is a leaf)
	children := optionInst.Children()
	if children != nil {
		t.Error("Option children should be nil")
	}
}

// TestOptionGroup_Measurable tests that OptionGroup still works with layout.
func TestOptionGroup_Measurable(t *testing.T) {
	props := rtui.Props{
		"key":   "test-group",
		"label": "Test Group",
		"mode":  ModeSingle,
	}
	inst := NewInstance(props)

	// Set some basic props
	inst.SetProps(rtui.Props{
		"label": "Test Label",
		"style": style.Style{FG: "white", BG: "black"},
	})

	// Test Measure
	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 100,
	}

	size := inst.Measure(constraints)

	// Should have a reasonable size
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("Invalid size: %dx%d", size.Width, size.Height)
	}
}
