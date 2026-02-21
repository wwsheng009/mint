// Package layout provides positioning types for absolute and relative layout
package layout

// =============================================================================
// Position Types
// =============================================================================

// PositionType defines the positioning scheme
type PositionType int

const (
	// PositionRelative normal flow positioning (default)
	PositionRelative PositionType = iota

	// PositionAbsolute positioned relative to nearest positioned ancestor
	PositionAbsolute

	// PositionFixed positioned relative to viewport (not yet implemented)
	PositionFixed
)

// String returns the string representation of PositionType
func (p PositionType) String() string {
	switch p {
	case PositionRelative:
		return "relative"
	case PositionAbsolute:
		return "absolute"
	case PositionFixed:
		return "fixed"
	default:
		return "unknown"
	}
}

// =============================================================================
// Anchor Types
// =============================================================================

// Anchor defines how a positioned element aligns to its position
type Anchor int

const (
	// AnchorTopLeft element's top-left corner at position (default)
	AnchorTopLeft Anchor = iota
	// AnchorTop element's top-center at position
	AnchorTop
	// AnchorTopRight element's top-right corner at position
	AnchorTopRight
	// AnchorLeft element's center-left at position
	AnchorLeft
	// AnchorCenter element's center at position
	AnchorCenter
	// AnchorRight element's center-right at position
	AnchorRight
	// AnchorBottomLeft element's bottom-left corner at position
	AnchorBottomLeft
	// AnchorBottom element's bottom-center at position
	AnchorBottom
	// AnchorBottomRight element's bottom-right corner at position
	AnchorBottomRight
)

// =============================================================================
// Position Value Types (for percentage support)
// =============================================================================

// PositionValue represents either absolute or relative (percentage) position
type PositionValue interface {
	isPositionValue()
	Resolve(containerSize int) int
}

// AbsolutePos is an absolute position in cells
type AbsolutePos int

func (p AbsolutePos) isPositionValue() {}
func (p AbsolutePos) Resolve(_ int) int { return int(p) }

// RelativePos is a relative position as percentage (0-100)
type RelativePos int

func (p RelativePos) isPositionValue() {}
func (p RelativePos) Resolve(containerSize int) int {
	return containerSize * int(p) / 100
}

// Position defines absolute positioning properties
// For absolute positioning, offsets specify distance from edges of containing block
type Position struct {
	// Type of positioning
	Type PositionType

	// Top offset from top edge (nil = auto)
	Top *int

	// Left offset from left edge (nil = auto)
	Left *int

	// Right offset from right edge (nil = auto)
	Right *int

	// Bottom offset from bottom edge (nil = auto)
	Bottom *int
}

// IsAbsolute returns true if this is absolute positioning
func (p Position) IsAbsolute() bool {
	return p.Type == PositionAbsolute
}

// IsRelative returns true if this is relative positioning
func (p Position) IsRelative() bool {
	return p.Type == PositionRelative
}

// HasTop returns true if top offset is set
func (p Position) HasTop() bool {
	return p.Top != nil
}

// HasLeft returns true if left offset is set
func (p Position) HasLeft() bool {
	return p.Left != nil
}

// HasRight returns true if right offset is set
func (p Position) HasRight() bool {
	return p.Right != nil
}

// HasBottom returns true if bottom offset is set
func (p Position) HasBottom() bool {
	return p.Bottom != nil
}

// NewRelativePosition creates a relative position
func NewRelativePosition() Position {
	return Position{Type: PositionRelative}
}

// NewAbsolutePosition creates an absolute position with all offsets as auto
func NewAbsolutePosition() Position {
	return Position{Type: PositionAbsolute}
}

// NewAbsolutePositionWithOffsets creates an absolute position with specified offsets
func NewAbsolutePositionWithOffsets(top, left, right, bottom *int) Position {
	return Position{
		Type:   PositionAbsolute,
		Top:    top,
		Left:   left,
		Right:  right,
		Bottom: bottom,
	}
}

// =============================================================================
// Positionable Interface
// =============================================================================

// Positionable nodes can provide position type and offsets
// This is an optional interface that nodes can implement
type Positionable interface {
	Node

	// GetPositionType returns the position configuration
	GetPositionType() Position
}

// =============================================================================
// Absolute Layout Helpers
// =============================================================================

