// Package tabs provides Fiber-first Tabs navigation component.
package tabs

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propActiveTab         = "activeTab"
	propActiveTabID       = "activeTabID"
	propActiveTabStyle    = "activeTabStyle"
	propChangeIntent      = "changeIntent"
	propChangeIntentField = "changeIntentField"
	propCloseIntent       = "closeIntent"
	propComponentID       = "componentID"
	propDisabledTabStyle  = "disabledTabStyle"
	propDivider           = "divider"
	propFlex              = "flex"
	propHeight            = "height"
	propKey               = "key"
	propLoopNavigation    = "loopNavigation"
	propPosition          = "position"
	propReorderIntent     = "reorderIntent"
	propReorderable       = "reorderable"
	propShowHotkeys       = "showHotkeys"
	propTabGap            = "tabGap"
	propTabVariant        = "tabVariant"
	propTabStyle          = "tabStyle"
	propTabs              = "tabs"
	propWidth             = "width"
	propWrapTabs          = "wrapTabs"
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

// TabVariant defines the visual treatment of tabs.
type TabVariant int

const (
	TabVariantLine TabVariant = iota
	TabVariantCard
)

// TabItem represents a single tab in a Tabs component
type TabItem struct {
	ID       string
	Label    string
	Icon     string
	Badge    string
	Hotkey   rune
	Closable bool
	Disabled bool
	Hidden   bool
}

// Item creates a tab item with the provided ID and label.
func Item(id, label string) TabItem {
	return TabItem{ID: id, Label: label}
}

// WithIcon adds an icon prefix to the tab label.
func (t TabItem) WithIcon(icon string) TabItem {
	t.Icon = icon
	return t
}

// WithBadge adds a badge suffix to the tab label.
func (t TabItem) WithBadge(badge string) TabItem {
	t.Badge = badge
	return t
}

// WithHotkey assigns a keyboard hotkey for direct tab activation.
func (t TabItem) WithHotkey(hotkey rune) TabItem {
	t.Hotkey = hotkey
	return t
}

// WithClosable toggles whether the tab can be closed by the user.
func (t TabItem) WithClosable(closable bool) TabItem {
	t.Closable = closable
	return t
}

// WithDisabled toggles the disabled state.
func (t TabItem) WithDisabled(disabled bool) TabItem {
	t.Disabled = disabled
	return t
}

// WithHidden toggles visibility without removing the tab from the data model.
func (t TabItem) WithHidden(hidden bool) TabItem {
	t.Hidden = hidden
	return t
}

// =============================================================================
// VNode - Pure Description (No State, No Closures, No Paint)
// =============================================================================

// VNode is the tabs component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key         string
	componentID string // Phase 7: Component ID for Intent routing

	// === Tab Props ===
	tabs           []TabItem
	position       TabPosition
	activeTab      int
	activeTabID    string
	wrapTabs       bool
	tabGap         int
	loopNavigation bool
	showHotkeys    bool
	divider        string
	tabVariant     TabVariant
	reorderable    bool

	// === Layout Props ===
	width  int
	height int
	flex   int

	// === Style ===
	tabStyle         style.Style
	activeTabStyle   style.Style
	disabledTabStyle style.Style

	// === Intent (No Closures!) ===
	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent // For FieldChangeIntent
	closeIntent       intent.Intent
	reorderIntent     intent.Intent
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
		ElementVNode:      rtui.NewElement("tabs"),
		componentID:       "", // Phase 7: Component ID
		tabs:              []TabItem{},
		position:          TabPositionTop,
		activeTab:         -1,
		wrapTabs:          false,
		tabGap:            1,
		loopNavigation:    false,
		showHotkeys:       false,
		divider:           " | ",
		tabVariant:        TabVariantLine,
		reorderable:       false,
		flex:              1,
		tabStyle:          style.Style{},
		activeTabStyle:    style.Style{},
		disabledTabStyle:  style.Style{},
		changeIntentField: nil, // Phase 7: FieldChangeIntent
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode { v.key = key; return v }
func (v *VNode) Tag() string                  { return "tabs" }
func (v *VNode) Type() rtui.VNodeType         { return rtui.VNodeElement }

func (v *VNode) Children() []rtui.VNode {
	// Children are built dynamically by Instance
	return []rtui.VNode{}
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// Tabs manages its own children
	return v
}

