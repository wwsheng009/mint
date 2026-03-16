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
// Tab changes are synchronous UI events and should update state immediately.
func (TabChangeIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// TabChangeIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (TabChangeIntent) IsGlobal() bool {
	return false
}

// =============================================================================
// TabNextIntent
// =============================================================================

// TabCloseIntent is emitted after a tab is closed.
type TabCloseIntent struct {
	// ComponentID identifies the tabs component emitting this intent.
	ComponentID string

	// ClosedTabIndex is the index of the tab before it was removed.
	ClosedTabIndex int

	// ClosedTabID is the ID of the removed tab.
	ClosedTabID string

	// ClosedTabLabel is the label of the removed tab.
	ClosedTabLabel string

	// ActiveTab is the new active tab index after the close settles.
	ActiveTab int

	// ActiveTabID is the ID of the new active tab, if any.
	ActiveTabID string

	// ActiveTabLabel is the label of the new active tab, if any.
	ActiveTabLabel string
}

// IntentType implements Intent interface.
func (TabCloseIntent) IntentType() string {
	return "Tabs:TabClose"
}

// Priority implements Intent interface.
func (TabCloseIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition implements Intent interface.
func (TabCloseIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
func (TabCloseIntent) IsGlobal() bool {
	return false
}

// =============================================================================
// TabReorderIntent
// =============================================================================

// TabReorderIntent is emitted after a drag operation changes tab order.
type TabReorderIntent struct {
	// ComponentID identifies the tabs component emitting this intent.
	ComponentID string

	// FromIndex is the original index when dragging started.
	FromIndex int

	// ToIndex is the final index after dropping.
	ToIndex int

	// TabID is the ID of the dragged tab.
	TabID string

	// TabLabel is the label of the dragged tab.
	TabLabel string

	// TabOrder is the final ordered tab ID list after reordering.
	TabOrder []string

	// ActiveTab is the active tab index after reorder settles.
	ActiveTab int

	// ActiveTabID is the active tab ID after reorder settles.
	ActiveTabID string

	// ActiveTabLabel is the active tab label after reorder settles.
	ActiveTabLabel string
}

// IntentType implements Intent interface.
func (TabReorderIntent) IntentType() string {
	return "Tabs:TabReorder"
}

// Priority implements Intent interface.
func (TabReorderIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition implements Intent interface.
func (TabReorderIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
func (TabReorderIntent) IsGlobal() bool {
	return false
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
// Tab navigation commands are handled synchronously by the component.
func (TabNextIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// TabNextIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (TabNextIntent) IsGlobal() bool {
	return false
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
// Tab navigation commands are handled synchronously by the component.
func (TabPreviousIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// TabPreviousIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (TabPreviousIntent) IsGlobal() bool {
	return false
}

// =============================================================================
// TabSelectIntent
// =============================================================================

// TabSelectIntent is a command intent to select a specific tab by ID or index.
type TabSelectIntent struct {
	// ComponentID identifies the target tabs component.
	ComponentID string

	// TabID selects the target tab by ID when provided.
	TabID string

	// TabIndex selects the target tab by index when TabID is empty.
	TabIndex int
}

// IntentType implements Intent interface.
func (TabSelectIntent) IntentType() string {
	return "Tabs:TabSelect"
}

// Priority implements Intent interface.
func (TabSelectIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition implements Intent interface.
// Selecting a tab is an immediate state change, not an async transition.
func (TabSelectIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
func (TabSelectIntent) IsGlobal() bool {
	return false
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

// TabClose creates a TabCloseIntent.
func TabClose(
	componentID string,
	closedTabIndex int,
	closedTabID, closedTabLabel string,
	activeTab int,
	activeTabID, activeTabLabel string,
) TabCloseIntent {
	return TabCloseIntent{
		ComponentID:    componentID,
		ClosedTabIndex: closedTabIndex,
		ClosedTabID:    closedTabID,
		ClosedTabLabel: closedTabLabel,
		ActiveTab:      activeTab,
		ActiveTabID:    activeTabID,
		ActiveTabLabel: activeTabLabel,
	}
}

// TabReorder creates a TabReorderIntent.
func TabReorder(
	componentID string,
	fromIndex, toIndex int,
	tabID, tabLabel string,
	tabOrder []string,
	activeTab int,
	activeTabID, activeTabLabel string,
) TabReorderIntent {
	orderCopy := append([]string(nil), tabOrder...)
	return TabReorderIntent{
		ComponentID:    componentID,
		FromIndex:      fromIndex,
		ToIndex:        toIndex,
		TabID:          tabID,
		TabLabel:       tabLabel,
		TabOrder:       orderCopy,
		ActiveTab:      activeTab,
		ActiveTabID:    activeTabID,
		ActiveTabLabel: activeTabLabel,
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

// TabSelect creates a TabSelectIntent targeting a tab by ID.
func TabSelect(componentID, tabID string) TabSelectIntent {
	return TabSelectIntent{
		ComponentID: componentID,
		TabID:       tabID,
		TabIndex:    -1,
	}
}

// TabSelectIndex creates a TabSelectIntent targeting a tab by index.
func TabSelectIndex(componentID string, tabIndex int) TabSelectIntent {
	return TabSelectIntent{
		ComponentID: componentID,
		TabIndex:    tabIndex,
	}
}
