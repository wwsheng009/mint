package ui

import (
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	menucomp "github.com/wwsheng009/mint/ui/components/menu"
)

type MenuItem = menucomp.MenuItem
type MenuTheme = menucomp.Theme
type MenuVariant = menucomp.Variant
type MenuPlacement = menucomp.Placement
type MenuPortalPosition = rttypes.PositionType
type MenuShortcut = menucomp.Shortcut
type MenuShortcutBinding = menucomp.ShortcutBinding
type OpenMenuIntent = menucomp.OpenMenuIntent
type CloseMenuIntent = menucomp.CloseMenuIntent
type ActivateMenuItemIntent = menucomp.ActivateMenuItemIntent
type NavigateMenuIntent = menucomp.NavigateMenuIntent

const (
	MenuVariantBar      = menucomp.VariantMenuBar
	MenuVariantDropdown = menucomp.VariantDropdown
	MenuVariantContext  = menucomp.VariantContext
	MenuVariantPopup    = menucomp.VariantPopup

	MenuPlacementAuto        = menucomp.PlacementAuto
	MenuPlacementBottomStart = menucomp.PlacementBottomStart
	MenuPlacementBottomEnd   = menucomp.PlacementBottomEnd
	MenuPlacementTopStart    = menucomp.PlacementTopStart
	MenuPlacementTopEnd      = menucomp.PlacementTopEnd
	MenuPlacementRightStart  = menucomp.PlacementRightStart
	MenuPlacementLeftStart   = menucomp.PlacementLeftStart

	MenuPortalRelative = rttypes.PositionRelative
	MenuPortalAbsolute = rttypes.PositionAbsolute
	MenuPortalFixed    = rttypes.PositionFixed
)

func NewMenuBuilder() *menucomp.Builder { return menucomp.NewBuilder() }
func NewMenuBarBuilder(items []menucomp.MenuItem) *menucomp.Builder {
	return menucomp.NewMenuBar(items)
}
func NewMenuPopupBuilder(items []menucomp.MenuItem) *menucomp.Builder {
	return menucomp.NewPopup(items)
}
func NewContextMenuBuilder(items []menucomp.MenuItem) *menucomp.Builder {
	return menucomp.NewContextMenu(items)
}
func MenuThemeDefault() menucomp.Theme  { return menucomp.DefaultTheme() }
func MenuThemeMuted() menucomp.Theme    { return menucomp.MutedTheme() }
func MenuThemeContrast() menucomp.Theme { return menucomp.ContrastTheme() }
func MenuAction(key, label string, pressIntent intent.Intent) menucomp.MenuItem {
	return menucomp.Action(key, label, pressIntent)
}
func MenuCheckbox(key, label string, checked bool, pressIntent intent.Intent) menucomp.MenuItem {
	return menucomp.Checkbox(key, label, checked, pressIntent)
}
func MenuRadio(key, label, group string, checked bool, pressIntent intent.Intent) menucomp.MenuItem {
	return menucomp.Radio(key, label, group, checked, pressIntent)
}
func MenuSubmenu(key, label string, children ...menucomp.MenuItem) menucomp.MenuItem {
	return menucomp.Submenu(key, label, children...)
}
func MenuSeparator() menucomp.MenuItem                         { return menucomp.Separator() }
func MenuLabel(key, label string) menucomp.MenuItem            { return menucomp.LabelItem(key, label) }
func MenuItems(items ...menucomp.MenuItem) []menucomp.MenuItem { return menucomp.Items(items...) }
func MenuCollectShortcuts(items []menucomp.MenuItem) []menucomp.ShortcutBinding {
	return menucomp.CollectShortcuts(items)
}
func BindMenuGlobalShortcuts(app *framework.App, builder *menucomp.Builder) int {
	if app == nil || builder == nil {
		return 0
	}
	return builder.BindGlobalShortcuts(app, func(i intent.Intent) {
		EmitIntentGlobal(i)
	})
}

func InstallMenu(app *framework.App, builders ...*menucomp.Builder) int {
	if app == nil {
		return 0
	}
	return menucomp.Install(app, func(i intent.Intent) {
		EmitIntentGlobal(i)
	}, builders...)
}
