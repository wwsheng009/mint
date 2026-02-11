package navigation

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/components/layout"
	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/framework/cmd"
	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/ui"
)

// Interface implementation assertions
var _ frameworkevent.Component = (*TabsVNode)(nil)
var _ component.Updater = (*TabsVNode)(nil) // Phase 3: Msg/Cmd support
var _ action.ActionTarget = (*TabsVNode)(nil)
var _ action.ScrollableActionTarget = (*TabsVNode)(nil)
var _ action.SelectableActionTarget = (*TabsVNode)(nil)

// TabPosition defines where tabs are positioned relative to content
type TabPosition int

const (
	TabPositionTop    TabPosition = iota // Tabs above content
	TabPositionBottom                    // Tabs below content
	TabPositionLeft                      // Tabs to the left of content
	TabPositionRight                     // Tabs to the right of content
)

// Tab represents a single tab in a Tabs component
type TabItem struct {
	ID       string
	Label    string
	Content  ui.VNode // Content to show when tab is active
	Key      string   // Optional key for diffing
	Disabled bool     // Whether the tab is disabled
}

// =============================================================================
// Enhanced Tabs Component
// =============================================================================

// TabsVNode represents a tabs component
type TabsVNode struct {
	*ui.ElementVNode
	tabs      []TabItem
	activeTab int
	position  TabPosition         // Tab position (top, bottom, left, right)
	onChange  func(string)        // Callback when tab changes
	vertical  bool                // DEPRECATED: Use position instead
	contents  map[string]ui.VNode // Tab content mapping
	wrapTabs  bool                // Enable automatic tab bar wrapping
	tabGap    int                 // Gap between tabs when wrapping

	// ActionTarget support
	supportedActions []action.ActionType // Supported action types
}

// NewTabs creates a new tabs component
func NewTabs() *TabsVNode {
	return &TabsVNode{
		ElementVNode: ui.NewElement("tabs"),
		tabs:         make([]TabItem, 0),
		activeTab:    0,
		position:     TabPositionTop, // Default position
		onChange:     nil,
		vertical:     false, // DEPRECATED
		contents:     make(map[string]ui.VNode),
		wrapTabs:     false, // Default: no wrapping (single row)
		tabGap:       1,     // Default gap when wrapping
		supportedActions: []action.ActionType{
			action.ActionNavigateNext,
			action.ActionNavigatePrev,
			action.ActionNavigateLeft,
			action.ActionNavigateRight,
			action.ActionNavigateHome,
			action.ActionNavigateEnd,
			action.ActionSelect,
			action.ActionEnter,
			action.ActionScroll,
			action.ActionClick,
		},
	}
}

// Tabs creates a new tabs node
func Tabs() ui.VNode {
	return NewTabs()
}

// Builder pattern
type TabsBuilderType struct {
	node     *TabsVNode
	contents map[string]ui.VNode
}

// TabsBuilder creates a new tabs builder
func TabsBuilder() *TabsBuilderType {
	return &TabsBuilderType{
		node:     NewTabs(),
		contents: make(map[string]ui.VNode),
	}
}

// Build returns the tabs ui.VNode
func (b *TabsBuilderType) Build() ui.VNode {
	// Ensure children reflect current active tab (tab bar + content)
	b.node.updateActiveContent()
	return b.node
}

// AddTab adds a tab
func (b *TabsBuilderType) AddTab(id, label string) *TabsBuilderType {
	b.node.tabs = append(b.node.tabs, TabItem{ID: id, Label: label})
	return b
}

// Content sets the content for a tab
func (b *TabsBuilderType) Content(id string, content ui.VNode) *TabsBuilderType {
	if b.contents == nil {
		b.contents = make(map[string]ui.VNode)
	}
	b.contents[id] = content
	b.node.contents[id] = content

	// If this is the active tab, set it as the child
	// This allows the framework's layout engine to render the content
	if len(b.node.tabs) > 0 && content != nil {
		activeTabID := b.node.tabs[b.node.activeTab].ID
		if id == activeTabID {
			// Set the active tab's content as the child
			b.node.SetChildren([]ui.VNode{content})
		}
	}

	return b
}

