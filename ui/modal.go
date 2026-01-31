package ui

// =============================================================================
// Modal Component
// =============================================================================

// ModalVNode represents a modal dialog component
type ModalVNode struct {
	*ElementVNode
	title    string
	content  VNode
	footer   VNode
	isOpen   bool
	width    int
	height   int
	centered bool
}

// NewModal creates a new modal
func NewModal() *ModalVNode {
	return &ModalVNode{
		ElementVNode: NewElement("modal"),
		title:        "",
		content:      nil,
		footer:       nil,
		isOpen:       false,
		width:        40,
		height:       15,
		centered:     true,
	}
}

// Modal creates a new modal node
func Modal() VNode {
	return NewModal()
}

// Builder pattern
type ModalBuilderType struct {
	node *ModalVNode
}

// ModalBuilder creates a new modal builder
func ModalBuilder() *ModalBuilderType {
	return &ModalBuilderType{node: NewModal()}
}

// Build returns the modal VNode
func (b *ModalBuilderType) Build() VNode {
	return b.node
}

// Title sets the modal title
func (b *ModalBuilderType) Title(title string) *ModalBuilderType {
	b.node.title = title
	return b
}

// Content sets the modal content
func (b *ModalBuilderType) Content(content VNode) *ModalBuilderType {
	b.node.content = content
	return b
}

// Footer sets the modal footer
func (b *ModalBuilderType) Footer(footer VNode) *ModalBuilderType {
	b.node.footer = footer
	return b
}

// Open sets the modal as open
func (b *ModalBuilderType) Open(open bool) *ModalBuilderType {
	b.node.isOpen = open
	return b
}

// Width sets the modal width
func (b *ModalBuilderType) Width(width int) *ModalBuilderType {
	b.node.width = width
	return b
}

// Height sets the modal height
func (b *ModalBuilderType) Height(height int) *ModalBuilderType {
	b.node.height = height
	return b
}

// Centered sets whether the modal is centered
func (b *ModalBuilderType) Centered(centered bool) *ModalBuilderType {
	b.node.centered = centered
	return b
}

// Getters
func (m *ModalVNode) Title() string        { return m.title }
func (m *ModalVNode) Content() VNode      { return m.content }
func (m *ModalVNode) Footer() VNode       { return m.footer }
func (m *ModalVNode) IsOpen() bool        { return m.isOpen }
func (m *ModalVNode) Width() int          { return m.width }
func (m *ModalVNode) Height() int         { return m.height }
func (m *ModalVNode) Centered() bool      { return m.centered }

// Setters
func (m *ModalVNode) SetTitle(title string)        { m.title = title }
func (m *ModalVNode) SetContent(content VNode)      { m.content = content }
func (m *ModalVNode) SetFooter(footer VNode)       { m.footer = footer }
func (m *ModalVNode) SetIsOpen(open bool)           { m.isOpen = open }
func (m *ModalVNode) SetWidth(width int)            { m.width = width }
func (m *ModalVNode) SetHeight(height int)           { m.height = height }
func (m *ModalVNode) SetCentered(centered bool)      { m.centered = centered }

// Toggle opens/closes the modal and returns the new state
func (m *ModalVNode) Toggle() bool {
	m.isOpen = !m.isOpen
	return m.isOpen
}

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
	*ElementVNode
	tabs      []TabItem
	activeTab int
	onChange  func(string)
	vertical  bool
}

// NewTabs creates a new tabs component
func NewTabs() *TabsVNode {
	return &TabsVNode{
		ElementVNode: NewElement("tabs"),
		tabs:         make([]TabItem, 0),
		activeTab:    0,
		onChange:     nil,
		vertical:     false,
	}
}

// Tabs creates a new tabs node
func Tabs() VNode {
	return NewTabs()
}

// Builder pattern
type TabsBuilderType struct {
	node     *TabsVNode
	contents map[string]VNode
}

// TabsBuilder creates a new tabs builder
func TabsBuilder() *TabsBuilderType {
	return &TabsBuilderType{
		node:     NewTabs(),
		contents: make(map[string]VNode),
	}
}

// Build returns the tabs VNode
func (b *TabsBuilderType) Build() VNode {
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
func (b *TabsBuilderType) Content(id string, content VNode) *TabsBuilderType {
	if b.contents == nil {
		b.contents = make(map[string]VNode)
	}
	b.contents[id] = content
	return b
}

// Getters
func (t *TabsVNode) Tabs() []TabItem     { return t.tabs }
func (t *TabsVNode) ActiveTab() int     { return t.activeTab }
func (t *TabsVNode) OnChange() func(string) { return t.onChange }
func (t *TabsVNode) Vertical() bool      { return t.vertical }

// Setters
func (t *TabsVNode) SetTabs(tabs []TabItem)      { t.tabs = tabs }
func (t *TabsVNode) SetActiveTab(index int)     { t.activeTab = index }
func (t *TabsVNode) SetOnChange(fn func(string)) { t.onChange = fn }
func (t *TabsVNode) SetVertical(v bool)          { t.vertical = v }

// =============================================================================
// Divider Component
// =============================================================================

// DividerStyle defines the visual style of a divider
type DividerStyle int

const (
	DividerSolid   DividerStyle = iota // ───────────
	DividerDashed                      // - - - - - -
	DividerDotted                      // ·· ·· ·· ··
	DividerDouble                      // ═══════════
)

// DividerVNode represents a divider component
type DividerVNode struct {
	*ElementVNode
	text          string
	dividerStyle DividerStyle
	thickness     int
}

// NewDivider creates a new divider
func NewDivider() *DividerVNode {
	return &DividerVNode{
		ElementVNode:  NewElement("divider"),
		text:          "",
		dividerStyle:  DividerSolid,
		thickness:     1,
	}
}

// Divider creates a new divider node
func Divider() VNode {
	return NewDivider()
}

// Builder pattern
type DividerBuilderType struct {
	node *DividerVNode
}

// DividerBuilder creates a new divider builder
func DividerBuilder() *DividerBuilderType {
	return &DividerBuilderType{node: NewDivider()}
}

// Build returns the divider VNode
func (b *DividerBuilderType) Build() VNode {
	return b.node
}

// Text sets the divider text (centered label)
func (b *DividerBuilderType) Text(text string) *DividerBuilderType {
	b.node.SetText(text)
	return b
}

// Style sets the divider style
func (b *DividerBuilderType) Style(style DividerStyle) *DividerBuilderType {
	b.node.SetDividerStyle(style)
	return b
}

// Thickness sets the divider thickness
func (b *DividerBuilderType) Thickness(thickness int) *DividerBuilderType {
	b.node.SetThickness(thickness)
	return b
}

// Getters
func (d *DividerVNode) Text() string        { return d.text }
func (d *DividerVNode) DividerStyle() DividerStyle { return d.dividerStyle }
func (d *DividerVNode) Thickness() int      { return d.thickness }

// Setters
func (d *DividerVNode) SetText(text string)               { d.text = text }
func (d *DividerVNode) SetDividerStyle(style DividerStyle) { d.dividerStyle = style }
func (d *DividerVNode) SetThickness(thickness int)         { d.thickness = thickness }
