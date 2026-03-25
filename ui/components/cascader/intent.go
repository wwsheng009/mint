package cascader

import "github.com/wwsheng009/mint/runtime/intent"

// ChangeIntent is emitted whenever the committed cascader path changes.
type ChangeIntent struct {
	Value       string
	Label       string
	Values      []string
	Labels      []string
	ComponentID string
}

func (ChangeIntent) IntentType() string              { return "cascader.ChangeIntent" }
func (ChangeIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (ChangeIntent) IsTransition() bool              { return false }
func (ChangeIntent) IsGlobal() bool                  { return false }
func (i ChangeIntent) GetComponentID() string        { return i.ComponentID }

// Change creates a change intent without component routing metadata.
func Change(value, label string, values, labels []string) ChangeIntent {
	return ChangeIntent{
		Value:  value,
		Label:  label,
		Values: append([]string(nil), values...),
		Labels: append([]string(nil), labels...),
	}
}

// ChangeWithID creates a change intent with component routing metadata.
func ChangeWithID(componentID, value, label string, values, labels []string) ChangeIntent {
	change := Change(value, label, values, labels)
	change.ComponentID = componentID
	return change
}
