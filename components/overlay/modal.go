package overlay

import (
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Modal Component
// =============================================================================

// ModalVNode represents a modal dialog component
type ModalVNode struct {
	*ui.ElementVNode
	title    string
	content  ui.VNode
	footer   ui.VNode
	isOpen   bool
	width    int
	height   int
	centered bool
}

// NewModal creates a new modal
func NewModal() *ModalVNode {
	return &ModalVNode{
		ElementVNode: ui.NewElement("modal"),
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
func Modal() ui.VNode {
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

// Build returns the modal ui.VNode
func (b *ModalBuilderType) Build() ui.VNode {
	return b.node
}

// Title sets the modal title
func (b *ModalBuilderType) Title(title string) *ModalBuilderType {
	b.node.title = title
	return b
}

// Content sets the modal content
func (b *ModalBuilderType) Content(content ui.VNode) *ModalBuilderType {
	b.node.content = content
	return b
}

// Footer sets the modal footer
func (b *ModalBuilderType) Footer(footer ui.VNode) *ModalBuilderType {
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

// Key sets the key for diffing
func (b *ModalBuilderType) Key(key string) *ModalBuilderType {
	b.node.SetKey(key)
	return b
}

// Getters
func (m *ModalVNode) Title() string    { return m.title }
func (m *ModalVNode) Content() ui.VNode  { return m.content }
func (m *ModalVNode) Footer() ui.VNode   { return m.footer }
func (m *ModalVNode) IsOpen() bool        { return m.isOpen }
func (m *ModalVNode) Width() int          { return m.width }
func (m *ModalVNode) Height() int         { return m.height }
func (m *ModalVNode) Centered() bool      { return m.centered }

// Setters
func (m *ModalVNode) SetTitle(title string)         { m.title = title }
func (m *ModalVNode) SetContent(content ui.VNode)    { m.content = content }
func (m *ModalVNode) SetFooter(footer ui.VNode)      { m.footer = footer }
func (m *ModalVNode) SetIsOpen(open bool)            { m.isOpen = open }
func (m *ModalVNode) SetWidth(width int)             { m.width = width }
func (m *ModalVNode) SetHeight(height int)            { m.height = height }
func (m *ModalVNode) SetCentered(centered bool)       { m.centered = centered }

// Toggle opens/closes the modal and returns the new state
func (m *ModalVNode) Toggle() bool {
	m.isOpen = !m.isOpen
	return m.isOpen
}
