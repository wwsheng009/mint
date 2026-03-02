package ui

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/cmd"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"

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
	direction    Direction
	align        Align
	crossAlign   Align
	gap          int
	flex         int    // Flex factor (0 = fixed size, >0 = grows to fill space)
	padding      [4]int // top, right, bottom, left
	stretchCross bool   // Stretch children to fill cross axis (width for VStack, height for HStack)
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

// HStackBuilder creates a horizontal layout container builder for method chaining
func HStackBuilder(children ...VNode) *LayoutBuilder {
	return &LayoutBuilder{
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

// VStackBuilder creates a vertical layout container builder for method chaining
func VStackBuilder(children ...VNode) *LayoutBuilder {
	return &LayoutBuilder{
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
	b.node.flex = n // Also set the field for LayoutNode.Flex()
	return b
}

// Stretch makes all children stretch to fill the cross axis
// For VStack: children stretch to fill width
// For HStack: children stretch to fill height
func (b *LayoutBuilder) Stretch() *LayoutBuilder {
	b.node.stretchCross = true
	return b
}

// FillWidth makes this component stretch to fill the parent's width
// This is typically used for children in a VStack (where cross-axis is horizontal)
// Unlike Stretch() which affects all children, FillWidth() applies to a single component
func (b *LayoutBuilder) FillWidth() *LayoutBuilder {
	b.node.SetProp("fillWidth", true)
	return b
}

// FillHeight makes this component stretch to fill the parent's height
// This is typically used for children in an HStack (where cross-axis is vertical)
// Unlike Stretch() which affects all children, FillHeight() applies to a single component
func (b *LayoutBuilder) FillHeight() *LayoutBuilder {
	b.node.SetProp("fillHeight", true)
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

	// Sync layout properties to Props so CreateFiber can read them
	// This fixes the component interference issue where layout config was lost
	props := b.node.Props()
	if props == nil {
		props = make(Props)
	}
	props["direction"] = b.node.direction
	props["align"] = b.node.align
	props["crossAlign"] = b.node.crossAlign
	props["gap"] = b.node.gap
	props["padding"] = b.node.padding
	props["flex"] = b.node.flex
	props["stretchCross"] = b.node.stretchCross
	b.node.SetProps(props)

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

// Flex returns the flex factor
func (l *LayoutNode) Flex() int {
	return l.flex
}

// Padding returns the padding
func (l *LayoutNode) Padding() [4]int {
	return l.padding
}

// StretchCross returns whether children stretch to fill cross axis
func (l *LayoutNode) StretchCross() bool {
	return l.stretchCross
}

// GetFlexStyle returns the flex layout style for the node
// This implements the layout.FlexStyleProvider interface, enabling Engine
// to use FlexLayout for LayoutNode instead of the default vertical layout
func (l *LayoutNode) GetFlexStyle() *layout.FlexStyle {
	if l == nil {
		return nil
	}

	flexStyle := &layout.FlexStyle{}

	// Set direction
	switch l.direction {
	case DirectionRow:
		flexStyle.Direction = layout.FlexRow
	case DirectionColumn:
		flexStyle.Direction = layout.FlexColumn
	default:
		flexStyle.Direction = layout.FlexColumn
	}

	// Set main axis alignment
	switch l.align {
	case AlignStart:
		flexStyle.MainAxis = layout.MainStart
	case AlignCenter:
		flexStyle.MainAxis = layout.Center
	case AlignEnd:
		flexStyle.MainAxis = layout.MainEnd
	default:
		flexStyle.MainAxis = layout.MainStart
	}

	// Set cross axis alignment
	switch l.crossAlign {
	case AlignStart:
		flexStyle.CrossAxis = layout.CrossStart
	case AlignCenter:
		flexStyle.CrossAxis = layout.CrossCenter
	case AlignEnd:
		flexStyle.CrossAxis = layout.CrossEnd
	default:
		flexStyle.CrossAxis = layout.CrossStart
	}

	// Set gap
	flexStyle.Gap = l.gap
	flexStyle.CrossGap = l.gap // LayoutNode uses the same gap for both axes

	// Set padding
	flexStyle.Padding = layout.Padding{
		Left:   l.padding[3],
		Right:  l.padding[1],
		Top:    l.padding[0],
		Bottom: l.padding[2],
	}

	// Flexible children are handled through FlexChildProvider on child nodes
	flexStyle.FlexibleChildren = make(map[int]*layout.Flex)

	return flexStyle
}

// GetBoxModel returns the box model for the LayoutNode
// Implements layout.BoxModelProvider interface for unified padding/border handling
func (l *LayoutNode) GetBoxModel() layout.BoxModel {
	boxModel := layout.BoxModel{}

	// Padding from layout node
	boxModel.Padding = layout.Padding{
		Left:   l.padding[3],
		Right:  l.padding[1],
		Top:    l.padding[0],
		Bottom: l.padding[2],
	}

	// Read border from props (if set)
	props := l.Props()
	if props != nil {
		// Read "border" prop (boolean)
		if hasBorder, ok := props["border"].(bool); ok && hasBorder {
			// Read border style
			borderStyle := layout.BorderSingle // default
			if style, ok := props["borderStyle"].(string); ok {
				switch style {
				case "double":
					borderStyle = layout.BorderDouble
				case "rounded":
					borderStyle = layout.BorderRounded
				case "dashed":
					borderStyle = layout.BorderDashed
				case "none":
					borderStyle = layout.BorderNone
				default:
					borderStyle = layout.BorderSingle
				}
			}
			boxModel.Border = layout.NewBorder(borderStyle)
		}
	}

	// Note: Margin is not currently supported on LayoutNode
	// If needed, it can be read from props in the future

	return boxModel
}

// =============================================================================
// Flex Wrapper - elegant API for making VNodes flexible
// =============================================================================

// Flex wraps a VNode to make it flexible in a layout
// Usage: ui.Flex(vnode, 1) or just ui.Flex(vnode)
func Flex(vnode VNode, flexFactors ...int) VNode {
	flex := 1 // Default flex factor
	if len(flexFactors) > 0 {
		flex = flexFactors[0]
	}
	// Try SetProp method first (for builders)
	if n, ok := vnode.(interface{ SetProp(string, interface{}) }); ok {
		n.SetProp("flex", flex)
		return vnode
	}
	// Fall back to SetProps (for VNode interface)
	props := vnode.Props()
	if props == nil {
		props = make(Props)
	}
	props["flex"] = flex
	vnode.SetProps(props)
	return vnode
}

// =============================================================================
// Measurable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
// Calculates the NATURAL size of the layout based on children only.
// This phase should NOT apply StretchCross - stretching is handled in the Layout phase.
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
	paddingWidth := l.padding[1] + l.padding[3]  // left + right
	paddingHeight := l.padding[0] + l.padding[2] // top + bottom

	var totalWidth, totalHeight int

	if l.direction == DirectionRow {
		// HStack: measure total width and max height
		// Children get unlimited width (to measure natural width)
		// Height is constrained by parent's MaxHeight (if bounded)
		maxChildHeight := 0

		// Calculate inner height constraint
		// Use parent's MaxHeight only if it's bounded, otherwise use Infinity
		innerMaxHeight := runtime.Infinity
		if constraints.HasBoundedHeight() {
			innerMaxHeight = max(0, constraints.MaxHeight-paddingHeight)
		}

		// First pass: identify flex children and measure non-flex children
		var flexChildren []struct {
			child  VNode
			factor int
		}
		var fixedWidth int
		flexTotalFactor := 0

		for i, child := range l.Children() {
			childInfo := GetLayoutInfo(child)
			if childInfo.Flex > 0 {
				flexChildren = append(flexChildren, struct {
					child  VNode
					factor int
				}{child, childInfo.Flex})
				flexTotalFactor += childInfo.Flex
			} else {
				// Non-flex child: measure with natural width
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  runtime.Infinity,
					MinHeight: 0,
					MaxHeight: innerMaxHeight,
				}
				childSize := l.measureChild(child, childConstraints)
				fixedWidth += childSize.Width
				if childSize.Height > maxChildHeight {
					maxChildHeight = childSize.Height
				}
			}
			// Account for gap (except after last child)
			if i < len(children)-1 {
				fixedWidth += l.gap
			}
		}

		totalWidth = fixedWidth

		// If we have flex children and bounded width, distribute remaining space
		if len(flexChildren) > 0 && constraints.HasBoundedWidth() {
			availableWidth := constraints.MaxWidth - paddingWidth - (len(children)-1)*l.gap
			remainingSpace := availableWidth - fixedWidth

			// Distribute remaining space to flex children
			for _, fc := range flexChildren {
				flexWidth := (remainingSpace * fc.factor) / flexTotalFactor
				if flexWidth < 0 {
					flexWidth = 0
				}

				childConstraints := runtime.BoxConstraints{
					MinWidth:  flexWidth,
					MaxWidth:  flexWidth,
					MinHeight: 0,
					MaxHeight: innerMaxHeight,
				}
				childSize := l.measureChild(fc.child, childConstraints)
				totalWidth += childSize.Width
				if childSize.Height > maxChildHeight {
					maxChildHeight = childSize.Height
				}
			}
		} else {
			// No flex or unbounded width: measure flex children naturally
			for _, fc := range flexChildren {
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  runtime.Infinity,
					MinHeight: 0,
					MaxHeight: innerMaxHeight,
				}
				childSize := l.measureChild(fc.child, childConstraints)
				totalWidth += childSize.Width
				if childSize.Height > maxChildHeight {
					maxChildHeight = childSize.Height
				}
			}
		}

		totalHeight = maxChildHeight
	} else {
		// VStack: measure max width and total height
		// Width is constrained by parent's MaxWidth (if bounded)
		// Children get unlimited height (to measure natural height)
		maxChildWidth := 0

		// Calculate inner width constraint
		// Use parent's MaxWidth only if it's bounded, otherwise use Infinity
		innerMaxWidth := runtime.Infinity
		if constraints.HasBoundedWidth() {
			innerMaxWidth = max(0, constraints.MaxWidth-paddingWidth)
		}

		// First pass: identify flex children and measure non-flex children
		var flexChildren []struct {
			child  VNode
			factor int
		}
		var fixedHeight int
		flexTotalFactor := 0

		for i, child := range l.Children() {
			childInfo := GetLayoutInfo(child)
			if childInfo.Flex > 0 {
				flexChildren = append(flexChildren, struct {
					child  VNode
					factor int
				}{child, childInfo.Flex})
				flexTotalFactor += childInfo.Flex
			} else {
				// Non-flex child: measure with natural height
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  innerMaxWidth,
					MinHeight: 0,
					MaxHeight: runtime.Infinity,
				}
				childSize := l.measureChild(child, childConstraints)
				if childSize.Width > maxChildWidth {
					maxChildWidth = childSize.Width
				}
				if childSize.Height < runtime.Infinity {
					fixedHeight += childSize.Height
				}
			}
			// Account for gap (except after last child)
			if i < len(children)-1 {
				fixedHeight += l.gap
			}
		}

		totalHeight = fixedHeight

		// If we have flex children and bounded height, distribute remaining space
		if len(flexChildren) > 0 && constraints.HasBoundedHeight() {
			availableHeight := constraints.MaxHeight - paddingHeight - (len(children)-1)*l.gap
			remainingSpace := availableHeight - fixedHeight

			// Distribute remaining space to flex children
			for _, fc := range flexChildren {
				flexHeight := (remainingSpace * fc.factor) / flexTotalFactor
				if flexHeight < 0 {
					flexHeight = 0
				}

				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  innerMaxWidth,
					MinHeight: flexHeight,
					MaxHeight: flexHeight,
				}
				childSize := l.measureChild(fc.child, childConstraints)
				if childSize.Width > maxChildWidth {
					maxChildWidth = childSize.Width
				}
				totalHeight += childSize.Height
			}
		} else {
			// No flex or unbounded height: measure flex children naturally
			for _, fc := range flexChildren {
				childConstraints := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  innerMaxWidth,
					MinHeight: 0,
					MaxHeight: runtime.Infinity,
				}
				childSize := l.measureChild(fc.child, childConstraints)
				if childSize.Width > maxChildWidth {
					maxChildWidth = childSize.Width
				}
				if childSize.Height < runtime.Infinity {
					totalHeight += childSize.Height
				}
			}
		}

		totalWidth = maxChildWidth
	}

	// Add padding
	totalWidth += paddingWidth
	totalHeight += paddingHeight

	// Respect explicit height prop when provided
	if props := l.Props(); props != nil {
		if explicitHeight := props.GetInt("height"); explicitHeight > 0 {
			totalHeight = explicitHeight
		}
	}

	// ⭐ IMPORTANT: Main-axis filling for HStack
	// When HStack has bounded width and is smaller than the bound, it should expand to fill
	// This ensures flex children properly distribute space in tight constraints
	if l.direction == DirectionRow && constraints.HasBoundedWidth() && totalWidth < constraints.MaxWidth {
		totalWidth = constraints.MaxWidth
	}

	// Apply MinWidth/MinHeight constraints
	if totalWidth < constraints.MinWidth {
		totalWidth = constraints.MinWidth
	}
	if totalHeight < constraints.MinHeight {
		totalHeight = constraints.MinHeight
	}

	// IMPORTANT: Cross-axis filling
	// - VStack: fill available width (MaxWidth) so children can stretch horizontally
	// - HStack: fill available height (MaxHeight) so children can stretch vertically
	// This is the NATURAL SIZE for layout containers - they expand to fill cross-axis space.
	if l.direction == DirectionColumn { // VStack
		// Fill available width so children can stretch
		if constraints.HasBoundedWidth() && totalWidth < constraints.MaxWidth {
			totalWidth = constraints.MaxWidth
		}
	} else { // HStack
		// Fill available height so children can stretch
		if constraints.HasBoundedHeight() && totalHeight < constraints.MaxHeight {
			totalHeight = constraints.MaxHeight
		}
	}

	// Clamp to MaxWidth/MaxHeight if exceeded
	if constraints.HasBoundedWidth() && totalWidth > constraints.MaxWidth {
		totalWidth = constraints.MaxWidth
	}
	if constraints.HasBoundedHeight() && totalHeight > constraints.MaxHeight {
		totalHeight = constraints.MaxHeight
	}

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

