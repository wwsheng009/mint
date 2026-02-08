package navigation

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/ui"
)

// TabPosition defines where tabs are positioned relative to content
type TabPosition int

const (
	TabPositionTop    TabPosition = iota // Tabs above content
	TabPositionBottom                       // Tabs below content
	TabPositionLeft                         // Tabs to the left of content
	TabPositionRight                        // Tabs to the right of content
)

// Tab represents a single tab in a Tabs component
type TabItem struct {
	ID       string
	Label    string
	Content  ui.VNode    // Content to show when tab is active
	Key      string      // Optional key for diffing
	Disabled bool        // Whether the tab is disabled
}

// =============================================================================
// Enhanced Tabs Component
// =============================================================================

// TabsVNode represents a tabs component
type TabsVNode struct {
	*ui.ElementVNode
	tabs      []TabItem
	activeTab int
	position  TabPosition     // Tab position (top, bottom, left, right)
	onChange  func(string)     // Callback when tab changes
	vertical  bool            // DEPRECATED: Use position instead
	contents  map[string]ui.VNode // Tab content mapping
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

// Getters
func (t *TabsVNode) Tabs() []TabItem     { return t.tabs }
func (t *TabsVNode) ActiveTab() int        { return t.activeTab }
func (t *TabsVNode) OnChange() func(string) { return t.onChange }
func (t *TabsVNode) Vertical() bool         { return t.vertical }
func (t *TabsVNode) Contents() map[string]ui.VNode { return t.contents }

// Setters
func (t *TabsVNode) SetTabs(tabs []TabItem)      { t.tabs = tabs }
func (t *TabsVNode) SetActiveTab(index int) {
	t.activeTab = index
	// Update children to reflect the new active tab
	t.updateActiveContent()
}
func (t *TabsVNode) SetOnChange(fn func(string)) { t.onChange = fn }
func (t *TabsVNode) SetVertical(v bool)           { t.vertical = v }

// =============================================================================
// Internal Helper Methods
// =============================================================================

// updateActiveContent updates the children to reflect the active tab's content
func (t *TabsVNode) updateActiveContent() {
	if t.activeTab < 0 || t.activeTab >= len(t.tabs) {
		return
	}

	// Create tab bar (label row)
	var parts []string
	for i, tab := range t.tabs {
		if i == t.activeTab {
			parts = append(parts, "["+tab.Label+"]")
		} else {
			parts = append(parts, tab.Label)
		}
	}
	tabLine := strings.Join(parts, " | ")
	tabBarNode := ui.Text(tabLine)

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

	height := 1 // Tab bar height

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

	// Measure content height if active tab has content
	if t.activeTab >= 0 && t.activeTab < len(t.tabs) {
		activeTabID := t.tabs[t.activeTab].ID
		if content, ok := t.contents[activeTabID]; ok && content != nil {
			// If content is measurable, measure it
			if measurable, ok := content.(interface{ Measure(runtime.BoxConstraints) runtime.Size }); ok {
				// Create constraints for content (unbounded height, but respect width)
				contentConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  constraints.MaxWidth,
					MinHeight: 0,
					MaxHeight: constraints.MaxHeight,
				}
				if height > 0 && !t.vertical {
					contentConstraints.MaxHeight = constraints.MaxHeight - height
					if contentConstraints.MaxHeight < 0 {
						contentConstraints.MaxHeight = 0
					}
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

	// Apply constraints
	if totalWidth < constraints.MinWidth {
		totalWidth = constraints.MinWidth
	}
	if totalWidth > constraints.MaxWidth && constraints.MaxWidth > 0 {
		totalWidth = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
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

