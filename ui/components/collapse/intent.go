package collapse

import "github.com/wwsheng009/mint/runtime/intent"

// CollapseToggleIntent is emitted by internal headers and can also be used externally.
type CollapseToggleIntent struct {
	ComponentID string
	ItemKey     string
	ItemHeader  string
	Index       int
}

func (CollapseToggleIntent) IntentType() string {
	return "collapse.CollapseToggleIntent"
}

func (CollapseToggleIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (CollapseToggleIntent) IsTransition() bool {
	return false
}

func (CollapseToggleIntent) IsGlobal() bool {
	return false
}

func (CollapseToggleIntent) StayPressed() bool {
	return true
}

func (i CollapseToggleIntent) GetComponentID() string {
	return i.ComponentID
}

// CollapseChangeIntent is emitted after the expanded key set changes.
type CollapseChangeIntent struct {
	ComponentID string
	ActiveKeys  []string
	ToggledKey  string
	ItemHeader  string
	Expanded    bool
	Index       int
	Accordion   bool
}

func (CollapseChangeIntent) IntentType() string {
	return "collapse.CollapseChangeIntent"
}

func (CollapseChangeIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (CollapseChangeIntent) IsTransition() bool {
	return false
}

func (CollapseChangeIntent) IsGlobal() bool {
	return false
}

func (i CollapseChangeIntent) GetComponentID() string {
	return i.ComponentID
}

// Toggle creates a component-agnostic toggle intent.
func Toggle(itemKey, itemHeader string, index int) CollapseToggleIntent {
	return CollapseToggleIntent{
		ItemKey:    itemKey,
		ItemHeader: itemHeader,
		Index:      index,
	}
}

// ToggleWithID creates a toggle intent bound to a component ID.
func ToggleWithID(componentID, itemKey, itemHeader string, index int) CollapseToggleIntent {
	toggle := Toggle(itemKey, itemHeader, index)
	toggle.ComponentID = componentID
	return toggle
}

// Change creates a local change intent.
func Change(activeKeys []string, toggledKey, itemHeader string, expanded bool, index int, accordion bool) CollapseChangeIntent {
	return CollapseChangeIntent{
		ActiveKeys: cloneStrings(activeKeys),
		ToggledKey: toggledKey,
		ItemHeader: itemHeader,
		Expanded:   expanded,
		Index:      index,
		Accordion:  accordion,
	}
}

// ChangeWithID creates a local change intent with component scoping.
func ChangeWithID(componentID string, activeKeys []string, toggledKey, itemHeader string, expanded bool, index int, accordion bool) CollapseChangeIntent {
	change := Change(activeKeys, toggledKey, itemHeader, expanded, index, accordion)
	change.ComponentID = componentID
	return change
}
