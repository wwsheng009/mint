package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	menucomp "github.com/wwsheng009/mint/ui/components/menu"
	"github.com/wwsheng009/mint/ui/components/toolbar"
)

type ToolbarBuilder = toolbar.Builder
type ToolbarVNode = toolbar.VNode
type ToolbarItem = toolbar.Item
type ToolbarItemKind = toolbar.ItemKind

const (
	ToolbarItemText      = toolbar.ItemText
	ToolbarItemBadge     = toolbar.ItemBadge
	ToolbarItemButton    = toolbar.ItemButton
	ToolbarItemMenu      = toolbar.ItemMenu
	ToolbarItemSeparator = toolbar.ItemSeparator
	ToolbarItemCustom    = toolbar.ItemCustom
)

// NewToolbarBuilder creates a Toolbar builder.
func NewToolbarBuilder() *toolbar.Builder {
	return toolbar.NewBuilder()
}

// Toolbar creates a Toolbar from left, center, and right items.
func Toolbar(left, center, right []toolbar.Item) rtui.VNode {
	return toolbar.Of(left, center, right)
}

// ToolbarText creates a plain toolbar item.
func ToolbarText(key, label string) toolbar.Item {
	return toolbar.Text(key, label)
}

// ToolbarBadge creates a highlighted toolbar item.
func ToolbarBadge(key, label string) toolbar.Item {
	return toolbar.Badge(key, label)
}

// ToolbarButton creates a command toolbar item.
func ToolbarButton(key, label string, pressIntent intent.Intent) toolbar.Item {
	return toolbar.Button(key, label, pressIntent)
}

// ToolbarDropdown creates a controlled toolbar menu item.
func ToolbarDropdown(key, label string, items []menucomp.MenuItem, open bool) toolbar.Item {
	return toolbar.Dropdown(key, label, items, open)
}

// ToolbarSeparator creates a toolbar separator item.
func ToolbarSeparator(key string) toolbar.Item {
	return toolbar.Separator(key)
}

// ToolbarCustom creates a custom toolbar item.
func ToolbarCustom(key string, node rtui.VNode) toolbar.Item {
	return toolbar.Custom(key, node)
}

// ToolbarKeyValue creates a compact "label: value" toolbar item.
func ToolbarKeyValue(key, label, value string) toolbar.Item {
	return toolbar.KeyValue(key, label, value)
}

// ToolbarMutedKeyValue creates a low-emphasis "label: value" toolbar item.
func ToolbarMutedKeyValue(key, label, value string) toolbar.Item {
	return toolbar.MutedKeyValue(key, label, value)
}

// ToolbarStateBadge creates a highlighted toolbar item using operational tone mapping.
func ToolbarStateBadge(key, status string) toolbar.Item {
	return toolbar.StateBadge(key, status)
}

// ToolbarBusyBadge creates a warning-colored toolbar item for running operations.
func ToolbarBusyBadge(key, label string) toolbar.Item {
	return toolbar.BusyBadge(key, label)
}

// ToolbarErrorBadge creates an error-colored toolbar item.
func ToolbarErrorBadge(key, label string) toolbar.Item {
	return toolbar.ErrorBadge(key, label)
}

// ToolbarEndpoint creates a standard toolbar item for the active API endpoint.
func ToolbarEndpoint(value string) toolbar.Item {
	return toolbar.Endpoint(value)
}

// ToolbarScope creates a standard toolbar item for the current page scope.
func ToolbarScope(value string) toolbar.Item {
	return toolbar.Scope(value)
}

// ToolbarSelection creates a low-emphasis toolbar item for the current selection.
func ToolbarSelection(value string) toolbar.Item {
	return toolbar.Selection(value)
}
