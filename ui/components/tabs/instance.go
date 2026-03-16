package tabs

import (
	"fmt"
	"strings"
	"unicode"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for tabs components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key         string
	componentID string // Phase 7: Component ID for Intent routing

	// === Props (from VNode, may change each render) ===
	tabs              []TabItem
	position          TabPosition
	wrapTabs          bool
	tabGap            int
	loopNavigation    bool
	showHotkeys       bool
	divider           string
	tabStyle          style.Style
	activeTabStyle    style.Style
	disabledTabStyle  style.Style
	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent // For FieldChangeIntent extraction

	// === Layout Props ===
	width  int
	height int
	flex   int

	// === Runtime State ===
	activeTab            int // Current active tab index in inst.tabs
	requestedActiveTab   int
	requestedActiveTabID string
	bounds               [4]int // x, y, w, h
	focused              bool
	dirty                bool

	// === Tab Bar Bounds (for click detection, local coordinates) ===
	tabBarBounds []tabBounds

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

type tabPaintItem struct {
	x         int
	y         int
	text      string
	style     style.Style
	clickable bool
	tabID     string
	index     int
}

type tabLayout struct {
	items  []tabPaintItem
	width  int
	height int
}

type activeBaselineSource int

const (
	activeBaselineComponentTheme activeBaselineSource = iota
	activeBaselineSemanticTheme
	activeBaselineLocalFallback
)

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
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
		key:                  proputil.GetString(props, "key", ""),
		componentID:          proputil.GetString(props, "componentID", ""), // Phase 7
		tabs:                 getTabsProp(props, []TabItem{}),
		position:             getTabPositionProp(props, TabPositionTop),
		wrapTabs:             proputil.GetBool(props, "wrapTabs", false),
		tabGap:               proputil.GetInt(props, "tabGap", 1),
		loopNavigation:       proputil.GetBool(props, "loopNavigation", false),
		showHotkeys:          proputil.GetBool(props, "showHotkeys", false),
		divider:              proputil.GetString(props, "divider", " | "),
		tabStyle:             proputil.GetStyle(props, "tabStyle", style.New()),
		activeTabStyle:       proputil.GetStyle(props, "activeTabStyle", style.New()),
		disabledTabStyle:     proputil.GetStyle(props, "disabledTabStyle", style.New()),
		changeIntent:         proputil.GetIntent(props, "changeIntent", nil),
		changeIntentField:    getChangeIntentFieldProp(props, nil),
		width:                proputil.GetInt(props, "width", 0),
		height:               proputil.GetInt(props, "height", 0),
		flex:                 proputil.GetInt(props, "flex", 1),
		activeTab:            -1,
		requestedActiveTab:   proputil.GetInt(props, "activeTab", -1),
		requestedActiveTabID: proputil.GetString(props, "activeTabID", ""),
		dirty:                true,
	}
	inst.normalizeActiveTab()
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }

// Parent implements TreeComponent interface (Phase 7: Intent Bubble).
// Returns nil as Tabs is a leaf component without parent tracking.
func (inst *Instance) Parent() interface{} {
	return nil
}

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              { inst.tabs = nil }
func (inst *Instance) OnMount()              { inst.dirty = true }
func (inst *Instance) OnUnmount()            {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldTabs := append([]TabItem(nil), inst.tabs...)
	oldPosition := inst.position
	oldWrapTabs := inst.wrapTabs
	oldTabGap := inst.tabGap
	oldLoopNavigation := inst.loopNavigation
	oldShowHotkeys := inst.showHotkeys
	oldDivider := inst.divider
	oldTabStyle := inst.tabStyle
	oldActiveTabStyle := inst.activeTabStyle
	oldDisabledTabStyle := inst.disabledTabStyle
	oldWidth := inst.width
	oldHeight := inst.height
	oldFlex := inst.flex
	oldActiveTab := inst.activeTab
	oldRequestedActiveTab := inst.requestedActiveTab
	oldRequestedActiveTabID := inst.requestedActiveTabID

	inst.componentID = proputil.GetString(props, "componentID", inst.componentID)
	inst.tabs = getTabsProp(props, inst.tabs)
	inst.position = getTabPositionProp(props, inst.position)
	inst.wrapTabs = proputil.GetBool(props, "wrapTabs", inst.wrapTabs)
	inst.tabGap = proputil.GetInt(props, "tabGap", inst.tabGap)
	inst.loopNavigation = proputil.GetBool(props, "loopNavigation", inst.loopNavigation)
	inst.showHotkeys = proputil.GetBool(props, "showHotkeys", inst.showHotkeys)
	inst.divider = proputil.GetString(props, "divider", inst.divider)
	inst.tabStyle = proputil.GetStyle(props, "tabStyle", inst.tabStyle)
	inst.activeTabStyle = proputil.GetStyle(props, "activeTabStyle", inst.activeTabStyle)
	inst.disabledTabStyle = proputil.GetStyle(props, "disabledTabStyle", inst.disabledTabStyle)
	inst.changeIntent = proputil.GetIntent(props, "changeIntent", inst.changeIntent)
	inst.changeIntentField = getChangeIntentFieldProp(props, inst.changeIntentField)
	inst.width = proputil.GetInt(props, "width", inst.width)
	inst.height = proputil.GetInt(props, "height", inst.height)
	inst.flex = proputil.GetInt(props, "flex", inst.flex)
	inst.requestedActiveTab = proputil.GetInt(props, "activeTab", inst.requestedActiveTab)
	inst.requestedActiveTabID = proputil.GetString(props, "activeTabID", inst.requestedActiveTabID)

	inst.normalizeActiveTab()

	changed := !tabsEqual(oldTabs, inst.tabs) ||
		oldPosition != inst.position ||
		oldWrapTabs != inst.wrapTabs ||
		oldTabGap != inst.tabGap ||
		oldLoopNavigation != inst.loopNavigation ||
		oldShowHotkeys != inst.showHotkeys ||
		oldDivider != inst.divider ||
		oldTabStyle != inst.tabStyle ||
		oldActiveTabStyle != inst.activeTabStyle ||
		oldDisabledTabStyle != inst.disabledTabStyle ||
		oldWidth != inst.width ||
		oldHeight != inst.height ||
		oldFlex != inst.flex ||
		oldActiveTab != inst.activeTab ||
		oldRequestedActiveTab != inst.requestedActiveTab ||
		oldRequestedActiveTabID != inst.requestedActiveTabID
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:       inst.key,
		propTabs:      inst.tabs,
		propPosition:  inst.position,
		propActiveTab: inst.activeTab,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()                        { inst.dirty = false }

// =============================================================================
// FocusableInstance Interface
// =============================================================================

func (inst *Instance) SetFocus(focused bool) {
	if inst.focused == focused {
		return
	}
	inst.focused = focused
	inst.dirty = true
}

func (inst *Instance) HasFocus() bool { return inst.focused }

func (inst *Instance) IsDisabled() bool {
	return inst.firstEnabledVisibleIndex() < 0
}

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout measurement.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	layoutWidthHint := 0
	if inst.wrapTabs && (inst.position == TabPositionTop || inst.position == TabPositionBottom) {
		layoutWidthHint = inst.width
		if layoutWidthHint == 0 && constraints.MaxWidth > 0 {
			layoutWidthHint = constraints.MaxWidth
		}
	}

	tabLayout := inst.computeTabLayout(layoutWidthHint)

	width := inst.width
	height := inst.height

	if width == 0 {
		width = tabLayout.width
	}
	if height == 0 {
		height = tabLayout.height
	}

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

// calculateTabBarWidth calculates the natural width of the tab bar.
func (inst *Instance) calculateTabBarWidth() int {
	return inst.computeTabLayout(0).width
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint renders the tabs component.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	return inst.paintTabBar(x, y)
}

// paintTabBar renders the tab bar.
func (inst *Instance) paintTabBar(x, y int) []paint.DrawCmd {
	inst.tabBarBounds = inst.tabBarBounds[:0]

	availableWidth, availableHeight := inst.currentRenderSpace()
	layoutWidthHint := availableWidth
	if !inst.wrapTabs || !(inst.position == TabPositionTop || inst.position == TabPositionBottom) {
		layoutWidthHint = 0
	}

	tabLayout := inst.computeTabLayout(layoutWidthHint)
	if len(tabLayout.items) == 0 {
		return nil
	}

	renderWidth, renderHeight := inst.resolveRenderSpace(tabLayout.width, tabLayout.height)
	offsetX, offsetY := inst.layoutOrigin(tabLayout, renderWidth, renderHeight)

	cmds := make([]paint.DrawCmd, 0, len(tabLayout.items))
	for _, item := range tabLayout.items {
		localX := offsetX + item.x
		localY := offsetY + item.y
		cmds = append(cmds, paint.NewTextCmd(x+localX, y+localY, item.text, item.style))
		if item.clickable {
			inst.tabBarBounds = append(inst.tabBarBounds, tabBounds{
				x:      localX,
				y:      localY,
				width:  paint.StringWidth(item.text),
				height: 1,
				tabID:  item.tabID,
				index:  item.index,
			})
		}
	}

	_ = availableHeight // keeps parity with local render computation for future expansion
	return cmds
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	switch act.Type {
	case action.ActionClick:
		return inst.handleClick(act.Payload)
	case action.ActionNavigateLeft, action.ActionNavigateUp, action.ActionNavigatePrev:
		return inst.PreviousTab()
	case action.ActionNavigateRight, action.ActionNavigateDown, action.ActionNavigateNext:
		return inst.NextTab()
	case action.ActionNavigateFirst, action.ActionNavigateHome:
		return inst.FirstTab()
	case action.ActionNavigateLast, action.ActionNavigateEnd:
		return inst.LastTab()
	case action.ActionInputText:
		switch payload := act.Payload.(type) {
		case *runtimemsg.KeyMsg:
			return inst.HandleKeyMessage(payload)
		case runtimemsg.KeyMsg:
			return inst.HandleKeyMessage(&payload)
		case string:
			runes := []rune(payload)
			if len(runes) != 1 {
				return false
			}
			return inst.HandleKeyMessage(runtimemsg.NewKeyMsg(runes[0], platform.KeyUnknown, runtimemsg.Modifiers{}))
		}
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
		if shouldHandleIntent(inst.componentID, v.ComponentID) {
			return inst.NextTab()
		}
		return false

	case TabPreviousIntent:
		if shouldHandleIntent(inst.componentID, v.ComponentID) {
			return inst.PreviousTab()
		}
		return false

	case TabSelectIntent:
		if !shouldHandleIntent(inst.componentID, v.ComponentID) {
			return false
		}
		oldActive := inst.activeTab
		var changed bool
		if v.TabID != "" {
			changed = inst.SetActiveTabByID(v.TabID)
		} else {
			changed = inst.SetActiveTab(v.TabIndex)
		}
		if changed && inst.activeTab != oldActive {
			inst.emitChangeIntent(inst.tabs[inst.activeTab].ID)
		}
		return changed

	default:
		return false
	}
}

// shouldHandleIntent checks if this tabs instance should handle the intent.
// Returns true if componentID matches or if both are empty.
func shouldHandleIntent(myID, intentID string) bool {
	if myID == "" || intentID == "" {
		return true
	}
	return myID == intentID
}

// =============================================================================
// Mouse and Keyboard Handling
// =============================================================================

// handleClick handles mouse clicks on tabs.
func (inst *Instance) handleClick(payload interface{}) bool {
	localX, localY, ok := inst.localClickCoordinates(payload)
	if !ok {
		return false
	}

	for _, tb := range inst.tabBarBounds {
		if localX >= tb.x && localX < tb.x+tb.width &&
			localY >= tb.y && localY < tb.y+tb.height {
			tab := inst.tabs[tb.index]
			if tab.Disabled || tab.Hidden || tb.index == inst.activeTab {
				return false
			}
			if inst.SetActiveTab(tb.index) {
				inst.emitChangeIntent(tab.ID)
				return true
			}
		}
	}

	return false
}

func (inst *Instance) localClickCoordinates(payload interface{}) (int, int, bool) {
	switch v := payload.(type) {
	case *runtimemsg.MouseMsg:
		if v == nil {
			return 0, 0, false
		}
		return v.LocalX, v.LocalY, true
	case runtimemsg.MouseMsg:
		return v.LocalX, v.LocalY, true
	case map[string]interface{}:
		localX, okX := v["localX"].(int)
		localY, okY := v["localY"].(int)
		if !okX || !okY {
			return 0, 0, false
		}
		return localX, localY, true
	default:
		return 0, 0, false
	}
}

// HandleKeyMessage handles keyboard navigation.
func (inst *Instance) HandleKeyMessage(keyMsg *runtimemsg.KeyMsg) bool {
	if keyMsg == nil {
		return false
	}

	if keyMsg.IsTab() && keyMsg.HasCtrl() {
		if keyMsg.HasShift() {
			return inst.PreviousTab()
		}
		return inst.NextTab()
	}

	switch keyMsg.Special {
	case platform.KeyLeft, platform.KeyUp:
		return inst.PreviousTab()
	case platform.KeyRight, platform.KeyDown:
		return inst.NextTab()
	case platform.KeyHome:
		return inst.FirstTab()
	case platform.KeyEnd:
		return inst.LastTab()
	}

	if keyMsg.HasModifier() {
		return false
	}

	if keyMsg.Rune != 0 && inst.selectByHotkey(keyMsg.Rune) {
		return true
	}

	if keyMsg.Rune >= '1' && keyMsg.Rune <= '9' {
		return inst.SetActiveVisibleOrdinal(int(keyMsg.Rune - '1'))
	}

	return false
}

// HandleMouseMessage handles mouse messages.
func (inst *Instance) HandleMouseMessage(mouseMsg *runtimemsg.MouseMsg) bool {
	if mouseMsg == nil || mouseMsg.Action != runtimemsg.MouseActionPress {
		return false
	}

	return inst.handleClick(map[string]interface{}{
		"localX": mouseMsg.LocalX,
		"localY": mouseMsg.LocalY,
	})
}

// =============================================================================
// Tab Navigation Methods
// =============================================================================

// SetActiveTab sets the active tab by index.
func (inst *Instance) SetActiveTab(index int) bool {
	if !inst.isSelectableIndex(index) {
		return false
	}

	if inst.activeTab == index {
		return true
	}

	inst.activeTab = index
	inst.dirty = true
	return true
}

// SetActiveVisibleOrdinal selects the nth visible enabled tab (0-based).
func (inst *Instance) SetActiveVisibleOrdinal(ordinal int) bool {
	visible := inst.visibleEnabledIndices()
	if ordinal < 0 || ordinal >= len(visible) {
		return false
	}
	tabIndex := visible[ordinal]
	if !inst.SetActiveTab(tabIndex) {
		return false
	}
	inst.emitChangeIntent(inst.tabs[tabIndex].ID)
	return true
}

// SetActiveTabByID sets the active tab by ID.
func (inst *Instance) SetActiveTabByID(id string) bool {
	idx := inst.FindTabByID(id)
	if idx < 0 {
		return false
	}
	return inst.SetActiveTab(idx)
}

// SetActiveTabByLabel sets the active tab by label.
func (inst *Instance) SetActiveTabByLabel(label string) bool {
	idx := inst.FindTabByLabel(label)
	if idx < 0 {
		return false
	}
	return inst.SetActiveTab(idx)
}

// NextTab switches to the next enabled visible tab.
func (inst *Instance) NextTab() bool {
	nextIndex, ok := inst.findAdjacentSelectableIndex(1, inst.loopNavigation)
	if !ok {
		return false
	}
	if !inst.SetActiveTab(nextIndex) {
		return false
	}
	inst.emitChangeIntent(inst.tabs[nextIndex].ID)
	return true
}

// PreviousTab switches to the previous enabled visible tab.
func (inst *Instance) PreviousTab() bool {
	prevIndex, ok := inst.findAdjacentSelectableIndex(-1, inst.loopNavigation)
	if !ok {
		return false
	}
	if !inst.SetActiveTab(prevIndex) {
		return false
	}
	inst.emitChangeIntent(inst.tabs[prevIndex].ID)
	return true
}

// FirstTab switches to the first enabled visible tab.
func (inst *Instance) FirstTab() bool {
	first := inst.firstEnabledVisibleIndex()
	if first < 0 {
		return false
	}
	if !inst.SetActiveTab(first) {
		return false
	}
	inst.emitChangeIntent(inst.tabs[first].ID)
	return true
}

// LastTab switches to the last enabled visible tab.
func (inst *Instance) LastTab() bool {
	last := inst.lastEnabledVisibleIndex()
	if last < 0 {
		return false
	}
	if !inst.SetActiveTab(last) {
		return false
	}
	inst.emitChangeIntent(inst.tabs[last].ID)
	return true
}

// GetActiveTab returns the active tab index.
func (inst *Instance) GetActiveTab() int { return inst.activeTab }

// GetActiveTabID returns the active tab ID.
func (inst *Instance) GetActiveTabID() string {
	if !inst.isSelectableOrPresentIndex(inst.activeTab) {
		return ""
	}
	return inst.tabs[inst.activeTab].ID
}

// GetActiveTabLabel returns the active tab label.
func (inst *Instance) GetActiveTabLabel() string {
	if !inst.isSelectableOrPresentIndex(inst.activeTab) {
		return ""
	}
	return inst.tabs[inst.activeTab].Label
}

// GetTabCount returns the total number of tabs, including hidden ones.
func (inst *Instance) GetTabCount() int {
	return len(inst.tabs)
}

// GetVisibleTabCount returns the number of visible tabs.
func (inst *Instance) GetVisibleTabCount() int {
	count := 0
	for _, tab := range inst.tabs {
		if !tab.Hidden {
			count++
		}
	}
	return count
}

// CanGoNext reports whether a next enabled visible tab exists.
func (inst *Instance) CanGoNext() bool {
	_, ok := inst.findAdjacentSelectableIndex(1, inst.loopNavigation)
	return ok
}

// CanGoPrevious reports whether a previous enabled visible tab exists.
func (inst *Instance) CanGoPrevious() bool {
	_, ok := inst.findAdjacentSelectableIndex(-1, inst.loopNavigation)
	return ok
}

// FindTabByID returns the matching tab index or -1.
func (inst *Instance) FindTabByID(id string) int {
	for i, tab := range inst.tabs {
		if tab.ID == id {
			return i
		}
	}
	return -1
}

// FindTabByLabel returns the matching tab index or -1.
func (inst *Instance) FindTabByLabel(label string) int {
	for i, tab := range inst.tabs {
		if tab.Label == label {
			return i
		}
	}
	return -1
}

// GetTabByIndex returns the tab and whether it exists.
func (inst *Instance) GetTabByIndex(index int) (TabItem, bool) {
	if index < 0 || index >= len(inst.tabs) {
		return TabItem{}, false
	}
	return inst.tabs[index], true
}

// IsTabEnabled reports whether a tab index is enabled.
func (inst *Instance) IsTabEnabled(index int) bool {
	if index < 0 || index >= len(inst.tabs) {
		return false
	}
	return !inst.tabs[index].Disabled
}

// SetTabEnabled toggles a tab's enabled state.
func (inst *Instance) SetTabEnabled(index int, enabled bool) bool {
	if index < 0 || index >= len(inst.tabs) {
		return false
	}

	disabled := !enabled
	if inst.tabs[index].Disabled == disabled {
		return true
	}

	oldActive := inst.activeTab
	inst.tabs[index].Disabled = disabled
	inst.normalizeActiveTab()
	inst.dirty = true

	if oldActive != inst.activeTab && inst.isSelectableOrPresentIndex(inst.activeTab) {
		inst.emitChangeIntent(inst.tabs[inst.activeTab].ID)
	}
	return true
}

// GetPosition returns the current tab bar position.
func (inst *Instance) GetPosition() TabPosition {
	return inst.position
}

// SetPosition updates the runtime tab bar position.
func (inst *Instance) SetPosition(pos TabPosition) bool {
	if inst.position == pos {
		return false
	}
	inst.position = pos
	inst.dirty = true
	return true
}

// GetBounds returns the component bounds.
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

// SetBounds stores the component bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// SetIntentEmitter sets the runtime intent emitter.
func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

// =============================================================================
// Intent Emission
// =============================================================================

// emitChangeIntent emits the change intent.
// Phase 7: Use intent.Emit to enable Intent Bubble.
func (inst *Instance) emitChangeIntent(tabID string) {
	if inst.componentID != "" {
		changeIntent := TabChange(
			inst.componentID,
			inst.activeTab,
			tabID,
			inst.GetActiveTabLabel(),
		)
		intent.Emit(inst, changeIntent)
		inst.emitOptionalGlobalIntent(changeIntent)
	}

	if inst.intentEmitter == nil {
		return
	}

	if inst.changeIntentField != nil {
		changeIntent := intent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: fmt.Sprintf("%d", inst.activeTab),
		}
		inst.intentEmitter(changeIntent)
		return
	}

	if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
}

