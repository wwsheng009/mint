package input

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestMouseClickFocusChange verifies that Input component responds correctly
// to focus state changes, simulating mouse click focus switching
func TestMouseClickFocusChange(t *testing.T) {
	// Create an input instance
	inst := NewInstance(rtui.Props{
		"key":        "test-input",
		"placeholder": "Click to focus",
		"width":      20,
	})

	// Test initial state
	if inst.HasFocus() {
		t.Error("Input should not have focus initially")
	}

	// Simulate mouse click by setting focus (this is what FocusManager does)
	inst.SetFocus(true)

	// Verify focus state is updated
	if !inst.HasFocus() {
		t.Error("Input should have focus after SetFocus(true)")
	}

	// Simulate clicking another input (losing focus)
	inst.SetFocus(false)

	// Verify focus state is cleared
	if inst.HasFocus() {
		t.Error("Input should not have focus after SetFocus(false)")
	}
}

// TestFocusStyle verifies that the style changes correctly when focused
func TestFocusStyle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key":        "test-input",
		"placeholder": "Focus style test",
		"width":      20,
	})

	// Get style when not focused
	styleBefore := inst.resolveStyle()
	// Check if style has bold modification (Bold() when true adds to the style)
	// In our implementation, Bold(true) is called when focused
	if styleBefore.IsBold() || styleBefore.IsUnderline() {
		t.Error("Style should not be bold or underlined when not focused")
	}

	// Set focus
	inst.SetFocus(true)

	// Get style when focused
	styleAfter := inst.resolveStyle()
	if !styleAfter.IsBold() || !styleAfter.IsUnderline() {
		t.Error("Style should be bold and underlined when focused")
	}
}

// TestFocusWhileDisabled verifies that disabled input's focus state
// Note: FocusManager filters out disabled components, so they shouldn't be in focusable list
// This test verifies that IsDisabled() correctly reports disabled state
func TestFocusWhileDisabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key":        "disabled-input",
		"placeholder": "Disabled input",
		"width":      20,
		"disabled":   true,
	})

	// Verify that disabled state is correctly set
	if !inst.IsDisabled() {
		t.Error("Input should be disabled")
	}

	// Note: SetFocus may still change the internal state when called directly,
	// but FocusManager doesn't include disabled components in focusable list.
	// This is correct design: FocusManager is responsible for filtering,
	// not individual components.
}

// TestFocusWhileReadOnly verifies that read-only input can receive focus
func TestFocusWhileReadOnly(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key":        "readonly-input",
		"placeholder": "Read-only input",
		"width":      20,
		"readOnly":   true,
	})

	// Read-only input CAN receive focus (common UI pattern)
	inst.SetFocus(true)

	// Verify that read-only input can receive focus
	if !inst.HasFocus() {
		t.Error("Read-only input should be able to receive focus")
	}
}

// TestHandleActionIgnoresUnknown focuses only on relevant actions
func TestHandleActionIgnoresUnknown(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"key":        "test-input",
		"placeholder": "Test",
	})
	inst.SetFocus(true) // Set initial focus

	// Unknown action should return false
	unknownAct := action.NewAction("unknown_action")
	handled := inst.HandleAction(unknownAct)
	if handled {
		t.Error("Unknown action should return false")
	}

	// Focus should remain unchanged
	if !inst.HasFocus() {
		t.Error("Focus should remain after processing unknown action")
	}
}
