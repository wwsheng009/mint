package overlay

import (
	"strings"

	"github.com/wwsheng009/mint/framework/cmd"
	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/ui"
)

// Interface implementation assertions
var _ frameworkevent.Component = (*ModalVNode)(nil)
var _ component.Updater = (*ModalVNode)(nil) // Msg/Cmd support

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
	onClose  func() // Callback when modal is closed
	// Bounds for hit testing (x, y, width, height)
	bounds [4]int
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
		onClose:      nil,
		bounds:       [4]int{0, 0, 0, 0},
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

// OnClose sets the close callback
func (b *ModalBuilderType) OnClose(onClose func()) *ModalBuilderType {
	b.node.onClose = onClose
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

// Children returns all child nodes for HitMap building
// This is CRITICAL for Msg/Cmd routing to work with modal buttons
func (m *ModalVNode) Children() []ui.VNode {
	var children []ui.VNode

	// Add content if present
	if m.content != nil {
		children = append(children, m.content)
	}

	// Add footer if present
	if m.footer != nil {
		children = append(children, m.footer)
	}

	// Also include any children set via ElementVNode
	baseChildren := m.ElementVNode.Children()
	if len(baseChildren) > 0 {
		children = append(children, baseChildren...)
	}

	return children
}
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
func (m *ModalVNode) SetOnClose(onClose func())       { m.onClose = onClose }

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

	// CRITICAL: Set bounds for mouse hit testing
	m.bounds = [4]int{x, y, width, height}

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

// =============================================================================
// Event Handling (Msg/Cmd Architecture)
// =============================================================================

// HandleEvent processes events (legacy, for backward compatibility)
func (m *ModalVNode) HandleEvent(e frameworkevent.Event) bool {
	if !m.isOpen {
		return false
	}

	// Handle ESC key to close
	if keyEvent, ok := e.(*frameworkevent.KeyEvent); ok {
		if keyEvent.Special == frameworkevent.KeyEscape {
			m.isOpen = false
			if m.onClose != nil {
				m.onClose()
			}
			return true
		}
	}

	// Handle click outside modal
	if mouseEvent, ok := e.(*frameworkevent.MouseEvent); ok {
		if mouseEvent.Type() == frameworkevent.EventMousePress {
			// Check if click is outside modal bounds
			if !m.containsPoint(mouseEvent.X, mouseEvent.Y) {
				m.isOpen = false
				if m.onClose != nil {
					m.onClose()
				}
				return true
			}
			// Click is INSIDE modal - don't handle it here
			// Let it be routed to child components via HitMap
			return false
		}
	}

	return false
}

// Update implements component.Updater interface for Msg/Cmd architecture
func (m *ModalVNode) Update(message runtimemsg.Msg) cmd.Cmd {
	if !m.isOpen {
		return nil
	}

	switch msg := message.(type) {
	case *runtimemsg.KeyMsg:
		return m.updateKey(msg)
	case *runtimemsg.MouseMsg:
		return m.updateMouse(msg)
	}

	return nil
}

// updateKey handles keyboard messages (ESC to close)
func (m *ModalVNode) updateKey(keyMsg *runtimemsg.KeyMsg) cmd.Cmd {
	// ESC to close
	if keyMsg.Special == runtimeplatform.KeyEscape {
		m.isOpen = false
		if m.onClose != nil {
			m.onClose()
		}
		return nil
	}

	return nil
}

// updateMouse handles mouse messages (click outside to close)
func (m *ModalVNode) updateMouse(mouseMsg *runtimemsg.MouseMsg) cmd.Cmd {
	if mouseMsg.Action == runtimemsg.MouseActionPress {
		// Check if click is outside modal bounds
		if !m.containsPoint(mouseMsg.X, mouseMsg.Y) {
			m.isOpen = false
			if m.onClose != nil {
				m.onClose()
			}
			return nil
		}
		// Click is INSIDE modal - don't handle it here
		// Let it be routed to child components via HitMap
	}

	return nil
}

// containsPoint checks if a point is within the modal bounds
func (m *ModalVNode) containsPoint(x, y int) bool {
	if m.bounds[2] <= 0 || m.bounds[3] <= 0 {
		return false
	}
	return x >= m.bounds[0] && x < m.bounds[0]+m.bounds[2] &&
		y >= m.bounds[1] && y < m.bounds[1]+m.bounds[3]
}

// =============================================================================
// Bounds and Hit Testing
// =============================================================================

// Bounds returns the modal bounds for hit testing
func (m *ModalVNode) Bounds() [4]int {
	return m.bounds
}

// SetBounds sets the modal bounds (typically called during layout)
func (m *ModalVNode) SetBounds(x, y, width, height int) {
	m.bounds = [4]int{x, y, width, height}
}