// FillWidth makes the box stretch to fill parent's width
func (b *BoxLayoutBuilder) FillWidth() *BoxLayoutBuilder {
	b.node.SetProp("fillWidth", true)
	return b
}

// FillHeight makes the box stretch to fill parent's height
func (b *BoxLayoutBuilder) FillHeight() *BoxLayoutBuilder {
	b.node.SetProp("fillHeight", true)
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

// BorderStyle is an alias to layout.BorderStyle for unified type definition.
// See runtime/layout/border.go for the canonical definition.
type BorderStyle = layout.BorderStyle

// BorderStyle constants - aliases to layout package
const (
	BorderSingle  = layout.BorderSingle  // Single line: ┌───┐
	BorderDouble  = layout.BorderDouble  // Double line: ╔═══╗
	BorderRounded = layout.BorderRounded // Rounded: ╭───╮
	BorderDashed  = layout.BorderDashed  // Dashed: +---+
	BorderNone    = layout.BorderNone    // No border
)

// BorderedNode represents a container with auto-rendered border
// The border is rendered outside the content area, not taking up content space
type BorderedNode struct {
	*ElementVNode
	borderStyle BorderStyle
	borderColor string
	borderLabel string // Optional title shown on top border
}

// Bordered creates a container with auto-rendered border
// Usage:
//
//	ui.Bordered().Child(content).Build()
//	ui.Bordered().Style("double").Label("Title").Child(content).Build()
func Bordered() *BorderedBuilder {
	return &BorderedBuilder{
		node: &BorderedNode{
			ElementVNode: NewElement("bordered"),
			borderStyle:  BorderSingle,
			borderColor:  "blue",
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

// Flex sets the flex factor
func (b *BorderedBuilder) Flex(n int) *BorderedBuilder {
	b.node.SetProp("flex", n)
	return b
}

// FillWidth makes the bordered container stretch to fill parent's width
func (b *BorderedBuilder) FillWidth() *BorderedBuilder {
	b.node.SetProp("fillWidth", true)
	return b
}

// FillHeight makes the bordered container stretch to fill parent's height
func (b *BorderedBuilder) FillHeight() *BorderedBuilder {
	b.node.SetProp("fillHeight", true)
	return b
}

// Build returns the VNode
func (b *BorderedBuilder) Build() VNode {
	// Sync border properties to Props so CreateFiber can read them
	props := b.node.Props()
	if props == nil {
		props = make(Props)
	}
	// Map borderStyle string value (e.g., "single", "double")
	switch b.node.borderStyle {
	case BorderDouble:
		props["borderStyle"] = "double"
	case BorderRounded:
		props["borderStyle"] = "rounded"
	case BorderDashed:
		props["borderStyle"] = "dashed"
	case BorderNone:
		props["borderStyle"] = "none"
	default:
		props["borderStyle"] = "single"
	}
	if b.node.borderLabel != "" {
		props["label"] = b.node.borderLabel
		props["borderLabel"] = b.node.borderLabel
	}
	// Sync border color to Props so CreateFiber can read it
	if b.node.borderColor != "" {
		props["borderColor"] = b.node.borderColor
	}
	b.node.SetProps(props)

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

// Update returns nil as BorderedNode doesn't handle messages
func (bn *BorderedNode) Update(message runtimemsg.Msg) cmd.Cmd {
	return nil
}

// SetBounds sets the bounds on BorderedNode
func (bn *BorderedNode) SetBounds(x, y, w, h int) {
	if bn.ElementVNode != nil {
		bn.ElementVNode.SetBounds(x, y, w, h)
	}
}

// SetLayer sets the rendering layer for this BorderedNode
// This overrides the embedded ElementVNode.SetLayer to preserve the BorderedNode type
func (bn *BorderedNode) SetLayer(l Layer) VNode {
	if bn == nil {
		return nil
	}
	// Set the layer in props
	if bn.Props() == nil {
		bn.SetProps(make(Props))
	}
	bn.Props().Set("_layer", l)
	return bn // Return bn (BorderedNode) instead of the generic ElementVNode
}

// GetLayer returns the rendering layer for this BorderedNode
func (bn *BorderedNode) GetLayer() Layer {
	if bn == nil {
		return LayerBase
	}
	props := bn.Props()
	if props == nil {
		return LayerBase
	}
	if layer, ok := props["_layer"].(Layer); ok {
		return layer
	}
	return LayerBase
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

	// Total width = border (2 chars) + contentWidth + padding
	// The label should be centered with equal horizontal lines on both sides
	// Total available space for horizontal lines = contentWidth
	horizontalLineTotal := contentWidth - labelWidth
	if horizontalLineTotal < 0 {
		horizontalLineTotal = 0
	}

	// Split evenly: left line gets half (rounded down), right line gets the rest
	leftLineLen := horizontalLineTotal / 2
	rightLineLen := horizontalLineTotal - leftLineLen

	// Build border parts
	var leftLine, rightLine string
	if leftLineLen > 0 {
		leftLine = strings.Repeat(string(horizontal), leftLineLen)
	}
	if rightLineLen > 0 {
		rightLine = strings.Repeat(string(horizontal), rightLineLen)
	}

	return []VNode{
		Element("text").Prop("content", string(cornerTL)+leftLine).Style(textStyle).Build(),
		Element("text").Prop("content", " "+label+" ").Style(textStyle.Bold(true)).Build(),
		Element("text").Prop("content", rightLine+string(cornerTR)).Style(textStyle).Build(),
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
	// Check if this Bordered node has flex, fillWidth, or explicit width/height
	hasFlex := false
	hasFillWidth := false
	props := bn.Props()
	if props != nil {
		if f, ok := props["flex"].(int); ok && f > 0 {
			hasFlex = true
		}
		if fillWidth, ok := props["fillWidth"].(bool); ok && fillWidth {
			hasFillWidth = true
		}
		// Check for explicit width/height props and apply as constraints
		if width, ok := props["width"].(int); ok && width > 0 {
			constraints.MaxWidth = width
			if constraints.MinWidth > width {
				constraints.MinWidth = width
			}
		}
		if height, ok := props["height"].(int); ok && height > 0 {
			constraints.MaxHeight = height
			if constraints.MinHeight > height {
				constraints.MinHeight = height
			}
		}
	}

	if len(children) > 0 {
		child := children[0]
		if measurable, ok := child.(interface {
			Measure(runtime.BoxConstraints) runtime.Size
		}); ok {
			// Child implements Measurable - measure with inner constraints
			// Use SubtractPadding helper to properly handle bounded/unbounded constraints
			innerConstraints := constraints.SubtractPadding(borderWidth, borderHeight)
			contentSize := measurable.Measure(innerConstraints)
			contentWidth = contentSize.Width
			contentHeight = contentSize.Height

			// If this Bordered node has flex or fillWidth, expand child to fill available space
			if hasFlex || hasFillWidth {
				if innerConstraints.HasBoundedWidth() && contentWidth < innerConstraints.MaxWidth {
					contentWidth = innerConstraints.MaxWidth
				}
				if hasFlex && innerConstraints.HasBoundedHeight() && contentHeight < innerConstraints.MaxHeight {
					contentHeight = innerConstraints.MaxHeight
				}
			}
		} else {
			// Fallback: estimate child size
			contentWidth = 10 // Default minimum
			contentHeight = 1
			// If fillWidth, expand to fill available space
			if hasFillWidth && constraints.HasBoundedWidth() {
				contentWidth = max(0, constraints.MaxWidth-borderWidth)
			}
		}
	} else {
		// No child - minimum size
		contentWidth = 1
		contentHeight = 1
		// If flex or fillWidth, expand to fill available space
		if hasFlex || hasFillWidth {
			if constraints.HasBoundedWidth() {
				contentWidth = max(0, constraints.MaxWidth-borderWidth)
			}
			if hasFlex && constraints.HasBoundedHeight() {
				contentHeight = max(0, constraints.MaxHeight-borderHeight)
			}
		}
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
		totalWidth += 2 // Extra padding for label rendering (see renderTopBorder)
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
