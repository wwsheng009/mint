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
