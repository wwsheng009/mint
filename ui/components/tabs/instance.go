package tabs

import (
	"fmt"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime/action"
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
	key          string
	componentID string // Phase 7: Component ID for Intent routing

	// === Props (from VNode, may change each render) ===
	tabs      []TabItem
	position  TabPosition
	wrapTabs  bool
	tabGap    int
	tabStyle      style.Style
	activeTabStyle style.Style
	changeIntent intent.Intent
	changeIntentField intent.FieldIntent  // For FieldChangeIntent extraction

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
	_ intent.IntentHandler       = (*Instance)(nil) // Phase 7: Intent Bubble
	_ intent.TreeComponent       = (*Instance)(nil) // Phase 7: Intent Bubble
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
		key:          getStringProp(props, "key", ""),
		componentID:  getStringProp(props, "componentID", ""), // Phase 7
		tabs:         getTabsProp(props, []TabItem{}),
		position:     getTabPositionProp(props, TabPositionTop),
		wrapTabs:     getBoolProp(props, "wrapTabs", false),
		tabGap:       getIntProp(props, "tabGap", 1),
		tabStyle:     getStyleProp(props, "tabStyle"),
		activeTabStyle: getStyleProp(props, "activeTabStyle"),
		changeIntent: getIntentProp(props),
		changeIntentField: getChangeIntentFieldProp(props),
		width:        getIntProp(props, "width", 0),
		height:       getIntProp(props, "height", 0),
		flex:         getIntProp(props, "flex", 1),
		activeTab:    0,
		dirty:        true,
	}
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }

// Parent implements TreeComponent interface (Phase 7: Intent Bubble).
// Returns nil as Tabs is a leaf component without parent tracking.
func (inst *Instance) Parent() interface{} {
	return nil
}

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()             { inst.tabs = nil }
func (inst *Instance) OnMount()             { inst.dirty = true }
func (inst *Instance) OnUnmount()           {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldTabs := inst.tabs
	oldPosition := inst.position

	inst.componentID = getStringProp(props, "componentID", inst.componentID) // Phase 7
	inst.tabs = getTabsProp(props, inst.tabs)
	inst.position = getTabPositionProp(props, inst.position)
	inst.wrapTabs = getBoolProp(props, "wrapTabs", inst.wrapTabs)
	inst.tabGap = getIntProp(props, "tabGap", inst.tabGap)
	inst.tabStyle = getStyleProp(props, "tabStyle")
	inst.activeTabStyle = getStyleProp(props, "activeTabStyle")
	inst.changeIntent = getIntentProp(props)
	inst.changeIntentField = getChangeIntentFieldProp(props)
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

// Paint renders the tabs component.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd

	// Draw tab bar
	cmds = append(cmds, inst.paintTabBar(x, y)...)

	return cmds
}

