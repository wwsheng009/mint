package input

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
	"github.com/wwsheng009/mint/ui/components/cursor"
	"github.com/wwsheng009/mint/ui/components/form"
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
	formID       string // Form ID for Form integration (Phase 6)
	maxLen       int
	cursorConfig cursor.Config

	// === Runtime State (managed by instance) ===
	state       control.InteractionState
	value       string
	cursorPos   int // cursor position for editing
	cursorModel *cursor.Model
	bounds      [4]int // x, y, w, h
	dirty       bool

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
	_ rtui.TickableInstance      = (*Instance)(nil)
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
	cursorCfg := getCursorConfigProp(props, "cursorConfig", cursor.DefaultConfig())
	inst := &Instance{
		key:          proputil.GetString(props, "key", ""),
		placeholder:  proputil.GetString(props, "placeholder", ""),
		inputType:    getTypeProp(props, TypeText),
		inputStyle:   proputil.GetStyle(props, "style", style.Style{}),
		width:        proputil.GetInt(props, "width", 0),
		borderStyle:  getBorderStyleProp(props, "borderStyle", layout.BorderSingle),
		changeIntent: proputil.GetIntent(props, "changeIntent", nil),
		submitIntent: proputil.GetIntent(props, "submitIntent", nil),
		formID:       proputil.GetString(props, "formID", ""),
		value:        proputil.GetString(props, "value", ""),
		maxLen:       proputil.GetInt(props, "maxLen", 0),
		cursorConfig: cursorCfg,
		cursorModel:  cursor.NewModel(cursorCfg),
		dirty:        true,
	}

	// Initialize cursor position at end of value
	inst.cursorPos = utf8.RuneCountInString(inst.value)

	// Initialize state
	inst.state = control.InteractionState{
		Disabled: proputil.GetBool(props, "disabled", false),
		Active:   proputil.GetBool(props, "readOnly", false), // Use Active for readOnly
	}

	// Initialize behaviors
	inst.initBehaviors()
	inst.syncCursorVisibility()

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