// ActiveTab sets the active tab index
func (b *TabsBuilderType) ActiveTab(index int) *TabsBuilderType {
	b.node.SetActiveTab(index)
	return b
}

// OnChange sets the change handler
func (b *TabsBuilderType) OnChange(fn func(string)) *TabsBuilderType {
	b.node.SetOnChange(fn)
	return b
}

// Vertical sets vertical orientation
func (b *TabsBuilderType) Vertical(v bool) *TabsBuilderType {
	b.node.SetVertical(v)
	return b
}

// Position sets where tabs are positioned relative to content
func (b *TabsBuilderType) Position(pos TabPosition) *TabsBuilderType {
	b.node.position = pos
	return b
}

// Width sets the total width
func (b *TabsBuilderType) Width(n int) *TabsBuilderType {
	b.node.SetProp("width", n)
	return b
}

// Height sets the content height
func (b *TabsBuilderType) Height(n int) *TabsBuilderType {
	b.node.SetProp("height", n)
	return b
}

// Key sets the key for diffing
func (b *TabsBuilderType) Key(key string) *TabsBuilderType {
	b.node.SetKey(key)
	return b
}

// WrapTabs enables automatic tab bar wrapping
// When enabled, tabs will wrap to multiple rows if they don't fit in one line
func (b *TabsBuilderType) WrapTabs(wrap bool) *TabsBuilderType {
	b.node.wrapTabs = wrap
	return b
}

// TabGap sets the gap between tabs when wrapping
func (b *TabsBuilderType) TabGap(gap int) *TabsBuilderType {
	b.node.tabGap = gap
	return b
}

// Getters
func (t *TabsVNode) Tabs() []TabItem               { return t.tabs }
func (t *TabsVNode) ActiveTab() int                { return t.activeTab }
func (t *TabsVNode) OnChange() func(string)        { return t.onChange }
func (t *TabsVNode) Vertical() bool                { return t.vertical }
func (t *TabsVNode) Contents() map[string]ui.VNode { return t.contents }

// Setters
func (t *TabsVNode) SetTabs(tabs []TabItem) { t.tabs = tabs }
func (t *TabsVNode) SetActiveTab(index int) {
	t.activeTab = index
	// Update children to reflect the new active tab
	t.updateActiveContent()
}
func (t *TabsVNode) SetOnChange(fn func(string)) { t.onChange = fn }
func (t *TabsVNode) SetVertical(v bool) { t.vertical = v }

// HandleEvent enables mouse tab switching using LocalX/LocalY from HitMap.
// The Inspector's HitMap-based event routing delivers events directly to target components.
func (t *TabsVNode) HandleEvent(ev frameworkevent.Event) bool {
	me, ok := ev.(*frameworkevent.MouseEvent)
	if !ok || ev.Type() != frameworkevent.EventMousePress {
		return false
	}

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		fmt.Fprintf(os.Stderr, "[Tabs] HandleEvent: LocalX=%d, LocalY=%d\n", me.LocalX, me.LocalY)
	}

	// Use LocalX/LocalY directly (pre-calculated by HitMap)
	localX := me.LocalX
	localY := me.LocalY

	// Calculate tab bar height (1 row for normal, multiple for wrap mode)
	tabBarHeight := 1
	if t.wrapTabs && !t.vertical {
		// Estimate rows (same logic as in Measure)
		containerWidth := 80
		if props := t.Props(); props != nil {
			if w, ok := props["width"].(int); ok && w > 0 {
				containerWidth = w
			}
		}

		currentRowWidth := 0
		rowCount := 1
		for i, tab := range t.tabs {
			tabWidth := utf8.RuneCountInString(tab.Label) + 2
			if i > 0 {
				tabWidth += t.tabGap
			}

			if currentRowWidth+tabWidth > containerWidth && currentRowWidth > 0 {
				rowCount++
				currentRowWidth = tabWidth
			} else {
				currentRowWidth += tabWidth
			}
		}
		tabBarHeight = rowCount
	}

	// Check if click is in tab bar area
	if localY >= 0 && localY < tabBarHeight {
		// Handle tab bar click - switch tabs
		return t.handleTabBarClick(localX, localY)
	}

	// Click is below tab bar - forward to active tab content
	if t.activeTab < 0 || t.activeTab >= len(t.tabs) {
		return false
	}

	activeTabID := t.tabs[t.activeTab].ID
	if content, ok := t.contents[activeTabID]; ok && content != nil {
		// Try to forward event to content component
		if contentComponent, ok := content.(frameworkevent.Component); ok {
			// Forward the event with the same LocalX/LocalY
			// Content should use these coordinates to determine what was clicked
			return contentComponent.HandleEvent(ev)
		}
	}

	return false
}

