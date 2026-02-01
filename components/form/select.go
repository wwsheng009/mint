package form

import (
	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// SelectOption represents a single option in a select
type SelectOption struct {
	Value string
	Label string
}

// SelectVNode represents a select dropdown component
type SelectVNode struct {
	*ui.ElementVNode
	options    []SelectOption
	selected   int    // Index of selected option
	disabled   bool
	isFocused  bool
	isOpen     bool // Whether dropdown is open
	onChange   func(string)
	focusIndex int // Index for focus management, set during collection
	// Mouse interaction state
	isHovered bool
	// Bounds for hit testing (x, y, width, height)
	bounds [4]int
}

// NewSelect creates a new select
func NewSelect() *SelectVNode {
	return &SelectVNode{
		ElementVNode: ui.NewElement("select"),
		options:      []SelectOption{},
		selected:     -1, // No selection
		disabled:     false,
		isFocused:    false,
		isOpen:       false,
	}
}

// Select creates a new select node
func Select() ui.VNode {
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
// Mouse Event Support
// =============================================================================

// IsHovered returns whether the select is currently hovered
func (s *SelectVNode) IsHovered() bool {
	return s.isHovered
}

// SetHovered sets the hover state
func (s *SelectVNode) SetHovered(hovered bool) *SelectVNode {
	s.isHovered = hovered
	return s
}

// SetBounds sets the select bounds for hit testing
func (s *SelectVNode) SetBounds(x, y, width, height int) {
	s.bounds = [4]int{x, y, width, height}
}

// Bounds returns the select bounds
func (s *SelectVNode) Bounds() [4]int {
	return s.bounds
}

// ContainsPoint checks if a point is within the select bounds
func (s *SelectVNode) ContainsPoint(x, y int) bool {
	if s.bounds[2] <= 0 || s.bounds[3] <= 0 {
		return false
	}
	return x >= s.bounds[0] && x < s.bounds[0]+s.bounds[2] &&
		y >= s.bounds[1] && y < s.bounds[1]+s.bounds[3]
}

// HandleEvent processes mouse events for the select
func (s *SelectVNode) HandleEvent(e event.Event) bool {
	if s.disabled {
		return false
	}

	mouseEvent, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}

	switch mouseEvent.Type() {
	case event.EventMouseEnter:
		if !s.isHovered {
			s.isHovered = true
		}
		return true

	case event.EventMouseLeave:
		if s.isHovered {
			s.isHovered = false
		}
		return true

	case event.EventMousePress, event.EventClick:
		if s.isHovered && mouseEvent.Button == event.MouseLeft {
			// Cycle to next option on click
			nextIdx := s.selected + 1
			if nextIdx >= len(s.options) {
				nextIdx = 0
			}
			s.SetSelected(nextIdx)
			if s.onChange != nil {
				s.onChange(s.SelectedValue())
			}
			return true
		}
	}

	return false
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

// Build returns the ui.VNode
func (b *SelectBuilderType) Build() ui.VNode {
	return b.node
}
