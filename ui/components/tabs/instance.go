package tabs

import (
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for tabs components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	tabs      []TabItem
	position  TabPosition
	wrapTabs  bool
	tabGap    int
	tabStyle      style.Style
	activeTabStyle style.Style
	changeIntent intent.Intent

	// === Layout Props ===
	width  int
	height int
	flex   int

	// === Runtime State ===
	activeTab int  // Current active tab index
	bounds    [4]int // x, y, w, h
	dirty     bool

	// === Tab Bar Bounds (for click detection) ===
	tabBarBounds []tabBounds // Bounds for each tab in the bar

	// === Intent Emitter ===
	intentEmitter func(intent.Intent)
}

type tabBounds struct {
	x      int
	y      int
	width  int
	height int
	tabID  string
	index  int
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new TabsInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:            getStringProp(props, "key", ""),
		tabs:           getTabsProp(props, []TabItem{}),
		position:       getTabPositionProp(props, TabPositionTop),
		wrapTabs:       getBoolProp(props, "wrapTabs", false),
		tabGap:         getIntProp(props, "tabGap", 1),
		tabStyle:       getStyleProp(props, "tabStyle"),
		activeTabStyle: getStyleProp(props, "activeTabStyle"),
		changeIntent:   getIntentProp(props),
		width:          getIntProp(props, "width", 0),
		height:         getIntProp(props, "height", 0),
		flex:           getIntProp(props, "flex", 1),
		activeTab:      0,
		dirty:          true,
	}
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()             { inst.tabs = nil }
func (inst *Instance) OnMount()             { inst.dirty = true }
func (inst *Instance) OnUnmount()           {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldTabs := inst.tabs
	oldPosition := inst.position

	inst.tabs = getTabsProp(props, inst.tabs)
	inst.position = getTabPositionProp(props, inst.position)
	inst.wrapTabs = getBoolProp(props, "wrapTabs", inst.wrapTabs)
	inst.tabGap = getIntProp(props, "tabGap", inst.tabGap)
	inst.tabStyle = getStyleProp(props, "tabStyle")
	inst.activeTabStyle = getStyleProp(props, "activeTabStyle")
	inst.changeIntent = getIntentProp(props)
	inst.width = getIntProp(props, "width", inst.width)
	inst.height = getIntProp(props, "height", inst.height)
	inst.flex = getIntProp(props, "flex", inst.flex)

	changed := !tabsEqual(oldTabs, inst.tabs) || oldPosition != inst.position
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":      inst.key,
		"tabs":     inst.tabs,
		"position": inst.position,
		"activeTab": inst.activeTab,
	}
}

func (inst *Instance) MarkDirty()    { inst.dirty = true }
func (inst *Instance) IsDirty() bool { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()   { inst.dirty = false }

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout measurement.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	tabBarHeight := 1

	minWidth := inst.calculateTabBarWidth()
	minHeight := tabBarHeight + inst.height

	// Use explicit width/height if set
	width := inst.width
	height := inst.height

	if width == 0 {
		width = minWidth
	}
	if height == 0 {
		height = minHeight
	}

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	return layout.Size{Width: width, Height: height}
}

// calculateTabBarWidth calculates the total width of the tab bar
func (inst *Instance) calculateTabBarWidth() int {
	if inst.wrapTabs || len(inst.tabs) == 0 {
		return 0
	}

	cursor := 0
	for i, tab := range inst.tabs {
		if i > 0 {
			cursor += 3 // " | "
		}

		labelWidth := utf8.RuneCountInString(tab.Label)
		width := labelWidth
		if i == inst.activeTab {
			width += 2 // brackets []
		}
		cursor += width
	}

	return cursor
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements drawing logic for the tabs.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd

	// Update bounds
	inst.bounds = [4]int{x, y, inst.width, inst.height}

	// Draw tab bar
	cmds = append(cmds, inst.paintTabBar(x, y)...)

	return cmds
}