func (v *VNode) GetLayer() rtui.Layer             { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Style() style.Style                { return v.tabStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.tabStyle = s; return v }

func (v *VNode) TextContent() string { return "" }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:               v.key,
		propComponentID:       v.componentID, // Phase 7
		propTabs:              v.tabs,
		propPosition:          v.position,
		propActiveTab:         v.activeTab,
		propActiveTabID:       v.activeTabID,
		propWrapTabs:          v.wrapTabs,
		propTabGap:            v.tabGap,
		propLoopNavigation:    v.loopNavigation,
		propShowHotkeys:       v.showHotkeys,
		propDivider:           v.divider,
		propTabVariant:        v.tabVariant,
		propReorderable:       v.reorderable,
		propWidth:             v.width,
		propHeight:            v.height,
		propFlex:              v.flex,
		propTabStyle:          v.tabStyle,
		propActiveTabStyle:    v.activeTabStyle,
		propDisabledTabStyle:  v.disabledTabStyle,
		propChangeIntent:      v.changeIntent,
		propChangeIntentField: v.changeIntentField, // Phase 7
		propCloseIntent:       v.closeIntent,
		propReorderIntent:     v.reorderIntent,
	}
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p[propKey].(string); ok {
		v.key = val
	}
	if val, ok := p[propComponentID].(string); ok {
		v.componentID = val // Phase 7
	}
	if val, ok := p[propTabs].([]TabItem); ok {
		v.tabs = val
	}
	if val, ok := p[propPosition].(TabPosition); ok {
		v.position = val
	}
	if val, ok := p[propActiveTab].(int); ok {
		v.activeTab = val
	}
	if val, ok := p[propActiveTabID].(string); ok {
		v.activeTabID = val
	}
	if val, ok := p[propWrapTabs].(bool); ok {
		v.wrapTabs = val
	}
	if val, ok := p[propTabGap].(int); ok {
		v.tabGap = val
	}
	if val, ok := p[propLoopNavigation].(bool); ok {
		v.loopNavigation = val
	}
	if val, ok := p[propShowHotkeys].(bool); ok {
		v.showHotkeys = val
	}
	if val, ok := p[propDivider].(string); ok {
		v.divider = val
	}
	if val, ok := p[propTabVariant].(TabVariant); ok {
		v.tabVariant = val
	}
	if val, ok := p[propReorderable].(bool); ok {
		v.reorderable = val
	}
	if val, ok := p[propWidth].(int); ok {
		v.width = val
	}
	if val, ok := p[propHeight].(int); ok {
		v.height = val
	}
	if val, ok := p[propFlex].(int); ok {
		v.flex = val
	}
	if val, ok := p[propTabStyle].(style.Style); ok {
		v.tabStyle = val
	}
	if val, ok := p[propActiveTabStyle].(style.Style); ok {
		v.activeTabStyle = val
	}
	if val, ok := p[propDisabledTabStyle].(style.Style); ok {
		v.disabledTabStyle = val
	}
	if val, ok := p[propChangeIntent].(intent.Intent); ok {
		v.changeIntent = val
	}
	if val, ok := p[propChangeIntentField].(intent.FieldIntent); ok {
		v.changeIntentField = val // Phase 7
	}
	if val, ok := p[propCloseIntent].(intent.Intent); ok {
		v.closeIntent = val
	}
	if val, ok := p[propReorderIntent].(intent.Intent); ok {
		v.reorderIntent = val
	}
	return v
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		propKey:               v.key,
		propComponentID:       v.componentID, // Phase 7
		propTabs:              v.tabs,
		propPosition:          v.position,
		propActiveTab:         v.activeTab,
		propActiveTabID:       v.activeTabID,
		propWrapTabs:          v.wrapTabs,
		propTabGap:            v.tabGap,
		propLoopNavigation:    v.loopNavigation,
		propShowHotkeys:       v.showHotkeys,
		propDivider:           v.divider,
		propTabVariant:        v.tabVariant,
		propReorderable:       v.reorderable,
		propWidth:             v.width,
		propHeight:            v.height,
		propFlex:              v.flex,
		propTabStyle:          v.tabStyle,
		propActiveTabStyle:    v.activeTabStyle,
		propDisabledTabStyle:  v.disabledTabStyle,
		propChangeIntent:      v.changeIntent,
		propChangeIntentField: v.changeIntentField, // Phase 7
		propCloseIntent:       v.closeIntent,
		propReorderIntent:     v.reorderIntent,
	})
}

// =============================================================================
// Fluent Setters - Tabs Configuration
// =============================================================================

// Identification setters
func (v *VNode) SetComponentID(id string) *VNode { v.componentID = id; return v } // Phase 7