// =============================================================================
// Msg/Cmd Architecture Support (Phase 3)
// =============================================================================

// Update implements component.Updater interface for Msg/Cmd architecture
//
// Handles:
// - MouseMsg: Direct routing via TargetID from HitMap
// - KeyMsg: Keyboard navigation (arrows, Enter, Tab)
func (t *TabsVNode) Update(message runtimemsg.Msg) cmd.Cmd {
	switch msg := message.(type) {
	case *runtimemsg.MouseMsg:
		return t.updateMouse(msg)
	case *runtimemsg.KeyMsg:
		return t.updateKey(msg)
	}
	return nil
}

// updateMouse handles mouse messages (direct routing via TargetID)
func (t *TabsVNode) updateMouse(mouseMsg *runtimemsg.MouseMsg) cmd.Cmd {
	// Only handle mouse press
	if mouseMsg.Action != runtimemsg.MouseActionPress {
		return nil
	}

	// Use LocalX/LocalY directly (pre-calculated by HitMap)
	localX := mouseMsg.LocalX
	localY := mouseMsg.LocalY

	// Check if click is in tab bar (first row)
	if localY == 0 {
		// Handle tab bar click - switch tabs
		t.handleTabBarClick(localX, localY)
		return nil // TODO: Return Cmd to trigger re-render
	}

	// Click is below tab bar - forward to active tab content
	// TODO: Need to support forwarding MouseMsg to child components
	// For now, return nil and let Event system handle it
	return nil
}

// updateKey handles keyboard messages (when focused)
func (t *TabsVNode) updateKey(keyMsg *runtimemsg.KeyMsg) cmd.Cmd {
	// Arrow keys for navigation
	switch keyMsg.Special {
	case runtimeplatform.KeyLeft:
		t.PreviousTab()
		return nil
	case runtimeplatform.KeyRight:
		t.NextTab()
		return nil
	case runtimeplatform.KeyHome:
		t.FirstTab()
		return nil
	case runtimeplatform.KeyEnd:
		t.LastTab()
		return nil
	case runtimeplatform.KeyEnter:
		// Enter on current tab - trigger onChange callback
		if t.activeTab >= 0 && t.activeTab < len(t.tabs) && t.onChange != nil {
			t.onChange(t.tabs[t.activeTab].ID)
		}
		return nil
	}

	return nil
}

// handleTabBarClick handles clicks on the tab bar
func (t *TabsVNode) handleTabBarClick(localX, localY int) bool {
	// For wrapping tabs, we need to calculate multi-row layout
	if t.wrapTabs {
		return t.handleWrappedTabBarClick(localX, localY)
	}

	// Single-line tab bar (original behavior)
	cursor := 0
	for i, tab := range t.tabs {
		if i > 0 {
			cursor += 3 // " | "
		}

		labelWidth := utf8.RuneCountInString(tab.Label)
		width := labelWidth
		if i == t.activeTab {
			width += 2 // brackets []
		}

		if localX >= cursor && localX < cursor+width {
			if tab.Disabled {
				return true // consume but no action
			}
			if i != t.activeTab {
				t.SetActiveTab(i)
				if t.onChange != nil {
					t.onChange(tab.ID)
				}
			}
			return true
		}

		cursor += width
	}

	return false
}

