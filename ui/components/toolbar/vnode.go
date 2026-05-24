// Package toolbar provides a Fiber-first operation toolbar for data and admin pages.
package toolbar

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	menucomp "github.com/wwsheng009/mint/ui/components/menu"
)

const (
	propActions      = "actions"
	propCenterItems  = "centerItems"
	propDense        = "dense"
	propGap          = "gap"
	propKey          = "key"
	propLeftItems    = "leftItems"
	propRightItems   = "rightItems"
	propSeparator    = "separator"
	propStyle        = "style"
	propTitle        = "title"
	propTitleWidth   = "titleWidth"
	propUseStatusBar = "useStatusBar"
	propWidth        = "width"
)

// ItemKind controls how a Toolbar item renders.
type ItemKind string

const (
	ItemText      ItemKind = "text"
	ItemBadge     ItemKind = "badge"
	ItemButton    ItemKind = "button"
	ItemMenu      ItemKind = "menu"
	ItemSeparator ItemKind = "separator"
	ItemCustom    ItemKind = "custom"
)

// Item describes one visible Toolbar entry.
type Item struct {
	Key            string
	Label          string
	Kind           ItemKind
	PressIntent    intent.Intent
	Variant        button.Variant
	Disabled       bool
	HelpText       string
	DisabledReason string
	Width          int
	FgColor        string
	BgColor        string
	Bold           bool
	Custom         rtui.VNode

	MenuID               string
	MenuItems            []menucomp.MenuItem
	MenuOpen             bool
	MenuPlacement        menucomp.Placement
	MenuActivePath       []int
	MenuMinWidth         int
	MenuMaxHeight        int
	MenuShowShortcuts    bool
	MenuShowDescriptions bool
}

