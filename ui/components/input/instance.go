package input

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Input components.
// It persists across renders and holds all state including cursor position.
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	placeholder  string
	inputType    Type
	inputStyle   style.Style
	width        int
	borderStyle  layout.BorderStyle
	changeIntent intent.Intent
	submitIntent intent.Intent
	maxLen       int

	// === Runtime State (managed by instance) ===
	state     control.InteractionState
	value     string
	cursorPos int    // cursor position for editing
	bounds    [4]int // x, y, w, h
	dirty     bool

	// === Intent Emitter ===
	intentEmitter func(intent.Intent)

	// === Behaviors ===
	behaviors *control.BehaviorList
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ control.Instance           = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new InputInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          getStringProp(props, "key", ""),
		placeholder:  getStringProp(props, "placeholder", ""),
		inputType:    getTypeProp(props, TypeText),
		inputStyle:   getStyleProp(props),
		width:        getIntProp(props, "width", 0),
		borderStyle:  getBorderStyleProp(props, "borderStyle", layout.BorderSingle),
		changeIntent: getIntentProp(props, "changeIntent"),
		submitIntent: getIntentProp(props, "submitIntent"),
		value:        getStringProp(props, "value", ""),
		maxLen:       getIntProp(props, "maxLen", 0),
		dirty:        true,
	}

	// Initialize cursor position at end of value
	inst.cursorPos = utf8.RuneCountInString(inst.value)

	// Initialize state
	inst.state = control.InteractionState{
		Disabled: getBoolProp(props, "disabled", false),
		Active:   getBoolProp(props, "readOnly", false), // Use Active for readOnly
	}

	// Initialize behaviors
	inst.initBehaviors()

	return inst
}

// initBehaviors initializes the behavior composition.
func (inst *Instance) initBehaviors() {
	inst.behaviors = control.NewBehaviorList(
		&control.FocusableBehavior{},
		&control.HoverableBehavior{},
		&control.DisableableBehavior{},
	)
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

// Key implements ComponentInstance.
func (inst *Instance) Key() string {
	return inst.key
}

// SetKey implements ComponentInstance.
func (inst *Instance) SetKey(key string) {
	inst.key = key
}

// Init implements ComponentInstance.
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}

// Destroy implements ComponentInstance.
func (inst *Instance) Destroy() {
	inst.behaviors.OnUnmount(inst)
}

// OnMount implements ComponentInstance.
func (inst *Instance) OnMount() {
	inst.behaviors.OnMount(inst)
}

// OnUnmount implements ComponentInstance.
func (inst *Instance) OnUnmount() {
	inst.behaviors.OnUnmount(inst)
}

// SetProps implements ComponentInstance.
func (inst *Instance) SetProps(props rtui.Props) bool {
	oldValue := inst.value
	oldDisabled := inst.state.Disabled
	oldPlaceholder := inst.placeholder

	inst.placeholder = getStringProp(props, "placeholder", inst.placeholder)
	inst.inputType = getTypeProp(props, inst.inputType)
	inst.inputStyle = getStyleProp(props)
	inst.width = getIntProp(props, "width", inst.width)
	inst.borderStyle = getBorderStyleProp(props, "borderStyle", inst.borderStyle)
	inst.changeIntent = getIntentProp(props, "changeIntent")
	inst.submitIntent = getIntentProp(props, "submitIntent")
	inst.value = getStringProp(props, "value", inst.value)
	inst.maxLen = getIntProp(props, "maxLen", inst.maxLen)

	newDisabled := getBoolProp(props, "disabled", inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}

	// Check if props changed
	changed := oldValue != inst.value ||
		oldDisabled != inst.state.Disabled ||
		oldPlaceholder != inst.placeholder

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":         inst.key,
		"placeholder": inst.placeholder,
		"value":       inst.value,
		"disabled":    inst.state.Disabled,
	}
}

// MarkDirty implements ComponentInstance.
func (inst *Instance) MarkDirty() {
	inst.dirty = true
}

// IsDirty implements ComponentInstance.
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}

