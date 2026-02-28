package control

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// InteractionState Tests
// =============================================================================

func TestInteractionState_IsIdle(t *testing.T) {
	tests := []struct {
		state    InteractionState
		expected bool
	}{
		{InteractionState{}, true},
		{InteractionState{Focused: true}, false},
		{InteractionState{Hovered: true}, false},
		{InteractionState{Pressed: true}, false},
		{InteractionState{Disabled: true}, false},
		{InteractionState{Active: true}, false},
		{InteractionState{Focused: true, Hovered: true}, false},
	}

	for i, tt := range tests {
		if got := tt.state.IsIdle(); got != tt.expected {
			t.Errorf("[%d] IsIdle() = %v, want %v", i, got, tt.expected)
		}
	}
}

func TestInteractionState_Reduce(t *testing.T) {
	tests := []struct {
		action     string
		before     InteractionState
		after      InteractionState
	}{
		{"Focus", InteractionState{}, InteractionState{Focused: true}},
		{"Blur", InteractionState{Focused: true}, InteractionState{}},
		{"MouseEnter", InteractionState{}, InteractionState{Hovered: true}},
		{"MouseLeave", InteractionState{Hovered: true}, InteractionState{}},
		{"PressStart", InteractionState{}, InteractionState{Pressed: true}},
		{"PressEnd", InteractionState{Pressed: true}, InteractionState{}},
		{"Enable", InteractionState{Disabled: true}, InteractionState{}},
		{"Disable", InteractionState{}, InteractionState{Disabled: true}},
		{"Activate", InteractionState{}, InteractionState{Active: true}},
		{"Deactivate", InteractionState{Active: true}, InteractionState{}},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			state := tt.before
			state.Reduce(action.ActionType(tt.action))
			if state != tt.after {
				t.Errorf("Reduce(%q) = %+v, want %+v", tt.action, state, tt.after)
			}
		})
	}
}

func TestInteractionState_Clone(t *testing.T) {
	original := InteractionState{Focused: true, Hovered: false, Pressed: true}
	cloned := original.Clone()

	if cloned != original {
		t.Errorf("Clone() = %+v, want %+v", cloned, original)
	}

	// Modify clone should not affect original
	cloned.Focused = false
	if original.Focused != true {
		t.Error("Modifying clone affected original")
	}
}

// =============================================================================
// Mock Instance for Testing Behaviors
// =============================================================================

type mockInstance struct {
	key        string
	state      InteractionState
	dirty      bool
	emitted    []intent.Intent
	props      map[string]interface{}
}

func newMockInstance() *mockInstance {
	return &mockInstance{
		props: make(map[string]interface{}),
	}
}

func (m *mockInstance) Key() string                                { return m.key }
func (m *mockInstance) GetState() *InteractionState                { return &m.state }
func (m *mockInstance) SetState(s InteractionState)                { m.state = s }
func (m *mockInstance) MarkDirty()                                 { m.dirty = true }
func (m *mockInstance) EmitIntent(i intent.Intent)                 { m.emitted = append(m.emitted, i) }
func (m *mockInstance) GetBounds() (x, y, w, h int)                { return 0, 0, 0, 0 }
func (m *mockInstance) SetBounds(x, y, w, h int)                   {}
func (m *mockInstance) GetStyle() style.Style                      { return style.Style{} }
func (m *mockInstance) SetStyle(s style.Style)                     {}
func (m *mockInstance) GetProp(key string) (interface{}, bool)     { v, ok := m.props[key]; return v, ok }
func (m *mockInstance) SetProp(key string, value interface{})      { m.props[key] = value }

// =============================================================================
// FocusableBehavior Tests
// =============================================================================

func TestFocusableBehavior_Name(t *testing.T) {
	b := &FocusableBehavior{}
	if b.Name() != "Focusable" {
		t.Errorf("Name() = %q, want %q", b.Name(), "Focusable")
	}
}

func TestFocusableBehavior_OnAction(t *testing.T) {
	inst := newMockInstance()
	b := &FocusableBehavior{}

	// Focus action
	act := action.NewAction("Focus")
	if !b.OnAction(inst, act) {
		t.Error("Focus action should return true")
	}
	if !inst.state.Focused {
		t.Error("Focus action should set Focused state")
	}
	if !inst.dirty {
		t.Error("Focus action should mark dirty")
	}

	// Blur action
	inst.dirty = false
	act = action.NewAction("Blur")
	if !b.OnAction(inst, act) {
		t.Error("Blur action should return true")
	}
	if inst.state.Focused {
		t.Error("Blur action should clear Focused state")
	}
}

func TestFocusableBehavior_OnAction_Disabled(t *testing.T) {
	inst := newMockInstance()
	inst.state.Disabled = true
	b := &FocusableBehavior{}

	act := action.NewAction("Focus")
	if b.OnAction(inst, act) {
		t.Error("Focus action on disabled should return false")
	}
}

// =============================================================================
// PressableBehavior Tests
// =============================================================================

func TestPressableBehavior_Name(t *testing.T) {
	b := &PressableBehavior{}
	if b.Name() != "Pressable" {
		t.Errorf("Name() = %q, want %q", b.Name(), "Pressable")
	}
}

