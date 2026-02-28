package control

import (
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// StayPressedIntent - Intent Visual Feedback Control
// =============================================================================

// StayPressedIntent is an interface that intents can implement to control
// whether the pressed visual state should be maintained after the intent is emitted.
//
// When a user triggers a button via keyboard (Enter/Submit), the terminal UI
// doesn't send key release events. This interface allows authors to specify:
//
//   - StayPressed() == true:  Keep the pressed state so用户 can see visual feedback
//     Useful for navigation, state updates, etc.
//   - StayPressed() == false: Immediately reset pressed state
//     Useful for Quit, Delete, etc.
//
// Example:
//
//	type UpdateStepIntent struct { Step int }
//
//	func (UpdateStepIntent) IntentType() string { return "UpdateStep" }
//	func (UpdateStepIntent) StayPressed() bool  { return true }  // 保持按下视觉反馈
type StayPressedIntent interface {
	intent.Intent

	// StayPressed returns true if the pressed visual state should be maintained
	// after emitting this intent (e.g., for navigation/state changes).
	// Returns false for immediate reset (e.g., for destructive actions).
	StayPressed() bool
}

// =============================================================================
// InteractionState - Unified Interaction State
// =============================================================================

// InteractionState represents the unified interaction state for all controls.
// All interactive components share this state definition.
type InteractionState struct {
	Focused  bool
	Hovered  bool
	Pressed  bool
	Disabled bool
	Active   bool
}

// IsIdle returns true if no interaction state is active.
func (s *InteractionState) IsIdle() bool {
	return !s.Focused && !s.Hovered && !s.Pressed && !s.Disabled && !s.Active
}

// Reduce applies an action type to update the state.
func (s *InteractionState) Reduce(actionType action.ActionType) {
	switch actionType {
	case action.ActionFocus:
		s.Focused = true
	case action.ActionBlur:
		s.Focused = false
	case action.ActionMouseEnter:
		s.Hovered = true
	case action.ActionMouseLeave:
		s.Hovered = false
	case action.ActionPressStart:
		s.Pressed = true
	case action.ActionPressEnd:
		s.Pressed = false
	case action.ActionEnable:
		s.Disabled = false
	case action.ActionDisable:
		s.Disabled = true
	case action.ActionActivate:
		s.Active = true
	case action.ActionDeactivate:
		s.Active = false
	}
}

// ReduceByAction applies an action to update the state.
func (s *InteractionState) ReduceByAction(act *action.Action) {
	switch act.Type {
	case action.ActionFocus:
		s.Focused = true
	case action.ActionBlur:
		s.Focused = false
	case action.ActionMouseEnter:
		s.Hovered = true
	case action.ActionMouseLeave:
		s.Hovered = false
	case action.ActionPressStart:
		s.Pressed = true
	case action.ActionPressEnd:
		s.Pressed = false
	case action.ActionEnable:
		s.Disabled = false
	case action.ActionDisable:
		s.Disabled = true
	case action.ActionActivate:
		s.Active = true
	case action.ActionDeactivate:
		s.Active = false
	}
}

// Clone creates a copy of the state.
func (s *InteractionState) Clone() InteractionState {
	return InteractionState{
		Focused:  s.Focused,
		Hovered:  s.Hovered,
		Pressed:  s.Pressed,
		Disabled: s.Disabled,
		Active:   s.Active,
	}
}

// =============================================================================
// Behavior Interface - Composable Interaction Logic
// =============================================================================

// Behavior is the interface for composable interaction behaviors.
// Components compose behaviors instead of implementing all logic directly.
type Behavior interface {
	// Name returns the behavior name for debugging.
	Name() string

	// OnMount is called when the instance is mounted.
	OnMount(inst Instance)

	// OnUnmount is called when the instance is unmounted.
	OnUnmount(inst Instance)

	// OnAction handles an action. Returns true if the action was consumed.
	OnAction(inst Instance, act *action.Action) bool

	// OnStateChange is called when interaction state changes.
	OnStateChange(inst Instance, oldState, newState InteractionState)
}

// Instance is the interface that behaviors use to interact with the component.
type Instance interface {
	// Identification
	Key() string

	// State
	GetState() *InteractionState
	SetState(state InteractionState)
	MarkDirty()

	// Intent
	EmitIntent(intent intent.Intent)

	// Layout
	GetBounds() (x, y, w, h int)
	SetBounds(x, y, w, h int)

	// Style
	GetStyle() style.Style
	SetStyle(s style.Style)

	// Props
	GetProp(key string) (interface{}, bool)
	SetProp(key string, value interface{})
}

// =============================================================================
// FocusableBehavior - Focus Handling
// =============================================================================

// FocusableBehavior handles focus/blur interactions.
type FocusableBehavior struct {
	focused bool
}

// Name returns the behavior name.
func (b *FocusableBehavior) Name() string {
	return "Focusable"
}

// OnMount is called on mount.
func (b *FocusableBehavior) OnMount(inst Instance) {
	// No-op
}

// OnUnmount is called on unmount.
func (b *FocusableBehavior) OnUnmount(inst Instance) {
	// No-op
}

// OnAction handles focus/blur actions.
func (b *FocusableBehavior) OnAction(inst Instance, act *action.Action) bool {
	state := inst.GetState()
	if state.Disabled {
		return false
	}

	switch act.Type {
	case action.ActionFocus:
		if !b.focused {
			b.focused = true
			state.Focused = true
			inst.MarkDirty()
			inst.EmitIntent(intent.Focus(inst.Key()))
			return true
		}
	case action.ActionBlur:
		if b.focused {
			b.focused = false
			state.Focused = false
			inst.MarkDirty()
			inst.EmitIntent(intent.Blur(inst.Key()))
			return true
		}
	}
	return false
}

// OnStateChange handles state changes.
func (b *FocusableBehavior) OnStateChange(inst Instance, oldState, newState InteractionState) {
	b.focused = newState.Focused
}

// IsFocused returns the focused state.
func (b *FocusableBehavior) IsFocused() bool {
	return b.focused
}

// =============================================================================
// PressableBehavior - Press/Click Handling
// =============================================================================

// PressableBehavior handles press/click interactions.
type PressableBehavior struct {
	pressed     bool
	pressIntent intent.Intent
}

// NewPressableBehavior creates a new PressableBehavior with an intent.
func NewPressableBehavior(pressIntent intent.Intent) *PressableBehavior {
	return &PressableBehavior{
		pressIntent: pressIntent,
	}
}

// Name returns the behavior name.
func (b *PressableBehavior) Name() string {
	return "Pressable"
}

// OnMount is called on mount.
func (b *PressableBehavior) OnMount(inst Instance) {
	// No-op
}

// OnUnmount is called on unmount.
func (b *PressableBehavior) OnUnmount(inst Instance) {
	// No-op
}

// OnAction handles press actions.
func (b *PressableBehavior) OnAction(inst Instance, act *action.Action) bool {
	state := inst.GetState()
	if state.Disabled {
		return false
	}

	switch act.Type {
	case action.ActionPress, action.ActionClick, action.ActionEnter,
		action.ActionSubmit, action.ActionMousePress:
		// For terminal UI, button press triggers intent immediately
		// Only handle if not already pressed (prevent double-trigger)
		if !b.pressed {
			b.pressed = true
			state.Pressed = true

			// Emit intent on press
			if b.pressIntent != nil {
				inst.EmitIntent(b.pressIntent)

				// Check if intent implements StayPressedIntent
				// If StayPressed() returns false, reset immediately
				if stayPressed, ok := b.pressIntent.(StayPressedIntent); ok && !stayPressed.StayPressed() {
					if act.Type == action.ActionEnter || act.Type == action.ActionSubmit {
						b.pressed = false
						state.Pressed = false
					}
				}

				// If intent does NOT implement StayPressedIntent, reset immediately (for backward compatibility)
				if _, ok := b.pressIntent.(StayPressedIntent); !ok {
					if act.Type == action.ActionEnter || act.Type == action.ActionSubmit {
						b.pressed = false
						state.Pressed = false
					}
				}
			} else {
				// No intent: always reset pressed state immediately
				if act.Type == action.ActionEnter || act.Type == action.ActionSubmit {
					b.pressed = false
					state.Pressed = false
				}
			}

			// Mark dirty once at the end after all state changes
			inst.MarkDirty()
		}
		return true

	case action.ActionRelease, action.ActionPressEnd, action.ActionMouseRelease:
		// Handle release actions (mouse only, keyboard has no release events)
		if b.pressed {
			b.pressed = false
			state.Pressed = false
			inst.MarkDirty()
		}
		return true
	}
	return false
}

// OnStateChange handles state changes.
//
// Note: Pressed state management now follows the design from
// docs/event/PRESSED_STATE_COMPLETE_SOLUTION.md:
//
// - OnAction manages press events (Enter/Submit/MousePress)
// - OnAction manages release events (MouseRelease only)
// - ResetPressed() is called by InteractionContext when new keyboard input is detected
//
// This approach works for both mouse and keyboard interactions, even in single-focus UIs.
func (b *PressableBehavior) OnStateChange(inst Instance, oldState, newState InteractionState) {
	// No-op: pressed state is managed by OnAction and ResetPressed() only
	// Do NOT sync from newState.Pressed as it's managed externally
}

// ResetPressed implements the PressedResetHandler interface.
//
// Called by InteractionContext when new keyboard input is detected.
// This ensures pressed state resets correctly even in single-focus UIs.
//
// Note: This method doesn't have access to Instance, so it only resets the
// internal pressed flag. The component wrapper should call ResetPressedWithInstance
// to also update the InteractionState.Pressed.
func (b *PressableBehavior) ResetPressed() {
	if b.pressed {
		b.pressed = false
	}
}

// ResetPressedWithInstance resets pressed state and also updates the InteractionState.
//
// This is the preferred method when the component has access to its Instance.
func (b *PressableBehavior) ResetPressedWithInstance(inst Instance) {
	if b.pressed {
		b.pressed = false
		state := inst.GetState()
		state.Pressed = false
		inst.MarkDirty()
	}
}

// IsPressed returns the pressed state.
func (b *PressableBehavior) IsPressed() bool {
	return b.pressed
}

// SetIntent sets the press intent.
func (b *PressableBehavior) SetIntent(pressIntent intent.Intent) {
	b.pressIntent = pressIntent
}

// =============================================================================
// HoverableBehavior - Hover Handling
// =============================================================================

// HoverableBehavior handles mouse hover interactions.
type HoverableBehavior struct {
	hovered bool
}

// Name returns the behavior name.
func (b *HoverableBehavior) Name() string {
	return "Hoverable"
}

// OnMount is called on mount.
func (b *HoverableBehavior) OnMount(inst Instance) {
	// No-op
}

// OnUnmount is called on unmount.
func (b *HoverableBehavior) OnUnmount(inst Instance) {
	// No-op
}

// OnAction handles hover actions.
func (b *HoverableBehavior) OnAction(inst Instance, act *action.Action) bool {
	state := inst.GetState()

	switch act.Type {
	case action.ActionMouseEnter:
		if !b.hovered {
			b.hovered = true
			state.Hovered = true
			inst.MarkDirty()
			return true
		}
	case action.ActionMouseLeave:
		if b.hovered {
			b.hovered = false
			state.Hovered = false
			inst.MarkDirty()
			return true
		}
	}
	return false
}

// OnStateChange handles state changes.
func (b *HoverableBehavior) OnStateChange(inst Instance, oldState, newState InteractionState) {
	b.hovered = newState.Hovered
}

// IsHovered returns the hovered state.
func (b *HoverableBehavior) IsHovered() bool {
	return b.hovered
}

// =============================================================================
// DisableableBehavior - Disabled State Handling
// =============================================================================

// DisableableBehavior handles disabled state.
type DisableableBehavior struct {
	disabled bool
}

// Name returns the behavior name.
func (b *DisableableBehavior) Name() string {
	return "Disableable"
}

// OnMount is called on mount.
func (b *DisableableBehavior) OnMount(inst Instance) {
	// Check initial disabled prop
	if v, ok := inst.GetProp("disabled"); ok {
		if disabled, ok := v.(bool); ok {
			b.disabled = disabled
			inst.GetState().Disabled = disabled
		}
	}
}

// OnUnmount is called on unmount.
func (b *DisableableBehavior) OnUnmount(inst Instance) {
	// No-op
}

// OnAction handles disable/enable actions.
func (b *DisableableBehavior) OnAction(inst Instance, act *action.Action) bool {
	switch act.Type {
	case action.ActionDisable:
		if !b.disabled {
			b.disabled = true
			inst.GetState().Disabled = true
			inst.MarkDirty()
			return true
		}
	case action.ActionEnable:
		if b.disabled {
			b.disabled = false
			inst.GetState().Disabled = false
			inst.MarkDirty()
			return true
		}
	}
	return false
}

// OnStateChange handles state changes.
func (b *DisableableBehavior) OnStateChange(inst Instance, oldState, newState InteractionState) {
	b.disabled = newState.Disabled
}

// IsDisabled returns the disabled state.
func (b *DisableableBehavior) IsDisabled() bool {
	return b.disabled
}

// =============================================================================
// BehaviorList - Composable Behavior Container
// =============================================================================

// BehaviorList manages a list of behaviors.
type BehaviorList struct {
	behaviors []Behavior
}

// NewBehaviorList creates a new behavior list.
func NewBehaviorList(behaviors ...Behavior) *BehaviorList {
	return &BehaviorList{behaviors: behaviors}
}

// Add adds a behavior to the list.
func (bl *BehaviorList) Add(b Behavior) {
	bl.behaviors = append(bl.behaviors, b)
}

// OnMount calls OnMount on all behaviors.
func (bl *BehaviorList) OnMount(inst Instance) {
	for _, b := range bl.behaviors {
		b.OnMount(inst)
	}
}

// OnUnmount calls OnUnmount on all behaviors.
func (bl *BehaviorList) OnUnmount(inst Instance) {
	for _, b := range bl.behaviors {
		b.OnUnmount(inst)
	}
}

// OnAction dispatches action to behaviors until one consumes it.
func (bl *BehaviorList) OnAction(inst Instance, act *action.Action) bool {
	for _, b := range bl.behaviors {
		if b.OnAction(inst, act) {
			return true
		}
	}
	return false
}

// OnStateChange notifies all behaviors of state change.
func (bl *BehaviorList) OnStateChange(inst Instance, oldState, newState InteractionState) {
	for _, b := range bl.behaviors {
		b.OnStateChange(inst, oldState, newState)
	}
}

// Get returns a behavior by name.
func (bl *BehaviorList) Get(name string) Behavior {
	for _, b := range bl.behaviors {
		if b.Name() == name {
			return b
		}
	}
	return nil
}

// List returns all behaviors.
func (bl *BehaviorList) List() []Behavior {
	return bl.behaviors
}

// =============================================================================
// Style Resolution Helper
// =============================================================================

// ResolveStyle resolves the visual style based on interaction state.
func ResolveStyle(baseStyle style.Style, state InteractionState) style.Style {
	s := baseStyle

	// Priority: Disabled > Pressed > Focused > Hovered > Normal
	if state.Disabled {
		// Disabled overrides everything
		return s // Caller should apply disabled theme
	}

	if state.Pressed {
		// Pressed state
		return s // Caller may apply pressed style
	}

	if state.Focused {
		// Focus state
		return s // Caller should apply focus style
	}

	if state.Hovered {
		// Hover state
		return s // Caller may apply hover style
	}

	return s
}

// =============================================================================
// Paint Context Helper
// =============================================================================

// PaintContext provides context for painting controls.
type PaintContext struct {
	X, Y   int
	Width  int
	Height int
}

// Paintable is the interface for paintable controls.
type Paintable interface {
	Paint(ctx PaintContext) []paint.DrawCmd
}
