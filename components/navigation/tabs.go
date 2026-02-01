package navigation

import (
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
