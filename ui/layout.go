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

// HStackBuilder creates a horizontal layout container builder for method chaining
// Example: ui.HStackBuilder(item1, item2).Gap(0).Stretch().Build()
func HStackBuilder(children ...VNode) *rtui.LayoutBuilder {
	return rtui.HStackBuilder(children...)
}

// VStack creates a vertical layout container
func VStack(children ...VNode) VNode {
	return rtui.VStack(children...)
}

// VStackBuilder creates a vertical layout container builder for method chaining
// Example: ui.VStackBuilder(item1, item2).Stretch().Build()
func VStackBuilder(children ...VNode) *rtui.LayoutBuilder {
	return rtui.VStackBuilder(children...)
}

// Flex wraps a VNode to make it flexible in a layout
// The flex factor determines how much the child grows to fill available space
func Flex(vnode VNode, flexFactors ...int) VNode {
	return rtui.Flex(vnode, flexFactors...)
}

// Box creates a container box
func Box() *rtui.BoxLayoutBuilder {
	return rtui.Box()
}

// Spacer creates a flexible space
func Spacer() *rtui.SpacerBuilder {
	return rtui.Spacer()
}

// =============================================================================
// Table Layout Types (re-exported from runtime/ui)
// =============================================================================

// TableRow represents a row in a table layout
type TableRow = rtui.TableRow

// TableCell represents a cell in a table row
type TableCell = rtui.TableCell

// Table creates a table with row-based layout
func Table(rows ...VNode) VNode {
	return rtui.Table(rows...)
}

// Row creates a table row containing cells
func Row(cells ...VNode) VNode {
	return rtui.Row(cells...)
}

// Cell creates a table cell containing content
func Cell(content VNode) VNode {
	return rtui.Cell(content)
}

// =============================================================================
// Bordered Container (re-exported from runtime/ui)
// =============================================================================

// BorderStyle defines the visual style of borders
type BorderStyle = rtui.BorderStyle

const (
	BorderSingle = rtui.BorderSingle
	BorderDouble = rtui.BorderDouble
	BorderRounded = rtui.BorderRounded
	BorderDashed  = rtui.BorderDashed
	BorderNone    = rtui.BorderNone
)

// BorderedNode represents a container with auto-rendered border
type BorderedNode = rtui.BorderedNode

// Bordered creates a container with auto-rendered border
// The border is rendered outside the content area, not taking up content space
func Bordered() *rtui.BorderedBuilder {
	return rtui.Bordered()
}
