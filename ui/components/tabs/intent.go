package tabs

// =============================================================================
// Tabs Intent Definitions (Phase 7: Intent Bubble)
// =============================================================================

import "github.com/wwsheng009/mint/runtime/intent"

// =============================================================================
// TabChangeIntent
// =============================================================================

// TabChangeIntent is emitted when the active tab changes.
// This intent can be used in Intent Bubble pattern to notify parent components.
type TabChangeIntent struct {
	// ComponentID identifies the tabs component emitting this intent.
	ComponentID string

	// ActiveTab is the new active tab index.
	ActiveTab int

	// TabID is the ID of the newly active tab.
	TabID string

	// TabLabel is the label of the newly active tab.
	TabLabel string
}

// IntentType implements Intent interface.
func (TabChangeIntent) IntentType() string {
	return "Tabs:TabChange"
}

// Priority implements Intent interface.
func (TabChangeIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition implements Intent interface.
func (TabChangeIntent) IsTransition() bool {
	return true
}

// =============================================================================
// TabNextIntent
// =============================================================================

// TabNextIntent is a command intent to switch to the next tab.
// This can be handled by the Tabs Instance to navigate forward.
type TabNextIntent struct {
	// ComponentID identifies the target tabs component.
	ComponentID string
}

// IntentType implements Intent interface.
func (TabNextIntent) IntentType() string {
	return "Tabs:TabNext"
}

// Priority implements Intent interface.
func (TabNextIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition implements Intent interface.
func (TabNextIntent) IsTransition() bool {
	return true
}

// =============================================================================
// TabPreviousIntent
// =============================================================================

// TabPreviousIntent is a command intent to switch to the previous tab.
// This can be handled by the Tabs Instance to navigate backward.
type TabPreviousIntent struct {
	// ComponentID identifies the target tabs component.
	ComponentID string
}

// IntentType implements Intent interface.
func (TabPreviousIntent) IntentType() string {
	return "Tabs:TabPrevious"
}

// Priority implements Intent interface.
func (TabPreviousIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition implements Intent interface.
func (TabPreviousIntent) IsTransition() bool {
	return true
}

// ==============================================================================
// Helper Functions
// ==============================================================================

// TabChange creates a TabChangeIntent.
//
// Example:
//
//	emitIntent(TabChange("myTabs", 0, "tab1", "Home"))
func TabChange(componentID string, activeTab int, tabID, tabLabel string) TabChangeIntent {
	return TabChangeIntent{
		ComponentID: componentID,
		ActiveTab:   activeTab,
		TabID:       tabID,
		TabLabel:    tabLabel,
	}
}

// TabNext creates a TabNextIntent.
//
// Example:
//
//	intent.Emit(tabsComponent, TabNext("myTabs"))
func TabNext(componentID string) TabNextIntent {
	return TabNextIntent{ComponentID: componentID}
}

// TabPrevious creates a TabPreviousIntent.
//
// Example:
//
//	intent.Emit(tabsComponent, TabPrevious("myTabs"))
func TabPrevious(componentID string) TabPreviousIntent {
	return TabPreviousIntent{ComponentID: componentID}
}
