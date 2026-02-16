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
