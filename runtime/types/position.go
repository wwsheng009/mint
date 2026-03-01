// Package types provides common type definitions shared across all runtime packages.
// This package has zero dependencies on other runtime packages to avoid import cycles.
package types

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

	// PositionFixed positioned relative to viewport
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

// String returns the string representation of Anchor
func (a Anchor) String() string {
	switch a {
	case AnchorTopLeft:
		return "topleft"
	case AnchorTop:
		return "top"
	case AnchorTopRight:
		return "topright"
	case AnchorLeft:
		return "left"
	case AnchorCenter:
		return "center"
	case AnchorRight:
		return "right"
	case AnchorBottomLeft:
		return "bottomleft"
	case AnchorBottom:
		return "bottom"
	case AnchorBottomRight:
		return "bottomright"
	default:
		return "unknown"
	}
}
