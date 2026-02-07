package form

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime/dimension"
	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Textarea component
// =============================================================================

// TextareaVNode represents a textarea component (multi-line input)
type TextareaVNode struct {
	*ui.ElementVNode
	value       string
	placeholder string
	rows        int
	cols        int
	maxLength   int
	disabled    bool
	resize      bool
	onChange    func(string)
	onFocus     func()
	onBlur      func()
	onSubmit    func() // Ctrl+Enter or Alt+Enter
	isFocused   bool
	// Mouse interaction state
	isHovered bool
	// Bounds for hit testing (x, y, width, height)
	bounds [4]int
}

// NewTextarea creates a new textarea
func NewTextarea() *TextareaVNode {
	return &TextareaVNode{
		ElementVNode: ui.NewElement("textarea"),
		value:        "",
		placeholder:  "",
		rows:         3,
		cols:         40,
		maxLength:    0,
		disabled:     false,
		resize:       true,
		isFocused:    false,
	}
}

// Textarea creates a new textarea node
func Textarea() ui.VNode {
	return NewTextarea()
}

// TextareaBuilder creates a textarea builder for chained calls
func TextareaBuilder() *TextareaBuilderType {
	return &TextareaBuilderType{
		node: NewTextarea(),
	}
}

// Value returns the textarea value
func (t *TextareaVNode) Value() string {
	return t.value
}

// SetValue sets the textarea value
func (t *TextareaVNode) SetValue(value string) *TextareaVNode {
	t.value = value
	t.SetProp("value", value)
	return t
}

// Placeholder returns the placeholder text
func (t *TextareaVNode) Placeholder() string {
	return t.placeholder
}

// SetPlaceholder sets the placeholder text
func (t *TextareaVNode) SetPlaceholder(text string) *TextareaVNode {
	t.placeholder = text
	t.SetProp("placeholder", text)
	return t
}

// Rows returns the number of rows
func (t *TextareaVNode) Rows() int {
	return t.rows
}

// SetRows sets the number of rows
func (t *TextareaVNode) SetRows(rows int) *TextareaVNode {
	t.rows = rows
	t.SetProp("rows", rows)
	return t
}

// Cols returns the number of columns
func (t *TextareaVNode) Cols() int {
	return t.cols
}

// SetCols sets the number of columns
func (t *TextareaVNode) SetCols(cols int) *TextareaVNode {
	t.cols = cols
	t.SetProp("cols", cols)
	return t
}

// MaxLength returns the max length
func (t *TextareaVNode) MaxLength() int {
	return t.maxLength
}

// SetMaxLength sets the max length (0 = no limit)
func (t *TextareaVNode) SetMaxLength(len int) *TextareaVNode {
	t.maxLength = len
	t.SetProp("maxLength", len)
	return t
}

// Disabled returns whether the textarea is disabled
func (t *TextareaVNode) Disabled() bool {
	return t.disabled
}

// SetDisabled sets the disabled state
func (t *TextareaVNode) SetDisabled(v bool) *TextareaVNode {
	t.disabled = v
	t.SetProp("disabled", v)
	return t
}

// Resize returns whether the textarea is resizable
func (t *TextareaVNode) Resize() bool {
	return t.resize
}

// SetResize sets the resizable state
func (t *TextareaVNode) SetResize(v bool) *TextareaVNode {
	t.resize = v
	t.SetProp("resize", v)
	return t
}

// OnChange returns the change handler
func (t *TextareaVNode) OnChange() func(string) {
	return t.onChange
}

// SetOnChange sets the change handler
func (t *TextareaVNode) SetOnChange(fn func(string)) *TextareaVNode {
	t.onChange = fn
	t.SetProp("onChange", fn)
	return t
}

// OnFocusFunc returns the focus handler
func (t *TextareaVNode) OnFocusFunc() func() {
	return t.onFocus
}

// SetOnFocus sets the focus handler
func (t *TextareaVNode) SetOnFocus(fn func()) *TextareaVNode {
	t.onFocus = fn
	t.SetProp("onFocus", fn)
	return t
}

// OnBlurFunc returns the blur handler
func (t *TextareaVNode) OnBlurFunc() func() {
	return t.onBlur
}

// SetOnBlur sets the blur handler
func (t *TextareaVNode) SetOnBlur(fn func()) *TextareaVNode {
	t.onBlur = fn
	t.SetProp("onBlur", fn)
	return t
}

// OnSubmitFunc returns the submit handler
func (t *TextareaVNode) OnSubmitFunc() func() {
	return t.onSubmit
}

// SetOnSubmit sets the submit handler
func (t *TextareaVNode) SetOnSubmit(fn func()) *TextareaVNode {
	t.onSubmit = fn
	t.SetProp("onSubmit", fn)
	return t
}

// IsFocused returns whether the textarea is focused
func (t *TextareaVNode) IsFocused() bool {
	return t.isFocused
}

