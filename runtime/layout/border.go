// Package layout provides border types for bordered container layout
package layout

// =============================================================================
// Border Types
// =============================================================================

// BorderStyle defines the visual style of a border
type BorderStyle int

const (
	// BorderNone no border
	BorderNone BorderStyle = iota

	// BorderSingle single line border
	BorderSingle

	// BorderDouble double line border
	BorderDouble

	// BorderRounded rounded corner border
	BorderRounded

	// BorderDashed dashed line border
	BorderDashed
)

// String returns the string representation of BorderStyle
func (s BorderStyle) String() string {
	switch s {
	case BorderNone:
		return "none"
	case BorderSingle:
		return "single"
	case BorderDouble:
		return "double"
	case BorderRounded:
		return "rounded"
	case BorderDashed:
		return "dashed"
	default:
		return "unknown"
	}
}

// HasBorder returns true if this style has a visible border
func (s BorderStyle) HasBorder() bool {
	return s != BorderNone
}

// =============================================================================
// Border Configuration
// =============================================================================

// Border defines border configuration for a container
type Border struct {
	// Style of the border
	Style BorderStyle

	// Width of border in characters (typically 1 for single line, 2 for double)
	// This is the space the border occupies
	Width int

	// Optional label for the border (displayed on top edge)
	Label string
}

// NewBorder creates a new border with the specified style
// Note: Width is always 1 for visible borders because all border glyphs
// occupy a single character cell.
func NewBorder(style BorderStyle) Border {
	width := 0
	if style != BorderNone {
		width = 1
	}
	return Border{
		Style: style,
		Width: width,
	}
}

// NewBorderWithLabel creates a new border with a label
func NewBorderWithLabel(style BorderStyle, label string) Border {
	b := NewBorder(style)
	b.Label = label
	return b
}

// HasBorder returns true if this border has a visible style
func (b Border) HasBorder() bool {
	return b.Style.HasBorder()
}

// HorizontalPadding returns the horizontal space taken by border (left + right)
// Note: Each border side uses 1 character cell, regardless of visual "width"
func (b Border) HorizontalPadding() int {
	if !b.HasBorder() {
		return 0
	}
	// Left border (1) + Right border (1) = 2
	return 2
}

// VerticalPadding returns the vertical space taken by border (top + bottom)
// Note: Each border side uses 1 character cell, regardless of visual "width"
func (b Border) VerticalPadding() int {
	if !b.HasBorder() {
		return 0
	}
	// Top border (1) + Bottom border (1) = 2
	return 2
}

// ContentOffset returns the x,y offset for content inside the border
// Note: Border characters (┌─┐│└┘ or ╔═╗║╚╝) each occupy 1 character cell,
// so content offset is always 1 for bordered containers, regardless of visual "width"
func (b Border) ContentOffset() (x, y int) {
	if !b.HasBorder() {
		return 0, 0
	}
	// All border styles use 1-char-wide glyphs, so offset is always 1
	return 1, 1
}

// LabelPadding returns the extra horizontal padding needed for labels
// Labels require an additional 2 characters of horizontal space for visual balance
func (b Border) LabelPadding() int {
	if !b.HasBorder() || b.Label == "" {
		return 0
	}
	return 2
}

// TotalHorizontalPadding returns the total horizontal space taken by border including label
// This = HorizontalPadding() + LabelPadding()
func (b Border) TotalHorizontalPadding() int {
	return b.HorizontalPadding() + b.LabelPadding()
}

// =============================================================================
// Bordered Interface
// =============================================================================

// Bordered nodes can provide border configuration
// This is an optional interface that nodes can implement
type Bordered interface {
	Node

	// GetBorder returns the border configuration
	GetBorder() Border
}

// =============================================================================
// BorderedNode Wrapper
// =============================================================================

// BorderedNode wraps a node to add border layout behavior
// This is useful for adding borders to nodes that don't implement Bordered
type BorderedNode struct {
	child  Node
	border Border
	id     string
}

