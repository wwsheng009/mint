package ui

import rtui "github.com/wwsheng009/mint/runtime/ui"

// =============================================================================
// Layout Types (re-exported from runtime/ui)
// =============================================================================

// Direction represents layout direction
type Direction = rtui.Direction

const (
	// DirectionRow is horizontal layout
	DirectionRow = rtui.DirectionRow
	// DirectionColumn is vertical layout
	DirectionColumn = rtui.DirectionColumn
)

// Align represents alignment options
type Align = rtui.Align

const (
	// AlignStart aligns to the start
	AlignStart = rtui.AlignStart
	// AlignCenter aligns to center
	AlignCenter = rtui.AlignCenter
	// AlignEnd aligns to end
	AlignEnd = rtui.AlignEnd
	// AlignSpaceBetween adds space between items
	AlignSpaceBetween = rtui.AlignSpaceBetween
	// AlignSpaceAround adds space around items
	AlignSpaceAround = rtui.AlignSpaceAround
)

// LayoutNode represents a layout container
type LayoutNode = rtui.LayoutNode

// HStack creates a horizontal layout container
func HStack(children ...VNode) VNode {
	return rtui.HStack(children...)
}

// VStack creates a vertical layout container
func VStack(children ...VNode) VNode {
	return rtui.VStack(children...)
}

// Box creates a container box
func Box() *rtui.BoxLayoutBuilder {
	return rtui.Box()
}

// Spacer creates a flexible space
func Spacer() *rtui.SpacerBuilder {
	return rtui.Spacer()
}