// handleWrappedTabBarClick handles clicks on a wrapped (multi-row) tab bar
func (t *TabsVNode) handleWrappedTabBarClick(localX, localY int) bool {
	// Get container width
	containerWidth := 80
	if props := t.Props(); props != nil {
		if w, ok := props["width"].(int); ok && w > 0 {
			containerWidth = w
		}
	}

	// Simulate wrap layout to find which row and tab was clicked
	currentRowWidth := 0
	currentRow := 0
	rowStartX := make(map[int]int) // Starting X position for each row

	for i, tab := range t.tabs {
		tabWidth := utf8.RuneCountInString(tab.Label) + 2 // +2 for brackets
		if i > 0 {
			tabWidth += t.tabGap
		}

		// Check if we need to wrap
		if currentRowWidth+tabWidth > containerWidth && currentRowWidth > 0 {
			currentRow++
			currentRowWidth = tabWidth
			rowStartX[currentRow] = 0
		} else {
			if currentRow == 0 {
				rowStartX[currentRow] = 0
			}
			currentRowWidth += tabWidth
		}

		// Check if click is in this row and matches this tab
		if currentRow == localY {
			tabStart := rowStartX[currentRow]
			tabEnd := tabStart + tabWidth

			if localX >= tabStart && localX < tabEnd {
				if tab.Disabled {
					return true // consume but no action
				}
				if i != t.activeTab {
					t.SetActiveTab(i)
					if t.onChange != nil {
						t.onChange(tab.ID)
					}
				}
				return true
			}

			// Update next tab's start position for this row
			rowStartX[currentRow] = tabEnd
		}
	}

	return false
}

// =============================================================================
// Internal Helper Methods
// =============================================================================

// updateActiveContent updates the children to reflect the active tab's content
func (t *TabsVNode) updateActiveContent() {
	if t.activeTab < 0 || t.activeTab >= len(t.tabs) {
		return
	}

	var tabBarNode ui.VNode

	// Create tab bar with or without wrapping
	if t.wrapTabs {
		// Use Wrap component for automatic wrapping
		var tabNodes []ui.VNode
		for i, tab := range t.tabs {
			var tabLabel string
			if i == t.activeTab {
				tabLabel = fmt.Sprintf("[%s]", tab.Label)
			} else {
				tabLabel = tab.Label
			}
			tabNodes = append(tabNodes, ui.Text(tabLabel))
		}

		// Get container width from props or default to 80
		containerWidth := 80
		if props := t.Props(); props != nil {
			if w, ok := props["width"].(int); ok {
				containerWidth = w
			}
		}

		tabBarNode = layout.NewWrapBuilder(tabNodes...).
			Gap(t.tabGap).
			ScreenWidth(containerWidth).
			Align(ui.AlignStart).
			Build()
	} else {
		// Original single-line behavior
		var parts []string
		for i, tab := range t.tabs {
			if i == t.activeTab {
				parts = append(parts, "["+tab.Label+"]")
			} else {
				parts = append(parts, tab.Label)
			}
		}
		tabLine := strings.Join(parts, " | ")
		tabBarNode = ui.Text(tabLine)
	}

	// Get active tab content
	activeTabID := t.tabs[t.activeTab].ID
	var contentNode ui.VNode
	if content, ok := t.contents[activeTabID]; ok && content != nil {
		contentNode = content
	} else {
		contentNode = ui.Text("")
	}

	// Set children as [tabBar, content]
	// This allows the framework's layout engine to render both
	t.SetChildren([]ui.VNode{tabBarNode, contentNode})
}