// NewBorderedNode creates a new bordered wrapper around a child node
func NewBorderedNode(id string, child Node, border Border) *BorderedNode {
	return &BorderedNode{
		id:     id,
		child:  child,
		border: border,
	}
}

// ID returns the node identifier
func (n *BorderedNode) ID() string {
	return n.id
}

// Type returns the node type
func (n *BorderedNode) Type() string {
	return "bordered"
}

// Children returns the child nodes
func (n *BorderedNode) Children() []Node {
	if n.child == nil {
		return nil
	}
	return []Node{n.child}
}

// GetPosition returns the current position
func (n *BorderedNode) GetPosition() (x, y int) {
	return 0, 0
}

// SetPosition sets the position
func (n *BorderedNode) SetPosition(x, y int) {
	// Position is handled by parent layout
}

// GetSize returns the current size
func (n *BorderedNode) GetSize() (width, height int) {
	if n.child == nil {
		return 0, 0
	}
	w, h := n.child.GetSize()
	return w + n.border.HorizontalPadding(), h + n.border.VerticalPadding()
}

// SetSize sets the size
func (n *BorderedNode) SetSize(width, height int) {
	// Size is calculated during layout
}

// GetWidth returns the width
func (n *BorderedNode) GetWidth() int {
	w, _ := n.GetSize()
	return w
}

// GetHeight returns the height
func (n *BorderedNode) GetHeight() int {
	_, h := n.GetSize()
	return h
}

// GetBorder returns the border configuration
func (n *BorderedNode) GetBorder() Border {
	return n.border
}

// GetChild returns the wrapped child node
func (n *BorderedNode) GetChild() Node {
	return n.child
}

// MeasureInner returns the size needed for inner content given outer constraints
func (n *BorderedNode) MeasureInner(constraints Constraints) Constraints {
	if !n.border.HasBorder() {
		return constraints
	}

	// Reduce constraints by border size
	return Constraints{
		MinWidth:  max(0, constraints.MinWidth-n.border.HorizontalPadding()),
		MaxWidth:  max(0, constraints.MaxWidth-n.border.HorizontalPadding()),
		MinHeight: max(0, constraints.MinHeight-n.border.VerticalPadding()),
		MaxHeight: max(0, constraints.MaxHeight-n.border.VerticalPadding()),
	}
}

// MeasureOuter returns the outer size given inner content size
func (n *BorderedNode) MeasureOuter(innerWidth, innerHeight int) (int, int) {
	if !n.border.HasBorder() {
		return innerWidth, innerHeight
	}
	return innerWidth + n.border.HorizontalPadding(), innerHeight + n.border.VerticalPadding()
}

// =============================================================================
// Helper Functions
// =============================================================================

// isBordered checks if a node implements Bordered interface
func isBordered(node Node) bool {
	_, ok := node.(Bordered)
	return ok
}

// GetBorderFromNode safely gets border from a node
func GetBorderFromNode(node Node) Border {
	if node == nil {
		return Border{Style: BorderNone}
	}
	if b, ok := node.(Bordered); ok {
		return b.GetBorder()
	}
	return Border{Style: BorderNone}
}

// HasBorder checks if a node has a visible border
func HasBorder(node Node) bool {
	border := GetBorderFromNode(node)
	return border.HasBorder()
}

// CalculateBorderConstraints adjusts constraints for border content
func CalculateBorderConstraints(constraints Constraints, border Border) Constraints {
	if !border.HasBorder() {
		return constraints
	}

	return Constraints{
		MinWidth:  max(0, constraints.MinWidth-border.HorizontalPadding()),
		MaxWidth:  max(0, constraints.MaxWidth-border.HorizontalPadding()),
		MinHeight: max(0, constraints.MinHeight-border.VerticalPadding()),
		MaxHeight: max(0, constraints.MaxHeight-border.VerticalPadding()),
	}
}

// CalculateBorderBoxSize calculates the outer box size including border
func CalculateBorderBoxSize(contentWidth, contentHeight int, border Border) (int, int) {
	if !border.HasBorder() {
		return contentWidth, contentHeight
	}
	return contentWidth + border.HorizontalPadding(), contentHeight + border.VerticalPadding()
}