// AbsoluteLayoutResult contains the calculated absolute position
type AbsoluteLayoutResult struct {
	// NodeID identifies which node this result belongs to
	NodeID string

	// X is the final absolute X position
	X int

	// Y is the final absolute Y position
	Y int

	// Width is the node's width (from layout)
	Width int

	// Height is the node's height (from layout)
	Height int
}

// CalculateAbsolutePosition computes the absolute position for a positioned node
// parentWidth/Height: dimensions of the containing block
// nodeWidth/Height: dimensions of the node (from layout)
// position: position configuration
func CalculateAbsolutePosition(parentWidth, parentHeight, nodeWidth, nodeHeight int, position Position) (x, y int) {
	x, y = 0, 0

	// Calculate X position
	if position.HasLeft() {
		x = *position.Left
	} else if position.HasRight() {
		// Right offset: X = ParentWidth - Right - NodeWidth
		x = parentWidth - *position.Right - nodeWidth
	}

	// Calculate Y position
	if position.HasTop() {
		y = *position.Top
	} else if position.HasBottom() {
		// Bottom offset: Y = ParentHeight - Bottom - NodeHeight
		y = parentHeight - *position.Bottom - nodeHeight
	}

	return x, y
}

// =============================================================================
// LayoutBox Position Extensions
// =============================================================================

// PositionedLayoutBox extends LayoutBox with position information
type PositionedLayoutBox struct {
	*LayoutBox

	// PositionType indicates how this box is positioned
	PositionType PositionType

	// AbsoluteX is the absolute X coordinate (for absolute positioning)
	AbsoluteX int

	// AbsoluteY is the absolute Y coordinate (for absolute positioning)
	AbsoluteY int
}

// IsAbsolute returns true if this box uses absolute positioning
func (b *PositionedLayoutBox) IsAbsolute() bool {
	return b.PositionType == PositionAbsolute
}

// GetEffectivePosition returns the effective position (absolute or relative)
func (b *PositionedLayoutBox) GetEffectivePosition() (x, y int) {
	if b.IsAbsolute() {
		return b.AbsoluteX, b.AbsoluteY
	}
	return b.LayoutBox.X, b.LayoutBox.Y
}

// =============================================================================
// Helper Functions
// =============================================================================

// isPositionable checks if a node implements Positionable interface
func isPositionable(node Node) bool {
	_, ok := node.(Positionable)
	return ok
}

// getPositionType returns the position of a node, or relative if not implemented
func getPositionType(node Node) Position {
	if p, ok := node.(Positionable); ok {
		return p.GetPositionType()
	}
	return NewRelativePosition()
}

// IsPositionAbsolute checks if a node has absolute positioning
func IsPositionAbsolute(node Node) bool {
	return getPositionType(node).IsAbsolute()
}

// GetPositionFromNode safely gets position from a node
func GetPositionFromNode(node Node) Position {
	if node == nil {
		return NewRelativePosition()
	}
	return getPositionType(node)
}

// =============================================================================
// Absolute Style (for new absolute positioning)
// =============================================================================

// AbsoluteStyle defines absolute positioning style
type AbsoluteStyle struct {
	// Left position (absolute or percentage)
	Left PositionValue

	// Top position (absolute or percentage)
	Top PositionValue

	// Right position (absolute or percentage)
	Right PositionValue

	// Bottom position (absolute or percentage)
	Bottom PositionValue

	// Anchor alignment point
	Anchor Anchor

	// Width explicit width (0 = auto)
	Width int

	// Height explicit height (0 = auto)
	Height int

	// ZIndex stacking order
	ZIndex int
}

// AbsoluteStyleProvider defines the interface for absolute positioned nodes
type AbsoluteStyleProvider interface {
	Node

	// GetAbsoluteStyle returns the absolute positioning style
	GetAbsoluteStyle() *AbsoluteStyle
}

// NewAbsoluteStyle creates default absolute style
func NewAbsoluteStyle() *AbsoluteStyle {
	return &AbsoluteStyle{
		Anchor: AnchorTopLeft,
	}
}

