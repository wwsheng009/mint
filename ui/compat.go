package ui

// =============================================================================
// Compatibility Types for components/ migration
// =============================================================================
// This file provides minimal type definitions for backward compatibility.
// These are stub types that will be phased out as code migrates to use components/ directly.

// Basic component types (minimal stub - content field required for diff)
type TextVNode struct {
	*ElementVNode
	content string
}

// Content returns the text content (for compatibility with diff.go)
func (t *TextVNode) Content() string {
	if t == nil {
		return ""
	}
	return t.content
}

// Component stub types - these allow compilation but full functionality
// requires using the actual components from components/ package
type ButtonVNode struct {
	*ElementVNode
	label      string
	onClick    func()
	disabled   bool
	focusIndex int
}

// Label returns the button label
func (b *ButtonVNode) Label() string {
	if b == nil {
		return ""
	}
	return b.label
}

// OnClick returns the click handler
func (b *ButtonVNode) OnClick() func() {
	if b == nil {
		return nil
	}
	return b.onClick
}

// Disabled returns whether the button is disabled
func (b *ButtonVNode) Disabled() bool {
	if b == nil {
		return false
	}
	return b.disabled
}

// ContainsPoint checks if a point is within the button's bounds
func (b *ButtonVNode) ContainsPoint(x, y int) bool {
	// Stub implementation - always returns false
	// Actual implementation would check bounds
	return false
}

type InputVNode struct {
	*ElementVNode
	value      string
	placeholder string
}

// Value returns the input value
func (i *InputVNode) Value() string {
	if i == nil {
		return ""
	}
	return i.value
}

// Placeholder returns the input placeholder
func (i *InputVNode) Placeholder() string {
	if i == nil {
		return ""
	}
	return i.placeholder
}

type TextareaVNode struct {
	*ElementVNode
	value      string
	placeholder string
	rows       int
}

// Value returns the textarea value
func (t *TextareaVNode) Value() string {
	if t == nil {
		return ""
	}
	return t.value
}

// Placeholder returns the textarea placeholder
func (t *TextareaVNode) Placeholder() string {
	if t == nil {
		return ""
	}
	return t.placeholder
}

type CheckboxVNode struct {
	*ElementVNode
	checked bool
	label   string
}

// Label returns the checkbox label
func (c *CheckboxVNode) Label() string {
	if c == nil {
		return ""
	}
	return c.label
}

// Checked returns whether the checkbox is checked
func (c *CheckboxVNode) Checked() bool {
	if c == nil {
		return false
	}
	return c.checked
}

type SelectVNode struct {
	*ElementVNode
	selected int
	options  []SelectOption
}

// Selected returns the selected index
func (s *SelectVNode) Selected() int {
	if s == nil {
		return 0
	}
	return s.selected
}

// SelectOption represents an option in a select dropdown
type SelectOption struct {
	Value string
	Label string
}

// Options returns the select options (stub)
func (s *SelectVNode) Options() []SelectOption {
	return nil
}

type ProgressVNode struct {
	*ElementVNode
	value   int
	max     int
	percent int
	width   int
}

// Value returns the progress value
func (p *ProgressVNode) Value() int {
	if p == nil {
		return 0
	}
	return p.value
}

// Max returns the progress max value
func (p *ProgressVNode) Max() int {
	if p == nil {
		return 0
	}
	return p.max
}

// Percent returns the progress percentage
func (p *ProgressVNode) Percent() int {
	if p == nil {
		return 0
	}
	return p.percent
}

// Width returns the progress width
func (p *ProgressVNode) Width() int {
	if p == nil {
		return 0
	}
	return p.width
}

type SpinnerVNode struct {
	*ElementVNode
}

type TableVNode struct {
	*ElementVNode
	columns []string
	rows    [][]string
}

// Columns returns the table columns
func (t *TableVNode) Columns() []string {
	if t == nil {
		return nil
	}
	return t.columns
}

// Rows returns the table rows
func (t *TableVNode) Rows() [][]string {
	if t == nil {
		return nil
	}
	return t.rows
}

type VirtualListVNode struct {
	*ElementVNode
	listHeight int
	itemCount  int
	itemHeight int
}

// ListHeight returns the virtual list height
func (v *VirtualListVNode) ListHeight() int {
	if v == nil {
		return 0
	}
	return v.listHeight
}

// ItemCount returns the virtual list item count
func (v *VirtualListVNode) ItemCount() int {
	if v == nil {
		return 0
	}
	return v.itemCount
}

// ItemHeight returns the virtual list item height
func (v *VirtualListVNode) ItemHeight() int {
	if v == nil {
		return 0
	}
	return v.itemHeight
}

type TabsVNode struct {
	*ElementVNode
	activeTab string
}

// ActiveTab returns the active tab
func (t *TabsVNode) ActiveTab() string {
	if t == nil {
		return ""
	}
	return t.activeTab
}

type ModalVNode struct {
	*ElementVNode
	width   int
	height  int
	title   string
	isOpen  bool
	content VNode
	footer  VNode
}

// Width returns the modal width
func (m *ModalVNode) Width() int {
	if m == nil {
		return 0
	}
	return m.width
}

// Height returns the modal height
func (m *ModalVNode) Height() int {
	if m == nil {
		return 0
	}
	return m.height
}

// Title returns the modal title
func (m *ModalVNode) Title() string {
	if m == nil {
		return ""
	}
	return m.title
}

// IsOpen returns whether the modal is open
func (m *ModalVNode) IsOpen() bool {
	if m == nil {
		return false
	}
	return m.isOpen
}

// Content returns the modal content
func (m *ModalVNode) Content() VNode {
	if m == nil {
		return nil
	}
	return m.content
}

// Footer returns the modal footer
func (m *ModalVNode) Footer() VNode {
	if m == nil {
		return nil
	}
	return m.footer
}

type TooltipVNode struct {
	*ElementVNode
}

type ToastVNode struct {
	*ElementVNode
}

type GridVNode struct {
	*ElementVNode
}

type AbsoluteVNode struct {
	*ElementVNode
}

// StackVNode is now defined in components/layout
// This stub allows compilation for now
type StackVNode struct {
	*ElementVNode
}