// Measure implements runtime.Measurable interface
func (t *TabsVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if t == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Calculate total width of all tabs
	totalWidth := 0
	for _, tab := range t.tabs {
		totalWidth += utf8.RuneCountInString(tab.Label) + 3 // +3 for " | " separators
	}
	if totalWidth > 0 {
		totalWidth += 2 // Adjust for first/last formatting
	}

	height := 1 // Tab bar height (default single line)

	// For vertical tabs, width is the max tab width
	if t.vertical {
		maxWidth := 0
		for _, tab := range t.tabs {
			labelWidth := utf8.RuneCountInString(tab.Label)
			if labelWidth > maxWidth {
				maxWidth = labelWidth
			}
		}
		totalWidth = maxWidth + 2 // +2 for padding
		height = len(t.tabs)
	}

	// For wrapping tabs, estimate height based on container width
	if t.wrapTabs && !t.vertical {
		containerWidth := constraints.MaxWidth
		if props := t.Props(); props != nil {
			if w, ok := props["width"].(int); ok && w > 0 {
				containerWidth = w
			}
		}

		// Estimate rows by simulating wrap
		currentRowWidth := 0
		rowCount := 1
		for i, tab := range t.tabs {
			tabWidth := utf8.RuneCountInString(tab.Label) + 2 // +2 for brackets
			if i > 0 {
				tabWidth += t.tabGap // Add gap between tabs
			}

			if currentRowWidth+tabWidth > containerWidth && currentRowWidth > 0 {
				rowCount++
				currentRowWidth = tabWidth
			} else {
				currentRowWidth += tabWidth
			}
		}
		height = rowCount
	}

	// ADDED: Check for explicit width/height props (like measureLayoutChildren does)
	props := t.Props()
	explicitWidth := 0
	hasWidthProp := false
	explicitHeight := 0
	hasHeightProp := false

	if props != nil {
		if w, ok := props["width"].(int); ok && w > 0 {
			explicitWidth = w
			hasWidthProp = true
		}
		if h, ok := props["height"].(int); ok && h > 0 {
			explicitHeight = h
			hasHeightProp = true
		}
	}

	// Measure content height if active tab has content
	if t.activeTab >= 0 && t.activeTab < len(t.tabs) {
		activeTabID := t.tabs[t.activeTab].ID
		if content, ok := t.contents[activeTabID]; ok && content != nil {
			// If content is measurable, measure it
			if measurable, ok := content.(interface {
				Measure(runtime.BoxConstraints) runtime.Size
			}); ok {
				// MODIFIED: Use bounded height from prop or constraint
				maxContentHeight := runtime.Infinity
				if hasHeightProp {
					maxContentHeight = explicitHeight - height // Subtract tab bar
				} else if constraints.HasBoundedHeight() {
					maxContentHeight = constraints.MaxHeight - height
				}

				// Create constraints for content
				contentConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  constraints.MaxWidth,
					MinHeight: 0,
					MaxHeight: maxContentHeight,
				}
				if contentConstraints.MaxHeight < 0 {
					contentConstraints.MaxHeight = 0
				}

				contentSize := measurable.Measure(contentConstraints)
				// Add content height to total height
				height += contentSize.Height

				// Use max width for non-vertical tabs
				if !t.vertical && contentSize.Width > totalWidth {
					totalWidth = contentSize.Width
				}
			} else {
				// Content is not measurable, assume default height
				height += 10 // Default content height
			}
		}
	}

	// Apply constraints (respect explicit width/height props first)
	if hasWidthProp {
		totalWidth = explicitWidth
	} else {
		if totalWidth < constraints.MinWidth {
			totalWidth = constraints.MinWidth
		}
		if totalWidth > constraints.MaxWidth && constraints.MaxWidth > 0 {
			totalWidth = constraints.MaxWidth
		}
	}

	if hasHeightProp {
		height = explicitHeight
	} else {
		if height < constraints.MinHeight {
			height = constraints.MinHeight
		}
		if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
			height = constraints.MaxHeight
		}
	}

	return runtime.Size{Width: totalWidth, Height: height}
}