// GetContext implements ComponentInstance (no hooks for Input).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd

	// Get actual content size from bounds, or calculate from props
	_, _, boxWidth, boxHeight := inst.GetBounds()
	if boxWidth == 0 {
		// Calculate size from props if bounds not set
		boxWidth = 12 // Default minimum with border
		if inst.width > 0 {
			boxWidth = inst.width + 2 // Add bracket/border padding
		}
	}

	// Determine if we need brackets (BorderNone still uses [ ] brackets)
	useBrackets := inst.borderStyle == layout.BorderNone
	useFullBorder := inst.borderStyle != layout.BorderNone

	// Set appropriate height based on border style
	if boxHeight == 0 {
		if useFullBorder {
			boxHeight = 3 // 1 content + 2 border lines
		} else {
			boxHeight = 1 // Just content line for bracket style
		}
	}

	// Calculate border offset
	borderWidth := 0
	if useFullBorder {
		borderWidth = 1
	} else if useBrackets {
		borderWidth = 1 // Brackets also take 1 char on each side
	}

	// Draw full border if present
	if useFullBorder {
		cmds = append(cmds, inst.paintBorder(x, y, boxWidth, boxHeight)...)
	}

	// Calculate content position (inside border/brackets)
	contentX := x + borderWidth
	contentY := y + borderWidth
	contentWidth := boxWidth - borderWidth*2
	if contentWidth < 1 {
		contentWidth = 1
	}

	// Determine what to display
	displayValue := inst.value
	if displayValue == "" {
		displayValue = inst.placeholder
	}

	// Format based on type
	var text string
	switch inst.inputType {
	case TypePassword:
		if inst.value == "" {
			text = inst.placeholder
		} else {
			text = strings.Repeat("*", utf8.RuneCountInString(inst.value))
		}
	default:
		text = displayValue
	}

	// Apply width constraint
	if contentWidth > 0 {
		text = inst.padText(text, contentWidth)
	}

	// Resolve style
	inputStyle := inst.resolveStyle()

	// Draw with brackets if BorderNone
	if useBrackets {
		// Draw [text] format
		bracketStyle := inst.resolveBorderColor()
		cmds = append(cmds,
			paint.DrawCmd{X: x, Y: contentY, Text: "[", Style: style.Style{FG: bracketStyle}},
			paint.DrawCmd{X: x + 1, Y: contentY, Text: text, Style: inputStyle},
			paint.DrawCmd{X: x + 1 + contentWidth, Y: contentY, Text: "]", Style: style.Style{FG: bracketStyle}},
		)
	} else {
		// Add text draw command (for full border mode)
		cmds = append(cmds, paint.DrawCmd{
			X:     contentX,
			Y:     contentY,
			Text:  text,
			Style: inputStyle,
		})
	}

	return cmds
}

// paintBorder draws the border around the input field.
func (inst *Instance) paintBorder(x, y, width, height int) []paint.DrawCmd {
	if inst.borderStyle == layout.BorderNone || width < 2 || height < 2 {
		return nil
	}

	var cmds []paint.DrawCmd
	borderColor := inst.resolveBorderColor()

	// Get border characters based on state
	cornerTL, cornerTR, cornerBL, cornerBR, horizontal, vertical := inst.getBorderChars()

	// Calculate content area (total size minus border)
	contentWidth := width - 2
	contentHeight := height - 2

	// Top border
	topLine := string(cornerTL) + strings.Repeat(string(horizontal), contentWidth) + string(cornerTR)
	cmds = append(cmds, paint.DrawCmd{
		X:     x,
		Y:     y,
		Text:  topLine,
		Style: style.Style{FG: borderColor},
	})

	// Side borders for content rows
	for i := 0; i < contentHeight; i++ {
		cmds = append(cmds,
			paint.DrawCmd{
				X:     x,
				Y:     y + 1 + i,
				Text:  string(vertical),
				Style: style.Style{FG: borderColor},
			},
			paint.DrawCmd{
				X:     x + width - 1,
				Y:     y + 1 + i,
				Text:  string(vertical),
				Style: style.Style{FG: borderColor},
			},
		)
	}

	// Bottom border
	bottomLine := string(cornerBL) + strings.Repeat(string(horizontal), contentWidth) + string(cornerBR)
	cmds = append(cmds, paint.DrawCmd{
		X:     x,
		Y:     y + height - 1,
		Text:  bottomLine,
		Style: style.Style{FG: borderColor},
	})

	return cmds
}