func (inst *Instance) emitOptionalGlobalIntent(i intent.Intent) {
	if i == nil {
		return
	}
	runtime := rtui.GetGlobalIntentRuntime()
	if runtime == nil || runtime.Registry == nil {
		return
	}
	intentType := i.IntentType()
	if !runtime.Registry.HasHandler(intentType) && !runtime.Registry.HasFallback() {
		return
	}
	rtui.EmitIntentGlobal(i)
}

// =============================================================================
// Layout Helpers
// =============================================================================

func (inst *Instance) computeTabLayout(maxWidth int) tabLayout {
	visible := inst.visibleTabIndices()
	if len(visible) == 0 {
		return tabLayout{}
	}

	switch inst.position {
	case TabPositionLeft, TabPositionRight:
		return inst.computeVerticalLayout(visible)
	default:
		return inst.computeHorizontalLayout(visible, maxWidth)
	}
}

func (inst *Instance) computeHorizontalLayout(visible []int, maxWidth int) tabLayout {
	layout := tabLayout{}
	cursorX := 0
	cursorY := 0
	lineWidth := 0
	separator := inst.divider
	separatorWidth := paint.StringWidth(separator)

	for _, tabIndex := range visible {
		tab := inst.tabs[tabIndex]
		text := inst.renderTabText(tabIndex)
		textWidth := paint.StringWidth(text)

		needsSeparator := cursorX > 0 && separator != ""
		nextWidth := textWidth
		if needsSeparator {
			nextWidth += separatorWidth
		}

		if inst.wrapTabs && maxWidth > 0 && cursorX > 0 && cursorX+nextWidth > maxWidth {
			lineWidth = max(lineWidth, cursorX)
			cursorX = 0
			cursorY++
			needsSeparator = false
		}

		if needsSeparator {
			layout.items = append(layout.items, tabPaintItem{
				x:     cursorX,
				y:     cursorY,
				text:  separator,
				style: inst.resolveDividerStyle(),
			})
			cursorX += separatorWidth
		}

		layout.items = append(layout.items, tabPaintItem{
			x:         cursorX,
			y:         cursorY,
			text:      text,
			style:     inst.resolveTabStyle(tabIndex),
			clickable: true,
			tabID:     tab.ID,
			index:     tabIndex,
		})
		cursorX += textWidth
		lineWidth = max(lineWidth, cursorX)
	}

	layout.width = lineWidth
	layout.height = cursorY + 1
	return layout
}

