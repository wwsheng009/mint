package popover

import "github.com/wwsheng009/mint/runtime/intent"

// PopoverToggleIntent toggles popover visibility.
type PopoverToggleIntent struct {
	ComponentID string
}

func (PopoverToggleIntent) IntentType() string { return "popover.PopoverToggleIntent" }

func (PopoverToggleIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }

func (PopoverToggleIntent) IsTransition() bool { return false }

func (PopoverToggleIntent) IsGlobal() bool { return false }

func (PopoverToggleIntent) StayPressed() bool { return true }

func (i PopoverToggleIntent) GetComponentID() string { return i.ComponentID }

// PopoverOpenIntent opens the popover.
type PopoverOpenIntent struct {
	ComponentID string
}

func (PopoverOpenIntent) IntentType() string { return "popover.PopoverOpenIntent" }

func (PopoverOpenIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }

func (PopoverOpenIntent) IsTransition() bool { return false }

func (PopoverOpenIntent) IsGlobal() bool { return false }

func (PopoverOpenIntent) StayPressed() bool { return true }

func (i PopoverOpenIntent) GetComponentID() string { return i.ComponentID }

// PopoverCloseIntent closes the popover.
type PopoverCloseIntent struct {
	ComponentID string
}

func (PopoverCloseIntent) IntentType() string { return "popover.PopoverCloseIntent" }

func (PopoverCloseIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }

func (PopoverCloseIntent) IsTransition() bool { return false }

func (PopoverCloseIntent) IsGlobal() bool { return false }

func (PopoverCloseIntent) StayPressed() bool { return true }

func (i PopoverCloseIntent) GetComponentID() string { return i.ComponentID }

// PopoverChangeIntent is emitted whenever open state changes.
type PopoverChangeIntent struct {
	ComponentID string
	Open        bool
	Trigger     TriggerMode
}

func (PopoverChangeIntent) IntentType() string { return "popover.PopoverChangeIntent" }

func (PopoverChangeIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }

func (PopoverChangeIntent) IsTransition() bool { return false }

func (PopoverChangeIntent) IsGlobal() bool { return false }

func (i PopoverChangeIntent) GetComponentID() string { return i.ComponentID }

// Toggle creates a component-agnostic toggle intent.
func Toggle() PopoverToggleIntent { return PopoverToggleIntent{} }

// ToggleWithID creates a component-scoped toggle intent.
func ToggleWithID(componentID string) PopoverToggleIntent {
	return PopoverToggleIntent{ComponentID: componentID}
}

// Open creates a component-agnostic open intent.
func Open() PopoverOpenIntent { return PopoverOpenIntent{} }

// OpenWithID creates a component-scoped open intent.
func OpenWithID(componentID string) PopoverOpenIntent {
	return PopoverOpenIntent{ComponentID: componentID}
}

// Close creates a component-agnostic close intent.
func Close() PopoverCloseIntent { return PopoverCloseIntent{} }

// CloseWithID creates a component-scoped close intent.
func CloseWithID(componentID string) PopoverCloseIntent {
	return PopoverCloseIntent{ComponentID: componentID}
}
