package ui

import (
	"strings"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/style"
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
// Measurable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
// Calculates the size of the layout based on children and constraints
func (l *LayoutNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if l == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	children := l.Children()
	if len(children) == 0 {
		// Empty layout takes minimum space specified by constraints
		return runtime.Size{
			Width:  constraints.MinWidth,
			Height: constraints.MinHeight,
		}
	}

	// Add padding to content size
	paddingWidth := l.padding[1] + l.padding[3] // left + right
	paddingHeight := l.padding[0] + l.padding[2] // top + bottom

	// Calculate inner constraints (subtract padding from available space)
	innerMaxWidth := constraints.MaxWidth - paddingWidth
	innerMaxHeight := constraints.MaxHeight - paddingHeight
	if innerMaxWidth < 0 {
		innerMaxWidth = 0
	}
	if innerMaxHeight < 0 {
		innerMaxHeight = 0
	}

	var totalWidth, totalHeight int

	if l.direction == DirectionRow {
		// HStack: measure total width and max height
		maxChildHeight := 0
		for i, child := range l.Children() {
			// For HStack, children get unlimited width (to measure natural size)
			// but height is constrained to the container's height
			childConstraints := runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  runtime.Infinity, // Let children expand to natural width
				MinHeight: 0,
				MaxHeight: innerMaxHeight,
			}
			childSize := l.measureChild(child, childConstraints)
			totalWidth += childSize.Width
			if childSize.Height > maxChildHeight {
				maxChildHeight = childSize.Height
			}
			// Add gap between children
			if i < len(children)-1 {
				totalWidth += l.gap
			}
		}
		totalHeight = maxChildHeight
	} else {
		// VStack: measure max width and total height
		maxChildWidth := 0
		for i, child := range l.Children() {
			// For VStack, children get width constraint but unlimited height
			// (to measure natural height)
			childConstraints := runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  innerMaxWidth,
				MinHeight: 0,
				MaxHeight: runtime.Infinity, // Let children expand to natural height
			}
			childSize := l.measureChild(child, childConstraints)
			if childSize.Width > maxChildWidth {
				maxChildWidth = childSize.Width
			}
			totalHeight += childSize.Height
			// Add gap between children
			if i < len(children)-1 {
				totalHeight += l.gap
			}
		}
		totalWidth = maxChildWidth
	}

	// Add padding
	totalWidth += paddingWidth
	totalHeight += paddingHeight

	// Apply constraints using the helper method
	totalWidth, totalHeight = constraints.Constrain(totalWidth, totalHeight)

	return runtime.Size{Width: totalWidth, Height: totalHeight}
}

// measureChild measures a single child, returning its size
func (l *LayoutNode) measureChild(child VNode, constraints runtime.BoxConstraints) runtime.Size {
	if child == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Check if child implements Measurable
	type measurable interface {
		Measure(constraints runtime.BoxConstraints) runtime.Size
	}
	if m, ok := child.(measurable); ok {
		return m.Measure(constraints)
	}

	// Fallback to estimation
	width := l.estimateChildWidth(child)
	height := 1 // Default height

	return runtime.Size{Width: width, Height: height}
}

