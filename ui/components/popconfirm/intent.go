package popconfirm

import "github.com/wwsheng009/mint/runtime/intent"

// PopconfirmToggleIntent toggles overlay visibility locally.
type PopconfirmToggleIntent struct {
	ComponentID string
}

func (PopconfirmToggleIntent) IntentType() string              { return "popconfirm.PopconfirmToggleIntent" }
func (PopconfirmToggleIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (PopconfirmToggleIntent) IsTransition() bool              { return false }
func (PopconfirmToggleIntent) IsGlobal() bool                  { return false }
func (PopconfirmToggleIntent) StayPressed() bool               { return true }
func (i PopconfirmToggleIntent) GetComponentID() string        { return i.ComponentID }

// PopconfirmOpenIntent opens the overlay locally.
type PopconfirmOpenIntent struct {
	ComponentID string
}

func (PopconfirmOpenIntent) IntentType() string              { return "popconfirm.PopconfirmOpenIntent" }
func (PopconfirmOpenIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (PopconfirmOpenIntent) IsTransition() bool              { return false }
func (PopconfirmOpenIntent) IsGlobal() bool                  { return false }
func (PopconfirmOpenIntent) StayPressed() bool               { return true }
func (i PopconfirmOpenIntent) GetComponentID() string        { return i.ComponentID }

// PopconfirmCloseIntent closes the overlay locally.
type PopconfirmCloseIntent struct {
	ComponentID string
}

func (PopconfirmCloseIntent) IntentType() string              { return "popconfirm.PopconfirmCloseIntent" }
func (PopconfirmCloseIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (PopconfirmCloseIntent) IsTransition() bool              { return false }
func (PopconfirmCloseIntent) IsGlobal() bool                  { return false }
func (PopconfirmCloseIntent) StayPressed() bool               { return true }
func (i PopconfirmCloseIntent) GetComponentID() string        { return i.ComponentID }

// PopconfirmChangeIntent is emitted when open state changes.
type PopconfirmChangeIntent struct {
	ComponentID string
	Open        bool
	Trigger     TriggerMode
}

func (PopconfirmChangeIntent) IntentType() string              { return "popconfirm.PopconfirmChangeIntent" }
func (PopconfirmChangeIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (PopconfirmChangeIntent) IsTransition() bool              { return false }
func (PopconfirmChangeIntent) IsGlobal() bool                  { return false }
func (i PopconfirmChangeIntent) GetComponentID() string        { return i.ComponentID }

// PopconfirmConfirmIntent is emitted globally after user confirms.
type PopconfirmConfirmIntent struct {
	ComponentID string
}

func (PopconfirmConfirmIntent) IntentType() string              { return "popconfirm.PopconfirmConfirmIntent" }
func (PopconfirmConfirmIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (PopconfirmConfirmIntent) IsTransition() bool              { return false }
func (PopconfirmConfirmIntent) IsGlobal() bool                  { return true }
func (i PopconfirmConfirmIntent) GetComponentID() string        { return i.ComponentID }

// PopconfirmCancelIntent is emitted globally after user cancels.
type PopconfirmCancelIntent struct {
	ComponentID string
}

func (PopconfirmCancelIntent) IntentType() string              { return "popconfirm.PopconfirmCancelIntent" }
func (PopconfirmCancelIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (PopconfirmCancelIntent) IsTransition() bool              { return false }
func (PopconfirmCancelIntent) IsGlobal() bool                  { return true }
func (i PopconfirmCancelIntent) GetComponentID() string        { return i.ComponentID }

func ToggleWithID(componentID string) PopconfirmToggleIntent {
	return PopconfirmToggleIntent{ComponentID: componentID}
}
func OpenWithID(componentID string) PopconfirmOpenIntent {
	return PopconfirmOpenIntent{ComponentID: componentID}
}
func CloseWithID(componentID string) PopconfirmCloseIntent {
	return PopconfirmCloseIntent{ComponentID: componentID}
}
