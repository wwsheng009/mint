package layout

import (
	"os"
	"unicode/utf8"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/ui"
)

// WrapNode represents a wrapping layout container
// Similar to CSS flex-wrap: wrap
type WrapNode struct {
	*ui.ElementVNode

	// Configuration
	gap         int      // Spacing between items in the same row
	rowGap      int      // Spacing between rows (0 = use gap)
	align       ui.Align // Main-axis alignment for each row
	screenWidth int      // Container width for wrapping calculation

	// Internal state (calculated during Build)
	rows [][]ui.VNode // Pre-calculated rows (for debugging)
}

// WrapBuilder provides fluent API for building wrap layouts
type WrapBuilder struct {
	node       *WrapNode
	children   []ui.VNode
	widthCache map[ui.VNode]int // Cache for estimated widths
}

// Wrap creates a wrapping layout container
func Wrap(children ...ui.VNode) ui.VNode {
	builder := NewWrapBuilder(children...)
	return builder.Build()
}

// NewWrapBuilder creates a wrap layout builder
func NewWrapBuilder(children ...ui.VNode) *WrapBuilder {
	return &WrapBuilder{
		node: &WrapNode{
			ElementVNode: ui.NewElement("wrap"),
			gap:          1,             // Default gap
			rowGap:       0,             // Default row gap (use gap)
			align:        ui.AlignStart, // Default alignment
			screenWidth:  80,            // Default width
		},
		children:   children,
		widthCache: make(map[ui.VNode]int),
	}
}

// Gap sets spacing between items in the same row
func (b *WrapBuilder) Gap(n int) *WrapBuilder {
	b.node.gap = n
	return b
}

// RowGap sets spacing between rows (0 = use gap value)
func (b *WrapBuilder) RowGap(n int) *WrapBuilder {
	b.node.rowGap = n
	return b
}

// Align sets main-axis alignment for each row
func (b *WrapBuilder) Align(a ui.Align) *WrapBuilder {
	b.node.align = a
	return b
}

// ScreenWidth sets container width for wrap calculation
// This determines when to break to a new row
func (b *WrapBuilder) ScreenWidth(width int) *WrapBuilder {
	b.node.screenWidth = width
	return b
}

// Width sets explicit width (alias for ScreenWidth)
func (b *WrapBuilder) Width(n int) *WrapBuilder {
	return b.ScreenWidth(n)
}

// Style sets visual style
func (b *WrapBuilder) Style(s style.Style) *WrapBuilder {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *WrapBuilder) FgColor(c interface{}) *WrapBuilder {
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
func (b *WrapBuilder) BgColor(c interface{}) *WrapBuilder {
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

// Key sets key for diffing
func (b *WrapBuilder) Key(key string) *WrapBuilder {
	b.node.SetKey(key)
	return b
}

// FillWidth makes each row stretch to fill the container width
// This is useful for control panels where buttons should distribute evenly
func (b *WrapBuilder) FillWidth() *WrapBuilder {
	b.node.SetProp("fillWidth", true)
	return b
}

// FillHeight makes the wrap container stretch to fill parent's height
func (b *WrapBuilder) FillHeight() *WrapBuilder {
	b.node.SetProp("fillHeight", true)
	return b
}

// estimateWidth estimates the display width of a VNode
// Priority order:
// 1. Explicit width prop
// 2. Measure interface (if available)
// 3. Component-specific logic (Button, Text, Input, etc.)
// 4. Default fallback
func (b *WrapBuilder) estimateWidth(child ui.VNode) int {
	if child == nil {
		return 0
	}

	// Check cache first
	if w, ok := b.widthCache[child]; ok {
		return w
	}

	var width int

	// PRIORITY 1: Check explicit width prop
	if props := child.Props(); props != nil {
		if w := props.GetInt("width"); w > 0 {
			width = w
			goto cache
		}
	}

	// PRIORITY 2: Use Measure interface if available
	if measurable, ok := child.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	}); ok {
		// Measure with unbounded constraints
		size := measurable.Measure(runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  runtime.Infinity,
			MinHeight: 0,
			MaxHeight: runtime.Infinity,
		})
		width = size.Width
		goto cache
	}

	// PRIORITY 3: Component-specific logic

	// Button: label + brackets and focus indicator
	if btn, ok := child.(interface{ Label() string }); ok {
		label := btn.Label()
		if label == "" {
			label = " "
		}
		width = utf8.RuneCountInString(label) + 4 // "[label]" + focus
		goto cache
	}

	// Text: content length
	if text := ui.GetTextContent(child); text != "" {
		width = utf8.RuneCountInString(text)
		goto cache
	}

	// Input: value or placeholder width
	if input, ok := child.(interface {
		Value() string
		Placeholder() string
	}); ok {
		value := input.Value()
		if value == "" {
			value = input.Placeholder()
		}
		if value != "" {
			width = utf8.RuneCountInString(value) + 2 // ":value:"
		}
		goto cache
	}

	// PRIORITY 4: Default fallback
	width = 10 // Minimum reasonable width

cache:
	// Cache result
	b.widthCache[child] = width
	return width
}