// paintTabBar paints the tab bar
func (inst *Instance) paintTabBar(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd
	inst.tabBarBounds = []tabBounds{}

	if !inst.wrapTabs && len(inst.tabs) > 0 {
		// Single-line tab bar
		cursor := x
		for i, tab := range inst.tabs {
			if i > 0 {
				separator := style.Style{FG: style.Color("white")}
				sepText := " | "
				cmds = append(cmds, paint.NewTextCmd(cursor, y, sepText, separator))
				cursor += 3
			}

			labelWidth := utf8.RuneCountInString(tab.Label)
			width := labelWidth
			if i == inst.activeTab {
				width += 2 // brackets []
			}

			// Store tab bounds for click detection
			inst.tabBarBounds = append(inst.tabBarBounds, tabBounds{
				x:      cursor,
				y:      y,
				width:  width,
				height: 1,
				tabID:  tab.ID,
				index:  i,
			})

			// Build tab label
			var tabText string
			if i == inst.activeTab {
				tabText = "[" + tab.Label + "]"
			} else {
				tabText = tab.Label
			}

			// Apply style
			tabStyle := inst.tabStyle
			if i == inst.activeTab && inst.activeTabStyle.FG != "" {
				tabStyle = inst.activeTabStyle
			}
			if tab.Disabled {
				tabStyle.FG = style.Color("gray")
			}

			cmds = append(cmds, paint.NewTextCmd(cursor, y, tabText, tabStyle))
			cursor += width
		}
	}

	return cmds
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

func (inst *Instance) CanHandleAction(actionType string) bool {
	return actionType == string(action.ActionClick) ||
		actionType == "tab_change" ||
		actionType == "navigate" ||
		actionType == "select"
}

func (inst *Instance) HandleAction(actionType string, payload interface{}) bool {
	switch actionType {
	case string(action.ActionClick):
		return inst.handleClick(payload)
	case "tab_change":
		return inst.handleTabChange(payload)
	}
	return false
}

// handleClick handles mouse clicks on tabs
func (inst *Instance) handleClick(payload interface{}) bool {
	// Extract coordinates from payload
 coords, ok := payload.(map[string]interface{})
	if !ok {
		return false
	}

	localX, _ := coords["localX"].(int)
	localY, _ := coords["localY"].(int)

	// Check if click is in tab bar
	if localY != 0 {
		return false
	}

	// Find clicked tab
	for _, tb := range inst.tabBarBounds {
		if localX >= tb.x && localX < tb.x+tb.width {
			tab := inst.tabs[tb.index]
			if !tab.Disabled && tb.index != inst.activeTab {
				inst.SetActiveTab(tb.index)
				inst.emitChangeIntent(tab.ID)
				return true
			}
		}
	}

	return false
}

// handleTabChange handles tab change events
func (inst *Instance) handleTabChange(payload interface{}) bool {
	if index, ok := payload.(int); ok {
		if index >= 0 && index < len(inst.tabs) && !inst.tabs[index].Disabled {
			if index != inst.activeTab {
				inst.SetActiveTab(index)
				inst.emitChangeIntent(inst.tabs[index].ID)
				return true
			}
		}
	}
	return false
}

// HandleKeyMessage handles keyboard navigation
func (inst *Instance) HandleKeyMessage(keyMsg *runtimemsg.KeyMsg) bool {
	switch keyMsg.Special {
	case platform.KeyLeft:
		return inst.PreviousTab()
	case platform.KeyRight:
		return inst.NextTab()
	case platform.KeyHome:
		return inst.FirstTab()
	case platform.KeyEnd:
		return inst.LastTab()
	}
	return false
}

// HandleMouseMessage handles mouse messages
func (inst *Instance) HandleMouseMessage(mouseMsg *runtimemsg.MouseMsg) bool {
	if mouseMsg.Action != runtimemsg.MouseActionPress {
		return false
	}

	if mouseMsg.LocalY == 0 {
		// Check tab bar click
		return inst.handleClick(map[string]interface{}{
			"localX": mouseMsg.LocalX,
			"localY": mouseMsg.LocalY,
		})
	}

	return false
}

// =============================================================================
// Tab Navigation Methods
// =============================================================================

// SetActiveTab sets the active tab by index
func (inst *Instance) SetActiveTab(index int) bool {
	if index < 0 || index >= len(inst.tabs) {
		return false
	}

	tab := inst.tabs[index]
	if tab.Disabled {
		return false
	}

	inst.activeTab = index
	inst.dirty = true
	return true
}

// SetActiveTabByID sets the active tab by ID
func (inst *Instance) SetActiveTabByID(id string) bool {
	for i, tab := range inst.tabs {
		if tab.ID == id && !tab.Disabled {
			return inst.SetActiveTab(i)
		}
	}
	return false
}

// NextTab switches to the next enabled tab
func (inst *Instance) NextTab() bool {
	for i := inst.activeTab + 1; i < len(inst.tabs); i++ {
		if !inst.tabs[i].Disabled {
			tabID := inst.tabs[i].ID
			inst.activeTab = i
			inst.dirty = true
			inst.emitChangeIntent(tabID)
			return true
		}
	}
	return false
}

// PreviousTab switches to the previous enabled tab
func (inst *Instance) PreviousTab() bool {
	for i := inst.activeTab - 1; i >= 0; i-- {
		if !inst.tabs[i].Disabled {
			tabID := inst.tabs[i].ID
			inst.activeTab = i
			inst.dirty = true
			inst.emitChangeIntent(tabID)
			return true
		}
	}
	return false
}

// FirstTab switches to the first enabled tab
func (inst *Instance) FirstTab() bool {
	for i := 0; i < len(inst.tabs); i++ {
		if !inst.tabs[i].Disabled {
			tabID := inst.tabs[i].ID
			inst.activeTab = i
			inst.dirty = true
			inst.emitChangeIntent(tabID)
			return true
		}
	}
	return false
}

// LastTab switches to the last enabled tab
func (inst *Instance) LastTab() bool {
	for i := len(inst.tabs) - 1; i >= 0; i-- {
		if !inst.tabs[i].Disabled {
			tabID := inst.tabs[i].ID
			inst.activeTab = i
			inst.dirty = true
			inst.emitChangeIntent(tabID)
			return true
		}
	}
	return false
}

// emitChangeIntent emits the change intent
func (inst *Instance) emitChangeIntent(tabID string) {
	if inst.changeIntent != nil && inst.intentEmitter != nil {
		inst.intentEmitter(inst.changeIntent)
	}
}

// =============================================================================
// Getters
// =============================================================================

func (inst *Instance) GetActiveTab() int          { return inst.activeTab }
func (inst *Instance) GetActiveTabID() string {
	if inst.activeTab < 0 || inst.activeTab >= len(inst.tabs) {
		return ""
	}
	return inst.tabs[inst.activeTab].ID
}
func (inst *Instance) GetActiveTabLabel() string {
	if inst.activeTab < 0 || inst.activeTab >= len(inst.tabs) {
		return ""
	}
	return inst.tabs[inst.activeTab].Label
}

func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getStringProp(props rtui.Props, key, def string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if v, ok := props[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if v, ok := props[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

func getStyleProp(props rtui.Props, key string) style.Style {
	v, ok := props[key]
	if !ok {
		return style.Style{}
	}
	if s, ok := v.(style.Style); ok {
		return s
	}
	return style.Style{}
}

func getIntentProp(props rtui.Props) intent.Intent {
	v, ok := props["changeIntent"]
	if !ok {
		return nil
	}
	if i, ok := v.(intent.Intent); ok {
		return i
	}
	return nil
}

func getTabsProp(props rtui.Props, def []TabItem) []TabItem {
	v, ok := props["tabs"]
	if !ok {
		return def
	}
	if tabs, ok := v.([]TabItem); ok {
		return tabs
	}
	return def
}

func getTabPositionProp(props rtui.Props, def TabPosition) TabPosition {
	v, ok := props["position"]
	if !ok {
		return def
	}
	if pos, ok := v.(TabPosition); ok {
		return pos
	}
	return def
}

// tabsEqual compares two tab slices
func tabsEqual(a, b []TabItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
