package ui

// =============================================================================
// Component VNode Types
// =============================================================================
// These types define the structure for various UI component VNodes.
// They are used by the reconciler for type assertions and conversions.

// TextVNode represents a text node with content
type TextVNode struct {
	*ElementVNode
	content string
}

// Content returns the text content
func (t *TextVNode) Content() string {
	if t == nil {
		return ""
	}
	return t.content
}

// Tag returns "text" (implements VNode interface)
func (t *TextVNode) Tag() string {
	return "text"
}

// ButtonVNode represents a button component
//
// Deprecated: Use github.com/wwsheng009/mint/ui/components/button.Builder instead.
// The Builder API provides typed intent-based event handling via OnPress().
// Example: button.NewBuilder("Click Me").OnPress(someIntent).Build()
type ButtonVNode struct {
	*ElementVNode
	label    string
	onClick  func() // Deprecated: Use button.Builder().OnPress() instead
	disabled bool
}

// Label returns the button label
func (b *ButtonVNode) Label() string {
	if b == nil {
		return ""
	}
	return b.label
}

// OnClick returns the click handler
//
// Deprecated: Use button.Builder().OnPress() with intent.Intent instead.
// Closure-based event handlers will be removed in v2.0.
// Example: button.NewBuilder("Submit").OnPress(SubmitIntent{}).Build()
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
	return false
}

// InputVNode represents an input field
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

// TextareaVNode represents a textarea field
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

// CheckboxVNode represents a checkbox component
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

// SelectVNode represents a select dropdown
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

// Options returns the select options
func (s *SelectVNode) Options() []SelectOption {
	if s == nil {
		return nil
	}
	return s.options
}

// ProgressVNode represents a progress bar
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

// SpinnerVNode represents a loading spinner
type SpinnerVNode struct {
	*ElementVNode
}

// TableVNode represents a table
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

// VirtualListVNode represents a virtualized list
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

// TabsVNode represents a tab container
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

// ModalVNode represents a modal dialog
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

// TooltipVNode represents a tooltip
type TooltipVNode struct {
	*ElementVNode
}

// ToastVNode represents a toast notification
type ToastVNode struct {
	*ElementVNode
}

// GridVNode represents a grid layout
type GridVNode struct {
	*ElementVNode
}

// AbsoluteVNode represents absolute positioning
type AbsoluteVNode struct {
	*ElementVNode
}

// StackVNode represents a stack layout
type StackVNode struct {
	*ElementVNode
}
