package form

import (
	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// InputType represents the type of input
type InputType int

const (
	// InputTypeText is a standard text input
	InputTypeText InputType = iota
	// InputTypePassword is a password input (chars hidden)
	InputTypePassword
	// InputTypeNumber is a numeric input
	InputTypeNumber
	// InputTypeEmail is an email input
	InputTypeEmail
)

// InputVNode represents an input component
type InputVNode struct {
	*ui.ElementVNode
	value       string
	placeholder string
	inputType   InputType
	maxLength   int
	disabled    bool
	readOnly    bool
	onChange    func(string)
	onFocus     func()
	onBlur      func()
	onSubmit    func()
	isFocused   bool // Internal focus state
	focusIndex  int // Index for focus management, set during collection
	// Mouse interaction state
	isHovered bool
	// Bounds for hit testing (x, y, width, height)
	bounds [4]int
}

// NewInput creates a new input
func NewInput() *InputVNode {
	return &InputVNode{
		ElementVNode: ui.NewElement("input"),
		value:        "",
		placeholder:  "",
		inputType:    InputTypeText,
		maxLength:    0, // 0 = no limit
		disabled:     false,
		readOnly:     false,
		isFocused:    false,
	}
}

// Input creates a new input node
func Input() ui.VNode {
	return NewInput()
}

// InputBuilder creates a input builder for chained calls
func InputBuilder() *InputBuilderType {
	return &InputBuilderType{
		node: NewInput(),
	}
}

// =============================================================================
// InputVNode methods
// =============================================================================

// Value returns the input value
func (i *InputVNode) Value() string {
	return i.value
}

// SetValue sets the input value
func (i *InputVNode) SetValue(value string) *InputVNode {
	i.value = value
	i.SetProp("value", value)
	return i
}

// Placeholder returns the placeholder text
func (i *InputVNode) Placeholder() string {
	return i.placeholder
}

// SetPlaceholder sets the placeholder text
func (i *InputVNode) SetPlaceholder(text string) *InputVNode {
	i.placeholder = text
	i.SetProp("placeholder", text)
	return i
}

// InputType returns the input type
func (i *InputVNode) InputType() InputType {
	return i.inputType
}

// SetInputType sets the input type
func (i *InputVNode) SetInputType(t InputType) *InputVNode {
	i.inputType = t
	i.SetProp("inputType", t)
	return i
}

// MaxLength returns the max length
func (i *InputVNode) MaxLength() int {
	return i.maxLength
}

// SetMaxLength sets the max length (0 = no limit)
func (i *InputVNode) SetMaxLength(len int) *InputVNode {
	i.maxLength = len
	i.SetProp("maxLength", len)
	return i
}

// Disabled returns whether the input is disabled
func (i *InputVNode) Disabled() bool {
	return i.disabled
}

// SetDisabled sets the disabled state
func (i *InputVNode) SetDisabled(v bool) *InputVNode {
	i.disabled = v
	i.SetProp("disabled", v)
	return i
}

// ReadOnly returns whether the input is read-only
func (i *InputVNode) ReadOnly() bool {
	return i.readOnly
}

// SetReadOnly sets the read-only state
func (i *InputVNode) SetReadOnly(v bool) *InputVNode {
	i.readOnly = v
	i.SetProp("readOnly", v)
	return i
}

// OnChange returns the change handler
func (i *InputVNode) OnChange() func(string) {
	return i.onChange
}

// SetOnChange sets the change handler
func (i *InputVNode) SetOnChange(fn func(string)) *InputVNode {
	i.onChange = fn
	i.SetProp("onChange", fn)
	return i
}

// OnFocusFunc returns the focus handler
func (i *InputVNode) OnFocusFunc() func() {
	return i.onFocus
}

// SetOnFocus sets the focus handler
func (i *InputVNode) SetOnFocus(fn func()) *InputVNode {
	i.onFocus = fn
	i.SetProp("onFocus", fn)
	return i
}

// OnBlurFunc returns the blur handler
func (i *InputVNode) OnBlurFunc() func() {
	return i.onBlur
}

// SetOnBlur sets the blur handler
func (i *InputVNode) SetOnBlur(fn func()) *InputVNode {
	i.onBlur = fn
	i.SetProp("onBlur", fn)
	return i
}

// OnSubmitFunc returns the submit handler
func (i *InputVNode) OnSubmitFunc() func() {
	return i.onSubmit
}

// SetOnSubmit sets the submit handler
func (i *InputVNode) SetOnSubmit(fn func()) *InputVNode {
	i.onSubmit = fn
	i.SetProp("onSubmit", fn)
	return i
}

// IsFocused returns whether the input is focused
func (i *InputVNode) IsFocused() bool {
	return i.isFocused
}

// SetFocus sets the focused state (used internally)
func (i *InputVNode) SetFocus(focused bool) *InputVNode {
	i.isFocused = focused
	return i
}

// =============================================================================
// InputBuilderType provides fluent API for building inputs
// =============================================================================

// InputBuilderType is the builder for Input
type InputBuilderType struct {
	node *InputVNode
}

// Value sets the initial value
func (b *InputBuilderType) Value(value string) *InputBuilderType {
	b.node.SetValue(value)
	return b
}

// Placeholder sets the placeholder text
func (b *InputBuilderType) Placeholder(text string) *InputBuilderType {
	b.node.SetPlaceholder(text)
	return b
}

// Type sets the input type (Text, Password, Number, Email)
func (b *InputBuilderType) Type(t InputType) *InputBuilderType {
	b.node.SetInputType(t)
	return b
}

// Password sets the input type to password
func (b *InputBuilderType) Password() *InputBuilderType {
	b.node.SetInputType(InputTypePassword)
	return b
}

// MaxLength sets the maximum length
func (b *InputBuilderType) MaxLength(len int) *InputBuilderType {
	b.node.SetMaxLength(len)
	return b
}

// Disabled sets the disabled state
func (b *InputBuilderType) Disabled(v bool) *InputBuilderType {
	b.node.SetDisabled(v)
	return b
}

// ReadOnly sets the read-only state
func (b *InputBuilderType) ReadOnly(v bool) *InputBuilderType {
	b.node.SetReadOnly(v)
	return b
}

// OnChange sets the change handler
func (b *InputBuilderType) OnChange(fn func(string)) *InputBuilderType {
	b.node.SetOnChange(fn)
	return b
}

// OnFocus sets the focus handler
func (b *InputBuilderType) OnFocus(fn func()) *InputBuilderType {
	b.node.SetOnFocus(fn)
	return b
}

// OnBlur sets the blur handler
func (b *InputBuilderType) OnBlur(fn func()) *InputBuilderType {
	b.node.SetOnBlur(fn)
	return b
}

// OnSubmit sets the submit handler (Enter key)
func (b *InputBuilderType) OnSubmit(fn func()) *InputBuilderType {
	b.node.SetOnSubmit(fn)
	return b
}

// Key sets the key for diffing
func (b *InputBuilderType) Key(key string) *InputBuilderType {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *InputBuilderType) Style(s style.Style) *InputBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *InputBuilderType) FgColor(c interface{}) *InputBuilderType {
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
func (b *InputBuilderType) BgColor(c interface{}) *InputBuilderType {
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

// Bold sets the bold attribute
func (b *InputBuilderType) Bold(v bool) *InputBuilderType {
	s := b.node.Style()
	s = s.Bold(v)
	b.node.SetStyle(s)
	return b
}

// Build returns the ui.VNode
func (b *InputBuilderType) Build() ui.VNode {
	return b.node
}

// =============================================================================
// Mouse Event Support
// =============================================================================

// IsHovered returns whether the input is currently hovered
func (i *InputVNode) IsHovered() bool {
	return i.isHovered
}

// SetHovered sets the hover state
func (i *InputVNode) SetHovered(hovered bool) *InputVNode {
	i.isHovered = hovered
	return i
}

// SetBounds sets the input bounds for hit testing
func (i *InputVNode) SetBounds(x, y, width, height int) {
	i.bounds = [4]int{x, y, width, height}
}

// Bounds returns the input bounds
func (i *InputVNode) Bounds() [4]int {
	return i.bounds
}

// ContainsPoint checks if a point is within the input bounds
func (i *InputVNode) ContainsPoint(x, y int) bool {
	if i.bounds[2] <= 0 || i.bounds[3] <= 0 {
		return false
	}
	return x >= i.bounds[0] && x < i.bounds[0]+i.bounds[2] &&
		y >= i.bounds[1] && y < i.bounds[1]+i.bounds[3]
}

// HandleEvent processes mouse events for the input
func (i *InputVNode) HandleEvent(e event.Event) bool {
	if i.disabled || i.readOnly {
		return false
	}

	mouseEvent, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}

	switch mouseEvent.Type() {
	case event.EventMouseEnter:
		if !i.isHovered {
			i.isHovered = true
		}
		return true

	case event.EventMouseLeave:
		if i.isHovered {
			i.isHovered = false
		}
		return true

	case event.EventMousePress, event.EventClick:
		if i.isHovered && mouseEvent.Button == event.MouseLeft {
			// Focus the input - the actual focus is managed by the framework
			// Return true to indicate the event was handled
			return true
		}
	}

	return false
}
