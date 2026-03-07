package selectcomp

import (
	"github.com/wwsheng009/mint/runtime/intent"
)

// =============================================================================
// Select Intents (Phase 10: Intent Bubble)
// =============================================================================

// SelectChangeIntent is emitted when the selected option changes
// This is similar to OptionGroup's OptionSelectIntent but for Select component
type SelectChangeIntent struct {
	// SelectedIndex is the newly selected option index
	SelectedIndex int

	// SelectedValue is the value of the selected option
	SelectedValue string

	// SelectedLabel is the label of the selected option
	SelectedLabel string

	// ComponentID is the select component ID for routing (optional)
	ComponentID string
}

// IntentType implements intent.Intent
func (i SelectChangeIntent) IntentType() string {
	return "select.SelectChangeIntent"
}

// Priority implements PriorityAware
func (i SelectChangeIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition returns false (synchronous intent)
func (i SelectChangeIntent) IsTransition() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing
func (i SelectChangeIntent) GetComponentID() string {
	return i.ComponentID
}

// SelectNextIntent requests to select the next option
type SelectNextIntent struct {
	ComponentID string // Optional component ID for routing
}

// IntentType implements intent.Intent
func (i SelectNextIntent) IntentType() string {
	return "select.SelectNextIntent"
}

// Priority implements PriorityAware
func (i SelectNextIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition returns false (synchronous intent)
func (i SelectNextIntent) IsTransition() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing
func (i SelectNextIntent) GetComponentID() string {
	return i.ComponentID
}

// SelectPrevIntent requests to select the previous option
type SelectPrevIntent struct {
	ComponentID string // Optional component ID for routing
}

// IntentType implements intent.Intent
func (i SelectPrevIntent) IntentType() string {
	return "select.SelectPrevIntent"
}

// Priority implements PriorityAware
func (i SelectPrevIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition returns false (synchronous intent)
func (i SelectPrevIntent) IsTransition() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing
func (i SelectPrevIntent) GetComponentID() string {
	return i.ComponentID
}

// SelectByIndexIntent requests to select an option by index
type SelectByIndexIntent struct {
	// Index is the option index to select
	Index int

	// ComponentID is the select component ID for routing (optional)
	ComponentID string
}

// IntentType implements intent.Intent
func (i SelectByIndexIntent) IntentType() string {
	return "select.SelectByIndexIntent"
}

// Priority implements PriorityAware
func (i SelectByIndexIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition returns false (synchronous intent)
func (i SelectByIndexIntent) IsTransition() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing
func (i SelectByIndexIntent) GetComponentID() string {
	return i.ComponentID
}

// SelectByValueIntent requests to select an option by value
type SelectByValueIntent struct {
	// Value is the option value to select
	Value string

	// ComponentID is the select component ID for routing (optional)
	ComponentID string
}

// IntentType implements intent.Intent
func (i SelectByValueIntent) IntentType() string {
	return "select.SelectByValueIntent"
}

// Priority implements PriorityAware
func (i SelectByValueIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition returns false (synchronous intent)
func (i SelectByValueIntent) IsTransition() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing
func (i SelectByValueIntent) GetComponentID() string {
	return i.ComponentID
}

// =============================================================================
// Intent Constructors
// =============================================================================

// SelectChange creates a SelectChangeIntent
func SelectChange(selectedIndex int, selectedValue, selectedLabel string) SelectChangeIntent {
	return SelectChangeIntent{
		SelectedIndex:  selectedIndex,
		SelectedValue:  selectedValue,
		SelectedLabel:  selectedLabel,
	}
}

// SelectChangeWithID creates a SelectChangeIntent with component ID
func SelectChangeWithID(componentID string, selectedIndex int, selectedValue, selectedLabel string) SelectChangeIntent {
	return SelectChangeIntent{
		ComponentID:    componentID,
		SelectedIndex:  selectedIndex,
		SelectedValue:  selectedValue,
		SelectedLabel:  selectedLabel,
	}
}

// SelectNext creates a SelectNextIntent
func NewSelectNextIntent() SelectNextIntent {
	return SelectNextIntent{}
}

// SelectNextWithID creates a SelectNextIntent with component ID
func NewSelectNextIntentWithID(componentID string) SelectNextIntent {
	return SelectNextIntent{
		ComponentID: componentID,
	}
}

// SelectPrev creates a SelectPrevIntent
func NewSelectPrevIntent() SelectPrevIntent {
	return SelectPrevIntent{}
}

// SelectPrevWithID creates a SelectPrevIntent with component ID
func NewSelectPrevIntentWithID(componentID string) SelectPrevIntent {
	return SelectPrevIntent{
		ComponentID: componentID,
	}
}

// SelectByIndex creates a SelectByIndexIntent
func SelectByIndex(index int) SelectByIndexIntent {
	return SelectByIndexIntent{
		Index: index,
	}
}

// SelectByIndexWithID creates a SelectByIndexIntent with component ID
func SelectByIndexWithID(componentID string, index int) SelectByIndexIntent {
	return SelectByIndexIntent{
		ComponentID: componentID,
		Index:       index,
	}
}

// SelectByValue creates a SelectByValueIntent
func SelectByValue(value string) SelectByValueIntent {
	return SelectByValueIntent{
		Value: value,
	}
}

// SelectByValueWithID creates a SelectByValueIntent with component ID
func SelectByValueWithID(componentID, value string) SelectByValueIntent {
	return SelectByValueIntent{
		ComponentID: componentID,
		Value:       value,
	}
}