func (inst *Instance) computeVerticalLayout(visible []int) tabLayout {
	layout := tabLayout{}
	gap := max(0, inst.tabGap)
	cursorY := 0

	for i, tabIndex := range visible {
		text := inst.renderTabText(tabIndex)
		textWidth := paint.StringWidth(text)

		layout.items = append(layout.items, tabPaintItem{
			x:         0,
			y:         cursorY,
			text:      text,
			style:     inst.resolveTabStyle(tabIndex),
			clickable: true,
			tabID:     inst.tabs[tabIndex].ID,
			index:     tabIndex,
		})

		layout.width = max(layout.width, textWidth)
		cursorY++
		if i < len(visible)-1 {
			cursorY += gap
		}
	}

	layout.height = max(1, cursorY)
	return layout
}

func (inst *Instance) resolveDividerStyle() style.Style {
	return style.NewStyle().Foreground(style.BrightBlack).Merge(inst.tabStyle)
}

func (inst *Instance) resolveActiveBaseline() (style.Style, activeBaselineSource) {
	componentSelected := style.GetStyle("tabs", "select")
	if inst.isProtectableComponentSelected(componentSelected) {
		return componentSelected, activeBaselineComponentTheme
	}

	selectColor := fwtheme.Select()
	backgroundColor := fwtheme.BG()
	if isUsableStyleColor(selectColor) && isUsableStyleColor(backgroundColor) {
		return style.NewStyle().
			Foreground(backgroundColor).
			Background(selectColor).
			Bold(true), activeBaselineSemanticTheme
	}

	return style.NewStyle().Reverse(true).Bold(true), activeBaselineLocalFallback
}

