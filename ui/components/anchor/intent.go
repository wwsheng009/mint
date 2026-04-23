package anchor

import "github.com/wwsheng009/mint/runtime/intent"

// ActivateIntent requests activating an anchor item by key.
type ActivateIntent struct {
	Key         string
	ComponentID string
}

func (ActivateIntent) IntentType() string              { return "anchor.ActivateIntent" }
func (ActivateIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (ActivateIntent) IsTransition() bool              { return false }
func (ActivateIntent) IsGlobal() bool                  { return false }
func (i ActivateIntent) GetComponentID() string        { return i.ComponentID }

// ChangeIntent is emitted whenever the active anchor changes.
type ChangeIntent struct {
	Key         string
	Href        string
	Title       string
	ComponentID string
}

func (ChangeIntent) IntentType() string              { return "anchor.ChangeIntent" }
func (ChangeIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (ChangeIntent) IsTransition() bool              { return false }
func (ChangeIntent) IsGlobal() bool                  { return false }
func (i ChangeIntent) GetComponentID() string        { return i.ComponentID }

// Activate creates an activate intent without component routing metadata.
func Activate(key string) ActivateIntent {
	return ActivateIntent{Key: key}
}

// ActivateWithID creates an activate intent with component routing metadata.
func ActivateWithID(componentID, key string) ActivateIntent {
	return ActivateIntent{Key: key, ComponentID: componentID}
}

// Change creates a change intent without component routing metadata.
func Change(key, href, title string) ChangeIntent {
	return ChangeIntent{
		Key:   key,
		Href:  href,
		Title: title,
	}
}

// ChangeWithID creates a change intent with component routing metadata.
func ChangeWithID(componentID, key, href, title string) ChangeIntent {
	change := Change(key, href, title)
	change.ComponentID = componentID
	return change
}
