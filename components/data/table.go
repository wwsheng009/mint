package data

import (
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// TableColumn represents a column in a table
type TableColumn struct {
	Title string
	Width int
}

// TableVNode represents a table component
type TableVNode struct {
	*ui.ElementVNode
	columns     []TableColumn
	rows        [][]string
	headerStyle style.Style
}

// NewTable creates a new table
func NewTable() *TableVNode {
	return &TableVNode{
		ElementVNode: ui.NewElement("table"),
		columns:      []TableColumn{},
		rows:         [][]string{},
		headerStyle:  style.Style{}.Bold(true),
	}
}

// Table creates a new table node
func Table() ui.VNode {
	return NewTable()
}

// TableBuilder creates a table builder for chained calls
func TableBuilder() *TableBuilderType {
	return &TableBuilderType{
		node: NewTable(),
	}
}

// Columns returns the columns
func (t *TableVNode) Columns() []TableColumn {
	return t.columns
}

// SetColumns sets the columns
func (t *TableVNode) SetColumns(cols []TableColumn) *TableVNode {
	t.columns = cols
	t.SetProp("columns", cols)
	return t
}

// Rows returns the rows
func (t *TableVNode) Rows() [][]string {
	return t.rows
}

// SetRows sets the rows
func (t *TableVNode) SetRows(rows [][]string) *TableVNode {
	t.rows = rows
	t.SetProp("rows", rows)
	return t
}

// AddRow adds a single row
func (t *TableVNode) AddRow(row []string) *TableVNode {
	t.rows = append(t.rows, row)
	return t
}

// HeaderStyle returns the header style
func (t *TableVNode) HeaderStyle() style.Style {
	return t.headerStyle
}

// SetHeaderStyle sets the header style
func (t *TableVNode) SetHeaderStyle(s style.Style) *TableVNode {
	t.headerStyle = s
	return t
}

// =============================================================================
// TableBuilderType provides fluent API for building tables
// =============================================================================

// TableBuilderType is the builder for Table
type TableBuilderType struct {
	node *TableVNode
}

// Columns sets the columns
func (b *TableBuilderType) Columns(cols []TableColumn) *TableBuilderType {
	b.node.SetColumns(cols)
	return b
}

// Rows sets the rows
func (b *TableBuilderType) Rows(rows [][]string) *TableBuilderType {
	b.node.SetRows(rows)
	return b
}

// AddRow adds a single row
func (b *TableBuilderType) AddRow(row ...string) *TableBuilderType {
	b.node.AddRow(row)
	return b
}

// HeaderStyle sets the header style
func (b *TableBuilderType) HeaderStyle(s style.Style) *TableBuilderType {
	b.node.SetHeaderStyle(s)
	return b
}

// Key sets the key for diffing
func (b *TableBuilderType) Key(key string) *TableBuilderType {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *TableBuilderType) Style(s style.Style) *TableBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *TableBuilderType) FgColor(c interface{}) *TableBuilderType {
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
func (b *TableBuilderType) BgColor(c interface{}) *TableBuilderType {
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

// Build returns the ui.VNode
func (b *TableBuilderType) Build() ui.VNode {
	return b.node
}
