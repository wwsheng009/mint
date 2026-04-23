package datepicker

import "github.com/wwsheng009/mint/runtime/intent"

// DateChangeIntent is emitted when the selected date changes.
type DateChangeIntent struct {
	Value       string
	Year        int
	Month       int
	Day         int
	ComponentID string
}

func (i DateChangeIntent) IntentType() string              { return "datepicker.DateChangeIntent" }
func (i DateChangeIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (i DateChangeIntent) IsTransition() bool              { return false }
func (i DateChangeIntent) IsGlobal() bool                  { return false }
func (i DateChangeIntent) GetComponentID() string          { return i.ComponentID }

// DateChange creates a DateChangeIntent without component routing metadata.
func DateChange(value string, year, month, day int) DateChangeIntent {
	return DateChangeIntent{
		Value: value,
		Year:  year,
		Month: month,
		Day:   day,
	}
}

// DateChangeWithID creates a DateChangeIntent with component routing metadata.
func DateChangeWithID(componentID, value string, year, month, day int) DateChangeIntent {
	return DateChangeIntent{
		Value:       value,
		Year:        year,
		Month:       month,
		Day:         day,
		ComponentID: componentID,
	}
}
