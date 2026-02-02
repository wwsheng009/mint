package navigation

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Tabs Component
// =============================================================================

// TabItem represents a single tab
type TabItem struct {
	ID    string
	Label string
}

// TabsVNode represents a tabs component
type TabsVNode struct {
	*ui.ElementVNode
	tabs      []TabItem
	activeTab int
	onChange  func(string)
	vertical  bool
	contents  map[string]ui.VNode // Tab content mapping
}

// NewTabs creates a new tabs component
func NewTabs() *TabsVNode {
	return &TabsVNode{
		ElementVNode: ui.NewElement("tabs"),
		tabs:         make([]TabItem, 0),
		activeTab:    0,
		onChange:     nil,
		vertical:     false,
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

// Content sets the content for a tab
func (b *TabsBuilderType) Content(id string, content ui.VNode) *TabsBuilderType {
	if b.contents == nil {
		b.contents = make(map[string]ui.VNode)
	}
	b.contents[id] = content
	b.node.contents[id] = content
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
func (t *TabsVNode) SetActiveTab(index int)      { t.activeTab = index }
func (t *TabsVNode) SetOnChange(fn func(string)) { t.onChange = fn }
func (t *TabsVNode) SetVertical(v bool)           { t.vertical = v }

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

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

	height := 1

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

// Paint implements paint.Paintable interface
func (t *TabsVNode) Paint(x, y int) []paint.DrawCmd {
	if t == nil || len(t.tabs) == 0 {
		return nil
	}

	tabStyle := t.Style()

	if t.vertical {
		// Vertical tabs rendering
		var cmds []paint.DrawCmd
		for i, tab := range t.tabs {
			var prefix string
			if i == t.activeTab {
				prefix = "> "
				tabStyle = tabStyle.Bold(true)
			} else {
				prefix = "  "
				tabStyle = tabStyle.Bold(false)
			}
			tabLabel := prefix + tab.Label
			cmds = append(cmds, paint.NewTextCmd(x, y+i, tabLabel, tabStyle))
		}
		return cmds
	}

	// Horizontal tabs rendering
	var parts []string
	for i, tab := range t.tabs {
		if i == t.activeTab {
			parts = append(parts, "["+tab.Label+"]")
		} else {
			parts = append(parts, tab.Label)
		}
	}
	tabLine := strings.Join(parts, " | ")

	return []paint.DrawCmd{
		paint.NewTextCmd(x, y, tabLine, tabStyle),
	}
}