// calculateRows divides children into rows based on screenWidth
func (b *WrapBuilder) calculateRows() [][]ui.VNode {
	var rows [][]ui.VNode
	currentRow := []ui.VNode{}
	currentWidth := 0
	screenWidth := b.node.screenWidth

	for i, child := range b.children {
		childWidth := b.estimateWidth(child)

		// Check if we need to wrap
		// Wrap if:
		// 1. Current row is not empty
		// 2. Adding this child would exceed screen width (including gap)
		shouldWrap := len(currentRow) > 0 &&
			(currentWidth+childWidth+b.node.gap > screenWidth)

		if shouldWrap {
			// Finish current row
			rows = append(rows, currentRow)
			// Start new row with this child
			currentRow = []ui.VNode{child}
			currentWidth = childWidth
		} else {
			// Add to current row
			currentRow = append(currentRow, child)
			if len(currentRow) > 1 {
				// Add gap for all but first item
				currentWidth += b.node.gap
			}
			currentWidth += childWidth
		}

		// Debug logging (optional)
		if os.Getenv("TUI_WRAP_DEBUG") == "true" {
			log.UILogger.Debug("[Wrap] child %d: width=%d, currentWidth=%d, shouldWrap=%v\n",
				i, childWidth, currentWidth, shouldWrap)
		}
	}

	// Don't forget the last row
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return rows
}

// createRowHStack creates an HStack for a row with proper configuration
func createRowHStack(row []ui.VNode, align ui.Align, gap int, stretchCross bool, width int) ui.VNode {
	hstackBuilder := &LayoutBuilder{
		node: &LayoutNode{
			ElementVNode: ui.NewElement("hstack"),
			direction:    DirectionRow,
			align:        Align(align), // Convert ui.Align to layout.Align type
			crossAlign:   AlignStart,
			gap:          gap,
			stretchCross: stretchCross,
		},
		children: row,
	}
	// Set props with ui.Align values (not layout.Align)
	// GetLayoutInfo expects runtime/ui.Align values
	hstackBuilder.node.SetProp("align", int(align))
	hstackBuilder.node.SetProp("crossAlign", int(ui.AlignStart))
	hstackBuilder.node.SetProp("gap", gap)
	// Set explicit width if provided (width > 0)
	if width > 0 {
		hstackBuilder.node.SetProp("width", width)
	}
	return hstackBuilder.Build()
}

// Build converts the Wrap node into a VStack of HStacks
// This is where the wrapping logic happens
func (b *WrapBuilder) Build() ui.VNode {
	// Handle empty children
	if len(b.children) == 0 {
		// Return empty VStack
		return &LayoutNode{
			ElementVNode: ui.NewElement("vstack"),
			direction:    DirectionColumn,
		}
	}

	// Calculate rows based on child widths and screen width
	rows := b.calculateRows()

	// Store rows for debugging
	b.node.rows = rows

	// Check if fillWidth is enabled - we need this for HStack configuration
	fillWidth := false
	if props := b.node.Props(); props != nil {
		if fw := props.GetBool("fillWidth"); fw {
			fillWidth = true
		}
	}

	// Convert each row to an HStack using LayoutBuilder
	var rowNodes []ui.VNode
	numRows := len(rows)
	if numRows == 0 {
		numRows = 1 // Avoid division by zero
	}

	for _, row := range rows {
		if fillWidth {
			// When fillWidth is enabled, set fixed width on each HStack
			// Calculate width as screenWidth divided by number of rows
			// This allows VStack's stretchCross to evenly distribute the HStacks
			hstackWidth := b.node.screenWidth / numRows
			rowNodes = append(rowNodes, createRowHStack(row, b.node.align, b.node.gap, false, hstackWidth))
		} else {
			// No fillWidth - use original row directly without fixed width
			rowNodes = append(rowNodes, createRowHStack(row, b.node.align, b.node.gap, false, 0))
		}
	}

	// Determine row gap
	rowGap := b.node.rowGap
	if rowGap == 0 {
		rowGap = b.node.gap // Use gap if rowGap not specified
	}

	// Return VStack containing all rows using LayoutBuilder
	vstackBuilder := &LayoutBuilder{
		node: &LayoutNode{
			ElementVNode: ui.NewElement("vstack"),
			direction:    DirectionColumn,
			align:        AlignStart,
			crossAlign:   AlignStart,
			gap:          rowGap,
		},
		children: rowNodes,
	}

	// If fillWidth is enabled, enable stretchCross so children stretch
	if fillWidth {
		vstackBuilder.node.stretchCross = true
		// Set props so GetLayoutInfo can read them
		vstackBuilder.node.SetProp("stretchCross", true)
		vstackBuilder.node.SetProp("gap", rowGap)
		vstackBuilder.node.SetProp("align", int(AlignStart))
		vstackBuilder.node.SetProp("crossAlign", int(AlignStart))
	}

	result := vstackBuilder.Build()

	// Copy style and key from WrapNode to result
	if b.node.Style() != (style.Style{}) {
		result.SetStyle(b.node.Style())
	}
	if b.node.Key() != "" {
		result.SetKey(b.node.Key())
	}

	return result
}