// getBorderChars returns border characters based on style and state.
func (inst *Instance) getBorderChars() (cornerTL, cornerTR, cornerBL, cornerBR, horizontal, vertical rune) {
	switch inst.borderStyle {
	case layout.BorderDouble:
		return '╔', '╗', '╚', '╝', '═', '║'
	case layout.BorderRounded:
		return '╭', '╮', '╰', '╯', '─', '│'
	default: // BorderSingle
		return '┌', '┐', '└', '┘', '─', '│'
	}
}

// resolveBorderColor returns the border color based on state.
func (inst *Instance) resolveBorderColor() style.Color {
	// Priority: Disabled > Focused > Hovered > Normal
	if inst.state.Disabled {
		return theme.DisabledFG()
	}
	if inst.state.Focused {
		return theme.Focus()
	}
	if inst.state.Hovered {
		return theme.Select()
	}
	return theme.Border()
}

// padText pads or truncates text to fit the specified display width.
// Correctly handles wide characters like Chinese (width=2).
func (inst *Instance) padText(text string, width int) string {
	textWidth := paint.StringWidth(text)
	if textWidth > width {
		// Truncate by display width
		return truncateByDisplayWidth(text, width)
	}
	if textWidth < width {
		return text + strings.Repeat(" ", width-textWidth)
	}
	return text
}

// truncateByDisplayWidth truncates text to fit within max display width.
func truncateByDisplayWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(text)
	var result []rune
	currentWidth := 0

	for _, r := range runes {
		runeWidth := paint.RuneWidth(r)
		if currentWidth+runeWidth > maxWidth {
			break
		}
		result = append(result, r)
		currentWidth += runeWidth
	}

	return string(result)
}

// resolveStyle resolves the visual style based on state.
func (inst *Instance) resolveStyle() style.Style {
	s := inst.inputStyle

	// Apply default colors
	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}

	// Placeholder styling
	if inst.value == "" && inst.placeholder != "" {
		s = s.Foreground(theme.Placeholder())
	}

	// State priority: Disabled > Focused > Hovered > Normal
	if inst.state.Disabled {
		s = s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	} else if inst.state.Focused {
		s = s.Foreground(theme.Focus()).Underline(true).Bold(true)
	} else if inst.state.Hovered {
		s = s.Underline(true)
	}

	return s
}

// =============================================================================
// FocusableInstance Interface
// =============================================================================

// SetFocus implements FocusableInstance.
func (inst *Instance) SetFocus(focused bool) {
	if inst.state.Focused != focused {
		oldState := inst.state
		inst.state.Focused = focused
		inst.dirty = true
		inst.behaviors.OnStateChange(inst, oldState, inst.state)
	}
}

// HasFocus implements FocusableInstance.
func (inst *Instance) HasFocus() bool {
	return inst.state.Focused
}