// Paint is removed to let framework handle children rendering
// The tab bar and content are now set as children, so the framework's
// default layout engine will render them correctly.
//
// If custom styling is needed for tabs, it should be applied via SetStyle()
// and the tab bar should be created with styled Text nodes.

// =============================================================================
// Enhanced Runtime Methods
// =============================================================================

// SetActiveTabByLabel switches to a tab by its label
// Returns true if successful, false if label not found
func (t *TabsVNode) SetActiveTabByLabel(label string) bool {
	for i, tab := range t.tabs {
		if strings.EqualFold(tab.Label, label) {
			if tab.Disabled {
				return false
			}
			t.activeTab = i

			// Trigger onChange callback
			if t.onChange != nil {
				t.onChange(tab.ID)
			}

			// Update children to reflect the new active tab
			t.updateActiveContent()

			return true
		}
	}
	return false
}

// SetActiveTabByID switches to a tab by its ID
// Returns true if successful, false if ID not found
func (t *TabsVNode) SetActiveTabByID(id string) bool {
	for i, tab := range t.tabs {
		if tab.ID == id {
			if tab.Disabled {
				return false
			}
			t.activeTab = i

			// Trigger onChange callback
			if t.onChange != nil {
				t.onChange(tab.ID)
			}

			// Update children to reflect the new active tab
			t.updateActiveContent()

			return true
		}
	}
	return false
}

// GetActiveTabLabel returns the label of the currently active tab
func (t *TabsVNode) GetActiveTabLabel() string {
	if t.activeTab < 0 || t.activeTab >= len(t.tabs) {
		return ""
	}
	return t.tabs[t.activeTab].Label
}

// GetActiveTabID returns the ID of the currently active tab
func (t *TabsVNode) GetActiveTabID() string {
	if t.activeTab < 0 || t.activeTab >= len(t.tabs) {
		return ""
	}
	return t.tabs[t.activeTab].ID
}

// GetActiveTabContent returns the content of the currently active tab
func (t *TabsVNode) GetActiveTabContent() ui.VNode {
	if t.activeTab < 0 || t.activeTab >= len(t.tabs) {
		return nil
	}

	// Try contents map first
	if t.contents != nil {
		if content, ok := t.contents[t.tabs[t.activeTab].ID]; ok {
			return content
		}
	}

	// Fallback to Content field in TabItem
	return t.tabs[t.activeTab].Content
}

// Bounds returns the tab's bounds for hit testing
func (t *TabsVNode) Bounds() [4]int {
	return t.ElementVNode.GetBounds()
}

// NextTab switches to the next enabled tab
// Returns true if successful, false if already at the last tab
func (t *TabsVNode) NextTab() bool {
	for i := t.activeTab + 1; i < len(t.tabs); i++ {
		if !t.tabs[i].Disabled {
			oldTab := t.activeTab
			t.activeTab = i

			// Trigger onChange callback
			if t.onChange != nil {
				t.onChange(t.tabs[i].ID)
			}

			// Update old tab's style
			if oldTab >= 0 && oldTab < len(t.tabs) {
				// Style update will happen on next render
			}

			return true
		}
	}
	return false
}

// PreviousTab switches to the previous enabled tab
// Returns true if successful, false if already at the first tab
func (t *TabsVNode) PreviousTab() bool {
	for i := t.activeTab - 1; i >= 0; i-- {
		if !t.tabs[i].Disabled {
			oldTab := t.activeTab
			t.activeTab = i

			// Trigger onChange callback
			if t.onChange != nil {
				t.onChange(t.tabs[i].ID)
			}

			// Update old tab's style
			if oldTab >= 0 && oldTab < len(t.tabs) {
				// Style update will happen on next render
			}

			return true
		}
	}
	return false
}

