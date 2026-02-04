package form

import (
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// CheckboxVNode represents a checkbox component
type CheckboxVNode struct {
	*ui.ElementVNode
	checked      bool
	disabled     bool
	label        string
	onChange     func(bool)
	isFocused    bool // Internal focus state
	// Mouse interaction state
	isHovered    bool
	onMouseEnter func()
	onMouseLeave func()
	// Bounds for hit testing (x, y, width, height)
	bounds [4]int
}

// NewCheckbox creates a new checkbox
func NewCheckbox() *CheckboxVNode {
	return &CheckboxVNode{
		ElementVNode: ui.NewElement("checkbox"),
		checked:      false,
		disabled:     false,
		label:        "",
		isFocused:    false,
	}
}

// Checkbox creates a new checkbox node
func Checkbox() ui.VNode {
	return NewCheckbox()
}

// CheckboxBuilder creates a checkbox builder for chained calls
func CheckboxBuilder() *CheckboxBuilderType {
	return &CheckboxBuilderType{
		node: NewCheckbox(),
	}
}

// =============================================================================
// CheckboxVNode methods
// =============================================================================

// Checked returns whether the checkbox is checked
func (c *CheckboxVNode) Checked() bool {
	return c.checked
}

// SetChecked sets the checked state
func (c *CheckboxVNode) SetChecked(v bool) *CheckboxVNode {
	c.checked = v
	c.SetProp("checked", v)
	return c
}

// Toggle toggles the checked state
func (c *CheckboxVNode) Toggle() bool {
	c.checked = !c.checked
	c.SetProp("checked", c.checked)
	return c.checked
}

// Disabled returns whether the checkbox is disabled
func (c *CheckboxVNode) Disabled() bool {
	return c.disabled
}

// SetDisabled sets the disabled state
func (c *CheckboxVNode) SetDisabled(v bool) *CheckboxVNode {
	c.disabled = v
	c.SetProp("disabled", v)
	return c
}

// Label returns the label text
func (c *CheckboxVNode) Label() string {
	return c.label
}

// SetLabel sets the label text
func (c *CheckboxVNode) SetLabel(text string) *CheckboxVNode {
	c.label = text
	c.SetProp("label", text)
	return c
}

// OnChange returns the change handler
func (c *CheckboxVNode) OnChange() func(bool) {
	return c.onChange
}

// SetOnChange sets the change handler
func (c *CheckboxVNode) SetOnChange(fn func(bool)) *CheckboxVNode {
	c.onChange = fn
	c.SetProp("onChange", fn)
	return c
}

// IsFocused returns whether the checkbox is focused
func (c *CheckboxVNode) IsFocused() bool {
	return c.isFocused
}

// SetFocus sets the focused state (implements FocusableVNode)
func (c *CheckboxVNode) SetFocus(focused bool) {
	c.isFocused = focused
}

// =============================================================================
// CheckboxBuilderType provides fluent API for building checkboxes
// =============================================================================

// CheckboxBuilderType is the builder for Checkbox
type CheckboxBuilderType struct {
	node *CheckboxVNode
}

// Checked sets the initial checked state
func (b *CheckboxBuilderType) Checked(v bool) *CheckboxBuilderType {
	b.node.SetChecked(v)
	return b
}

// Disabled sets the disabled state
func (b *CheckboxBuilderType) Disabled(v bool) *CheckboxBuilderType {
	b.node.SetDisabled(v)
	return b
}

// Label sets the label text
func (b *CheckboxBuilderType) Label(text string) *CheckboxBuilderType {
	b.node.SetLabel(text)
	return b
}

// OnChange sets the change handler
func (b *CheckboxBuilderType) OnChange(fn func(bool)) *CheckboxBuilderType {
	b.node.SetOnChange(fn)
	return b
}

// Key sets the key for diffing
func (b *CheckboxBuilderType) Key(key string) *CheckboxBuilderType {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *CheckboxBuilderType) Style(s style.Style) *CheckboxBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *CheckboxBuilderType) FgColor(c interface{}) *CheckboxBuilderType {
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
func (b *CheckboxBuilderType) BgColor(c interface{}) *CheckboxBuilderType {
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

// Bold sets the bold attribute
func (b *CheckboxBuilderType) Bold(v bool) *CheckboxBuilderType {
	s := b.node.Style()
	s = s.Bold(v)
	b.node.SetStyle(s)
	return b
}

// Build returns the ui.VNode
func (b *CheckboxBuilderType) Build() ui.VNode {
	return b.node
}

// =============================================================================
// Mouse Event Support
// =============================================================================

// IsHovered returns whether the checkbox is currently hovered
func (c *CheckboxVNode) IsHovered() bool {
	return c.isHovered
}

// SetHovered sets the hover state
func (c *CheckboxVNode) SetHovered(hovered bool) *CheckboxVNode {
	c.isHovered = hovered
	return c
}

// SetBounds sets the checkbox bounds for hit testing
func (c *CheckboxVNode) SetBounds(x, y, width, height int) {
	c.bounds = [4]int{x, y, width, height}
}

// Bounds returns the checkbox bounds
func (c *CheckboxVNode) Bounds() [4]int {
	return c.bounds
}

// ContainsPoint checks if a point is within the checkbox bounds
func (c *CheckboxVNode) ContainsPoint(x, y int) bool {
	if c.bounds[2] <= 0 || c.bounds[3] <= 0 {
		return false
	}
	return x >= c.bounds[0] && x < c.bounds[0]+c.bounds[2] &&
		y >= c.bounds[1] && y < c.bounds[1]+c.bounds[3]
}

// SetOnMouseEnter sets the mouse enter handler
func (c *CheckboxVNode) SetOnMouseEnter(fn func()) *CheckboxVNode {
	c.onMouseEnter = fn
	return c
}

// SetOnMouseLeave sets the mouse leave handler
func (c *CheckboxVNode) SetOnMouseLeave(fn func()) *CheckboxVNode {
	c.onMouseLeave = fn
	return c
}

// HandleEvent processes mouse and keyboard events for the checkbox
func (c *CheckboxVNode) HandleEvent(e event.Event) bool {
	if c.disabled {
		return false
	}

	// Handle keyboard events (Space to toggle)
	keyEvent, ok := e.(*event.KeyEvent)
	if ok {
		// Only respond to keyboard events when focused
		if !c.isFocused {
			return false
		}

		// Space toggles the checkbox
		if keyEvent.Key.Rune == ' ' {
			newState := c.Toggle()
			if c.onChange != nil {
				c.onChange(newState)
			}
			return true
		}
		return false
	}

	mouseEvent, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}

	switch mouseEvent.Type() {
	case event.EventMouseEnter:
		if !c.isHovered {
			c.isHovered = true
			if c.onMouseEnter != nil {
				c.onMouseEnter()
			}
		}
		return true

	case event.EventMouseLeave:
		if c.isHovered {
			c.isHovered = false
			if c.onMouseLeave != nil {
				c.onMouseLeave()
			}
		}
		return true

	case event.EventMousePress, event.EventClick:
		if c.isHovered && mouseEvent.Button == event.MouseLeft {
			// Toggle the checkbox
			newState := c.Toggle()
			if c.onChange != nil {
				c.onChange(newState)
			}
			return true
		}
	}

	return false
}

// =============================================================================
// Mouse Event Builder Methods
// =============================================================================

// OnMouseEnter sets the mouse enter handler (builder)
func (b *CheckboxBuilderType) OnMouseEnter(fn func()) *CheckboxBuilderType {
	b.node.SetOnMouseEnter(fn)
	return b
}

// OnMouseLeave sets the mouse leave handler (builder)
func (b *CheckboxBuilderType) OnMouseLeave(fn func()) *CheckboxBuilderType {
	b.node.SetOnMouseLeave(fn)
	return b
}

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
// Calculates the size of the checkbox based on label and constraints
func (c *CheckboxVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if c == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Width: checkbox "[X]" (3) + space (1) + label length
	width := 4 + utf8.RuneCountInString(c.label)
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
	elemStyle := c.Style()
	if elemStyle.Width > 0 {
		width = elemStyle.Width
	}
	if elemStyle.Height > 0 {
		height = elemStyle.Height
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface
// Generates draw commands for rendering this checkbox component
func (c *CheckboxVNode) Paint(x, y int) []paint.DrawCmd {
	if c == nil {
		return nil
	}

	checkboxStyle := c.Style()

	// Checkbox indicator: [X] or [ ]
	var indicator string
	if c.checked {
		indicator = "[X]"
	} else {
		indicator = "[ ]"
	}

	// Build checkbox display: indicator + label
	var displayText string
	if c.label != "" {
		displayText = indicator + " " + c.label
	} else {
		displayText = indicator
	}

	// State priority: Focused > Hovered > Normal
	// Focus: blue background with white text for clear visibility
	// Hover: underline only
	if c.isFocused && !c.disabled {
		checkboxStyle = checkboxStyle.Foreground(style.Color("white")).Background(style.Color("blue")).Bold(true)
	} else if c.isHovered && !c.disabled {
		checkboxStyle = checkboxStyle.Underline(true)
	}

	// Apply disabled state
	if c.disabled {
		checkboxStyle = checkboxStyle.Foreground(style.Color("gray"))
	}

	return []paint.DrawCmd{
		paint.NewTextCmd(x, y, displayText, checkboxStyle),
	}
}

// =============================================================================
// FocusableVNode Interface Implementation
// =============================================================================

// IsFocusable returns whether this checkbox can receive focus.
// Disabled checkboxes cannot receive focus.
func (c *CheckboxVNode) IsFocusable() bool {
	return !c.disabled
}

// GetFocusID returns a unique identifier for focus persistence.
// Uses the checkbox's Key if set, otherwise generates a stable ID.
func (c *CheckboxVNode) GetFocusID() string {
	if key := c.Key(); key != "" {
		return "checkbox:" + key
	}
	// Generate stable ID based on label
	id := c.label
	if id == "" {
		id = "checkbox"
	}
	return "checkbox:" + id
}
