package timepicker

import "github.com/wwsheng009/mint/runtime/intent"

// TimeChangeIntent is emitted when the selected time changes.
type TimeChangeIntent struct {
	Value       string
	Hour        int
	Minute      int
	ComponentID string
}

func (i TimeChangeIntent) IntentType() string              { return "timepicker.TimeChangeIntent" }
func (i TimeChangeIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (i TimeChangeIntent) IsTransition() bool              { return false }
func (i TimeChangeIntent) IsGlobal() bool                  { return false }
func (i TimeChangeIntent) GetComponentID() string          { return i.ComponentID }

// TimeChange creates a TimeChangeIntent without component routing metadata.
func TimeChange(value string, hour, minute int) TimeChangeIntent {
	return TimeChangeIntent{
		Value:  value,
		Hour:   hour,
		Minute: minute,
	}
}

// TimeChangeWithID creates a TimeChangeIntent with component routing metadata.
func TimeChangeWithID(componentID, value string, hour, minute int) TimeChangeIntent {
	return TimeChangeIntent{
		Value:       value,
		Hour:        hour,
		Minute:      minute,
		ComponentID: componentID,
	}
}