// VNode is the declarative description of a Toolbar.
type VNode struct {
	*rtui.ElementVNode

	key          string
	title        string
	titleWidth   int
	width        int
	gap          int
	dense        bool
	separator    string
	useStatusBar bool
	leftItems    []Item
	centerItems  []Item
	rightItems   []Item
	rootStyle    style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a Toolbar VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("toolbar"),
		gap:          1,
		separator:    "|",
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to the key.
func (v *VNode) ID() string {
	if id := v.ElementVNode.ID(); id != "" {
		return id
	}
	return v.key
}

func (v *VNode) SetID(id string) rtui.VNode {
	v.ElementVNode.SetID(id)
	return v
}

func (v *VNode) Tag() string { return "toolbar" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propActions:      cloneItems(v.rightItems),
		propCenterItems:  cloneItems(v.centerItems),
		propDense:        v.dense,
		propGap:          v.gap,
		propKey:          v.key,
		propLeftItems:    cloneItems(v.leftItems),
		propRightItems:   cloneItems(v.rightItems),
		propSeparator:    v.separator,
		propStyle:        v.rootStyle,
		propTitle:        v.title,
		propTitleWidth:   v.titleWidth,
		propUseStatusBar: v.useStatusBar,
		propWidth:        v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	v.key = getStringProp(props, propKey, v.key)
	v.title = getStringProp(props, propTitle, v.title)
	v.titleWidth = getIntProp(props, propTitleWidth, v.titleWidth)
	v.width = getIntProp(props, propWidth, v.width)
	v.gap = getIntProp(props, propGap, v.gap)
	v.dense = getBoolProp(props, propDense, v.dense)
	v.separator = getStringProp(props, propSeparator, v.separator)
	v.useStatusBar = getBoolProp(props, propUseStatusBar, v.useStatusBar)
	v.leftItems = normalizeItems(getItemsProp(props, propLeftItems, v.leftItems))
	v.centerItems = normalizeItems(getItemsProp(props, propCenterItems, v.centerItems))
	v.rightItems = normalizeItems(getItemsProp(props, propRightItems, getItemsProp(props, propActions, v.rightItems)))
	v.rootStyle = getStyleProp(props, propStyle, v.rootStyle)
	v.normalize()
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

func (v *VNode) SetTitleWidth(width int) *VNode {
	v.titleWidth = width
	v.normalize()
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	v.normalize()
	return v
}

func (v *VNode) SetGap(gap int) *VNode {
	v.gap = gap
	v.normalize()
	return v
}

func (v *VNode) SetDense(dense bool) *VNode {
	v.dense = dense
	return v
}

func (v *VNode) SetSeparator(separator string) *VNode {
	v.separator = separator
	return v
}

func (v *VNode) SetUseStatusBar(use bool) *VNode {
	v.useStatusBar = use
	return v
}

func (v *VNode) SetRootStyle(s style.Style) *VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) SetLeftItems(items []Item) *VNode {
	v.leftItems = normalizeItems(items)
	return v
}

func (v *VNode) AddLeftItem(item Item) *VNode {
	v.leftItems = normalizeItems(append(v.leftItems, item))
	return v
}

func (v *VNode) SetCenterItems(items []Item) *VNode {
	v.centerItems = normalizeItems(items)
	return v
}

func (v *VNode) AddCenterItem(item Item) *VNode {
	v.centerItems = normalizeItems(append(v.centerItems, item))
	return v
}

func (v *VNode) SetRightItems(items []Item) *VNode {
	v.rightItems = normalizeItems(items)
	return v
}

func (v *VNode) AddRightItem(item Item) *VNode {
	v.rightItems = normalizeItems(append(v.rightItems, item))
	return v
}

func (v *VNode) LeftItems() []Item { return cloneItems(v.leftItems) }

func (v *VNode) CenterItems() []Item { return cloneItems(v.centerItems) }

func (v *VNode) RightItems() []Item { return cloneItems(v.rightItems) }

func (v *VNode) normalize() {
	v.leftItems = normalizeItems(v.leftItems)
	v.centerItems = normalizeItems(v.centerItems)
	v.rightItems = normalizeItems(v.rightItems)
	if v.width < 0 {
		v.width = 0
	}
	if v.titleWidth < 0 {
		v.titleWidth = 0
	}
	if v.gap < 0 {
		v.gap = 0
	}
}

// Text creates a plain toolbar item.
func Text(key, label string) Item {
	return Item{Key: key, Label: label, Kind: ItemText}
}

// Badge creates a highlighted toolbar item.
func Badge(key, label string) Item {
	return Item{Key: key, Label: label, Kind: ItemBadge, Bold: true}
}

// Button creates a command item.
func Button(key, label string, pressIntent intent.Intent) Item {
	return Item{Key: key, Label: label, Kind: ItemButton, PressIntent: pressIntent}
}

// Dropdown creates a toolbar button that can render an anchored menu popup.
//
// The open flag is intentionally controlled by the application state. Pressing
// the dropdown emits menu.OpenMenuIntent by default; reducers should set
// MenuOpen(true) on the next render and close it through the existing menu
// close/outside-click flow.
func Dropdown(key, label string, items []menucomp.MenuItem, open bool) Item {
	return Item{
		Key:               key,
		Label:             label,
		Kind:              ItemMenu,
		MenuItems:         menucomp.NormalizeItems(items),
		MenuOpen:          open,
		MenuPlacement:     menucomp.PlacementBottomStart,
		MenuShowShortcuts: true,
	}
}

// Menu is an alias for Dropdown.
func Menu(key, label string, items []menucomp.MenuItem, open bool) Item {
	return Dropdown(key, label, items, open)
}

// Separator creates a visual separator item.
func Separator(key string) Item {
	return Item{Key: key, Kind: ItemSeparator}
}

// Custom creates a custom toolbar item.
func Custom(key string, node rtui.VNode) Item {
	return Item{Key: key, Kind: ItemCustom, Custom: node}
}

func (i Item) WithKey(key string) Item {
	i.Key = key
	return i
}

func (i Item) WithLabel(label string) Item {
	i.Label = label
	return i
}

func (i Item) OnPress(pressIntent intent.Intent) Item {
	i.PressIntent = pressIntent
	if i.Kind == ItemText || i.Kind == ItemBadge {
		i.Kind = ItemButton
	}
	return i
}

func (i Item) WithVariant(variant button.Variant) Item {
	i.Variant = variant
	return i
}

func (i Item) Primary() Item {
	i.Variant = button.VariantPrimary
	return i
}

func (i Item) Secondary() Item {
	i.Variant = button.VariantSecondary
	return i
}

func (i Item) Danger() Item {
	i.Variant = button.VariantDanger
	return i
}

func (i Item) Success() Item {
	i.Variant = button.VariantSuccess
	return i
}

func (i Item) WithDisabled(disabled bool) Item {
	i.Disabled = disabled
	return i
}

func (i Item) WithDisabledReason(reason string) Item {
	i.Disabled = true
	i.DisabledReason = normalizeToolbarText(reason)
	return i
}

func (i Item) WithHelp(helpText string) Item {
	i.HelpText = normalizeToolbarText(helpText)
	return i
}

func (i Item) WithTooltip(tooltipText string) Item {
	return i.WithHelp(tooltipText)
}

func (i Item) WithWidth(width int) Item {
	i.Width = width
	return i
}

func (i Item) WithColors(fgColor, bgColor string) Item {
	i.FgColor = fgColor
	i.BgColor = bgColor
	return i
}

func (i Item) WithForeground(fgColor string) Item {
	i.FgColor = fgColor
	return i
}

func (i Item) WithBackground(bgColor string) Item {
	i.BgColor = bgColor
	return i
}

func (i Item) WithBold(bold bool) Item {
	i.Bold = bold
	return i
}

func (i Item) WithMenuID(menuID string) Item {
	i.MenuID = strings.TrimSpace(menuID)
	return i
}

func (i Item) WithMenuItems(items []menucomp.MenuItem) Item {
	i.MenuItems = menucomp.NormalizeItems(items)
	if i.Kind == "" || i.Kind == ItemText || i.Kind == ItemButton {
		i.Kind = ItemMenu
	}
	return i
}

func (i Item) WithMenuOpen(open bool) Item {
	i.MenuOpen = open
	return i
}

func (i Item) WithMenuPlacement(placement menucomp.Placement) Item {
	i.MenuPlacement = placement
	return i
}

func (i Item) WithMenuActivePath(path ...int) Item {
	i.MenuActivePath = append([]int(nil), path...)
	return i
}

func (i Item) WithMenuMinWidth(width int) Item {
	i.MenuMinWidth = width
	return i
}

func (i Item) WithMenuMaxHeight(height int) Item {
	i.MenuMaxHeight = height
	return i
}

func (i Item) WithMenuShortcuts(show bool) Item {
	i.MenuShowShortcuts = show
	return i
}

func (i Item) WithMenuDescriptions(show bool) Item {
	i.MenuShowDescriptions = show
	return i
}

func normalizeItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	normalized := cloneItems(items)
	seen := make(map[string]int, len(normalized))
	for index := range normalized {
		key := strings.TrimSpace(normalized[index].Key)
		if key == "" {
			key = fmt.Sprintf("item-%d", index)
		}
		base := key
		if count, exists := seen[base]; exists {
			count++
			seen[base] = count
			key = fmt.Sprintf("%s-%d", base, count)
		} else {
			seen[base] = 0
		}
		normalized[index].Key = key
		if normalized[index].Kind == "" {
			normalized[index].Kind = ItemText
		}
		switch normalized[index].Kind {
		case ItemText, ItemBadge, ItemButton, ItemMenu, ItemSeparator, ItemCustom:
		default:
			normalized[index].Kind = ItemText
		}
		if normalized[index].Width < 0 {
			normalized[index].Width = 0
		}
		normalized[index].Label = normalizeToolbarText(normalized[index].Label)
		normalized[index].HelpText = normalizeToolbarText(normalized[index].HelpText)
		normalized[index].DisabledReason = normalizeToolbarText(normalized[index].DisabledReason)
		normalized[index].MenuID = strings.TrimSpace(normalized[index].MenuID)
		normalized[index].MenuItems = menucomp.NormalizeItems(normalized[index].MenuItems)
		normalized[index].MenuActivePath = append([]int(nil), normalized[index].MenuActivePath...)
		if normalized[index].MenuPlacement == "" {
			normalized[index].MenuPlacement = menucomp.PlacementBottomStart
		}
		if normalized[index].MenuMinWidth < 0 {
			normalized[index].MenuMinWidth = 0
		}
		if normalized[index].MenuMaxHeight < 0 {
			normalized[index].MenuMaxHeight = 0
		}
	}
	return normalized
}

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]Item, len(items))
	copy(cloned, items)
	return cloned
}

func normalizeToolbarText(content string) string {
	return strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(content)
}

func menuAnchorForPlacement(placement menucomp.Placement) rttypes.Anchor {
	switch placement {
	case menucomp.PlacementTopStart, menucomp.PlacementTopEnd:
		return rttypes.AnchorTopLeft
	case menucomp.PlacementRightStart:
		return rttypes.AnchorTopRight
	case menucomp.PlacementLeftStart:
		return rttypes.AnchorTopLeft
	default:
		return rttypes.AnchorBottomLeft
	}
}
