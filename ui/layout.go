package ui

import (
	"strings"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	alertcomp "github.com/wwsheng009/mint/ui/components/alert"
	"github.com/wwsheng009/mint/ui/components/wrap"
)

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

// Example: ui.NewHStack().SetGap(2).SetChildrenList([]ui.VNode{...})
// This returns *LayoutBuilder which implements VNode interface, so it can be used as VNode
func NewHStack() *rtui.LayoutBuilder {
	return rtui.HStackBuilder()
}

// VStack creates a vertical layout container
func VStack(children ...VNode) VNode {
	return rtui.VStack(children...)
}

// SectionBreak creates an explicit blank row between sequential page sections.
//
// Use this when the visual gap is intentional. Use nil or OptionalSection for
// content that should disappear from PageStack and panel stacks.
func SectionBreak() VNode {
	return Text("")
}

// OptionalSection returns node when show is true, otherwise nil.
//
// It is a small readability helper for page composition where a stage is
// genuinely absent, not a visible blank row.
func OptionalSection(show bool, node VNode) VNode {
	if !show {
		return nil
	}
	return node
}

// PageStack stacks page sections vertically and skips nil optional sections.
func PageStack(children ...VNode) VNode {
	nodes := make([]VNode, 0, len(children))
	for _, child := range children {
		if child != nil {
			nodes = append(nodes, child)
		}
	}
	return rtui.VStack(nodes...)
}

// PageStackWithAlert stacks page sections and optionally inserts an alert after a stage.
func PageStackWithAlert(alertText string, alertAfter int, children ...VNode) VNode {
	nodes := make([]VNode, 0, len(children)+1)
	for _, child := range children {
		if child != nil {
			nodes = append(nodes, child)
		}
	}
	alertText = strings.TrimSpace(alertText)
	if alertText != "" {
		if alertAfter < 0 || alertAfter > len(nodes) {
			alertAfter = len(nodes)
		}
		alertNode := alertcomp.NewBuilder(alertText).Build()
		nodes = append(nodes[:alertAfter], append([]VNode{alertNode}, nodes[alertAfter:]...)...)
	}
	return rtui.VStack(nodes...)
}

// VStackBuilder creates a vertical layout container builder for method chaining
// Example: ui.VStackBuilder(item1, item2).Stretch().Build()
func VStackBuilder(children ...VNode) *rtui.LayoutBuilder {
	return rtui.VStackBuilder(children...)
}

// Example: ui.NewVStack().SetGap(0).SetChildrenList([]ui.VNode{...})
// This returns *LayoutBuilder which implements VNode interface, so it can be used as VNode
func NewVStack() *rtui.LayoutBuilder {
	return rtui.VStackBuilder()
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

// SpacerWithFlex creates a flexible space with specified flex value (compatibility with ui/components/stack)
func SpacerWithFlex(flex int) VNode {
	return rtui.Spacer().Flex(flex).Build()
}

// =============================================================================
// Wrap Layout (re-exported from ui/components/wrap)
// =============================================================================

// Wrap creates a wrapping layout container
// Automatically wraps children to multiple rows based on width
// This is a convenience re-export from ui/components/wrap
func Wrap(children ...VNode) VNode {
	return wrap.Wrap(children...)
}

// WrapBuilder creates a wrapping layout container builder for method chaining
// This is a convenience re-export from ui/components/wrap
func WrapBuilder(children ...VNode) *wrap.Builder {
	return wrap.NewBuilder().Children(children...)
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

// BorderStyle defines the visual style of borders (internal use)
type BorderStyle = rtui.BorderStyle

const (
	BorderSingle  = rtui.BorderSingle
	BorderDouble  = rtui.BorderDouble
	BorderRounded = rtui.BorderRounded
	BorderDashed  = rtui.BorderDashed
	BorderNone    = rtui.BorderNone
)
