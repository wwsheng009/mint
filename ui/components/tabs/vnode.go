// Package tabs provides Fiber-first Tabs navigation component.
package tabs

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types
// =============================================================================

// TabPosition defines where tabs are positioned relative to content
type TabPosition int

const (
	TabPositionTop    TabPosition = iota // Tabs above content
	TabPositionBottom                    // Tabs below content
	TabPositionLeft                      // Tabs to the left of content
	TabPositionRight                     // Tabs to the right of content
)

// TabItem represents a single tab in a Tabs component
type TabItem struct {
	ID       string
	Label    string
	Disabled bool
}

// =============================================================================
// VNode - Pure Description (No State, No Closures, No Paint)
// =============================================================================

// VNode is the tabs component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Tab Props ===
	tabs      []TabItem
	position  TabPosition
	wrapTabs  bool
	tabGap    int

	// === Layout Props ===
	width  int
	height int
	flex   int

	// === Style ===
	tabStyle      style.Style
	activeTabStyle style.Style

	// === Intent (No Closures!) ===
	changeIntent intent.Intent
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new tabs VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:     rtui.NewElement("tabs"),
		tabs:            []TabItem{},
		position:        TabPositionTop,
		wrapTabs:        false,
		tabGap:          1,
		flex:            1,
		tabStyle:        style.Style{},
		activeTabStyle:  style.Style{},
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string           { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode { v.key = key; return v }
func (v *VNode) Tag() string           { return "tabs" }
func (v *VNode) Type() rtui.VNodeType  { return rtui.VNodeElement }

func (v *VNode) Children() []rtui.VNode {
	// Children are built dynamically by Instance
	return []rtui.VNode{}
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// Tabs manages its own children
	return v
}

func (v *VNode) GetLayer() rtui.Layer   { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Style() style.Style    { return v.tabStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.tabStyle = s; return v }

func (v *VNode) TextContent() string   { return "" }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":            v.key,
		"tabs":           v.tabs,
		"position":       v.position,
		"wrapTabs":       v.wrapTabs,
		"tabGap":         v.tabGap,
		"width":          v.width,
		"height":         v.height,
		"flex":           v.flex,
		"tabStyle":       v.tabStyle,
		"activeTabStyle": v.activeTabStyle,
		"changeIntent":   v.changeIntent,
	}
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p["key"].(string); ok {
		v.key = val
	}
	if val, ok := p["tabs"].([]TabItem); ok {
		v.tabs = val
	}
	if val, ok := p["position"].(TabPosition); ok {
		v.position = val
	}
	if val, ok := p["wrapTabs"].(bool); ok {
		v.wrapTabs = val
	}
	if val, ok := p["tabGap"].(int); ok {
		v.tabGap = val
	}
	if val, ok := p["width"].(int); ok {
		v.width = val
	}
	if val, ok := p["height"].(int); ok {
		v.height = val
	}
	if val, ok := p["flex"].(int); ok {
		v.flex = val
	}
	if val, ok := p["tabStyle"].(style.Style); ok {
		v.tabStyle = val
	}
	if val, ok := p["activeTabStyle"].(style.Style); ok {
		v.activeTabStyle = val
	}
	if val, ok := p["changeIntent"].(intent.Intent); ok {
		v.changeIntent = val
	}
	return v
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		"key":            v.key,
		"tabs":           v.tabs,
		"position":       v.position,
		"wrapTabs":       v.wrapTabs,
		"tabGap":         v.tabGap,
		"width":          v.width,
		"height":         v.height,
		"flex":           v.flex,
		"tabStyle":       v.tabStyle,
		"activeTabStyle": v.activeTabStyle,
		"changeIntent":   v.changeIntent,
	})
}

// =============================================================================
// Fluent Setters - Tabs Configuration
// =============================================================================

func (v *VNode) SetTabs(tabs []TabItem) *VNode  { v.tabs = tabs; return v }
func (v *VNode) SetPosition(pos TabPosition) *VNode { v.position = pos; return v }
func (v *VNode) SetWrapTabs(wrap bool) *VNode   { v.wrapTabs = wrap; return v }
func (v *VNode) SetTabGap(gap int) *VNode       { v.tabGap = gap; return v }
func (v *VNode) SetIntent(i intent.Intent) *VNode { v.changeIntent = i; return v }

// Layout setters
func (v *VNode) SetWidth(w int) *VNode  { v.width = w; return v }
func (v *VNode) SetHeight(h int) *VNode { v.height = h; return v }
func (v *VNode) SetFlex(f int) *VNode   { v.flex = f; return v }
func (v *VNode) Size(w, h int) *VNode   { return v.SetWidth(w).SetHeight(h) }

// Style setters
func (v *VNode) SetTabStyle(s style.Style) *VNode        { v.tabStyle = s; return v }
func (v *VNode) SetActiveTabStyle(s style.Style) *VNode { v.activeTabStyle = s; return v }

// Tab position convenience methods
func (v *VNode) Top() *VNode    { return v.SetPosition(TabPositionTop) }
func (v *VNode) Bottom() *VNode { return v.SetPosition(TabPositionBottom) }
func (v *VNode) Left() *VNode   { return v.SetPosition(TabPositionLeft) }
func (v *VNode) Right() *VNode  { return v.SetPosition(TabPositionRight) }

// =============================================================================
// Accessors
// =============================================================================

func (v *VNode) Tabs() []TabItem  { return v.tabs }
func (v *VNode) Position() TabPosition { return v.position }
func (v *VNode) WrapTabs() bool  { return v.wrapTabs }
func (v *VNode) TabGap() int     { return v.tabGap }
func (v *VNode) Width() int      { return v.width }
func (v *VNode) Height() int     { return v.height }
func (v *VNode) Flex() int       { return v.flex }
func (v *VNode) TabStyle() style.Style      { return v.tabStyle }
func (v *VNode) ActiveTabStyle() style.Style { return v.activeTabStyle }

// =============================================================================
// Tab Management Helpers
// =============================================================================

// AddTab adds a new tab
func (v *VNode) AddTab(id, label string) *VNode {
	v.tabs = append(v.tabs, TabItem{ID: id, Label: label, Disabled: false})
	return v
}

// AddTabWithOptions adds a new tab with options
func (v *VNode) AddTabWithOptions(id, label string, disabled bool) *VNode {
	v.tabs = append(v.tabs, TabItem{ID: id, Label: label, Disabled: disabled})
	return v
}
