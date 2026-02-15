// Package core provides the abstract layout node interface.
// This allows VNode and Fiber to share the same layout logic.
package core

import "github.com/wwsheng009/mint/runtime"

// Node defines the interface for nodes that participate in layout.
// Both VNode and Fiber implement this interface, allowing shared layout logic.
//
// The interface provides:
// - Identity: Tag, Key for identification
// - Layout properties: Direction, Gap, Flex, Align, Padding, etc.
// - Children access: GetChildren returns child Nodes
// - Props access: GetProp for property access
type Node interface {
	// Identity
	GetTag() string
	GetKey() string

	// Layout properties
	GetDirection() Direction      // Column or Row
	GetGap() int
	GetFlex() int
	GetAlign() Align              // main axis alignment
	GetCrossAlign() Align         // cross axis alignment
	GetPadding() [4]int           // top, right, bottom, left
	GetStretchCross() bool
	GetFillWidth() bool
	GetFillHeight() bool

	// Children - returns slice of Node
	GetChildren() []Node

	// Props access
	GetProp(name string) interface{}

	// Measurement
	Measure(constraints runtime.BoxConstraints) runtime.Size
}

// Direction represents layout direction
type Direction int

const (
	DirectionColumn Direction = iota
	DirectionRow
)

// Align represents alignment options
type Align int

const (
	AlignStart Align = iota
	AlignCenter
	AlignEnd
	AlignSpaceBetween
	AlignSpaceAround
)

// Measurer is the single-pass layout interface.
// It returns both size and child constraints for efficient layout.
type Measurer interface {
	Node
	MeasureLayout(measurer ChildMeasurer, constraints runtime.BoxConstraints) runtime.LayoutMeasurement
}

// ChildMeasurer measures children using the Node interface.
type ChildMeasurer interface {
	MeasureChild(child Node, constraints runtime.BoxConstraints) runtime.Size
}

// Info contains layout properties extracted from a Node.
// This is the unified layout info structure used by the layout algorithm.
type Info struct {
	IsHorizontal  bool
	Gap           int
	Flex          int
	Align         Align
	CrossAlign    Align
	Padding       [4]int
	StretchCross  bool
	FillWidth     bool
	FillHeight    bool
	IsBordered    bool
	IsText        bool
	TextContent   string
	ExplicitWidth int  // 0 means not set
	ExplicitHeight int // 0 means not set
}

// GetInfo extracts layout information from a Node.
func GetInfo(node Node) Info {
	info := Info{}

	if node == nil {
		return info
	}

	// Get direction
	info.IsHorizontal = node.GetDirection() == DirectionRow

	// Get layout properties
	info.Gap = node.GetGap()
	info.Flex = node.GetFlex()
	info.Align = node.GetAlign()
	info.CrossAlign = node.GetCrossAlign()
	info.Padding = node.GetPadding()
	info.StretchCross = node.GetStretchCross()
	info.FillWidth = node.GetFillWidth()
	info.FillHeight = node.GetFillHeight()

	// Check tag for special handling
	tag := node.GetTag()
	info.IsBordered = tag == "bordered"
	info.IsText = tag == "text"

	// Get explicit width/height from props
	if w, ok := node.GetProp("width").(int); ok && w > 0 {
		info.ExplicitWidth = w
	}
	if h, ok := node.GetProp("height").(int); ok && h > 0 {
		info.ExplicitHeight = h
	}

	// Get text content for text nodes
	if info.IsText {
		if content, ok := node.GetProp("content").(string); ok {
			info.TextContent = content
		}
	}

	return info
}
