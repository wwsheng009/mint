package paint

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// PaintableNode - Abstract Node Interface for Paint Engine
// =============================================================================
// This interface decouples the Paint Engine from VNode/Fiber dependencies.
// The active renderer builds PaintableBox trees from Fiber layout results before
// paint, so the paint package does not depend on declarative node types.

// NodeType represents the type of a paintable node.
type NodeType int

const (
	// NodeTypeText represents a text node
	NodeTypeText NodeType = iota
	// NodeTypeElement represents an element node (e.g., button, input)
	NodeTypeElement
	// NodeTypeComponent represents a component node
	NodeTypeComponent
	// NodeTypeFragment represents a fragment node
	NodeTypeFragment
)

// String returns the string representation of the node type.
func (t NodeType) String() string {
	switch t {
	case NodeTypeText:
		return "Text"
	case NodeTypeElement:
		return "Element"
	case NodeTypeComponent:
		return "Component"
	case NodeTypeFragment:
		return "Fragment"
	default:
		return "Unknown"
	}
}

// PaintableNode is the interface that the Paint Engine operates on.
// This is the minimal abstraction needed for rendering, decoupled from VNode/Fiber.
//
// Components can implement this interface directly, or adapters can wrap
// existing VNode/Fiber types to satisfy this interface.
type PaintableNode interface {
	// ID returns the unique identifier for this node.
	// 重要，会使用在paint cache,需要注意ID要唯一
	ID() string

	// NodeType returns the type of this node.
	NodeType() NodeType

	// Tag returns the tag name for element nodes (e.g., "button", "text", "hstack").
	// Returns empty string for non-element nodes.
	Tag() string

	// Style returns the current style of this node.
	Style() style.Style

	// SetStyle sets the style for this node.
	SetStyle(s style.Style)

	// TextContent returns the text content of this node.
	// For text nodes, this is the text itself.
	// For elements, this may return combined text content.
	TextContent() string

	// Paint generates draw commands for rendering this node.
	// This is inherited from the Paintable interface.
	// Returns nil if the node has no custom paint logic.
	Paint(x, y int) []DrawCmd
}

// =============================================================================
// Border Information Interface (Optional)
// =============================================================================

// BorderStyle represents the style of a border.
// BorderStyle is an alias to layout.BorderStyle for unified type definition.
// See runtime/layout/border.go for the canonical definition.
type BorderStyle = layout.BorderStyle

// BorderStyle constants - aliases to layout package
const (
	BorderStyleNone    = layout.BorderNone
	BorderStyleSingle  = layout.BorderSingle
	BorderStyleDouble  = layout.BorderDouble
	BorderStyleRounded = layout.BorderRounded
	BorderStyleDashed  = layout.BorderDashed
)

// BorderInfo is an optional interface that nodes can implement
// to provide border information for rendering.
type BorderInfo interface {
	// GetBorderStyle returns the border style.
	GetBorderStyle() BorderStyle
	// GetBorderColor returns the border color.
	GetBorderColor() string
	// GetBorderLabel returns the optional border label.
	GetBorderLabel() string
}

// =============================================================================
// Layout Information Interface (Optional)
// =============================================================================

// LayoutDirection represents the direction of a layout container.
type LayoutDirection int

const (
	// DirectionColumn means children are laid out vertically
	DirectionColumn LayoutDirection = iota
	// DirectionRow means children are laid out horizontally
	DirectionRow
)

// LayoutAlignType represents alignment along the main or cross axis.
type LayoutAlignType int

const (
	// LayoutAlignStart aligns items to the start
	LayoutAlignStart LayoutAlignType = iota
	// LayoutAlignCenter centers items
	LayoutAlignCenter
	// LayoutAlignEnd aligns items to the end
	LayoutAlignEnd
	// LayoutAlignSpaceBetween distributes space between items
	LayoutAlignSpaceBetween
	// LayoutAlignSpaceAround distributes space around items
	LayoutAlignSpaceAround
)

// LayoutInfo is an optional interface that nodes can implement
// to provide layout information.
type LayoutInfo interface {
	// GetDirection returns the layout direction (row or column).
	GetDirection() LayoutDirection
	// GetGap returns the gap between children.
	GetGap() int
	// GetFlex returns the flex factor for this node.
	GetFlex() int
	// GetAlign returns the main axis alignment.
	GetAlign() LayoutAlignType
	// GetCrossAlign returns the cross axis alignment.
	GetCrossAlign() LayoutAlignType
	// GetPadding returns the padding [top, right, bottom, left].
	GetPadding() [4]int
}

// =============================================================================
// Helper Functions
// =============================================================================

// IsTextType checks if the node type is text.
func IsTextType(t NodeType) bool {
	return t == NodeTypeText
}

// IsElementType checks if the node type is element.
func IsElementType(t NodeType) bool {
	return t == NodeTypeElement
}

// IsContainerType checks if the node can contain children.
func IsContainerType(t NodeType) bool {
	return t == NodeTypeElement || t == NodeTypeComponent || t == NodeTypeFragment
}