func (inst *Instance) isProtectableComponentSelected(selected style.Style) bool {
	if selected.IsEmpty() {
		return false
	}
	if isUsableStyleColor(selected.FG) {
		return true
	}
	if isUsableStyleColor(selected.BG) && isUsableStyleColor(fwtheme.BG()) {
		return true
	}
	return selected.IsReverse()
}

func (inst *Instance) resolveDisabledBaseline() style.Style {
	if disabledFG := fwtheme.DisabledFG(); isUsableStyleColor(disabledFG) {
		return style.NewStyle().Foreground(disabledFG)
	}
	return style.NewStyle().Italic(true)
}

func (inst *Instance) protectedActiveForeground(source activeBaselineSource, baseline style.Style) style.Color {
	switch {
	case source == activeBaselineComponentTheme && isUsableStyleColor(baseline.FG):
		return baseline.FG
	case source == activeBaselineComponentTheme && isUsableStyleColor(baseline.BG):
		if contrastFG := fwtheme.BG(); isUsableStyleColor(contrastFG) {
			return contrastFG
		}
	case source == activeBaselineSemanticTheme:
		if contrastFG := fwtheme.BG(); isUsableStyleColor(contrastFG) {
			return contrastFG
		}
	}
	return style.NoColor
}

func (inst *Instance) isReverseOnlyBaseline(source activeBaselineSource, baseline style.Style) bool {
	return baseline.IsReverse() &&
		!isUsableStyleColor(baseline.FG) &&
		!isUsableStyleColor(baseline.BG) &&
		(source == activeBaselineComponentTheme || source == activeBaselineLocalFallback)
}

