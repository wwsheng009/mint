package ui

import "github.com/wwsheng009/mint/runtime/style"

// Direction represents layout direction
type Direction int

const (
	// DirectionRow is horizontal layout
	DirectionRow Direction = iota
	// DirectionColumn is vertical layout
	DirectionColumn
)

// Align represents alignment options
type Align int

const (
	// AlignStart aligns to the start
	AlignStart Align = iota
	// AlignCenter aligns to center
	AlignCenter
	// AlignEnd aligns to end
	AlignEnd
	// AlignSpaceBetween adds space between items
	AlignSpaceBetween
	// AlignSpaceAround adds space around items
	AlignSpaceAround
)

// LayoutNode represents a layout container
type LayoutNode struct {
	*ElementVNode
	direction Direction
	align     Align
	crossAlign Align
	gap       int
	padding   [4]int // top, right, bottom, left
}

// HStack creates a horizontal layout container
func HStack(children ...VNode) VNode {
	builder := &LayoutBuilder{
		node: &LayoutNode{
			ElementVNode: NewElement("hstack"),
			direction:    DirectionRow,
			align:        AlignStart,
			crossAlign:   AlignStart,
			gap:          1, // Default gap of 1 space between items
			padding:      [4]int{0, 0, 0, 0},
		},
		children: children,
	}
	return builder.Build()
}

// VStack creates a vertical layout container
func VStack(children ...VNode) VNode {
	builder := &LayoutBuilder{
		node: &LayoutNode{
			ElementVNode: NewElement("vstack"),
			direction:    DirectionColumn,
			align:        AlignStart,
			crossAlign:   AlignStart,
			gap:          0,
			padding:      [4]int{0, 0, 0, 0},
		},
		children: children,
	}
	return builder.Build()
}

// LayoutBuilder provides fluent API for building layouts
type LayoutBuilder struct {
	node     *LayoutNode
	children []VNode
}

// Align sets the main axis alignment
func (b *LayoutBuilder) Align(a Align) *LayoutBuilder {
	b.node.align = a
	return b
}

// AlignCross sets the cross axis alignment
func (b *LayoutBuilder) AlignCross(a Align) *LayoutBuilder {
	b.node.crossAlign = a
	return b
}

// Gap sets the spacing between children
func (b *LayoutBuilder) Gap(n int) *LayoutBuilder {
	b.node.gap = n
	return b
}

// Padding sets the padding (top, right, bottom, left)
func (b *LayoutBuilder) Padding(top, right, bottom, left int) *LayoutBuilder {
	b.node.padding = [4]int{top, right, bottom, left}
	return b
}

// Width sets the width
func (b *LayoutBuilder) Width(n int) *LayoutBuilder {
	b.node.SetProp("width", n)
	return b
}

// Height sets the height
func (b *LayoutBuilder) Height(n int) *LayoutBuilder {
	b.node.SetProp("height", n)
	return b
}

// Flex sets the flex factor
func (b *LayoutBuilder) Flex(n int) *LayoutBuilder {
	b.node.SetProp("flex", n)
	return b
}

// Style sets the visual style
func (b *LayoutBuilder) Style(s style.Style) *LayoutBuilder {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *LayoutBuilder) FgColor(c interface{}) *LayoutBuilder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.FG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.FG = color
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *LayoutBuilder) BgColor(c interface{}) *LayoutBuilder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.BG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.BG = color
		b.node.SetStyle(s)
	}
	return b
}

// Key sets the key for diffing
func (b *LayoutBuilder) Key(key string) *LayoutBuilder {
	b.node.SetKey(key)
	return b
}

// Build returns the VNode
func (b *LayoutBuilder) Build() VNode {
	// Set children on the element
	b.node.SetChildren(b.children)
	return b.node
}

// Direction returns the layout direction
func (l *LayoutNode) Direction() Direction {
	return l.direction
}

// Align returns the main axis alignment
func (l *LayoutNode) Align() Align {
	return l.align
}

// CrossAlign returns the cross axis alignment
func (l *LayoutNode) CrossAlign() Align {
	return l.crossAlign
}

// Gap returns the gap between children
func (l *LayoutNode) Gap() int {
	return l.gap
}

// Padding returns the padding
func (l *LayoutNode) Padding() [4]int {
	return l.padding
}

// =============================================================================
// Box Layout
// =============================================================================

// Box creates a container box
func Box() *BoxLayoutBuilder {
	return &BoxLayoutBuilder{
		node: NewElement("box"),
	}
}

// BoxLayoutBuilder provides fluent API for building boxes
type BoxLayoutBuilder struct {
	node *ElementVNode
}

// Border adds a border
func (b *BoxLayoutBuilder) Border(v bool) *BoxLayoutBuilder {
	b.node.SetProp("border", v)
	return b
}

// BorderStyle sets the border style
func (b *BoxLayoutBuilder) BorderStyle(s string) *BoxLayoutBuilder {
	b.node.SetProp("borderStyle", s)
	return b
}

// Padding sets the padding
func (b *BoxLayoutBuilder) Padding(n int) *BoxLayoutBuilder {
	b.node.SetProp("padding", n)
	return b
}

// Background sets the background color
func (b *BoxLayoutBuilder) Background(c interface{}) *BoxLayoutBuilder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.BG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.BG = color
		b.node.SetStyle(s)
	}
	return b
}

// Child sets the child
func (b *BoxLayoutBuilder) Child(child VNode) *BoxLayoutBuilder {
	b.node.SetChildren([]VNode{child})
	return b
}

// Width sets the width
func (b *BoxLayoutBuilder) Width(n int) *BoxLayoutBuilder {
	b.node.SetProp("width", n)
	return b
}

// Height sets the height
func (b *BoxLayoutBuilder) Height(n int) *BoxLayoutBuilder {
	b.node.SetProp("height", n)
	return b
}

// Flex sets the flex factor
func (b *BoxLayoutBuilder) Flex(n int) *BoxLayoutBuilder {
	b.node.SetProp("flex", n)
	return b
}

// Build returns the VNode
func (b *BoxLayoutBuilder) Build() VNode {
	return b.node
}

// =============================================================================
// Spacer
// =============================================================================

// Spacer creates a flexible space
func Spacer() *SpacerBuilder {
	return &SpacerBuilder{
		node: NewElement("spacer"),
	}
}

// SpacerBuilder provides fluent API for spacer
type SpacerBuilder struct {
	node *ElementVNode
}

// Flex sets the flex factor
func (b *SpacerBuilder) Flex(n int) *SpacerBuilder {
	b.node.SetProp("flex", n)
	return b
}

// Build returns the VNode
func (b *SpacerBuilder) Build() VNode {
	return b.node
}