// FirstTab switches to the first enabled tab
// Returns true if successful, false if no tabs available
func (t *TabsVNode) FirstTab() bool {
	for i := 0; i < len(t.tabs); i++ {
		if !t.tabs[i].Disabled {
			oldTab := t.activeTab
			t.activeTab = i

			// Trigger onChange callback
			if t.onChange != nil {
				t.onChange(t.tabs[i].ID)
			}

			// Update old tab's style
			if oldTab >= 0 && oldTab < len(t.tabs) {
				// Style update will happen on next render
			}

			return true
		}
	}
	return false
}

// LastTab switches to the last enabled tab
// Returns true if successful, false if no tabs available
func (t *TabsVNode) LastTab() bool {
	for i := len(t.tabs) - 1; i >= 0; i-- {
		if !t.tabs[i].Disabled {
			oldTab := t.activeTab
			t.activeTab = i

			// Trigger onChange callback
			if t.onChange != nil {
				t.onChange(t.tabs[i].ID)
			}

			// Update old tab's style
			if oldTab >= 0 && oldTab < len(t.tabs) {
				// Style update will happen on next render
			}

			return true
		}
	}
	return false
}

// CanGoNext returns true if there's a next enabled tab
func (t *TabsVNode) CanGoNext() bool {
	for i := t.activeTab + 1; i < len(t.tabs); i++ {
		if !t.tabs[i].Disabled {
			return true
		}
	}
	return false
}

// CanGoPrevious returns true if there's a previous enabled tab
func (t *TabsVNode) CanGoPrevious() bool {
	for i := t.activeTab - 1; i >= 0; i-- {
		if !t.tabs[i].Disabled {
			return true
		}
	}
	return false
}

// FindTabByLabel finds a tab index by its label (case-insensitive)
// Returns -1 if not found
func (t *TabsVNode) FindTabByLabel(label string) int {
	for i, tab := range t.tabs {
		if strings.EqualFold(tab.Label, label) {
			return i
		}
	}
	return -1
}

// FindTabByID finds a tab index by its ID
// Returns -1 if not found
func (t *TabsVNode) FindTabByID(id string) int {
	for i, tab := range t.tabs {
		if tab.ID == id {
			return i
		}
	}
	return -1
}

// GetTabCount returns the total number of tabs
func (t *TabsVNode) GetTabCount() int {
	return len(t.tabs)
}

// IsTabEnabled returns true if the tab at index is enabled
func (t *TabsVNode) IsTabEnabled(index int) bool {
	if index < 0 || index >= len(t.tabs) {
		return false
	}
	return !t.tabs[index].Disabled
}

// SetTabEnabled enables or disables a tab
// Returns true if successful, false if index is out of range
func (t *TabsVNode) SetTabEnabled(index int, enabled bool) bool {
	if index < 0 || index >= len(t.tabs) {
		return false
	}
	t.tabs[index].Disabled = !enabled

	// If we disabled the current active tab, switch to another tab
	if !enabled && index == t.activeTab {
		t.FirstTab()
	}

	return true
}

// GetTabByIndex returns a tab by its index
// Returns nil if index is out of range
func (t *TabsVNode) GetTabByIndex(index int) *TabItem {
	if index < 0 || index >= len(t.tabs) {
		return nil
	}
	return &t.tabs[index]
}

// GetPosition returns the tab position
func (t *TabsVNode) GetPosition() TabPosition {
	return t.position
}

// SetPosition sets where tabs are positioned relative to content
func (t *TabsVNode) SetPosition(pos TabPosition) {
	t.position = pos
}

// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction implements ActionTarget interface
func (t *TabsVNode) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	// Handle action based on type
	switch act.Type {
	// Navigation actions
	case action.ActionNavigateNext, action.ActionNavigateRight:
		return t.NextTab()
	case action.ActionNavigatePrev, action.ActionNavigateLeft:
		return t.PreviousTab()
	case action.ActionNavigateHome:
		return t.FirstTab()
	case action.ActionNavigateEnd:
		return t.LastTab()

	// Selection actions
	case action.ActionSelect, action.ActionEnter:
		// Select current tab (already active)
		return true

	// Scroll action
	case action.ActionScroll:
		if delta, ok := act.GetPayloadInt(); ok {
			if delta > 0 {
				return t.NextTab()
			} else if delta < 0 {
				return t.PreviousTab()
			}
		}
		return false

	// Mouse click
	case action.ActionClick:
		// Click action already handled by HandleEvent
		return true
	}

	return false
}

// GetSupportedActions implements ActionTarget interface
func (t *TabsVNode) GetSupportedActions() []action.ActionType {
	if t.supportedActions == nil {
		return []action.ActionType{
			action.ActionNavigateNext,
			action.ActionNavigatePrev,
			action.ActionNavigateLeft,
			action.ActionNavigateRight,
			action.ActionNavigateHome,
			action.ActionNavigateEnd,
			action.ActionSelect,
			action.ActionEnter,
			action.ActionScroll,
			action.ActionClick,
		}
	}
	return t.supportedActions
}

// CanHandleAction implements ActionTarget interface
func (t *TabsVNode) CanHandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	// Check if action type is supported
	supported := t.GetSupportedActions()
	for _, supportedType := range supported {
		if supportedType == act.Type {
			return true
		}
	}

	return false
}

// ============================================================================
// FocusableActionTarget 接口实现
// ============================================================================

// Focus implements FocusableActionTarget interface
func (t *TabsVNode) Focus() bool {
	return len(t.tabs) > 0
}

// Blur implements FocusableActionTarget interface
func (t *TabsVNode) Blur() {
	// Tabs don't have a distinct focus state
}

// IsFocused implements FocusableActionTarget interface
func (t *TabsVNode) IsFocused() bool {
	// Tabs are always "focused" when they have tabs
	return len(t.tabs) > 0
}

// IsFocusable implements FocusableActionTarget interface
func (t *TabsVNode) IsFocusable() bool {
	return len(t.tabs) > 0
}

// ============================================================================
// ScrollableActionTarget 接口实现
// ============================================================================

// CanScroll implements ScrollableActionTarget interface
func (t *TabsVNode) CanScroll(delta int) bool {
	if delta > 0 {
		return t.CanGoNext()
	} else if delta < 0 {
		return t.CanGoPrevious()
	}
	return false
}

// Scroll implements ScrollableActionTarget interface
func (t *TabsVNode) Scroll(delta int) bool {
	if !t.CanScroll(delta) {
		return false
	}
	if delta > 0 {
		return t.NextTab()
	} else {
		return t.PreviousTab()
	}
}

// GetScrollPosition implements ScrollableActionTarget interface
func (t *TabsVNode) GetScrollPosition() (int, int, int) {
	current := t.activeTab
	total := len(t.tabs)
	visible := 1 // Only one tab is visible at a time
	return current, total, visible
}

// ============================================================================
// SelectableActionTarget 接口实现
// ============================================================================

// Select implements SelectableActionTarget interface
func (t *TabsVNode) Select() bool {
	return len(t.tabs) > 0
}

// IsSelected implements SelectableActionTarget interface
func (t *TabsVNode) IsSelected() bool {
	// A tab is always considered "selected" (the active one)
	return len(t.tabs) > 0
}

// ToggleSelection implements SelectableActionTarget interface
func (t *TabsVNode) ToggleSelection() bool {
	// Tabs can't be toggled off, but we can navigate
	return t.NextTab()
}

// GetSelectedCount implements SelectableActionTarget interface
func (t *TabsVNode) GetSelectedCount() int {
	if len(t.tabs) > 0 {
		return 1 // One tab is always active
	}
	return 0
}
