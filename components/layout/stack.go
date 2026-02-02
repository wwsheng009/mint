package layout

import (
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

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
	// AlignSpaceBetween adds space between items
	AlignSpaceBetween
	// AlignSpaceAround adds space around items
	AlignSpaceAround
)

// LayoutNode represents a layout container
type LayoutNode struct {
	*ui.ElementVNode
	direction  Direction
	align      Align
	crossAlign Align
	gap        int
	padding    [4]int // top, right, bottom, left
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

// Paint implements paint.Paintable interface
// Generates draw commands for the layout and its children
func (l *LayoutNode) Paint(x, y int) []paint.DrawCmd {
	if l == nil {
		return nil
	}

	var cmds []paint.DrawCmd
	children := l.Children()

	// Apply padding
	currentX := x + l.padding[3] // left padding
	currentY := y + l.padding[0] // top padding

	for i, child := range children {
		// Set child bounds for hit testing (if child supports SetBounds)
		l.setChildBounds(child, currentX, currentY)

		// Check if child implements Paintable
		if paintable, ok := child.(interface{ Paint(int, int) []paint.DrawCmd }); ok {
			// Child has custom paint logic
			childCmds := paintable.Paint(currentX, currentY)
			cmds = append(cmds, childCmds...)
		} else {
			// Fallback: render as text element
			if props := child.Props(); props != nil {
				if content := props.GetString("content"); content != "" {
					cmds = append(cmds, paint.DrawCmd{
						X:     currentX,
						Y:     currentY,
						Text:  content,
						Style: child.Style(),
					})
				}
			}
		}

		// Update position based on direction
		if l.direction == DirectionRow {
			// HStack: move horizontally
			// Simple width estimation
			childWidth := l.estimateChildWidth(child)
			currentX += childWidth
			if i < len(children)-1 {
				currentX += l.gap
			}
		} else {
			// VStack: move vertically
			currentY++
			if i < len(children)-1 {
				currentY += l.gap
			}
		}
	}

	return cmds
}

// setChildBounds sets the bounds of a child for hit testing
func (l *LayoutNode) setChildBounds(child ui.VNode, x, y int) {
	if child == nil {
		return
	}

	// Check if child implements SetBounds
	if boundsAware, ok := child.(interface{ SetBounds(x, y, width, height int) }); ok {
		width := 0
		height := 0

		// Try to get measured size from Measurable interface
		type measurable interface {
			Measure(constraints runtime.BoxConstraints) runtime.Size
		}
		if m, ok := child.(measurable); ok {
			size := m.Measure(runtime.BoxConstraints{})
			width = size.Width
			height = size.Height
		}

		// Fallback to estimation
		if width == 0 {
			width = l.estimateChildWidth(child)
		}
		if height == 0 {
			height = 1 // Default height
		}

		boundsAware.SetBounds(x, y, width, height)
	}
}

// estimateChildWidth estimates the width of a child for layout
func (l *LayoutNode) estimateChildWidth(child ui.VNode) int {
	if child == nil {
		return 0
	}

	// Check if child has explicit width
	if props := child.Props(); props != nil {
		if w := props.GetInt("width"); w > 0 {
			return w
		}
	}

	// Check if child has explicit height (for Input)
	if props := child.Props(); props != nil {
		if h := props.GetInt("height"); h > 0 {
			// For Input, use height as width hint if not specified
			return h
		}
	}

	// Estimate from content
	if props := child.Props(); props != nil {
		if content := props.GetString("content"); content != "" {
			// Simple width: length of content
			runes := []rune(content)
			return len(runes)
		}
	}

	// Check if child is Button/Input (has label)
	if labelGetter, ok := child.(interface{ Label() string }); ok {
		label := labelGetter.Label()
		if label != "" {
			return len(label) + 4 // Add space for brackets
		}
	}

	// Check if child is Input (has Value/Placeholder)
	if valueGetter, ok := child.(interface{ Value() string }); ok {
		value := valueGetter.Value()
		if value != "" {
			return len(value) + 2 // Add space for colons
		}
	}
	if placeholderGetter, ok := child.(interface{ Placeholder() string }); ok {
		placeholder := placeholderGetter.Placeholder()
		if placeholder != "" {
			return len(placeholder) + 2 // Add space for colons
		}
	}

	// Default minimum width
	return 10
}

// HStack creates a horizontal layout container
func HStack(children ...ui.VNode) ui.VNode {
	builder := &LayoutBuilder{
		node: &LayoutNode{
			ElementVNode: ui.NewElement("hstack"),
			direction:    DirectionRow,
			align:        AlignStart,
			crossAlign:   AlignStart,
			gap:          0,
			padding:      [4]int{0, 0, 0, 0},
		},
		children: children,
	}
	return builder.Build()
}

// VStack creates a vertical layout container
func VStack(children ...ui.VNode) ui.VNode {
	builder := &LayoutBuilder{
		node: &LayoutNode{
			ElementVNode: ui.NewElement("vstack"),
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
	children []ui.VNode
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

// Build returns the ui.VNode
func (b *LayoutBuilder) Build() ui.VNode {
	// Set children on the element
	b.node.SetChildren(b.children)
	return b.node
}

// =============================================================================
// Box Component
// =============================================================================

// Box creates a container box
func Box() *BoxLayoutBuilder {
	return &BoxLayoutBuilder{
		node: ui.NewElement("box"),
	}
}

// BoxLayoutBuilder provides fluent API for building boxes
type BoxLayoutBuilder struct {
	node *ui.ElementVNode
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
func (b *BoxLayoutBuilder) Child(child ui.VNode) *BoxLayoutBuilder {
	b.node.SetChildren([]ui.VNode{child})
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

// Build returns the ui.VNode
func (b *BoxLayoutBuilder) Build() ui.VNode {
	return b.node
}

// =============================================================================
// Spacer Component
// =============================================================================

// Spacer creates a flexible space
func Spacer() *SpacerBuilder {
	return &SpacerBuilder{
		node: ui.NewElement("spacer"),
	}
}

// SpacerBuilder provides fluent API for spacer
type SpacerBuilder struct {
	node *ui.ElementVNode
}

// Flex sets the flex factor
func (b *SpacerBuilder) Flex(n int) *SpacerBuilder {
	b.node.SetProp("flex", n)
	return b
}

// Build returns the ui.VNode
func (b *SpacerBuilder) Build() ui.VNode {
	return b.node
}