// SetFocus sets the focused state (implements FocusableVNode)
func (t *TextareaVNode) SetFocus(focused bool) {
	t.isFocused = focused
	// Call focus handler if set
	if focused && t.onFocus != nil {
		t.onFocus()
	} else if !focused && t.onBlur != nil {
		t.onBlur()
	}
}

// =============================================================================
// Mouse Event Support
// =============================================================================

// IsHovered returns whether the textarea is currently hovered
func (t *TextareaVNode) IsHovered() bool {
	return t.isHovered
}

// SetHovered sets the hover state
func (t *TextareaVNode) SetHovered(hovered bool) *TextareaVNode {
	t.isHovered = hovered
	return t
}

// SetBounds sets the textarea bounds for hit testing
func (t *TextareaVNode) SetBounds(x, y, width, height int) {
	t.bounds = [4]int{x, y, width, height}
}

// Bounds returns the textarea bounds
func (t *TextareaVNode) Bounds() [4]int {
	return t.bounds
}

// ContainsPoint checks if a point is within the textarea bounds
func (t *TextareaVNode) ContainsPoint(x, y int) bool {
	if t.bounds[2] <= 0 || t.bounds[3] <= 0 {
		return false
	}
	return x >= t.bounds[0] && x < t.bounds[0]+t.bounds[2] &&
		y >= t.bounds[1] && y < t.bounds[1]+t.bounds[3]
}

