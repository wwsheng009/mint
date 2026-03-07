package optiongroup

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestFocusability tests that Option instances can receive focus.
func TestFocusability(t *testing.T) {
	options := []Option{
		{Value: "bj", Label: "Beijing"},
		{Value: "sh", Label: "Shanghai"},
	}

	// Create OptionGroup Instance
	props := rtui.Props{
		"key":   "test-group",
		"mode":  ModeSingle,
		"options": options,
	}
	inst := NewInstance(props)
	inst.SetProps(props)

	// Check that child instances can be focused
	for i, child := range inst.childInstances {
		if child.IsDisabled() {
			t.Errorf("Option %d should not be disabled", i)
		}

		// Try setting focus
		child.SetFocus(true)
		if !child.HasFocus() {
			t.Errorf("Option %d should have focus after SetFocus(true)", i)
		}

		// Try clearing focus
		child.SetFocus(false)
		if child.HasFocus() {
			t.Errorf("Option %d should not have focus after SetFocus(false)", i)
		}
	}
}

// TestActionHandling tests that Option instances handle click actions.
func TestActionHandling(t *testing.T) {
	options := []Option{
		{Value: "bj", Label: "Beijing"},
		{Value: "sh", Label: "Shanghai"},
	}

	// Create OptionGroup Instance
	props := rtui.Props{
		"key":   "test-group",
		"mode":  ModeSingle,
		"options": options,
	}
	inst := NewInstance(props)
	inst.SetProps(props)

	// Note: Intent handling now happens via HandleIntent, not via SetIntentEmitter
	// The intent bubbles up through the instance tree and is handled by OptionGroup

	// Simulate click on first option
	clickAction := &action.Action{Type: action.ActionClick}
	beijingOption := inst.childInstances[0]

	// Initial state: nothing selected
	if inst.selected != "" {
		t.Errorf("Expected no initial selection, got %q", inst.selected)
	}

	handled := beijingOption.HandleAction(clickAction)
	if !handled {
		t.Error("Click action should be handled by parent")
	}

	// Check that selection was updated
	if inst.selected != "bj" {
		t.Errorf("Expected selected='bj' after click, got %q", inst.selected)
	}

	// Simulate Enter key on second option
	enterAction := &action.Action{Type: action.ActionEnter}
	shanghaiOption := inst.childInstances[1]

	handled = shanghaiOption.HandleAction(enterAction)
	if !handled {
		t.Error("Enter action should be handled by parent")
	}

	if inst.selected != "sh" {
		t.Errorf("Expected selected='sh' after Enter on Shanghai, got %q", inst.selected)
	}
}

// TestSelectionStateChangeActions tests that selection updates when actions are handled.
func TestSelectionStateChangeActions(t *testing.T) {
	options := []Option{
		{Value: "bj", Label: "Beijing"},
		{Value: "sh", Label: "Shanghai"},
		{Value: "gz", Label: "Guangzhou"},
	}

	// Create OptionGroup Instance
	props := rtui.Props{
		"key":   "test-group",
		"mode":  ModeSingle,
		"options": options,
	}
	inst := NewInstance(props)
	inst.SetProps(props)

	// Note: Parent references should be set up by Instance.AddChild
	// For this test, we manually set them
	for _, child := range inst.childInstances {
		child.parent = inst
	}

	clickAction := &action.Action{Type: action.ActionClick}

	// Test 1: Click Beijing -> should select it
	if !inst.childInstances[0].HandleAction(clickAction) {
		t.Error("Beijing click should be handled")
	}

	if inst.selected != "bj" {
		t.Errorf("After Beijing click, expected selected='bj', got %q", inst.selected)
	}
	if !inst.childInstances[0].selected {
		t.Error("Beijing OptionInstance should show selected=true")
	}
	if inst.childInstances[1].selected || inst.childInstances[2].selected {
		t.Error("Other options should not be selected")
	}

	// Test 2: Click Shanghai -> should switch selection
	if !inst.childInstances[1].HandleAction(clickAction) {
		t.Error("Shanghai click should be handled")
	}

	if inst.selected != "sh" {
		t.Errorf("After Shanghai click, expected selected='sh', got %q", inst.selected)
	}
	if inst.childInstances[0].selected {
		t.Error("Beijing should no longer be selected")
	}
	if !inst.childInstances[1].selected {
		t.Error("Shanghai OptionInstance should show selected=true")
	}
}

