package overlay

import (
	"strings"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
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

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
func (m *ModalVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if m == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// If modal is not open, it takes no space
	if !m.isOpen {
		return runtime.Size{Width: 0, Height: 0}
	}

	width := m.width
	height := m.height

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

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface
// Generates draw commands for rendering this modal component
func (m *ModalVNode) Paint(x, y int) []paint.DrawCmd {
	if m == nil || !m.isOpen {
		return nil
	}

	modalStyle := m.Style()
	measured := m.Measure(runtime.BoxConstraints{})
	width := measured.Width
	height := measured.Height

	var cmds []paint.DrawCmd

	// Build top border
	topBorder := "┌" + strings.Repeat("─", width-2) + "┐"
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, modalStyle))

	// Draw title row if present
	if m.title != "" {
		titleRow := "│ " + m.title + strings.Repeat(" ", width-3-len(m.title)) + "│"
		cmds = append(cmds, paint.NewTextCmd(x, y+1, titleRow, modalStyle))

		// Separator after title
		separator := "├" + strings.Repeat("─", width-2) + "┤"
		cmds = append(cmds, paint.NewTextCmd(x, y+2, separator, modalStyle))

		// Draw content area (empty rows for now)
		for i := 3; i < height-2; i++ {
			contentRow := "│" + strings.Repeat(" ", width-2) + "│"
			cmds = append(cmds, paint.NewTextCmd(x, y+i, contentRow, modalStyle))
		}
	} else {
		// Draw content area (empty rows)
		for i := 1; i < height-1; i++ {
			contentRow := "│" + strings.Repeat(" ", width-2) + "│"
			cmds = append(cmds, paint.NewTextCmd(x, y+i, contentRow, modalStyle))
		}
	}

	// Draw bottom border
	bottomBorder := "└" + strings.Repeat("─", width-2) + "┘"
	cmds = append(cmds, paint.NewTextCmd(x, y+height-1, bottomBorder, modalStyle))

	return cmds
}
