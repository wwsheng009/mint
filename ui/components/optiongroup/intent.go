package optiongroup

import (
	"github.com/wwsheng009/mint/runtime/intent"
)

// =============================================================================
// OptionSelectIntent - Local Intent Bubble for OptionGroup
// =============================================================================//
// This intent is used for local communication between Option and OptionGroup
// via the instance tree intent bubble system (Phase 3).
//
// Usage:
//   - Option emits OptionSelectIntent when clicked/selected
//   - Intent bubbles up through Parent() references
//   - OptionGroup.HandleIntent() intercepts the intent
//   - OptionGroup updates its state accordingly
// =============================================================================//

// OptionSelectIntent is emitted when an option is selected.
// This intent bubbles up the instance tree to be handled by OptionGroup.
type OptionSelectIntent struct {
	// GroupKey is the key of the OptionGroup that should handle this intent.
	// This ensures that OptionGroup only handles intents for its own options.
	GroupKey string

	// Value is the value of the option being selected.
	Value string

	// IsSelected indicates whether the option is being selected (true) or
	// deselected (false in multi-select mode).
	IsSelected bool

	// Mode is the select mode of the OptionGroup (Single or Multiple).
	Mode SelectMode
}

// IntentType implements the intent.Intent interface.
// Returns a unique type identifier for this intent.
func (OptionSelectIntent) IntentType() string {
	return "OptionGroup:Select"
}

// Type is an alias for IntentType for compatibility with existing code.
// This matches the naming convention in the global intent system.
func (i OptionSelectIntent) Type() string {
	return i.IntentType()
}

// Priority returns the action priority for this intent.
// Option selection is a user-blocking action (should respond quickly).
func (i OptionSelectIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

// IsTransition indicates this is NOT an async operation.
func (i OptionSelectIntent) IsTransition() bool {
	return false
}

// OptionGroupDeselectIntent is emitted when an option is deselected
// (only relevant in multi-select mode).
type OptionGroupDeselectIntent struct {
	GroupKey string
	Value    string
	Mode     SelectMode
}

// IntentType implements the intent.Intent interface.
func (OptionGroupDeselectIntent) IntentType() string {
	return "OptionGroup:Deselect"
}

// Type is an alias for IntentType.
func (i OptionGroupDeselectIntent) Type() string {
	return i.IntentType()
}

// Priority returns the action priority.
func (i OptionGroupDeselectIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

// IsTransition indicates this is NOT an async operation.
func (i OptionGroupDeselectIntent) IsTransition() bool {
	return false
}