// HandleEvent processes mouse events for the textarea
func (t *TextareaVNode) HandleEvent(e event.Event) bool {
	if t.disabled {
		return false
	}

	mouseEvent, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}

	switch mouseEvent.Type() {
	case event.EventMouseEnter:
		if !t.isHovered {
			t.isHovered = true
		}
		return true

	case event.EventMouseLeave:
		if t.isHovered {
			t.isHovered = false
		}
		return true

	case event.EventMousePress, event.EventClick:
		if t.isHovered && mouseEvent.Button == event.MouseLeft {
			// Focus the textarea - the actual focus is managed by the framework
			// Return true to indicate the event was handled
			return true
		}
	}

	return false
}

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
// Calculates the size of the textarea based on content and constraints
// Per Ant Design spec: min-height=3, padding=1, line-gap=0
func (t *TextareaVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if t == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Calculate dimensions based on content per dimension spec
	content := t.value
	if content == "" && t.placeholder != "" {
		content = t.placeholder
	}

	// Split content into lines
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{" "}
	}

	// Calculate width based on longest line
	maxWidth := 0
	for _, line := range lines {
		lineWidth := utf8.RuneCountInString(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	// Width: content + padding per dimension spec (TextAreaPadding=1, so total +2)
	// + 2 for border on each side (total +4)
	width := maxWidth + 4 + (dimension.TextAreaPadding * 2)

	// Height: number of lines + 2 for top/bottom border
	height := len(lines) + 2

	// Apply minimum height per dimension spec
	if height < dimension.TextAreaMinHeight {
		height = dimension.TextAreaMinHeight
	}

	// Apply explicit rows/cols if set
	if t.rows > 0 && height < t.rows+2 {
		height = t.rows + 2
	}
	if t.cols > 0 && width < t.cols+4 {
		width = t.cols + 4
	}

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
	elemStyle := t.Style()
	if elemStyle.Width > 0 {
		width = elemStyle.Width
	}
	if elemStyle.Height > 0 {
		height = elemStyle.Height
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface
// Generates draw commands for rendering this textarea component
func (t *TextareaVNode) Paint(x, y int) []paint.DrawCmd {
	if t == nil {
		return nil
	}

	textareaStyle := t.Style()

	// Apply default colors based on component spec
	// Background: SURFACE, Text: TEXT
	if textareaStyle.FG == "" {
		textareaStyle = textareaStyle.Foreground(theme.Text())
	}
	if textareaStyle.BG == "" {
		textareaStyle = textareaStyle.Background(theme.Surface())
	}

	// Apply focus/hover/disabled styling first
	// Focus: FOCUS border for clear visibility
	// Disabled: DISABLED colors
	borderStyle := textareaStyle
	contentStyle := textareaStyle

	// Placeholder should use PLACEHOLDER color
	if t.value == "" && t.placeholder != "" {
		contentStyle = contentStyle.Foreground(theme.Placeholder())
	}

	if t.isFocused && !t.disabled {
		// Focused: FOCUS color borders
		borderStyle = borderStyle.Foreground(theme.Focus()).Bold(true)
	} else if t.disabled {
		// Disabled: DISABLED_BG, DISABLED_FG
		borderStyle = borderStyle.Foreground(theme.DisabledBG())
		contentStyle = contentStyle.Foreground(theme.DisabledFG())
	}

	// Determine what to display
	displayValue := t.value
	if displayValue == "" {
		displayValue = t.placeholder
	}

	// Split content into lines
	lines := strings.Split(displayValue, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	// Calculate dimensions
	measured := t.Measure(runtime.BoxConstraints{})
	width := measured.Width
	height := measured.Height

	// Build top border
	topBorder := "+" + strings.Repeat("-", width-2) + "+"

	// Build commands
	var cmds []paint.DrawCmd

	// Draw top border
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, borderStyle))

	// Draw each line with side borders
	for i, line := range lines {
		if i >= height-2 {
			break // Don't exceed height
		}

		// Pad line to fit width
		lineWidth := utf8.RuneCountInString(line)
		padding := width - 2 - lineWidth
		if padding < 0 {
			padding = 0
		}

		lineText := "|" + line + strings.Repeat(" ", padding) + "|"
		cmds = append(cmds, paint.NewTextCmd(x, y+1+i, lineText, borderStyle))
	}

	// Draw bottom border
	cmds = append(cmds, paint.NewTextCmd(x, y+height-1, topBorder, borderStyle))

	// CRITICAL: Set bounds for mouse hit testing
	t.SetBounds(x, y, width, height)

	return cmds
}

// =============================================================================
// TextareaBuilderType provides fluent API for building textareas
// =============================================================================

// TextareaBuilderType is the builder for Textarea
type TextareaBuilderType struct {
	node *TextareaVNode
}

// Value sets the initial value
func (b *TextareaBuilderType) Value(value string) *TextareaBuilderType {
	b.node.SetValue(value)
	return b
}

// Placeholder sets the placeholder text
func (b *TextareaBuilderType) Placeholder(text string) *TextareaBuilderType {
	b.node.SetPlaceholder(text)
	return b
}

// Rows sets the number of rows
func (b *TextareaBuilderType) Rows(rows int) *TextareaBuilderType {
	b.node.SetRows(rows)
	return b
}

// Cols sets the number of columns
func (b *TextareaBuilderType) Cols(cols int) *TextareaBuilderType {
	b.node.SetCols(cols)
	return b
}

// MaxLength sets the maximum length
func (b *TextareaBuilderType) MaxLength(len int) *TextareaBuilderType {
	b.node.SetMaxLength(len)
	return b
}

// Disabled sets the disabled state
func (b *TextareaBuilderType) Disabled(v bool) *TextareaBuilderType {
	b.node.SetDisabled(v)
	return b
}

// Resize sets the resizable state
func (b *TextareaBuilderType) Resize(v bool) *TextareaBuilderType {
	b.node.SetResize(v)
	return b
}

// OnChange sets the change handler
func (b *TextareaBuilderType) OnChange(fn func(string)) *TextareaBuilderType {
	b.node.SetOnChange(fn)
	return b
}

// OnFocus sets the focus handler
func (b *TextareaBuilderType) OnFocus(fn func()) *TextareaBuilderType {
	b.node.SetOnFocus(fn)
	return b
}

// OnBlur sets the blur handler
func (b *TextareaBuilderType) OnBlur(fn func()) *TextareaBuilderType {
	b.node.SetOnBlur(fn)
	return b
}

// OnSubmit sets the submit handler
func (b *TextareaBuilderType) OnSubmit(fn func()) *TextareaBuilderType {
	b.node.SetOnSubmit(fn)
	return b
}

// Key sets the key for diffing
func (b *TextareaBuilderType) Key(key string) *TextareaBuilderType {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *TextareaBuilderType) Style(s style.Style) *TextareaBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *TextareaBuilderType) FgColor(c interface{}) *TextareaBuilderType {
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
func (b *TextareaBuilderType) BgColor(c interface{}) *TextareaBuilderType {
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

// Width sets the width
func (b *TextareaBuilderType) Width(w int) *TextareaBuilderType {
	b.node.SetProp("width", w)
	return b
}

// Height sets the height
func (b *TextareaBuilderType) Height(h int) *TextareaBuilderType {
	b.node.SetProp("height", h)
	return b
}

// Build returns the ui.VNode
func (b *TextareaBuilderType) Build() ui.VNode {
	return b.node
}

// =============================================================================
// FocusableVNode Interface Implementation
// =============================================================================

// IsFocusable returns whether this textarea can receive focus.
// Disabled textareas cannot receive focus.
func (t *TextareaVNode) IsFocusable() bool {
	return !t.disabled
}

// GetFocusID returns a unique identifier for focus persistence.
// Uses the textarea's Key if set, otherwise generates a stable ID.
func (t *TextareaVNode) GetFocusID() string {
	if key := t.Key(); key != "" {
		return "textarea:" + key
	}
	// Generate stable ID based on placeholder or value
	id := t.placeholder
	if id == "" {
		id = t.value
	}
	if id == "" {
		id = "textarea"
	}
	return "textarea:" + id
}

// Label returns a label for this textarea for testing/debugging.
// Returns the placeholder if set, otherwise "textarea".
func (t *TextareaVNode) Label() string {
	if t.placeholder != "" {
		return t.placeholder
	}
	if t.value != "" {
		// Truncate long values for label
		if len(t.value) > 20 {
			return t.value[:20] + "..."
		}
		return t.value
	}
	return "textarea"
}
