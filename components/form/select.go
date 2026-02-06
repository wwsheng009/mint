package form

import (
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/theme"
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

// SetFocus sets the focused state (implements FocusableVNode)
func (s *SelectVNode) SetFocus(focused bool) {
	s.isFocused = focused
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

// HandleEvent processes mouse and keyboard events for the select
func (s *SelectVNode) HandleEvent(e event.Event) bool {
	if s.disabled {
		return false
	}

	// Handle keyboard events (Space/Enter to cycle)
	keyEvent, ok := e.(*event.KeyEvent)
	if ok {
		// Only respond to keyboard events when focused
		if !s.isFocused {
			return false
		}

		// Space or Enter cycles to next option
		if keyEvent.Key.Rune == ' ' || keyEvent.Special == event.KeyEnter {
			if len(s.options) > 0 {
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
		s = s.Foreground(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Foreground(color)
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *SelectBuilderType) BgColor(c interface{}) *SelectBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Background(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Background(color)
		b.node.SetStyle(s)
	}
	return b
}

// Build returns the ui.VNode
func (b *SelectBuilderType) Build() ui.VNode {
	return b.node
}

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
// Calculates the size of the select based on options and constraints
func (s *SelectVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if s == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Find the longest option label
	maxWidth := 0
	for _, opt := range s.options {
		labelWidth := utf8.RuneCountInString(opt.Label)
		if labelWidth > maxWidth {
			maxWidth = labelWidth
		}
	}

	// If no options, use default width
	if maxWidth == 0 {
		maxWidth = 10
	}

	// Width: longest label + 4 for "< " and " >"
	width := maxWidth + 4
	height := 1

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	// Apply explicit style dimensions if set
	elemStyle := s.Style()
	if elemStyle.Width > 0 {
		width = elemStyle.Width
	}
	if elemStyle.Height > 0 {
		height = elemStyle.Height
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface
// Generates draw commands for rendering this select component
func (s *SelectVNode) Paint(x, y int) []paint.DrawCmd {
	if s == nil {
		return nil
	}

	selectStyle := s.Style()

	// Get the selected label to display
	displayLabel := s.SelectedLabel()
	if displayLabel == "" {
		if len(s.options) > 0 {
			displayLabel = s.options[0].Label
		} else {
			displayLabel = "..."
		}
	}

	// Build select display: < label >
	// Truncate if too long
	measured := s.Measure(runtime.BoxConstraints{})
	maxLabelWidth := measured.Width - 4

	labelWidth := utf8.RuneCountInString(displayLabel)
	if labelWidth > maxLabelWidth {
		// Truncate label
		runes := []rune(displayLabel)
		displayLabel = string(runes[:maxLabelWidth-3]) + "..."
	}

	selectDisplay := "< " + displayLabel + " >"

	// Apply default colors based on component spec
	// Normal: BG=SURFACE, FG=TEXT
	if selectStyle.FG == "" {
		selectStyle = selectStyle.Foreground(theme.Text())
	}
	if selectStyle.BG == "" {
		selectStyle = selectStyle.Background(theme.Surface())
	}

	// State priority: Focused > Hovered > Normal
	// Focus: FOCUS border effect
	if s.isFocused && !s.disabled {
		selectStyle = selectStyle.Foreground(theme.Focus()).Bold(true)
	} else if s.isHovered && !s.disabled {
		selectStyle = selectStyle.Underline(true)
	}

	// Apply disabled state: DISABLED_BG, DISABLED_FG
	if s.disabled {
		selectStyle = selectStyle.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}

	return []paint.DrawCmd{
		paint.NewTextCmd(x, y, selectDisplay, selectStyle),
	}
}

// =============================================================================
// FocusableVNode Interface Implementation
// =============================================================================

// IsFocusable returns whether this select can receive focus.
// Disabled selects cannot receive focus.
func (s *SelectVNode) IsFocusable() bool {
	return !s.disabled && len(s.options) > 0
}

// GetFocusID returns a unique identifier for focus persistence.
// Uses the select's Key if set, otherwise generates a stable ID.
func (s *SelectVNode) GetFocusID() string {
	if key := s.Key(); key != "" {
		return "select:" + key
	}
	// Generate stable ID based on first option label
	id := ""
	if len(s.options) > 0 {
		id = s.options[0].Label
	}
	if id == "" {
		id = "select"
	}
	return "select:" + id
}

// Label returns a label for this select for testing/debugging.
// Returns the first option's label or "select".
func (s *SelectVNode) Label() string {
	if len(s.options) > 0 {
		return s.options[0].Label
	}
	return "select"
}
