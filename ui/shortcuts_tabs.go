package ui

import "github.com/wwsheng009/mint/ui/components/tabs"

// Tabs intent type re-exports.
type (
	TabChangeIntent   = tabs.TabChangeIntent
	TabNextIntent     = tabs.TabNextIntent
	TabPreviousIntent = tabs.TabPreviousIntent
	TabSelectIntent   = tabs.TabSelectIntent
)

// NewTabItem creates a tab item with the provided ID and label.
func NewTabItem(id, label string) tabs.TabItem {
	return tabs.Item(id, label)
}

// TabsChange creates a change event payload for custom intent handling.
func TabsChange(componentID string, activeTab int, tabID, tabLabel string) tabs.TabChangeIntent {
	return tabs.TabChange(componentID, activeTab, tabID, tabLabel)
}

// TabsNext creates an intent that advances a tabs component.
func TabsNext(componentID string) tabs.TabNextIntent {
	return tabs.TabNext(componentID)
}

// TabsPrevious creates an intent that moves a tabs component backward.
func TabsPrevious(componentID string) tabs.TabPreviousIntent {
	return tabs.TabPrevious(componentID)
}

// TabsSelect creates an intent that selects a tab by ID.
func TabsSelect(componentID, tabID string) tabs.TabSelectIntent {
	return tabs.TabSelect(componentID, tabID)
}

// TabsSelectIndex creates an intent that selects a tab by index.
func TabsSelectIndex(componentID string, tabIndex int) tabs.TabSelectIntent {
	return tabs.TabSelectIndex(componentID, tabIndex)
}
