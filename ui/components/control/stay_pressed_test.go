package control

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
)

// =============================================================================
// Test Intents
// =============================================================================

// Intent that should keep pressed state
type KeepPressedIntent struct{}

func (KeepPressedIntent) IntentType() string { return "KeepPressed" }
func (KeepPressedIntent) StayPressed() bool  { return true }

// Intent that should reset immediately
type ResetImmedatelyIntent struct{}

func (ResetImmedatelyIntent) IntentType() string { return "ResetImmediate" }
func (ResetImmedatelyIntent) StayPressed() bool  { return false }

// Intent without StayPressed() (backward compatibility)
type NoStayPressedIntent struct{}

func (NoStayPressedIntent) IntentType() string { return "NoStayPressed" }
// Note: Does NOT implement StayPressedIntent

// =============================================================================
// Tests
// =============================================================================

func TestPressableBehavior_StayPressedIntent_Keep(t *testing.T) {
	// Create a mock instance
	inst := newMockInstance()

	// Create PressableBehavior with KeepPressedIntent
	b := &PressableBehavior{
		pressIntent: KeepPressedIntent{},
	}

	// Simulate keyboard Enter action
	act := action.NewAction(action.ActionEnter)

	// Handle action
	handled := b.OnAction(inst, act)

	if !handled {
		t.Fatal("Action should be handled")
	}

	// After emitting StayPressedIntent, pressed state should be TRUE
	if !b.pressed {
		t.Errorf("Expected pressed=true for StayPressedIntent, got false")
	}
	if !inst.state.Pressed {
		t.Errorf("Expected instance state.Pressed=true for StayPressedIntent, got false")
	}
}

func TestPressableBehavior_StayPressedIntent_Reset(t *testing.T) {
	// Create a mock instance
	inst := newMockInstance()

	// Create PressableBehavior with ResetImmediateIntent
	b := &PressableBehavior{
		pressIntent: ResetImmedatelyIntent{},
	}

	// Simulate keyboard Enter action
	act := action.NewAction(action.ActionEnter)

	// Handle action
	handled := b.OnAction(inst, act)

	if !handled {
		t.Fatal("Action should be handled")
	}

	// After emitting ResetImmediateIntent, pressed state should be FALSE
	if b.pressed {
		t.Errorf("Expected pressed=false for ResetImmediateIntent, got true")
	}
	if inst.state.Pressed {
		t.Errorf("Expected instance state.Pressed=false for ResetImmediateIntent, got true")
	}
}

func TestPressableBehavior_NoStayPressedIntent(t *testing.T) {
	// Create a mock instance
	inst := newMockInstance()

	// Create PressableBehavior with NoStayPressedIntent (backward compatibility)
	b := &PressableBehavior{
		pressIntent: NoStayPressedIntent{},
	}

	// Simulate keyboard Enter action
	act := action.NewAction(action.ActionEnter)

	// Handle action
	handled := b.OnAction(inst, act)

	if !handled {
		t.Fatal("Action should be handled")
	}

	// Default behavior: should reset immediately for backward compatibility
	if b.pressed {
		t.Errorf("Expected pressed=false (backward compatibility), got true")
	}
	if inst.state.Pressed {
		t.Errorf("Expected instance state.Pressed=false (backward compatibility), got true")
	}
}

func TestPressableBehavior_NoIntent(t *testing.T) {
	// Create a mock instance
	inst := newMockInstance()

	// Create PressableBehavior with no intent
	b := &PressableBehavior{
		pressIntent: nil,
	}

	// Simulate keyboard Enter action
	act := action.NewAction(action.ActionEnter)

	// Handle action
	handled := b.OnAction(inst, act)

	if !handled {
		t.Fatal("Action should be handled")
	}

	// No intent: should reset immediately
	if b.pressed {
		t.Errorf("Expected pressed=false when no intent, got true")
	}
	if inst.state.Pressed {
		t.Errorf("Expected instance state.Pressed=false when no intent, got true")
	}
}

func TestPressableBehavior_MousePress(t *testing.T) {
	// Mouse events are NOT affected by StayPressedIntent
	// They always wait for ActionMouseRelease

	inst := newMockInstance()

	b := &PressableBehavior{
		pressIntent: KeepPressedIntent{},
	}

	// Simulate mouse press
	act := action.NewAction(action.ActionMousePress)
	handled := b.OnAction(inst, act)

	if !handled {
		t.Fatal("Mouse press should be handled")
	}

	// After mouse press, should stay pressed (wait for release)
	if !b.pressed {
		t.Errorf("Expected pressed=true after mouse press, got false")
	}

	// Simulate mouse release
	actRelease := action.NewAction(action.ActionMouseRelease)
	handled = b.OnAction(inst, actRelease)

	if !handled {
		t.Fatal("Mouse release should be handled")
	}

	// After mouse release, should be false
	if b.pressed {
		t.Errorf("Expected pressed=false after mouse release, got true")
	}
}

func TestPressableBehavior_MousePressWithResetIntent(t *testing.T) {
	// Verify that mouse press ignores StayPressedIntent (stays pressed until release)

	inst := newMockInstance()

	b := &PressableBehavior{
		pressIntent: ResetImmedatelyIntent{}, // Intent says reset immediately
	}

	// Simulate mouse press
	act := action.NewAction(action.ActionMousePress)
	handled := b.OnAction(inst, act)

	if !handled {
		t.Fatal("Mouse press should be handled")
	}

	// Mouse press should stay pressed even with ResetImmediateIntent
	if !b.pressed {
		t.Errorf("Expected pressed=true after mouse press (ignores StayPressed), got false")
	}
}