// paintTabBar renders the tab bar
func (inst *Instance) paintTabBar(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd

	// Clear tab bar bounds
	inst.tabBarBounds = inst.tabBarBounds[:0]

	if len(inst.tabs) == 0 {
		return cmds
	}

	cursor := x
	tabBarY := y

	switch inst.position {
	case TabPositionTop, TabPositionBottom:
		// Horizontal tab bar
		if !inst.wrapTabs {
			for i, tab := range inst.tabs {
				tabText := tab.Label
				if i == inst.activeTab {
					tabText = fmt.Sprintf("[%s]", tabText)
				}

				tabStyle := inst.tabStyle
				if i == inst.activeTab {
					tabStyle = inst.activeTabStyle
				}
				if tabStyle.FG == "" {
					tabStyle.FG = style.Color("cyan")
				}

				width := utf8.RuneCountInString(tabText)

				// Add separator
				if i > 0 {
					cmds = append(cmds, paint.NewTextCmd(cursor-2, tabBarY, " | ", tabStyle))
				}

				cmds = append(cmds, paint.NewTextCmd(cursor, tabBarY, tabText, tabStyle))

				// Store bounds for click detection
				boundX := cursor
				if i > 0 {
					boundX = cursor - 2 // Include separator
				}
				inst.tabBarBounds = append(inst.tabBarBounds, tabBounds{
					x:      boundX,
					y:      tabBarY,
					width:  width + 2, // Include separator
					height: 1,
					tabID:  tab.ID,
					index:  i,
				})

				cursor += width
			}
		}
	}

	return cmds
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

func (inst *Instance) HandleAction(act *action.Action) bool {
	switch act.Type {
	case action.ActionClick:
		return inst.handleClick(act.Payload)
	}
	return false
}

// =============================================================================
// IntentHandler Interface (Phase 7: Intent Bubble)
// =============================================================================

// HandleIntent implements IntentHandler to handle Tabs-specific intents.
func (inst *Instance) HandleIntent(i intent.Intent) bool {
	switch v := i.(type) {
	case TabNextIntent:
		// Handle TabNext: switch to the next tab
		if shouldHandleIntent(inst.componentID, v.ComponentID) {
			return inst.NextTab()
		}
		return false

	case TabPreviousIntent:
		// Handle TabPrevious: switch to the previous tab
		if shouldHandleIntent(inst.componentID, v.ComponentID) {
			return inst.PreviousTab()
		}
		return false

	default:
		// Intent not handled by this tabs instance
		return false
	}
}

// shouldHandleIntent checks if this tabs instance should handle the intent.
// Returns true if componentID matches or if both are empty.
func shouldHandleIntent(myID, intentID string) bool {
	if myID == "" || intentID == "" {
		// If no component ID is set, handle all intents (backward compatibility)
		return true
	}
	return myID == intentID
}

// =============================================================================
// Mouse and Keyboard Handling
// =============================================================================

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
// Phase 7: Use intent.Emit to enable Intent Bubble
func (inst *Instance) emitChangeIntent(tabID string) {
	if inst.intentEmitter != nil {
		// Priority: Emit Tabs-specific TabChangeIntent
		if inst.componentID != "" {
			changeIntent := TabChange(
				inst.componentID,
				inst.activeTab,
				tabID,
				inst.GetActiveTabLabel(),
			)
			intent.Emit(inst, changeIntent)
		}

		// Fallback: Use FieldChangeIntent mode
		if inst.changeIntentField != nil {
			changeIntent := intent.FieldChangeIntent{
				Field: inst.changeIntentField.GetField(),
				Value: fmt.Sprintf("%d", inst.activeTab),
			}
			inst.intentEmitter(changeIntent)
		} else if inst.changeIntent != nil {
			// Fallback to original intent mode
			inst.intentEmitter(inst.changeIntent)
		}
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
// Helper Functions
// =============================================================================

// tabsEqual checks if two slices of TabItem are equal
func tabsEqual(a, b []TabItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Label != b[i].Label || a[i].ID != b[i].ID || a[i].Disabled != b[i].Disabled {
			return false
		}
	}
	return true
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

func getTabPositionProp(props rtui.Props, def TabPosition) TabPosition {
	if v, ok := props["position"]; ok {
		if tp, ok := v.(TabPosition); ok {
			return tp
		}
	}
	return def
}

func getStyleProp(props rtui.Props, key string) style.Style {
	if v, ok := props[key]; ok {
		if s, ok := v.(style.Style); ok {
			return s
		}
	}
	return style.New()
}

func getIntentProp(props rtui.Props) intent.Intent {
	if v, ok := props["changeIntent"]; ok {
		if i, ok := v.(intent.Intent); ok {
			return i
		}
	}
	return nil
}

func getChangeIntentFieldProp(props rtui.Props) intent.FieldIntent {
	if v, ok := props["changeIntent"]; ok {
		if intentField, ok := v.(intent.FieldIntent); ok {
			return intentField
		}
	}
	return nil
}

func getTabsProp(props rtui.Props, def []TabItem) []TabItem {
	if v, ok := props["tabs"]; ok {
		if tabs, ok := v.([]TabItem); ok {
			return tabs
		}
	}
	return def
}