func (v *VNode) SetTabs(tabs []TabItem) *VNode      { v.tabs = tabs; return v }
func (v *VNode) SetPosition(pos TabPosition) *VNode { v.position = pos; return v }
func (v *VNode) SetActiveTab(index int) *VNode      { v.activeTab = index; return v }
func (v *VNode) SetActiveTabID(id string) *VNode    { v.activeTabID = id; return v }
func (v *VNode) SetWrapTabs(wrap bool) *VNode       { v.wrapTabs = wrap; return v }
func (v *VNode) SetTabGap(gap int) *VNode           { v.tabGap = gap; return v }
func (v *VNode) SetLoopNavigation(loop bool) *VNode { v.loopNavigation = loop; return v }
func (v *VNode) SetShowHotkeys(show bool) *VNode    { v.showHotkeys = show; return v }
func (v *VNode) SetDivider(divider string) *VNode   { v.divider = divider; return v }
func (v *VNode) SetReorderable(reorderable bool) *VNode {
	v.reorderable = reorderable
	return v
}
func (v *VNode) SetTabVariant(variant TabVariant) *VNode {
	v.tabVariant = variant
	return v
}

// Intent setters
func (v *VNode) SetIntent(i intent.Intent) *VNode           { v.changeIntent = i; return v }
func (v *VNode) SetFieldIntent(i intent.FieldIntent) *VNode { v.changeIntentField = i; return v } // Phase 7
func (v *VNode) SetCloseIntent(i intent.Intent) *VNode      { v.closeIntent = i; return v }
func (v *VNode) OnClose(i intent.Intent) *VNode             { return v.SetCloseIntent(i) }
func (v *VNode) SetReorderIntent(i intent.Intent) *VNode    { v.reorderIntent = i; return v }
func (v *VNode) OnReorder(i intent.Intent) *VNode           { return v.SetReorderIntent(i) }

// Layout setters
func (v *VNode) SetWidth(w int) *VNode  { v.width = w; return v }
func (v *VNode) SetHeight(h int) *VNode { v.height = h; return v }
func (v *VNode) SetFlex(f int) *VNode   { v.flex = f; return v }
func (v *VNode) Size(w, h int) *VNode   { return v.SetWidth(w).SetHeight(h) }

// Style setters
func (v *VNode) SetTabStyle(s style.Style) *VNode         { v.tabStyle = s; return v }
func (v *VNode) SetActiveTabStyle(s style.Style) *VNode   { v.activeTabStyle = s; return v }
func (v *VNode) SetDisabledTabStyle(s style.Style) *VNode { v.disabledTabStyle = s; return v }

// Tab position convenience methods
func (v *VNode) Top() *VNode    { return v.SetPosition(TabPositionTop) }
func (v *VNode) Bottom() *VNode { return v.SetPosition(TabPositionBottom) }
func (v *VNode) Left() *VNode   { return v.SetPosition(TabPositionLeft) }
func (v *VNode) Right() *VNode  { return v.SetPosition(TabPositionRight) }
func (v *VNode) Line() *VNode   { return v.SetTabVariant(TabVariantLine) }
func (v *VNode) Card() *VNode   { return v.SetTabVariant(TabVariantCard) }

// =============================================================================
// Accessors
// =============================================================================

func (v *VNode) Tabs() []TabItem               { return v.tabs }
func (v *VNode) Position() TabPosition         { return v.position }
func (v *VNode) ActiveTab() int                { return v.activeTab }
func (v *VNode) ActiveTabID() string           { return v.activeTabID }
func (v *VNode) WrapTabs() bool                { return v.wrapTabs }
func (v *VNode) TabGap() int                   { return v.tabGap }
func (v *VNode) LoopNavigation() bool          { return v.loopNavigation }
func (v *VNode) ShowHotkeys() bool             { return v.showHotkeys }
func (v *VNode) Divider() string               { return v.divider }
func (v *VNode) Reorderable() bool             { return v.reorderable }
func (v *VNode) TabVariant() TabVariant        { return v.tabVariant }
func (v *VNode) Width() int                    { return v.width }
func (v *VNode) Height() int                   { return v.height }
func (v *VNode) Flex() int                     { return v.flex }
func (v *VNode) TabStyle() style.Style         { return v.tabStyle }
func (v *VNode) ActiveTabStyle() style.Style   { return v.activeTabStyle }
func (v *VNode) DisabledTabStyle() style.Style { return v.disabledTabStyle }
func (v *VNode) CloseIntent() intent.Intent    { return v.closeIntent }
func (v *VNode) ReorderIntent() intent.Intent  { return v.reorderIntent }

// =============================================================================
// Tab Management Helpers
// =============================================================================

// AddTab adds a new tab
func (v *VNode) AddTab(id, label string) *VNode {
	v.tabs = append(v.tabs, TabItem{ID: id, Label: label, Disabled: false})
	return v
}

// AddTabItem adds a fully configured tab item.
func (v *VNode) AddTabItem(tab TabItem) *VNode {
	v.tabs = append(v.tabs, tab)
	return v
}

// AddTabWithOptions adds a new tab with options
func (v *VNode) AddTabWithOptions(id, label string, disabled bool) *VNode {
	v.tabs = append(v.tabs, TabItem{ID: id, Label: label, Disabled: disabled})
	return v
}
