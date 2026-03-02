package grid

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for creating Grid VNodes.
type Builder struct {
	grid *VNode
}

// NewBuilder creates a new Grid builder.
func NewBuilder() *Builder {
	return &Builder{
		grid: New(),
	}
}

// Key sets the component key.
func (b *Builder) Key(key string) *Builder {
	b.grid.SetKey(key)
	return b
}

// SetID sets the business identifier for positioning and Portal anchoring.
// This is separate from Key() which is used for list diffing.
func (b *Builder) SetID(id string) *Builder {
	b.grid.SetID(id)
	return b
}

// Columns sets the column definitions.
func (b *Builder) Columns(dims ...Dimension) *Builder {
	b.grid.SetColumns(dims...)
	return b
}

// Rows sets the row definitions.
func (b *Builder) Rows(dims ...Dimension) *Builder {
	b.grid.SetRows(dims...)
	return b
}

// Cell adds a cell at the specified position.
func (b *Builder) Cell(row, col int, child rtui.VNode) *Builder {
	b.grid.AddCell(row, col, child)
	return b
}

// CellSpan adds a cell that spans multiple rows and columns.
func (b *Builder) CellSpan(row, col, rowSpan, colSpan int, child rtui.VNode) *Builder {
	b.grid.AddCellSpan(row, col, rowSpan, colSpan, child)
	return b
}

// Gap sets the gap between columns and rows.
func (b *Builder) Gap(columnGap, rowGap int) *Builder {
	b.grid.SetGap(columnGap, rowGap)
	return b
}

// Padding sets the padding (top, right, bottom, left).
func (b *Builder) Padding(top, right, bottom, left int) *Builder {
	b.grid.SetPadding(top, right, bottom, left)
	return b
}

// AlignContent sets the grid alignment in container.
func (b *Builder) AlignContent(a rtui.Align) *Builder {
	b.grid.SetAlignContent(a)
	return b
}

// Width sets the explicit width.
func (b *Builder) Width(w int) *Builder {
	b.grid.SetWidth(w)
	return b
}

// Height sets the explicit height.
func (b *Builder) Height(h int) *Builder {
	b.grid.SetHeight(h)
	return b
}

// Flex sets the flex factor.
func (b *Builder) Flex(flex int) *Builder {
	b.grid.SetFlex(flex)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.grid.SetStyle(s)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		s := b.grid.Style()
		s.FG = style.Color(colorStr)
		b.grid.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.grid.Style()
		s.FG = color
		b.grid.SetStyle(s)
	}
	return b
}

// BgColor sets the background color.
func (b *Builder) BgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		s := b.grid.Style()
		s.BG = style.Color(colorStr)
		b.grid.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.grid.Style()
		s.BG = color
		b.grid.SetStyle(s)
	}
	return b
}

// Children sets children with auto-positioning in row-major order.
func (b *Builder) Children(children []rtui.VNode) *Builder {
	b.grid.SetChildrenAuto(children)
	return b
}

// Build returns the Grid VNode.
func (b *Builder) Build() rtui.VNode {
	return b.grid
}

// =============================================================================
// Convenience Functions
// =============================================================================

// SimpleGrid creates a simple grid with equal columns and auto-positioned children.
func SimpleGrid(numCols int, children ...rtui.VNode) rtui.VNode {
	cols := make([]Dimension, numCols)
	for i := range cols {
		cols[i] = Flex{Factor: 1}
	}

	return New().
		SetColumns(cols...).
		SetChildrenAuto(children)
}

// FixedGrid creates a grid with fixed-size columns.
func FixedGrid(colWidths []int, children ...rtui.VNode) rtui.VNode {
	cols := make([]Dimension, len(colWidths))
	for i, w := range colWidths {
		cols[i] = Fixed(w)
	}

	return New().
		SetColumns(cols...).
		SetChildrenAuto(children)
}

// TwoColumnGrid creates a 2-column equal width grid.
func TwoColumnGrid(children ...rtui.VNode) rtui.VNode {
	return New().
		SetColumns(Flex{Factor: 1}, Flex{Factor: 1}).
		SetChildrenAuto(children)
}

// ThreeColumnGrid creates a 3-column equal width grid.
func ThreeColumnGrid(children ...rtui.VNode) rtui.VNode {
	return New().
		SetColumns(Flex{Factor: 1}, Flex{Factor: 1}, Flex{Factor: 1}).
		SetChildrenAuto(children)
}