// IsDisabled implements FocusableInstance.
func (inst *Instance) IsDisabled() bool {
	return inst.state.Disabled
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

// CanHandleAction implements ActionHandlerInstance.
func (inst *Instance) CanHandleAction(actionType string) bool {
	if inst.state.Disabled || inst.state.Active { // Active = readOnly
		return false
	}
	return actionType == "input" || actionType == "backspace" ||
		actionType == "delete" || actionType == "submit" || actionType == "enter"
}

// HandleAction implements ActionHandlerInstance.
func (inst *Instance) HandleAction(actionType string, payload interface{}) bool {
	if inst.state.Disabled || inst.state.Active {
		return false
	}

	switch actionType {
	case "input":
		if text, ok := payload.(string); ok {
			return inst.InsertText(text)
		}
	case "backspace":
		return inst.DeleteText(-1)
	case "delete":
		return inst.DeleteText(1)
	case "submit", "enter":
		if inst.submitIntent != nil && inst.intentEmitter != nil {
			inst.intentEmitter(inst.submitIntent)
		}
		return true
	}
	return false
}

// =============================================================================
// Input-specific Methods
// =============================================================================

// InsertText inserts text at the current cursor position.
func (inst *Instance) InsertText(text string) bool {
	if inst.state.Disabled || inst.state.Active {
		return false
	}

	// Check max length
	if inst.maxLen > 0 {
		currentLen := utf8.RuneCountInString(inst.value)
		textLen := utf8.RuneCountInString(text)
		if currentLen+textLen > inst.maxLen {
			return false
		}
	}

	// Insert at cursor
	runes := []rune(inst.value)
	textRunes := []rune(text)

	// Create new slice with enough capacity to avoid shared array issues
	newRunes := make([]rune, 0, len(runes)+len(textRunes))
	newRunes = append(newRunes, runes[:inst.cursorPos]...)
	newRunes = append(newRunes, textRunes...)
	newRunes = append(newRunes, runes[inst.cursorPos:]...)

	inst.value = string(newRunes)
	inst.cursorPos += len(textRunes)
	inst.dirty = true

	// Emit change intent
	if inst.changeIntent != nil && inst.intentEmitter != nil {
		inst.intentEmitter(inst.changeIntent)
	}

	return true
}

// DeleteText deletes text relative to cursor.
func (inst *Instance) DeleteText(direction int) bool {
	if inst.state.Disabled || inst.state.Active {
		return false
	}

	runes := []rune(inst.value)
	if len(runes) == 0 {
		return false
	}

	if direction < 0 {
		// Backspace: delete before cursor
		if inst.cursorPos == 0 {
			return false
		}
		newRunes := make([]rune, 0, len(runes)-1)
		newRunes = append(newRunes, runes[:inst.cursorPos-1]...)
		newRunes = append(newRunes, runes[inst.cursorPos:]...)
		inst.value = string(newRunes)
		inst.cursorPos--
	} else {
		// Delete: delete at cursor
		if inst.cursorPos >= len(runes) {
			return false
		}
		newRunes := make([]rune, 0, len(runes)-1)
		newRunes = append(newRunes, runes[:inst.cursorPos]...)
		newRunes = append(newRunes, runes[inst.cursorPos+1:]...)
		inst.value = string(newRunes)
	}

	inst.dirty = true

	// Emit change intent
	if inst.changeIntent != nil && inst.intentEmitter != nil {
		inst.intentEmitter(inst.changeIntent)
	}

	return true
}

// SetValue sets the entire value.
func (inst *Instance) SetValue(value string) {
	if inst.value != value {
		inst.value = value
		inst.cursorPos = utf8.RuneCountInString(value)
		inst.dirty = true
	}
}

// GetValue returns the current value.
func (inst *Instance) GetValue() string {
	return inst.value
}

// CursorPos returns the cursor position.
func (inst *Instance) CursorPos() int {
	return inst.cursorPos
}

// SetCursorPos sets the cursor position.
func (inst *Instance) SetCursorPos(pos int) {
	max := utf8.RuneCountInString(inst.value)
	if pos < 0 {
		pos = 0
	} else if pos > max {
		pos = max
	}
	inst.cursorPos = pos
}

// =============================================================================
// control.Instance Interface (for Behaviors)
// =============================================================================

// GetState returns the interaction state.
func (inst *Instance) GetState() *control.InteractionState {
	return &inst.state
}

// SetState sets the interaction state.
func (inst *Instance) SetState(state control.InteractionState) {
	oldState := inst.state
	inst.state = state
	inst.behaviors.OnStateChange(inst, oldState, inst.state)
}

// EmitIntent emits an intent.
func (inst *Instance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil {
		inst.intentEmitter(i)
	}
}

// GetBounds returns the layout bounds.
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

// SetBounds sets the layout bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// GetStyle returns the visual style.
func (inst *Instance) GetStyle() style.Style {
	return inst.inputStyle
}

// SetStyle sets the visual style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.inputStyle = s
}

