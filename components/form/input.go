package form

import (
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/framework/cmd"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/runtime/dimension"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/framework/theme"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// Interface implementation assertions
var _ frameworkevent.Component = (*InputVNode)(nil)
var _ component.Updater = (*InputVNode)(nil) // Phase 3: Msg/Cmd support

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
	cursorPos   int // Cursor position for editing
	// Mouse interaction state
	isHovered bool
	// Bounds for hit testing (x, y, width, height)
	bounds [4]int
	// ActionTarget support
	supportedActions []action.ActionType // Supported action types
}

// NewInput creates a new input
func NewInput() *InputVNode {
	return &InputVNode{
		ElementVNode: ui.NewElement("input"),
		value:        "",
		placeholder: "",
		inputType:    InputTypeText,
		maxLength:    0, // 0 = no limit
		disabled:     false,
		readOnly:     false,
		isFocused:    false,
		cursorPos:    0, // Start at position 0
		supportedActions: []action.ActionType{
			action.ActionInputText,
			action.ActionBackspace,
			action.ActionDeleteChar,
			action.ActionEnter,
			action.ActionSubmit,
		},
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

// SetFocus sets the focused state (implements FocusableVNode)
func (i *InputVNode) SetFocus(focused bool) {
	i.isFocused = focused
	// Call focus handler if set
	if focused && i.onFocus != nil {
		i.onFocus()
	} else if !focused && i.onBlur != nil {
		i.onBlur()
	}
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
func (b *InputBuilderType) BgColor(c interface{}) *InputBuilderType {
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
func (b *InputBuilderType) Bold(v bool) *InputBuilderType {
	s := b.node.Style()
	s = s.Bold(v)
	b.node.SetStyle(s)
	return b
}

// Width sets the explicit width of the input
func (b *InputBuilderType) Width(w int) *InputBuilderType {
	s := b.node.Style()
	s.Width = w
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

// HandleEvent processes mouse and keyboard events for the input
func (i *InputVNode) HandleEvent(e frameworkevent.Event) bool {
	if i.disabled || i.readOnly {
		return false
	}

	// Handle keyboard events (character input)
	keyEvent, ok := e.(*frameworkevent.KeyEvent)
	if ok {
		// Only process keyboard input when focused
		if !i.isFocused {
			return false
		}

		// Handle special keys
		if keyEvent.Special == frameworkevent.KeyBackspace || keyEvent.Special == frameworkevent.KeyDelete {
			// Delete last character
			if len(i.value) > 0 {
				runes := []rune(i.value)
				i.value = string(runes[:len(runes)-1])
				if i.onChange != nil {
					i.onChange(i.value)
				}
			}
			return true
		}

		if keyEvent.Special == frameworkevent.KeyEnter {
			// Submit (if handler exists)
			if i.onSubmit != nil {
				i.onSubmit()
			}
			return true
		}

		// Handle character input
		if keyEvent.Key.Rune > 0 && keyEvent.Key.Rune >= 32 && keyEvent.Key.Rune <= 126 {
			// Check max length
			if i.maxLength > 0 && utf8.RuneCountInString(i.value) >= i.maxLength {
				return true // Reached max length
			}

			// Append character
			i.value += string(keyEvent.Key.Rune)
			if i.onChange != nil {
				i.onChange(i.value)
			}
			return true
		}

		return false
	}

	mouseEvent, ok := e.(*frameworkevent.MouseEvent)
	if !ok {
		return false
	}

	switch mouseEvent.Type() {
	case frameworkevent.EventMouseEnter:
		if !i.isHovered {
			i.isHovered = true
		}
		return true

	case frameworkevent.EventMouseLeave:
		if i.isHovered {
			i.isHovered = false
		}
		return true

	case frameworkevent.EventMousePress, frameworkevent.EventClick:
		if i.isHovered && mouseEvent.Button == frameworkevent.MouseLeft {
			// Focus the input - the actual focus is managed by the framework
			// Return true to indicate the event was handled
			return true
		}
	}

	return false
}

// =============================================================================
// Msg/Cmd Architecture Support (Phase 3)
// =============================================================================

// Update implements component.Updater interface for Msg/Cmd architecture
//
// Handles:
// - KeyMsg: Text input, deletion, cursor movement
// - MouseMsg: Focus handling
func (i *InputVNode) Update(message runtimemsg.Msg) cmd.Cmd {
	if i.disabled || i.readOnly {
		return nil
	}

	switch msg := message.(type) {
	case *runtimemsg.KeyMsg:
		// Only process keyboard input when focused
		if !i.isFocused {
			return nil
		}
		return i.updateKey(msg)

	case *runtimemsg.MouseMsg:
		return i.updateMouse(msg)
	}

	return nil
}

// updateKey handles keyboard messages for text input
func (i *InputVNode) updateKey(keyMsg *runtimemsg.KeyMsg) cmd.Cmd {
	// Handle special keys
	if keyMsg.Special == runtimeplatform.KeyBackspace || keyMsg.Special == runtimeplatform.KeyDelete {
		// Delete last character
		if len(i.value) > 0 {
			runes := []rune(i.value)
			i.value = string(runes[:len(runes)-1])
			if i.onChange != nil {
				i.onChange(i.value)
			}
		}
		return nil
	}

	if keyMsg.Special == runtimeplatform.KeyEnter {
		// Submit (if handler exists)
		if i.onSubmit != nil {
			i.onSubmit()
		}
		return nil
	}

	// Handle character input
	if keyMsg.Rune > 0 && keyMsg.Rune >= 32 && keyMsg.Rune <= 126 {
		// Check max length
		if i.maxLength > 0 && utf8.RuneCountInString(i.value) >= i.maxLength {
			return nil // Reached max length
		}

		// Append character
		i.value += string(keyMsg.Rune)
		if i.onChange != nil {
			i.onChange(i.value)
		}
		return nil
	}

	return nil
}

// updateMouse handles mouse messages for focus handling
func (i *InputVNode) updateMouse(mouseMsg *runtimemsg.MouseMsg) cmd.Cmd {
	switch mouseMsg.Action {
	case runtimemsg.MouseActionMove:
		// Update hover state
		// TODO: Need proper hover tracking in Pump
		i.isHovered = true
		return nil

	case runtimemsg.MouseActionPress:
		if mouseMsg.Button == runtimemsg.MouseLeft {
			// Focus the input - the actual focus is managed by the framework
			// Just mark as hovered, focus manager handles the rest
			i.isHovered = true
			return nil
		}
	}

	return nil
}

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
// Calculates the size of the input based on value/placeholder and constraints
// Per Ant Design spec: height=1, padding LR=1, min-width=10
func (i *InputVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if i == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Calculate content width per dimension spec
	content := i.value
	if content == "" && i.placeholder != "" {
		content = i.placeholder
	}
	if content == "" {
		content = " " // Empty input still has minimal width
	}

	// Width: content length + 2 for brackets ":"
	// + padding per dimension spec (InputPaddingLR=1, so total +2)
	width := utf8.RuneCountInString(content) + 2 + (dimension.InputPaddingLR * 2)
	height := dimension.InputHeight

	// Apply max length constraint
	if i.maxLength > 0 && width > i.maxLength+2 {
		width = i.maxLength + 2
	}

	// Apply minimum width per dimension spec
	if width < dimension.InputMinWidth {
		width = dimension.InputMinWidth
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
	elemStyle := i.Style()
	if elemStyle.Width > 0 {
		width = elemStyle.Width
	}
	if elemStyle.Height > 0 {
		height = elemStyle.Height
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface
// Generates draw commands for rendering this input component
func (i *InputVNode) Paint(x, y int) []paint.DrawCmd {
	if i == nil {
		return nil
	}

	inputStyle := i.Style()

	// Determine what to display
	displayValue := i.value
	if displayValue == "" {
		displayValue = i.placeholder
	}

	// Format input value based on type
	var displayText string
	switch i.inputType {
	case InputTypePassword:
		// Hide password characters
		maskedLen := utf8.RuneCountInString(i.value)
		if maskedLen == 0 && i.placeholder != "" {
			displayText = i.placeholder
		} else {
			displayText = ""
			for j := 0; j < maskedLen; j++ {
				displayText += "*"
			}
		}
	default:
		displayText = displayValue
	}

	// Build input display with brackets
	inputLabel := ":" + displayText + ":"

	// If explicit width is set, pad to fill the width
	// Width includes the brackets, so content area is Width - 2
	elemStyle := i.Style()
	if elemStyle.Width > 0 {
		targetContentWidth := elemStyle.Width - 2 // Account for brackets ":"
		if targetContentWidth > 0 {
			currentContentWidth := utf8.RuneCountInString(displayText)
			if currentContentWidth < targetContentWidth {
				// Pad with spaces to fill the width
				padding := targetContentWidth - currentContentWidth
				for k := 0; k < padding; k++ {
					inputLabel += " "
				}
			}
		}
	}

	// Apply focus/hover styling based on component spec
	// Background: SURFACE, Text: TEXT, Focus border: FOCUS
	if inputStyle.FG == "" {
		inputStyle = inputStyle.Foreground(theme.Text())
	}
	if inputStyle.BG == "" {
		inputStyle = inputStyle.Background(theme.Surface())
	}

	// Placeholder should use PLACEHOLDER color
	if i.value == "" && i.placeholder != "" {
		inputStyle = inputStyle.Foreground(theme.Placeholder())
	}

	// Focus state: FOCUS border (underline + bold)
	if i.isFocused {
		inputStyle = inputStyle.Foreground(theme.Focus()).Underline(true).Bold(true)
	} else if i.isHovered {
		inputStyle = inputStyle.Underline(true)
	}

	// Apply disabled state: DISABLED_BG, DISABLED_FG
	if i.disabled {
		inputStyle = inputStyle.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}

	// CRITICAL: Set bounds for mouse hit testing
	inputWidth := utf8.RuneCountInString(inputLabel)
	inputHeight := 1
	i.SetBounds(x, y, inputWidth, inputHeight)

	return []paint.DrawCmd{
		paint.NewTextCmd(x, y, inputLabel, inputStyle),
	}
}

// =============================================================================
// FocusableVNode Interface Implementation
// =============================================================================

// IsFocusable returns whether this input can receive focus.
// Disabled or read-only inputs cannot receive focus.
func (i *InputVNode) IsFocusable() bool {
	return !i.disabled && !i.readOnly
}

// GetFocusID returns a unique identifier for focus persistence.
// Uses the input's Key if set, otherwise generates a stable ID.
func (i *InputVNode) GetFocusID() string {
	if key := i.Key(); key != "" {
		return "input:" + key
	}
	// Generate stable ID based on placeholder or value
	id := i.placeholder
	if id == "" {
		id = i.value
	}
	if id == "" {
		id = "input"
	}
	return "input:" + id
}


// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction implements ActionTarget interface
func (i *InputVNode) HandleAction(act *action.Action) bool {
	if act == nil || i.disabled || i.readOnly {
		return false
	}

	// Handle action based on type
	switch act.Type {
	// Text input
	case action.ActionInputText:
		if text, ok := act.GetPayloadString(); ok {
			return i.InsertText(text)
		}
		return false

	// Deletion
	case action.ActionBackspace:
		return i.DeleteText(-1) // Backspace = delete backwards

	case action.ActionDeleteChar:
		return i.DeleteText(1) // Delete = delete forwards

	// Submission
	case action.ActionEnter, action.ActionSubmit:
		if i.onSubmit != nil {
			i.onSubmit()
			return true
		}
		return false
	}

	return false
}

// GetSupportedActions implements ActionTarget interface
func (i *InputVNode) GetSupportedActions() []action.ActionType {
	if i.supportedActions == nil {
		return []action.ActionType{
			action.ActionInputText,
			action.ActionBackspace,
			action.ActionDeleteChar,
			action.ActionEnter,
			action.ActionSubmit,
		}
	}
	return i.supportedActions
}

// CanHandleAction implements ActionTarget interface
func (i *InputVNode) CanHandleAction(act *action.Action) bool {
	if act == nil || i.disabled || i.readOnly {
		return false
	}

	// Check if action type is supported
	supported := i.GetSupportedActions()
	for _, supportedType := range supported {
		if supportedType == act.Type {
			return true
		}
	}

	return false
}

// ============================================================================
// FocusableActionTarget 接口实现
// ============================================================================

// Focus implements FocusableActionTarget interface
func (i *InputVNode) Focus() bool {
	if i.disabled || i.readOnly {
		return false
	}
	i.SetFocus(true)
	return true
}

// Blur implements FocusableActionTarget interface
func (i *InputVNode) Blur() {
	i.SetFocus(false)
}



// ============================================================================
// EditableActionTarget 接口实现
// ============================================================================

// InsertText inserts text at the current cursor position
func (i *InputVNode) InsertText(text string) bool {
	if i.disabled || i.readOnly {
		return false
	}

	// Check max length
	if i.maxLength > 0 {
		currentLen := utf8.RuneCountInString(i.value)
		textLen := utf8.RuneCountInString(text)
		if currentLen+textLen > i.maxLength {
			return false // Would exceed max length
		}
	}

	// Insert text at cursor position
	runes := []rune(i.value)
	before := runes[:i.cursorPos]
	after := runes[i.cursorPos:]
	textRunes := []rune(text)
	newRunes := append(append(before, textRunes...), after...)
	i.value = string(newRunes)
	i.cursorPos += len(textRunes)

	// Trigger onChange
	if i.onChange != nil {
		i.onChange(i.value)
	}

	return true
}

// DeleteText deletes text
// direction: -1 = backwards (Backspace), 1 = forwards (Delete)
func (i *InputVNode) DeleteText(direction int) bool {
	if i.disabled || i.readOnly {
		return false
	}

	runes := []rune(i.value)
	if len(runes) == 0 {
		return false // Nothing to delete
	}

	if direction < 0 {
		// Backspace: delete character before cursor
		if i.cursorPos == 0 {
			return false // Already at start
		}
		// Remove character at cursorPos-1
		newRunes := make([]rune, 0, len(runes)-1)
		newRunes = append(newRunes, runes[:i.cursorPos-1]...)
		newRunes = append(newRunes, runes[i.cursorPos:]...)
		i.value = string(newRunes)
		i.cursorPos--
	} else {
		// Delete: delete character at cursor
		if i.cursorPos >= len(runes) {
			return false // Already at end
		}
		// Remove character at cursorPos
		newRunes := make([]rune, 0, len(runes)-1)
		newRunes = append(newRunes, runes[:i.cursorPos]...)
		newRunes = append(newRunes, runes[i.cursorPos+1:]...)
		i.value = string(newRunes)
		// cursorPos stays the same
	}

	// Trigger onChange
	if i.onChange != nil {
		i.onChange(i.value)
	}

	return true
}

// ReplaceText replaces all text
func (i *InputVNode) ReplaceText(text string) bool {
	if i.disabled || i.readOnly {
		return false
	}

	// Check max length
	if i.maxLength > 0 && utf8.RuneCountInString(text) > i.maxLength {
		return false
	}

	i.value = text
	i.cursorPos = utf8.RuneCountInString(text)

	// Trigger onChange
	if i.onChange != nil {
		i.onChange(i.value)
	}

	return true
}

// GetText returns the current text content
func (i *InputVNode) GetText() string {
	return i.value
}

// GetCursorPosition returns the cursor position
func (i *InputVNode) GetCursorPosition() int {
	return i.cursorPos
}

// SetCursorPosition sets the cursor position
func (i *InputVNode) SetCursorPosition(pos int) bool {
	if pos < 0 || pos > utf8.RuneCountInString(i.value) {
		return false
	}
	i.cursorPos = pos
	return true
}