func isUsableStyleColor(color style.Color) bool {
	return color != style.NoColor
}

func (inst *Instance) resolveTabStyle(index int) style.Style {
	tab := inst.tabs[index]

	switch {
	case tab.Disabled:
		return inst.resolveDisabledBaseline().
			Merge(inst.tabStyle).
			Merge(inst.disabledTabStyle)
	case index == inst.activeTab:
		baseline, source := inst.resolveActiveBaseline()
		activeStyle := style.NewStyle().Merge(inst.tabStyle)
		if inst.isReverseOnlyBaseline(source, baseline) {
			activeStyle.FG = style.NoColor
			activeStyle.BG = style.NoColor
		}
		activeStyle = activeStyle.Merge(baseline).Merge(inst.activeTabStyle)
		if inst.isReverseOnlyBaseline(source, baseline) {
			activeStyle.FG = style.NoColor
			activeStyle.BG = style.NoColor
		}
		if protectedFG := inst.protectedActiveForeground(source, baseline); isUsableStyleColor(protectedFG) {
			activeStyle.FG = protectedFG
		}
		return activeStyle
	default:
		return style.NewStyle().
			Merge(inst.tabStyle)
	}
}

func (inst *Instance) renderTabText(index int) string {
	tab := inst.tabs[index]
	parts := make([]string, 0, 4)

	if inst.showHotkeys && tab.Hotkey != 0 {
		parts = append(parts, fmt.Sprintf("{%c}", unicode.ToUpper(tab.Hotkey)))
	}
	if tab.Icon != "" {
		parts = append(parts, tab.Icon)
	}

	label := tab.Label
	if label == "" {
		label = tab.ID
	}
	parts = append(parts, label)

	if tab.Badge != "" {
		parts = append(parts, fmt.Sprintf("(%s)", tab.Badge))
	}

	text := strings.Join(parts, " ")
	if index == inst.activeTab {
		return "[" + text + "]"
	}
	return text
}

