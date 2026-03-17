package steps

import "github.com/wwsheng009/mint/runtime/intent"

// StepChangeIntent is emitted when the active step changes.
type StepChangeIntent struct {
	ComponentID string
	FromIndex   int
	ToIndex     int
	StepCount   int
	StepKey     string
	StepTitle   string
}

func (StepChangeIntent) IntentType() string {
	return "steps.StepChangeIntent"
}

func (StepChangeIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (StepChangeIntent) IsTransition() bool {
	return false
}

func (StepChangeIntent) IsGlobal() bool {
	return false
}

func (i StepChangeIntent) GetComponentID() string {
	return i.ComponentID
}

// StepChange creates a step change intent.
func StepChange(fromIndex, toIndex, stepCount int, stepKey, stepTitle string) StepChangeIntent {
	return StepChangeIntent{
		FromIndex: fromIndex,
		ToIndex:   toIndex,
		StepCount: stepCount,
		StepKey:   stepKey,
		StepTitle: stepTitle,
	}
}

// StepChangeWithID creates a step change intent with component ID.
func StepChangeWithID(componentID string, fromIndex, toIndex, stepCount int, stepKey, stepTitle string) StepChangeIntent {
	change := StepChange(fromIndex, toIndex, stepCount, stepKey, stepTitle)
	change.ComponentID = componentID
	return change
}