// CalculatePosition calculates absolute position based on container size
func (s *AbsoluteStyle) CalculatePosition(containerWidth, containerHeight, nodeWidth, nodeHeight int) (x, y int) {
	x = 0
	y = 0

	// Calculate X position
	if s.Left != nil {
		x = s.Left.Resolve(containerWidth)
	} else if s.Right != nil {
		rightPos := s.Right.Resolve(containerWidth)
		x = containerWidth - rightPos - nodeWidth
	}

	// Calculate Y position
	if s.Top != nil {
		y = s.Top.Resolve(containerHeight)
	} else if s.Bottom != nil {
		bottomPos := s.Bottom.Resolve(containerHeight)
		y = containerHeight - bottomPos - nodeHeight
	}

	// Apply anchor adjustment
	w := nodeWidth
	if s.Width > 0 {
		w = s.Width
	}
	h := nodeHeight
	if s.Height > 0 {
		h = s.Height
	}

	switch s.Anchor {
	case AnchorTopLeft:
		// No adjustment
	case AnchorTop:
		x = x - w/2
	case AnchorTopRight:
		x = x - w
	case AnchorLeft:
		y = y - h/2
	case AnchorCenter:
		x = x - w/2
		y = y - h/2
	case AnchorRight:
		x = x - w
		y = y - h/2
	case AnchorBottomLeft:
		y = y - h
	case AnchorBottom:
		x = x - w/2
		y = y - h
	case AnchorBottomRight:
		x = x - w
		y = y - h
	}

	return x, y
}

// =============================================================================
// Absolute Layout Implementation
// =============================================================================

// AbsoluteLayout handles absolute positioned children
type AbsoluteLayout struct {
	id       string
	children []Node
	styles   []*AbsoluteStyle
	size     Size
	position Point
}

// NewAbsoluteLayout creates a new absolute layout container
func NewAbsoluteLayout(id string) *AbsoluteLayout {
	return &AbsoluteLayout{
		id:     id,
		styles: make([]*AbsoluteStyle, 0),
	}
}

// SetChildren sets children with their absolute styles
func (a *AbsoluteLayout) SetChildren(children []Node, styles []*AbsoluteStyle) {
	a.children = children
	a.styles = styles
}

// ID returns the node identifier
func (a *AbsoluteLayout) ID() string {
	return a.id
}

// Type returns the node type
func (a *AbsoluteLayout) Type() string {
	return "absolute"
}

// Children returns child nodes
func (a *AbsoluteLayout) Children() []Node {
	return a.children
}

// GetPosition returns the current position
func (a *AbsoluteLayout) GetPosition() (int, int) {
	return a.position.X, a.position.Y
}

// SetPosition sets the position
func (a *AbsoluteLayout) SetPosition(x, y int) {
	a.position.X = x
	a.position.Y = y
}

// GetSize returns the current size
func (a *AbsoluteLayout) GetSize() (int, int) {
	return a.size.Width, a.size.Height
}

// SetSize sets the size
func (a *AbsoluteLayout) SetSize(width, height int) {
	a.size.Width = width
	a.size.Height = height
}

// GetWidth returns the width
func (a *AbsoluteLayout) GetWidth() int {
	return a.size.Width
}

// GetHeight returns the height
func (a *AbsoluteLayout) GetHeight() int {
	return a.size.Height
}

// Measure measures the layout (returns container size, children are absolute)
func (a *AbsoluteLayout) Measure(constraints Constraints) Size {
	// Absolute positioned children don't affect container size
	// Return minimum size or explicit size
	width := constraints.ConstrainWidth(0)
	height := constraints.ConstrainHeight(0)
	return Size{Width: width, Height: height}
}

// LayoutChildren positions absolute children and returns their LayoutBoxes
func (a *AbsoluteLayout) LayoutChildren(containerWidth, containerHeight int) []LayoutBox {
	if len(a.children) == 0 {
		return nil
	}

	boxes := make([]LayoutBox, len(a.children))

	for i, child := range a.children {
		if child == nil {
			continue
		}

		// Get style for this child
		var style *AbsoluteStyle
		if i < len(a.styles) && a.styles[i] != nil {
			style = a.styles[i]
		} else {
			style = NewAbsoluteStyle()
		}

		// Measure child
		var childW, childH int
		if measurable, ok := child.(Measurable); ok {
			size := measurable.Measure(UnboundedConstraints())
			childW = size.Width
			childH = size.Height
		} else {
			childW, childH = child.GetSize()
		}

		// Use explicit size if set
		if style.Width > 0 {
			childW = style.Width
		}
		if style.Height > 0 {
			childH = style.Height
		}

		// Calculate position
		x, y := style.CalculatePosition(containerWidth, containerHeight, childW, childH)

		boxes[i] = LayoutBox{
			ID:     child.ID(),
			X:      x,
			Y:      y,
			Width:  childW,
			Height: childH,
			ZIndex: style.ZIndex,
		}

		child.SetPosition(x, y)
		child.SetSize(childW, childH)
	}

	return boxes
}