func (inst *Instance) layoutOrigin(layout tabLayout, renderWidth, renderHeight int) (int, int) {
	switch inst.position {
	case TabPositionBottom:
		return 0, max(0, renderHeight-layout.height)
	case TabPositionRight:
		return max(0, renderWidth-layout.width), 0
	default:
		return 0, 0
	}
}

func (inst *Instance) currentRenderSpace() (int, int) {
	width := inst.width
	height := inst.height
	if inst.bounds[2] > 0 {
		width = inst.bounds[2]
	}
	if inst.bounds[3] > 0 {
		height = inst.bounds[3]
	}
	return width, height
}

func (inst *Instance) resolveRenderSpace(fallbackWidth, fallbackHeight int) (int, int) {
	width, height := inst.currentRenderSpace()
	if width == 0 {
		width = fallbackWidth
	}
	if height == 0 {
		height = fallbackHeight
	}
	return width, height
}

// =============================================================================
// State Helpers
// =============================================================================

func (inst *Instance) normalizeActiveTab() {
	if len(inst.tabs) == 0 {
		inst.activeTab = -1
		return
	}

	if inst.requestedActiveTabID != "" {
		if idx := inst.FindTabByID(inst.requestedActiveTabID); inst.isSelectableIndex(idx) {
			inst.activeTab = idx
			return
		}
		inst.activeTab = inst.firstEnabledVisibleIndex()
		return
	}

	if inst.requestedActiveTab >= 0 {
		if inst.isSelectableIndex(inst.requestedActiveTab) {
			inst.activeTab = inst.requestedActiveTab
			return
		}
		inst.activeTab = inst.firstEnabledVisibleIndex()
		return
	}

	if inst.isSelectableIndex(inst.activeTab) {
		return
	}

	inst.activeTab = inst.firstEnabledVisibleIndex()
}