// TestParentReference tests that child instances have correct parent reference.
func TestParentReference(t *testing.T) {
	options := []Option{
		{Value: "bj", Label: "Beijing"},
		{Value: "sh", Label: "Shanghai"},
	}

	// Create OptionGroup Instance
	props := rtui.Props{
		"key":   "test-group",
		"mode":  ModeSingle,
		"options": options,
	}
	inst := NewInstance(props)
	inst.SetProps(props)

	// Check parent reference on all children
	for i, child := range inst.childInstances {
		if child.parent == nil {
			t.Errorf("Child %d has nil parent reference", i)
		} else if child.parent != inst {
			t.Errorf("Child %d has wrong parent reference (expected inst)", i)
		}
	}

	// Test that child can query parent for selection state
	beijingOption := inst.childInstances[0]
	shanghaiOption := inst.childInstances[1]

	// Select Beijing
	inst.SelectOption("bj")

	// Check that Beijing knows it's selected via parent
	if !beijingOption.isOptionSelected() {
		t.Error("Beijing isOptionSelected() should return true")
	}
	if shanghaiOption.isOptionSelected() {
		t.Error("Shanghai isOptionSelected() should return false")
	}

	// Switch to Shanghai
	inst.SelectOption("sh")

	if beijingOption.isOptionSelected() {
		t.Error("Beijing isOptionSelected() should return false after switch")
	}
	if !shanghaiOption.isOptionSelected() {
		t.Error("Shanghai isOptionSelected() should return true after switch")
	}
}

// TestMultiSelectActionHandling tests multi-select mode action handling.
func TestMultiSelectActionHandling(t *testing.T) {
	options := []Option{
		{Value: "dev", Label: "Development"},
		{Value: "design", Label: "Design"},
	}

	// Create multi-select OptionGroup Instance
	props := rtui.Props{
		"key":   "test-group",
		"mode":  ModeMultiple,
		"options": options,
	}
	inst := NewInstance(props)
	inst.SetProps(props)

	// Note: Parent references should be set up by Instance.AddChild
	// For this test, we manually set them
	for _, child := range inst.childInstances {
		child.parent = inst
	}

	clickAction := &action.Action{Type: action.ActionClick}

	// Click Development -> should add to selection
	if !inst.childInstances[0].HandleAction(clickAction) {
		t.Error("Development click should be handled")
	}

	if !contains(inst.selecteds, "dev") {
		t.Error("dev should be in selecteds")
	}
	if !inst.childInstances[0].selected {
		t.Error("Development OptionInstance should be selected")
	}

	// Click Design -> should add second selection
	if !inst.childInstances[1].HandleAction(clickAction) {
		t.Error("Design click should be handled")
	}

	if !contains(inst.selecteds, "dev") || !contains(inst.selecteds, "design") {
		t.Error("Both dev and design should be selected")
	}
	if !inst.childInstances[0].selected || !inst.childInstances[1].selected {
		t.Error("Both OptionInstances should be selected")
	}

	// Click Development again -> should remove selection
	if !inst.childInstances[0].HandleAction(clickAction) {
		t.Error("Development click should be handled")
	}

	if contains(inst.selecteds, "dev") {
		t.Error("dev should not be in selecteds after deselect")
	}
	if inst.childInstances[0].selected {
		t.Error("Development should not be selected after second click")
	}
}

// contains is a helper to check if string slice contains value.
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