// Parent implements TreeComponent interface (intent bubble).
// Returns nil as Input is a leaf component without parent tracking.
func (inst *Instance) Parent() interface{} {
	return nil
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
	oldReadOnly := inst.state.Active
	oldPlaceholder := inst.placeholder
	oldCursorConfig := inst.cursorConfig

	inst.placeholder = proputil.GetString(props, "placeholder", inst.placeholder)
	inst.inputType = getTypeProp(props, inst.inputType)
	inst.inputStyle = proputil.GetStyle(props, "style", style.Style{})
	inst.width = proputil.GetInt(props, "width", inst.width)
	inst.borderStyle = getBorderStyleProp(props, "borderStyle", inst.borderStyle)
	inst.changeIntent = proputil.GetIntent(props, "changeIntent", nil)
	inst.submitIntent = proputil.GetIntent(props, "submitIntent", nil)
	inst.formID = proputil.GetString(props, "formID", inst.formID)

	// ✨ CRITICAL: When value changes, update cursorPos to prevent out-of-bounds
	// This fixes panic in InsertText when cursorPos > len(value)
	newValue := proputil.GetString(props, "value", inst.value)
	if newValue != inst.value {
		inst.value = newValue
		inst.cursorPos = utf8.RuneCountInString(inst.value)
		inst.cursorModel.ResetBlink()
	}
	inst.maxLen = proputil.GetInt(props, "maxLen", inst.maxLen)

	newDisabled := proputil.GetBool(props, "disabled", inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}
	newReadOnly := proputil.GetBool(props, "readOnly", inst.state.Active)
	if newReadOnly != inst.state.Active {
		inst.state.Active = newReadOnly
	}

	newCursorConfig := getCursorConfigProp(props, "cursorConfig", inst.cursorConfig)
	if newCursorConfig != inst.cursorConfig {
		inst.cursorConfig = newCursorConfig
		inst.cursorModel.SetConfig(newCursorConfig)
	}
	inst.syncCursorVisibility()

	// Check if props changed
	changed := oldValue != inst.value ||
		oldDisabled != inst.state.Disabled ||
		oldReadOnly != inst.state.Active ||
		oldCursorConfig != inst.cursorConfig ||
		oldPlaceholder != inst.placeholder

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":          inst.key,
		"placeholder":  inst.placeholder,
		"value":        inst.value,
		"disabled":     inst.state.Disabled,
		"readOnly":     inst.state.Active,
		"cursorConfig": inst.cursorConfig,
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

	// Keep raw text for cursor calculation, then clamp/pad for rendering.
	rawText := text
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

	// Draw caret overlay when focused.
	if inst.shouldDrawCursor() {
		cursorCol, cursorChar := inst.computeCursorOverlay(rawText, contentWidth)
		if cursorCol >= 0 && cursorCol < contentWidth {
			if cmd, ok := inst.cursorModel.DrawCmd(contentX+cursorCol, contentY, cursorChar, inputStyle); ok {
				cmds = append(cmds, cmd)
			}
		}
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
		inst.syncCursorVisibility()
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
// TickableInstance Interface
// =============================================================================

// WantsTick reports whether the input caret needs periodic blink updates.
func (inst *Instance) WantsTick() bool {
	return inst.cursorModel != nil && inst.cursorModel.WantsTick()
}

// Tick advances caret blink state.
func (inst *Instance) Tick(now time.Time) bool {
	if inst.cursorModel == nil {
		return false
	}
	if inst.cursorModel.Tick(now) {
		inst.dirty = true
		return true
	}
	return false
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

// HandleAction implements ActionHandlerInstance.
func (inst *Instance) HandleAction(act *action.Action) bool {
	if inst.state.Disabled || inst.state.Active {
		return false
	}

	switch act.Type {
	case action.ActionInputText:
		if text, ok := act.GetPayloadString(); ok {
			return inst.InsertText(text)
		}
		return false
	case action.ActionBackspace:
		return inst.DeleteText(-1)
	case action.ActionDeleteChar:
		return inst.DeleteText(1)
	case action.ActionCursorLeft:
		return inst.MoveCursor(-1)
	case action.ActionCursorRight:
		return inst.MoveCursor(1)
	case action.ActionCursorHome:
		return inst.MoveCursorToHome()
	case action.ActionCursorEnd:
		return inst.MoveCursorToEnd()
	case action.ActionSubmit, action.ActionEnter:
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

	// ✨ CRITICAL: Clamp cursorPos to valid range to prevent panic
	// This handles edge cases where cursorPos was set incorrectly
	maxPos := len(runes)
	if inst.cursorPos < 0 {
		inst.cursorPos = 0
	} else if inst.cursorPos > maxPos {
		inst.cursorPos = maxPos
	}

	// Enforce input width (content area width), allowing partial insert when needed.
	textRunes = inst.clampInsertRunesByWidth(runes, textRunes)
	if len(textRunes) == 0 {
		return false
	}

	// Create new slice with enough capacity to avoid shared array issues
	newRunes := make([]rune, 0, len(runes)+len(textRunes))
	newRunes = append(newRunes, runes[:inst.cursorPos]...)
	newRunes = append(newRunes, textRunes...)
	newRunes = append(newRunes, runes[inst.cursorPos:]...)

	inst.value = string(newRunes)
	inst.cursorPos += len(textRunes)
	inst.cursorModel.ResetBlink()
	inst.dirty = true

	// ✨ MVP/Phase 6: Emit FieldChangeIntent or FormFieldChangeIntent with runtime value
	// State becomes the single source of truth
	inst.emitFieldValueChanged()

	return true
}

func (inst *Instance) clampInsertRunesByWidth(current []rune, inserted []rune) []rune {
	maxContentWidth := inst.editableContentWidth()
	if maxContentWidth <= 0 || len(inserted) == 0 {
		return inserted
	}

	currentWidth := paint.StringWidth(string(current))
	remaining := maxContentWidth - currentWidth
	if remaining <= 0 {
		return nil
	}

	allowed := make([]rune, 0, len(inserted))
	used := 0
	for _, r := range inserted {
		rw := paint.RuneWidth(r)
		if used+rw > remaining {
			break
		}
		allowed = append(allowed, r)
		used += rw
	}
	return allowed
}

func (inst *Instance) editableContentWidth() int {
	_, _, boxWidth, _ := inst.GetBounds()
	if boxWidth > 0 {
		contentWidth := boxWidth - 2 // left+right border/bracket
		if contentWidth < 1 {
			return 1
		}
		return contentWidth
	}
	if inst.width > 0 {
		return inst.width
	}
	return 0
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

	// ✨ CRITICAL: Clamp cursorPos to valid range to prevent panic
	maxPos := len(runes)
	if inst.cursorPos < 0 {
		inst.cursorPos = 0
	} else if inst.cursorPos > maxPos {
		inst.cursorPos = maxPos
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

	inst.cursorModel.ResetBlink()
	inst.dirty = true

	// ✨ MVP/Phase 6: Emit FieldChangeIntent or FormFieldChangeIntent with runtime value
	// State becomes the single source of truth
	inst.emitFieldValueChanged()

	return true
}

// SetValue sets the entire value.
func (inst *Instance) SetValue(value string) {
	if inst.value != value {
		inst.value = value
		inst.cursorPos = utf8.RuneCountInString(value)
		inst.cursorModel.ResetBlink()
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
	if inst.cursorPos != pos {
		inst.cursorPos = pos
		inst.cursorModel.ResetBlink()
		inst.dirty = true
	}
}

// MoveCursor moves the cursor by delta runes.
func (inst *Instance) MoveCursor(delta int) bool {
	oldPos := inst.cursorPos
	inst.SetCursorPos(inst.cursorPos + delta)
	return inst.cursorPos != oldPos
}

// MoveCursorToHome moves the cursor to the beginning.
func (inst *Instance) MoveCursorToHome() bool {
	oldPos := inst.cursorPos
	inst.SetCursorPos(0)
	return inst.cursorPos != oldPos
}

// MoveCursorToEnd moves the cursor to the end.
func (inst *Instance) MoveCursorToEnd() bool {
	oldPos := inst.cursorPos
	inst.SetCursorPos(utf8.RuneCountInString(inst.value))
	return inst.cursorPos != oldPos
}

func (inst *Instance) shouldDrawCursor() bool {
	return inst.cursorModel != nil && inst.cursorModel.ShouldPaint()
}

func (inst *Instance) wantsCursorVisible() bool {
	return inst.state.Focused && !inst.state.Disabled && !inst.state.Active
}

func (inst *Instance) syncCursorVisibility() bool {
	if inst.cursorModel == nil {
		return false
	}
	changed := inst.cursorModel.SetVisible(inst.wantsCursorVisible())
	if changed {
		inst.dirty = true
	}
	return changed
}

// computeCursorOverlay calculates the caret column and displayed glyph.
func (inst *Instance) computeCursorOverlay(rawText string, contentWidth int) (int, string) {
	if contentWidth <= 0 {
		return -1, ""
	}

	// Placeholder is visual hint; caret follows editable value.
	source := rawText
	if inst.value == "" {
		source = ""
	}

	runes := []rune(source)
	valueLen := utf8.RuneCountInString(inst.value)
	cursor := inst.cursorPos
	if cursor < 0 {
		cursor = 0
	}
	if cursor > valueLen {
		cursor = valueLen
	}

	cursorCol := displayWidthByRuneIndex(runes, cursor)
	if cursorCol >= contentWidth {
		cursorCol = contentWidth - 1
	}

	// Draw current rune under caret, or a space if caret is at end.
	if cursor < len(runes) {
		return cursorCol, string(runes[cursor])
	}
	return cursorCol, " "
}

func displayWidthByRuneIndex(runes []rune, runeIndex int) int {
	if runeIndex < 0 {
		return 0
	}
	if runeIndex > len(runes) {
		runeIndex = len(runes)
	}
	width := 0
	for i := 0; i < runeIndex; i++ {
		width += paint.RuneWidth(runes[i])
	}
	return width
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
	inst.syncCursorVisibility()
	if oldState != inst.state {
		inst.dirty = true
	}
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
			inst.syncCursorVisibility()
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
// Form Integration Methods (Phase 6)
// =============================================================================

// emitFieldValueChanged emits FieldChangeIntent or FormFieldChangeIntent
// depending on whether formID is set.
func (inst *Instance) emitFieldValueChanged() {
	if inst.intentEmitter == nil {
		return
	}

	// Phase 6: If formID is set, use FormFieldChangeIntent
	if inst.formID != "" {
		if fieldIntent, ok := inst.changeIntent.(intent.FieldIntent); ok {
			formIntent := form.FieldChange(
				inst.formID,
				fieldIntent.GetField(),
				inst.value,
				true, // isDirty
			)
			intent.Emit(inst, formIntent)
		}
		return
	}

	// Original MVP behavior: emit FieldChangeIntent
	if fieldIntent, ok := inst.changeIntent.(intent.FieldIntent); ok {
		changeIntent := intent.FieldChangeIntent{
			Field: fieldIntent.GetField(),
			Value: inst.value,
		}
		inst.intentEmitter(changeIntent)
	} else if inst.changeIntent != nil {
		// Fallback: emit the original intent for backward compatibility
		inst.intentEmitter(inst.changeIntent)
	}
}

// emitFieldBlur emits FormFieldBlurIntent to trigger validation (Phase 6)
// Only called when formID is set.
func (inst *Instance) emitFieldBlur() {
	if inst.intentEmitter == nil || inst.formID == "" {
		return
	}

	if fieldIntent, ok := inst.changeIntent.(intent.FieldIntent); ok {
		blurIntent := form.FieldBlur(
			inst.formID,
			fieldIntent.GetField(),
			inst.value,
		)
		intent.Emit(inst, blurIntent)
	}
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getTypeProp(props rtui.Props, def Type) Type {
	if v, ok := props["inputType"]; ok {
		if t, ok := v.(Type); ok {
			return t
		}
	}
	return def
}

func getBorderStyleProp(props rtui.Props, key string, def layout.BorderStyle) layout.BorderStyle {
	if v, ok := props[key]; ok {
		if s, ok := v.(layout.BorderStyle); ok {
			return s
		}
	}
	return def
}

func getCursorConfigProp(props rtui.Props, key string, def cursor.Config) cursor.Config {
	if v, ok := props[key]; ok {
		if cfg, ok := v.(cursor.Config); ok {
			return cursor.NormalizeConfig(cfg)
		}
	}
	return cursor.NormalizeConfig(def)
}
