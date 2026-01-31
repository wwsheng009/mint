package ui

import "github.com/wwsheng009/mint/runtime/style"

// SelectOption represents a single option in a select
type SelectOption struct {
	Value string
	Label string
}

// SelectVNode represents a select dropdown component
type SelectVNode struct {
	*ElementVNode
	options   []SelectOption
	selected  int    // Index of selected option
	disabled  bool
	isFocused bool
	isOpen    bool // Whether dropdown is open
	onChange  func(string)
}

// NewSelect creates a new select
func NewSelect() *SelectVNode {
	return &SelectVNode{
		ElementVNode: NewElement("select"),
		options:      []SelectOption{},
		selected:     -1, // No selection
		disabled:     false,
		isFocused:    false,
		isOpen:       false,
	}
}

// Select creates a new select node
func Select() VNode {
	return NewSelect()
}

// SelectBuilder creates a select builder for chained calls
func SelectBuilder() *SelectBuilderType {
	return &SelectBuilderType{
		node: NewSelect(),
	}
}

// =============================================================================
// SelectVNode methods
// =============================================================================

// Options returns the options list
func (s *SelectVNode) Options() []SelectOption {
	return s.options
}

// SetOptions sets the options list
func (s *SelectVNode) SetOptions(opts []SelectOption) *SelectVNode {
	s.options = opts
	s.SetProp("options", opts)
	return s
}

// AddOption adds a single option
func (s *SelectVNode) AddOption(value, label string) *SelectVNode {
	s.options = append(s.options, SelectOption{Value: value, Label: label})
	return s
}

// Selected returns the selected index
func (s *SelectVNode) Selected() int {
	return s.selected
}

// SetSelected sets the selected index
func (s *SelectVNode) SetSelected(idx int) *SelectVNode {
	if idx >= -1 && idx < len(s.options) {
		s.selected = idx
		s.SetProp("selected", idx)
	}
	return s
}

// SelectedValue returns the selected value
func (s *SelectVNode) SelectedValue() string {
	if s.selected >= 0 && s.selected < len(s.options) {
		return s.options[s.selected].Value
	}
	return ""
}

// SelectedLabel returns the selected label
func (s *SelectVNode) SelectedLabel() string {
	if s.selected >= 0 && s.selected < len(s.options) {
		return s.options[s.selected].Label
	}
	return ""
}

// Disabled returns whether the select is disabled
func (s *SelectVNode) Disabled() bool {
	return s.disabled
}

// SetDisabled sets the disabled state
func (s *SelectVNode) SetDisabled(v bool) *SelectVNode {
	s.disabled = v
	s.SetProp("disabled", v)
	return s
}

// IsFocused returns whether the select is focused
func (s *SelectVNode) IsFocused() bool {
	return s.isFocused
}

// SetFocus sets the focused state
func (s *SelectVNode) SetFocus(focused bool) *SelectVNode {
	s.isFocused = focused
	return s
}

// IsOpen returns whether the dropdown is open
func (s *SelectVNode) IsOpen() bool {
	return s.isOpen
}

// SetOpen sets the open state
func (s *SelectVNode) SetOpen(open bool) *SelectVNode {
	s.isOpen = open
	s.SetProp("open", open)
	return s
}

// OnChange returns the change handler
func (s *SelectVNode) OnChange() func(string) {
	return s.onChange
}

// SetOnChange sets the change handler
func (s *SelectVNode) SetOnChange(fn func(string)) *SelectVNode {
	s.onChange = fn
	s.SetProp("onChange", fn)
	return s
}

// SelectByValue selects an option by value
func (s *SelectVNode) SelectByValue(value string) *SelectVNode {
	for i, opt := range s.options {
		if opt.Value == value {
			s.SetSelected(i)
			break
		}
	}
	return s
}

// =============================================================================
// SelectBuilderType provides fluent API for building selects
// =============================================================================

// SelectBuilderType is the builder for Select
type SelectBuilderType struct {
	node *SelectVNode
}

// Options sets the options list
func (b *SelectBuilderType) Options(opts []SelectOption) *SelectBuilderType {
	b.node.SetOptions(opts)
	return b
}

// AddOption adds a single option
func (b *SelectBuilderType) AddOption(value, label string) *SelectBuilderType {
	b.node.AddOption(value, label)
	return b
}

// Selected sets the selected index
func (b *SelectBuilderType) Selected(idx int) *SelectBuilderType {
	b.node.SetSelected(idx)
	return b
}

// Disabled sets the disabled state
func (b *SelectBuilderType) Disabled(v bool) *SelectBuilderType {
	b.node.SetDisabled(v)
	return b
}

// OnChange sets the change handler
func (b *SelectBuilderType) OnChange(fn func(string)) *SelectBuilderType {
	b.node.SetOnChange(fn)
	return b
}

// Key sets the key for diffing
func (b *SelectBuilderType) Key(key string) *SelectBuilderType {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *SelectBuilderType) Style(s style.Style) *SelectBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *SelectBuilderType) FgColor(c interface{}) *SelectBuilderType {
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
func (b *SelectBuilderType) BgColor(c interface{}) *SelectBuilderType {
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

// Build returns the VNode
func (b *SelectBuilderType) Build() VNode {
	return b.node
}

// =============================================================================
// Table component
// =============================================================================

// TableColumn represents a column in a table
type TableColumn struct {
	Title string
	Width int
}

// TableVNode represents a table component
type TableVNode struct {
	*ElementVNode
	columns []TableColumn
	rows    [][]string
	headerStyle style.Style
}

// NewTable creates a new table
func NewTable() *TableVNode {
	return &TableVNode{
		ElementVNode:  NewElement("table"),
		columns:       []TableColumn{},
		rows:          [][]string{},
		headerStyle:   style.Style{}.Bold(true),
	}
}

// Table creates a new table node
func Table() VNode {
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

// Build returns the VNode
func (b *TableBuilderType) Build() VNode {
	return b.node
}
