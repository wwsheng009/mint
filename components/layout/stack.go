package layout

import (
	"os"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
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
	direction    Direction
	align        Align
	crossAlign   Align
	gap          int
	padding      [4]int // top, right, bottom, left
	stretchCross bool   // Stretch children to fill cross axis
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

// StretchCross returns whether children should stretch to fill cross axis
func (l *LayoutNode) StretchCross() bool {
	return l.stretchCross
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
	paddingWidth := l.padding[1] + l.padding[3]  // left + right
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

		// First pass: identify flex children and measure non-flex children
		type flexChild struct {
			child  ui.VNode
			index  int
			factor int
		}
		var flexChildren []flexChild
		fixedWidth := 0
		flexTotalFactor := 0

		for i, child := range l.Children() {
			childInfo := rtui.GetLayoutInfo(child)
			if childInfo.Flex > 0 {
				// Flex child: will be measured in second pass
				flexChildren = append(flexChildren, flexChild{
					child:  child,
					index:  i,
					factor: childInfo.Flex,
				})
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

		// Second pass: distribute space to flex children if we have bounded width
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
			// No bounded width or no flex children: measure flex children naturally
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

	// Respect explicit height prop if provided
	if props := l.Props(); props != nil {
		if explicitHeight := props.GetInt("height"); explicitHeight > 0 {
			totalHeight = explicitHeight
		}
	}

	// Apply constraints using the helper method
	totalWidth, totalHeight = constraints.Constrain(totalWidth, totalHeight)

	return runtime.Size{Width: totalWidth, Height: totalHeight}
}

// IsLayoutMeasurer marks LayoutNode as implementing the LayoutMeasurer interface
func (l *LayoutNode) IsLayoutMeasurer() {
}

// MeasureLayout implements runtime.LayoutMeasurer interface
// This provides single-pass measurement that preserves flex child constraints
func (l *LayoutNode) MeasureLayout(measurer runtime.ChildMeasurer, constraints runtime.BoxConstraints) runtime.LayoutMeasurement {
	children := l.Children()
	if len(children) == 0 {
		return runtime.NewLayoutMeasurement(
			runtime.Size{Width: constraints.MinWidth, Height: constraints.MinHeight},
			nil,
		)
	}

	// ⭐ IMPORTANT: Check if this node has an explicit width prop
	// If set, use it as the constraint MaxWidth instead of parent's constraint
	if props := l.Props(); props != nil {
		if explicitWidth := props.GetInt("width"); explicitWidth > 0 {
			// Override with explicit width
			constraints.MaxWidth = explicitWidth
			if os.Getenv("TUI_DEBUG_LAYOUT") == "true" {
				log.UILogger.Debug("[HStack.MeasureLayout] tag=%s, using explicit width=%d\n",
					l.Tag(), explicitWidth)
			}
		}
	}

	// Debug: log constraints received
	if os.Getenv("TUI_DEBUG_LAYOUT") == "true" {
		log.UILogger.Debug("[HStack.MeasureLayout] tag=%s, constraints.MinWidth=%d, MaxWidth=%d, gap=%d\n",
			l.Tag(), constraints.MinWidth, constraints.MaxWidth, l.gap)
	}

	// Add padding to content size
	paddingWidth := l.padding[1] + l.padding[3]  // left + right
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

	childConstraints := make([]runtime.BoxConstraints, len(children))
	var totalWidth, totalHeight int

	if l.direction == DirectionRow {
		// HStack: measure total width and max height
		maxChildHeight := 0

		// First pass: identify flex children and measure non-flex children
		type flexChild struct {
			index  int
			factor int
		}
		var flexChildren []flexChild
		fixedWidth := 0
		flexTotalFactor := 0

		for i, child := range children {
			childInfo := rtui.GetLayoutInfo(child)
			if childInfo.Flex > 0 {
				flexChildren = append(flexChildren, flexChild{
					index:  i,
					factor: childInfo.Flex,
				})
				flexTotalFactor += childInfo.Flex
			} else {
				// Non-flex child: measure with natural width
				cc := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  runtime.Infinity,
					MinHeight: 0,
					MaxHeight: innerMaxHeight,
				}
				childSize := measurer.MeasureChild(child, cc)
				childConstraints[i] = cc
				fixedWidth += childSize.Width
				if childSize.Height > maxChildHeight {
					maxChildHeight = childSize.Height
				}
			}
			// ⚠️ DON'T add gap to fixedWidth here!
			// Gaps are already subtracted from availableWidth in the second pass
			// Adding them here would cause remainingSpace to be calculated incorrectly
		}

		totalWidth = fixedWidth

		// Second pass: distribute space to flex children if we have bounded width
		if len(flexChildren) > 0 && constraints.HasBoundedWidth() {
			availableWidth := constraints.MaxWidth - paddingWidth - (len(children)-1)*l.gap
			remainingSpace := availableWidth - fixedWidth

			if os.Getenv("TUI_DEBUG_LAYOUT") == "true" {
				log.UILogger.Debug("[HStack] availableWidth=%d, fixedWidth=%d, remainingSpace=%d, flexChildren=%d\n",
					availableWidth, fixedWidth, remainingSpace, len(flexChildren))
			}

			// ✅ IMPROVED: Distribute remaining space with remainder handling
			// This ensures all available space is used (no wasted pixels)
			baseFlexWidth := remainingSpace / flexTotalFactor
			remainder := remainingSpace % flexTotalFactor

			if os.Getenv("TUI_DEBUG_LAYOUT") == "true" {
				log.UILogger.Debug("[HStack] baseFlexWidth=%d, remainder=%d\n", baseFlexWidth, remainder)
			}

			// Distribute to flex children
			for _, fc := range flexChildren {
				flexWidth := baseFlexWidth * fc.factor

				// Distribute remainder to first few children
				if remainder > 0 {
					extra := 1
					if remainder < fc.factor {
						extra = remainder
					}
					flexWidth += extra
					remainder -= extra
				}

				if flexWidth < 0 {
					flexWidth = 0
				}

				if os.Getenv("TUI_DEBUG_LAYOUT") == "true" {
					log.UILogger.Debug("[HStack]   child[%d]: flexWidth=%d\n", fc.index, flexWidth)
				}

				cc := runtime.BoxConstraints{
					MinWidth:  flexWidth,
					MaxWidth:  flexWidth,
					MinHeight: 0,
					MaxHeight: innerMaxHeight,
				}
				childConstraints[fc.index] = cc
				totalWidth += flexWidth

				// ✅ IMPORTANT: Measure flex children to get their heights!
				childSize := measurer.MeasureChild(children[fc.index], cc)
				if childSize.Height > maxChildHeight {
					maxChildHeight = childSize.Height
				}
			}
		} else {
			// No bounded width or no flex children: measure flex children naturally
			for _, fc := range flexChildren {
				cc := runtime.BoxConstraints{
					MinWidth:  0,
					MaxWidth:  runtime.Infinity,
					MinHeight: 0,
					MaxHeight: innerMaxHeight,
				}
				childConstraints[fc.index] = cc

				// ✅ IMPORTANT: Measure flex children to get their heights!
				childSize := measurer.MeasureChild(children[fc.index], cc)
				if childSize.Height > maxChildHeight {
					maxChildHeight = childSize.Height
				}
			}
		}

		totalHeight = maxChildHeight
	} else {
		// VStack: measure max width and total height
		maxChildWidth := 0
		for i := range children {
			// For VStack, children get width constraint but unlimited height
			cc := runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  innerMaxWidth,
				MinHeight: 0,
				MaxHeight: runtime.Infinity,
			}
			childConstraints[i] = cc
			// Note: We can't measure children here without access to their actual sizes
			// The Engine will use these constraints to measure children
			if cc.MaxWidth > maxChildWidth {
				maxChildWidth = cc.MaxWidth
			}
		}
		totalWidth = maxChildWidth
		// For VStack, we can't calculate totalHeight without measuring children
		// Use MaxHeight if bounded, otherwise use a reasonable default
		if constraints.HasBoundedHeight() {
			totalHeight = constraints.MaxHeight
		} else {
			totalHeight = runtime.Infinity
		}
	}

	// Add padding
	totalWidth += paddingWidth
	totalHeight += paddingHeight

	// Apply constraints
	totalWidth, totalHeight = constraints.Constrain(totalWidth, totalHeight)

	return runtime.LayoutMeasurement{
		Size:             runtime.Size{Width: totalWidth, Height: totalHeight},
		ChildConstraints: childConstraints,
	}
}

// measureChild measures a single child, returning its size
func (l *LayoutNode) measureChild(child ui.VNode, constraints runtime.BoxConstraints) runtime.Size {
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
		// ⚠️ IMPORTANT: In two-phase rendering, LayoutEngine.calculatePositions()
		// already set the correct bounds with flex widths. We should:
		// 1. Use the bounds x, y coordinates for Paint()
		// 2. Use the bounds width for position calculation
		// 3. NOT overwrite bounds with natural widths

		var paintX, paintY int
		var childWidth, childHeight int
		hasLayoutBounds := false

		// Check if child has layout bounds (set by LayoutEngine)
		if boundsAware, ok := child.(interface{ Bounds() [4]int }); ok {
			if bounds := boundsAware.Bounds(); bounds[2] > 0 {
				// Use layout engine's bounds
				paintX, paintY = bounds[0], bounds[1]
				childWidth, childHeight = bounds[2], bounds[3]
				hasLayoutBounds = true
			}
		}

		if !hasLayoutBounds {
			// Legacy path: use currentX/currentY
			paintX, paintY = currentX, currentY
			l.setChildBounds(child, currentX, currentY)
		}

		// Check if child implements Paintable
		if paintable, ok := child.(interface {
			Paint(int, int) []paint.DrawCmd
		}); ok {
			// Child has custom paint logic
			childCmds := paintable.Paint(paintX, paintY)
			cmds = append(cmds, childCmds...)
		} else {
			// Fallback: render as text element
			if props := child.Props(); props != nil {
				if content := props.GetString("content"); content != "" {
					cmds = append(cmds, paint.DrawCmd{
						X:     paintX,
						Y:     paintY,
						Text:  content,
						Style: child.Style(),
					})
				}
			}
		}

		// Update position based on direction
		if l.direction == DirectionRow {
			// HStack: move horizontally
			if hasLayoutBounds {
				// Use flex width from layout engine
				currentX = paintX + childWidth
			} else {
				// Legacy: estimate natural width
				childWidth = l.estimateChildWidth(child)
				currentX += childWidth
			}
			if i < len(children)-1 {
				currentX += l.gap
			}
		} else {
			// VStack: move vertically
			if hasLayoutBounds {
				currentY = paintY + childHeight
			} else {
				currentY++
			}
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

// Stretch makes all children stretch to fill the cross axis
// For VStack: children stretch to fill width
// For HStack: children stretch to fill height
func (b *LayoutBuilder) Stretch() *LayoutBuilder {
	b.node.stretchCross = true
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