func (inst *Instance) visibleTabIndices() []int {
	indices := make([]int, 0, len(inst.tabs))
	for i, tab := range inst.tabs {
		if !tab.Hidden {
			indices = append(indices, i)
		}
	}
	return indices
}

func (inst *Instance) visibleEnabledIndices() []int {
	indices := make([]int, 0, len(inst.tabs))
	for i, tab := range inst.tabs {
		if !tab.Hidden && !tab.Disabled {
			indices = append(indices, i)
		}
	}
	return indices
}

func (inst *Instance) firstEnabledVisibleIndex() int {
	for i, tab := range inst.tabs {
		if !tab.Hidden && !tab.Disabled {
			return i
		}
	}
	return -1
}

func (inst *Instance) lastEnabledVisibleIndex() int {
	for i := len(inst.tabs) - 1; i >= 0; i-- {
		tab := inst.tabs[i]
		if !tab.Hidden && !tab.Disabled {
			return i
		}
	}
	return -1
}

func (inst *Instance) findAdjacentSelectableIndex(direction int, wrap bool) (int, bool) {
	visible := inst.visibleEnabledIndices()
	if len(visible) == 0 {
		return -1, false
	}
	if len(visible) == 1 && visible[0] == inst.activeTab {
		return -1, false
	}

	currentPos := -1
	for pos, idx := range visible {
		if idx == inst.activeTab {
			currentPos = pos
			break
		}
	}

	if currentPos == -1 {
		if direction < 0 {
			return visible[len(visible)-1], true
		}
		return visible[0], true
	}

	nextPos := currentPos + direction
	if nextPos < 0 || nextPos >= len(visible) {
		if !wrap {
			return -1, false
		}
		if direction < 0 {
			nextPos = len(visible) - 1
		} else {
			nextPos = 0
		}
	}

	if visible[nextPos] == inst.activeTab {
		return -1, false
	}
	return visible[nextPos], true
}

func (inst *Instance) selectByHotkey(r rune) bool {
	target := unicode.ToLower(r)
	for i, tab := range inst.tabs {
		if tab.Hidden || tab.Disabled || tab.Hotkey == 0 {
			continue
		}
		if unicode.ToLower(tab.Hotkey) != target {
			continue
		}
		if !inst.SetActiveTab(i) {
			return false
		}
		inst.emitChangeIntent(tab.ID)
		return true
	}
	return false
}

func (inst *Instance) isSelectableIndex(index int) bool {
	if index < 0 || index >= len(inst.tabs) {
		return false
	}
	tab := inst.tabs[index]
	return !tab.Disabled && !tab.Hidden
}

func (inst *Instance) isSelectableOrPresentIndex(index int) bool {
	return index >= 0 && index < len(inst.tabs)
}

// =============================================================================
// Helper Functions
// =============================================================================

// tabsEqual checks if two slices of TabItem are equal.
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

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getTabPositionProp(props rtui.Props, def TabPosition) TabPosition {
	if v, ok := props[propPosition]; ok {
		if tp, ok := v.(TabPosition); ok {
			return tp
		}
	}
	return def
}

func getChangeIntentFieldProp(props rtui.Props, def intent.FieldIntent) intent.FieldIntent {
	if v, ok := props[propChangeIntentField]; ok {
		if intentField, ok := v.(intent.FieldIntent); ok {
			return intentField
		}
	}
	return def
}

func getTabsProp(props rtui.Props, def []TabItem) []TabItem {
	if v, ok := props[propTabs]; ok {
		if tabs, ok := v.([]TabItem); ok {
			return tabs
		}
	}
	return def
}