// GetProp returns a prop value.
func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case "disabled":
		return inst.state.Disabled, true
	case "value":
		return inst.value, true
	case "placeholder":
		return inst.placeholder, true
	case "inputType":
		return inst.inputType, true
	case "changeIntent":
		return inst.changeIntent, true
	case "submitIntent":
		return inst.submitIntent, true
	default:
		return nil, false
	}
}

// SetProp sets a prop value.
func (inst *Instance) SetProp(key string, value interface{}) {
	switch key {
	case "disabled":
		if v, ok := value.(bool); ok {
			inst.state.Disabled = v
			inst.dirty = true
		}
	case "value":
		if v, ok := value.(string); ok {
			inst.SetValue(v)
		}
	}
}

// SetIntentEmitter sets the intent emitter function.
func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

// ClearDirty clears the dirty flag.
func (inst *Instance) ClearDirty() {
	inst.dirty = false
}

// GetBorder returns the border configuration for the layout engine.
func (inst *Instance) GetBorder() layout.Border {
	if inst.borderStyle == layout.BorderNone {
		return layout.Border{Style: layout.BorderNone}
	}
	return layout.Border{
		Style: inst.borderStyle,
		Width: 1,
	}
}

// =============================================================================
// Measurable Interface (Two-Pass Layout)
// =============================================================================

// Measure implements layout.Measurable interface.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	// Calculate padding: full border uses 2, brackets also use 2
	padding := 2 // Always 2 for input (brackets or full border)

	// Calculate content width
	content := inst.value
	if content == "" {
		content = inst.placeholder
	}
	if content == "" {
		content = " "
	}

	contentWidth := paint.StringWidth(content)
	contentHeight := 1

	// Apply explicit width if set
	width := contentWidth
	if inst.width > 0 {
		width = inst.width
	}

	// Apply max length constraint
	if inst.maxLen > 0 && width > inst.maxLen {
		width = inst.maxLen
	}

	// Apply minimum width (for content only)
	if width < 10 {
		width = 10
	}

	// Add padding to get total size
	totalWidth := width + padding

	// Height: brackets use 1 line, full border uses 3 lines
	totalHeight := contentHeight
	if inst.borderStyle != layout.BorderNone {
		totalHeight = contentHeight + 2 // Add top and bottom border lines
	}

	// Apply constraints
	totalWidth = constraints.ConstrainWidth(totalWidth)
	totalHeight = constraints.ConstrainHeight(totalHeight)

	// Apply explicit style dimensions if set
	if inst.inputStyle.Width > 0 {
		totalWidth = constraints.ConstrainWidth(inst.inputStyle.Width)
	}
	if inst.inputStyle.Height > 0 {
		totalHeight = constraints.ConstrainHeight(inst.inputStyle.Height)
	}

	return layout.Size{Width: totalWidth, Height: totalHeight}
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getStringProp(props rtui.Props, key, def string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if v, ok := props[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if v, ok := props[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getTypeProp(props rtui.Props, def Type) Type {
	if v, ok := props["inputType"]; ok {
		if t, ok := v.(Type); ok {
			return t
		}
	}
	return def
}

func getStyleProp(props rtui.Props) style.Style {
	if v, ok := props["style"]; ok {
		if s, ok := v.(style.Style); ok {
			return s
		}
	}
	return style.Style{}
}

func getIntentProp(props rtui.Props, key string) intent.Intent {
	if v, ok := props[key]; ok {
		if i, ok := v.(intent.Intent); ok {
			return i
		}
	}
	return nil
}

func getBorderStyleProp(props rtui.Props, key string, def layout.BorderStyle) layout.BorderStyle {
	if v, ok := props[key]; ok {
		if s, ok := v.(layout.BorderStyle); ok {
			return s
		}
	}
	return def
}