// estimateChildWidth estimates the width of a child for layout
func (l *LayoutNode) estimateChildWidth(child VNode) int {
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

// =============================================================================
// Table Layout (Row-based layout for aligned columns)
// =============================================================================

// TableRow represents a row in a table layout
type TableRow struct {
	*ElementVNode
}

// TableCell represents a cell in a table row
type TableCell struct {
	*ElementVNode
}

// Table creates a table with row-based layout
// Each row is rendered independently, allowing different content in each column per row
func Table(rows ...VNode) VNode {
	node := NewElement("table")
	node.SetChildren(rows)
	return node
}

// Row creates a table row containing cells
func Row(cells ...VNode) VNode {
	row := &TableRow{
		ElementVNode: NewElement("tr"),
	}
	row.SetChildren(cells)
	return row
}

// Cell creates a table cell containing content
func Cell(content VNode) VNode {
	cell := &TableCell{
		ElementVNode: NewElement("td"),
	}
	if content != nil {
		cell.SetChildren([]VNode{content})
	}
	return cell
}

// GetTableCells returns the cells in a table row
func (tr *TableRow) GetTableCells() []VNode {
	return tr.Children()
}

// GetCellContent returns the content of a table cell
func (td *TableCell) GetCellContent() VNode {
	children := td.Children()
	if len(children) > 0 {
		return children[0]
	}
	return nil
}

// =============================================================================
// Bordered Container - Auto-rendered borders (don't occupy content space)
// =============================================================================

// BorderStyle defines the visual style of borders
type BorderStyle int

const (
	BorderSingle BorderStyle = iota // Single line: ┌───┐
	BorderDouble                      // Double line: ╔═══╗
	BorderRounded                     // Rounded: ╭───╮
	BorderDashed                      // Dashed: +---+
	BorderNone                        // No border
)

// BorderedNode represents a container with auto-rendered border
// The border is rendered outside the content area, not taking up content space
type BorderedNode struct {
	*ElementVNode
	borderStyle BorderStyle
	borderColor string
	borderLabel  string // Optional title shown on top border
}

// Bordered creates a container with auto-rendered border
// Usage:
//   ui.Bordered().Child(content).Build()
//   ui.Bordered().Style("double").Label("Title").Child(content).Build()
func Bordered() *BorderedBuilder {
	return &BorderedBuilder{
		node: &BorderedNode{
			ElementVNode: NewElement("bordered"),
			borderStyle:  BorderSingle,
			borderColor: "blue",
		},
	}
}

// BorderedBuilder builds a bordered container
type BorderedBuilder struct {
	node *BorderedNode
}

// Style sets the border style ("single", "double", "rounded", "dashed")
func (b *BorderedBuilder) Style(style string) *BorderedBuilder {
	switch style {
	case "double":
		b.node.borderStyle = BorderDouble
	case "rounded":
		b.node.borderStyle = BorderRounded
	case "dashed":
		b.node.borderStyle = BorderDashed
	case "none":
		b.node.borderStyle = BorderNone
	default:
		b.node.borderStyle = BorderSingle
	}
	return b
}

// Color sets the border color
func (b *BorderedBuilder) Color(c string) *BorderedBuilder {
	b.node.borderColor = c
	return b
}

// Label sets a title shown on the top border
func (b *BorderedBuilder) Label(label string) *BorderedBuilder {
	b.node.borderLabel = label
	return b
}

// Child sets the content inside the border
func (b *BorderedBuilder) Child(child VNode) *BorderedBuilder {
	b.node.SetChildren([]VNode{child})
	return b
}

// Width sets the content width (border adds 2 chars)
func (b *BorderedBuilder) Width(n int) *BorderedBuilder {
	b.node.SetProp("width", n)
	return b
}

// Height sets the content height (border adds 2 lines)
func (b *BorderedBuilder) Height(n int) *BorderedBuilder {
	b.node.SetProp("height", n)
	return b
}

// Build returns the VNode
func (b *BorderedBuilder) Build() VNode {
	return b.node
}

// GetBorderStyle returns the border style
func (bn *BorderedNode) GetBorderStyle() BorderStyle {
	return bn.borderStyle
}

// GetBorderColor returns the border color
func (bn *BorderedNode) GetBorderColor() string {
	return bn.borderColor
}

// GetBorderLabel returns the border label
func (bn *BorderedNode) GetBorderLabel() string {
	return bn.borderLabel
}

// RenderBorder returns the border VNodes for a bordered container
// This is called by the renderer to draw borders around content
func (bn *BorderedNode) RenderBorder(contentWidth, contentHeight int) []VNode {
	if bn.borderStyle == BorderNone {
		return nil
	}

	// Ensure minimum size
	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	result := make([]VNode, 0)

	// Get border characters
	cornerTL, cornerTR, cornerBL, cornerBR, horizontal, vertical := bn.GetBorderChars()

	// Top border (with optional label)
	topBorder := bn.renderTopBorder(cornerTL, cornerTR, horizontal, contentWidth)
	result = append(result, topBorder...)

	// Middle rows (left + right borders only, content painted between them)
	for i := 0; i < contentHeight; i++ {
		// Left border
		leftBorder := Element("text").Prop("content", string(vertical)).Style(style.Style{FG: style.Color(bn.borderColor)}).Build()
		result = append(result, leftBorder)

		// Right border (will be positioned at contentWidth + 1)
		rightBorder := Element("text").Prop("content", string(vertical)).Style(style.Style{FG: style.Color(bn.borderColor)}).Build()
		result = append(result, rightBorder)
	}

	// Bottom border
	result = append(result, bn.renderBottomBorder(cornerBL, cornerBR, horizontal, contentWidth)...)

	return result
}

// renderTopBorder renders the top border line with optional label
func (bn *BorderedNode) renderTopBorder(cornerTL, cornerTR, horizontal rune, contentWidth int) []VNode {
	color := bn.borderColor
	textStyle := style.Style{FG: style.Color(color)}

	if bn.borderLabel == "" {
		// Simple top border without label: cornerTL + horizontal fill + cornerTR
		// Return 3 separate VNodes for proper positioning
		horizontalFill := strings.Repeat(string(horizontal), contentWidth)
		if contentWidth > 0 {
			return []VNode{
				Element("text").Prop("content", string(cornerTL)).Style(textStyle).Build(),
				Element("text").Prop("content", horizontalFill).Style(textStyle).Build(),
				Element("text").Prop("content", string(cornerTR)).Style(textStyle).Build(),
			}
		}
		// Empty content - just corners
		return []VNode{
			Element("text").Prop("content", string(cornerTL)+string(horizontal)).Style(textStyle).Build(),
			Element("text").Prop("content", string(horizontal)+string(cornerTR)).Style(textStyle).Build(),
		}
	}

	// Top border with label: "+-- Label ----+"
	label := bn.borderLabel
	labelWidth := len(label) + 2 // +1 for space on each side
	padding := (contentWidth - labelWidth) / 2
	if padding < 0 {
		padding = 0
	}

	return []VNode{
		Element("text").Prop("content", string(cornerTL)+strings.Repeat(string(horizontal), padding+1)).Style(textStyle).Build(),
		Element("text").Prop("content", " "+label+" ").Style(textStyle.Bold(true)).Build(),
		Element("text").Prop("content", strings.Repeat(string(horizontal), contentWidth-padding-labelWidth+2)+string(cornerTR)).Style(textStyle).Build(),
	}
}

// renderBottomBorder renders the bottom border line
func (bn *BorderedNode) renderBottomBorder(cornerBL, cornerBR, horizontal rune, contentWidth int) []VNode {
	color := bn.borderColor
	textStyle := style.Style{FG: style.Color(color)}

	// Bottom border: cornerBL + horizontal fill + cornerBR
	horizontalFill := strings.Repeat(string(horizontal), contentWidth)
	if contentWidth > 0 {
		return []VNode{
			Element("text").Prop("content", string(cornerBL)).Style(textStyle).Build(),
			Element("text").Prop("content", horizontalFill).Style(textStyle).Build(),
			Element("text").Prop("content", string(cornerBR)).Style(textStyle).Build(),
		}
	}
	// Empty content - just corners
	return []VNode{
		Element("text").Prop("content", string(cornerBL)+string(horizontal)).Style(textStyle).Build(),
		Element("text").Prop("content", string(horizontal)+string(cornerBR)).Style(textStyle).Build(),
	}
}

// GetBorderChars returns the border characters for the current style
func (bn *BorderedNode) GetBorderChars() (cornerTL, cornerTR, cornerBL, cornerBR, horizontal, vertical rune) {
	switch bn.borderStyle {
	case BorderDouble:
		return '╔', '╗', '╚', '╝', '═', '║'
	case BorderRounded:
		return '╭', '╮', '╰', '╯', '─', '│'
	case BorderDashed:
		return '+', '+', '+', '+', '-', '|'
	default: // BorderSingle - continuous line style
		return '┌', '┐', '└', '┘', '─', '│'
	}
}

// =============================================================================
// Measurable Interface Implementation
// =============================================================================

// Measure calculates the size of the bordered container
// The border adds 2 characters to width and height (1 on each side)
// If a label is present, the width expands to accommodate it
func (bn *BorderedNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if bn == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	borderWidth := 0
	borderHeight := 0
	if bn.borderStyle != BorderNone {
		borderWidth = 2  // 1 char on left, 1 char on right
		borderHeight = 2 // 1 char on top, 1 char on bottom
	}

	// Calculate label width if present
	labelWidth := 0
	if bn.borderLabel != "" && bn.borderStyle != BorderNone {
		labelWidth = len(bn.borderLabel) + 2 // +2 for spaces around label
	}

	// Measure child content
	var contentWidth, contentHeight int
	children := bn.Children()
	if len(children) > 0 {
		child := children[0]
		if measurable, ok := child.(interface {
			Measure(runtime.BoxConstraints) runtime.Size
		}); ok {
			// Child implements Measurable - measure with inner constraints
			innerConstraints := runtime.BoxConstraints{
				MinWidth:  max(0, constraints.MinWidth-borderWidth),
				MaxWidth:  max(0, constraints.MaxWidth-borderWidth),
				MinHeight: max(0, constraints.MinHeight-borderHeight),
				MaxHeight: max(0, constraints.MaxHeight-borderHeight),
			}
			contentSize := measurable.Measure(innerConstraints)
			contentWidth = contentSize.Width
			contentHeight = contentSize.Height
		} else {
			// Fallback: estimate child size
			contentWidth = 10  // Default minimum
			contentHeight = 1
		}
	} else {
		// No child - minimum size
		contentWidth = 1
		contentHeight = 1
	}

	// Inner width is the larger of content and label
	innerWidth := contentWidth
	if labelWidth > innerWidth {
		innerWidth = labelWidth
	}

	// Total size = content + border
	// When label is present, renderTopBorder adds extra 2 chars for visual balance
	// This matches the actual rendering logic in renderTopBorder
	totalWidth := innerWidth + borderWidth
	if labelWidth > 0 {
		totalWidth += 2  // Extra padding for label rendering (see renderTopBorder)
	}
	totalHeight := contentHeight + borderHeight

	// Apply constraints
	if totalWidth < constraints.MinWidth {
		totalWidth = constraints.MinWidth
	}
	if totalWidth > constraints.MaxWidth && constraints.MaxWidth > 0 {
		totalWidth = constraints.MaxWidth
	}
	if totalHeight < constraints.MinHeight {
		totalHeight = constraints.MinHeight
	}
	if totalHeight > constraints.MaxHeight && constraints.MaxHeight > 0 {
		totalHeight = constraints.MaxHeight
	}

	// Apply explicit style dimensions if set
	elemStyle := bn.Style()
	if elemStyle.Width > 0 {
		totalWidth = elemStyle.Width
	}
	if elemStyle.Height > 0 {
		totalHeight = elemStyle.Height
	}

	return runtime.Size{Width: totalWidth, Height: totalHeight}
}