func TestPressableBehavior_OnAction(t *testing.T) {
	inst := newMockInstance()
	testIntent := intent.Click("test")
	b := NewPressableBehavior(testIntent)

	// Press action
	act := action.NewAction(action.ActionPress)
	if !b.OnAction(inst, act) {
		t.Error("Press action should return true")
	}
	if !inst.state.Pressed {
		t.Error("Press action should set Pressed state")
	}
	// Intent is emitted on press, not release
	if len(inst.emitted) != 1 {
		t.Errorf("Press should emit intent, got %d emitted", len(inst.emitted))
	}

	// Release action
	inst.dirty = false
	act = action.NewAction(action.ActionRelease)
	if !b.OnAction(inst, act) {
		t.Error("Release action should return true")
	}
	if inst.state.Pressed {
		t.Error("Release action should clear Pressed state")
	}
	// Intent is already emitted on press, no new intent on release
	if len(inst.emitted) != 1 {
		t.Errorf("No additional intent should be emitted on release, got %d emitted", len(inst.emitted))
	}
}

func TestPressableBehavior_OnAction_Disabled(t *testing.T) {
	inst := newMockInstance()
	inst.state.Disabled = true
	b := NewPressableBehavior(nil)

	act := action.NewAction("Press")
	if b.OnAction(inst, act) {
		t.Error("Press action on disabled should return false")
	}
}

// =============================================================================
// HoverableBehavior Tests
// =============================================================================

func TestHoverableBehavior_Name(t *testing.T) {
	b := &HoverableBehavior{}
	if b.Name() != "Hoverable" {
		t.Errorf("Name() = %q, want %q", b.Name(), "Hoverable")
	}
}

func TestHoverableBehavior_OnAction(t *testing.T) {
	inst := newMockInstance()
	b := &HoverableBehavior{}

	// MouseEnter action
	act := action.NewAction("MouseEnter")
	if !b.OnAction(inst, act) {
		t.Error("MouseEnter action should return true")
	}
	if !inst.state.Hovered {
		t.Error("MouseEnter should set Hovered state")
	}

	// MouseLeave action
	inst.dirty = false
	act = action.NewAction("MouseLeave")
	if !b.OnAction(inst, act) {
		t.Error("MouseLeave action should return true")
	}
	if inst.state.Hovered {
		t.Error("MouseLeave should clear Hovered state")
	}
}

// =============================================================================
// DisableableBehavior Tests
// =============================================================================

func TestDisableableBehavior_Name(t *testing.T) {
	b := &DisableableBehavior{}
	if b.Name() != "Disableable" {
		t.Errorf("Name() = %q, want %q", b.Name(), "Disableable")
	}
}

func TestDisableableBehavior_OnMount(t *testing.T) {
	inst := newMockInstance()
	inst.props["disabled"] = true
	b := &DisableableBehavior{}

	b.OnMount(inst)

	if !b.IsDisabled() {
		t.Error("OnMount should read disabled prop")
	}
	if !inst.state.Disabled {
		t.Error("OnMount should set state.Disabled")
	}
}

func TestDisableableBehavior_OnAction(t *testing.T) {
	inst := newMockInstance()
	b := &DisableableBehavior{}

	// Disable action
	act := action.NewAction("Disable")
	if !b.OnAction(inst, act) {
		t.Error("Disable action should return true")
	}
	if !inst.state.Disabled {
		t.Error("Disable action should set Disabled state")
	}

	// Enable action
	inst.dirty = false
	act = action.NewAction("Enable")
	if !b.OnAction(inst, act) {
		t.Error("Enable action should return true")
	}
	if inst.state.Disabled {
		t.Error("Enable action should clear Disabled state")
	}
}

// =============================================================================
// BehaviorList Tests
// =============================================================================

func TestBehaviorList(t *testing.T) {
	bl := NewBehaviorList(
		&FocusableBehavior{},
		&PressableBehavior{},
	)

	if len(bl.List()) != 2 {
		t.Errorf("List() length = %d, want 2", len(bl.List()))
	}

	if bl.Get("Focusable") == nil {
		t.Error("Get(Focusable) should not be nil")
	}

	if bl.Get("NonExistent") != nil {
		t.Error("Get(NonExistent) should be nil")
	}
}

func TestBehaviorList_OnAction(t *testing.T) {
	inst := newMockInstance()
	bl := NewBehaviorList(
		&FocusableBehavior{},
		&PressableBehavior{},
	)

	// First behavior should handle Focus
	act := action.NewAction("Focus")
	if !bl.OnAction(inst, act) {
		t.Error("BehaviorList should handle Focus")
	}

	// Second behavior should handle Press
	act = action.NewAction("Press")
	if !bl.OnAction(inst, act) {
		t.Error("BehaviorList should handle Press")
	}

	// Unknown action should return false
	act = action.NewAction("Unknown")
	if bl.OnAction(inst, act) {
		t.Error("BehaviorList should not handle Unknown")
	}
}

func TestBehaviorList_OnMountUnmount(t *testing.T) {
	inst := newMockInstance()
	bl := NewBehaviorList(
		&FocusableBehavior{},
	)

	// Should not panic
	bl.OnMount(inst)
	bl.OnUnmount(inst)
}

func TestBehaviorList_Add(t *testing.T) {
	bl := NewBehaviorList()
	bl.Add(&FocusableBehavior{})

	if len(bl.List()) != 1 {
		t.Errorf("Add() should add behavior, got %d", len(bl.List()))
	}
}
