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

const (
	searchVariantPrefix = "/ "
	addonGapWidth       = 1
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
	placeholder   string
	inputType     Type
	prefix        string
	suffix        string
	addonBefore   string
	addonAfter    string
	inputStyle    style.Style
	width         int
	borderStyle   layout.BorderStyle
	allowNegative bool
	allowDecimal  bool
	changeIntent  intent.Intent
	submitIntent  intent.Intent
	formID        string // Form ID for Form integration (Phase 6)
	maxLen        int
	searchVariant bool
	cursorConfig  cursor.Config

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
	inputType := getTypeProp(props, TypeText)
	allowNegative := proputil.GetBool(props, propAllowNegative, true)
	allowDecimal := proputil.GetBool(props, propAllowDecimal, true)
	value := proputil.GetString(props, "value", "")
	if inputType == TypeNumber {
		value = sanitizeEditableNumberValue(value, allowNegative, allowDecimal)
	}

	inst := &Instance{
		key:           proputil.GetString(props, "key", ""),
		placeholder:   proputil.GetString(props, "placeholder", ""),
		inputType:     inputType,
		prefix:        proputil.GetString(props, propPrefix, ""),
		suffix:        proputil.GetString(props, propSuffix, ""),
		addonBefore:   proputil.GetString(props, propAddonBefore, ""),
		addonAfter:    proputil.GetString(props, propAddonAfter, ""),
		inputStyle:    proputil.GetStyle(props, "style", style.Style{}),
		width:         proputil.GetInt(props, "width", 0),
		borderStyle:   getBorderStyleProp(props, "borderStyle", layout.BorderSingle),
		allowNegative: allowNegative,
		allowDecimal:  allowDecimal,
		changeIntent:  proputil.GetIntent(props, "changeIntent", nil),
		submitIntent:  proputil.GetIntent(props, "submitIntent", nil),
		formID:        proputil.GetString(props, "formID", ""),
		value:         value,
		maxLen:        proputil.GetInt(props, "maxLen", 0),
		searchVariant: proputil.GetBool(props, propSearchVariant, false),
		cursorConfig:  cursorCfg,
		cursorModel:   cursor.NewModel(cursorCfg),
		dirty:         true,
	}

	// Initialize cursor position at end of value
	inst.cursorPos = utf8.RuneCountInString(inst.value)
	if inst.searchVariant && inst.inputType == TypePassword {
		inst.inputType = TypeText
	}

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
	oldPrefix := inst.prefix
	oldSuffix := inst.suffix
	oldAddonBefore := inst.addonBefore
	oldAddonAfter := inst.addonAfter
	oldSearchVariant := inst.searchVariant
	oldCursorConfig := inst.cursorConfig
	oldInputType := inst.inputType
	oldAllowNegative := inst.allowNegative
	oldAllowDecimal := inst.allowDecimal

	inst.placeholder = proputil.GetString(props, "placeholder", inst.placeholder)
	inst.inputType = getTypeProp(props, inst.inputType)
	inst.prefix = proputil.GetString(props, propPrefix, inst.prefix)
	inst.suffix = proputil.GetString(props, propSuffix, inst.suffix)
	inst.addonBefore = proputil.GetString(props, propAddonBefore, inst.addonBefore)
	inst.addonAfter = proputil.GetString(props, propAddonAfter, inst.addonAfter)
	inst.inputStyle = proputil.GetStyle(props, "style", style.Style{})
	inst.width = proputil.GetInt(props, "width", inst.width)
	inst.borderStyle = getBorderStyleProp(props, "borderStyle", inst.borderStyle)
	inst.allowNegative = proputil.GetBool(props, propAllowNegative, inst.allowNegative)
	inst.allowDecimal = proputil.GetBool(props, propAllowDecimal, inst.allowDecimal)
	inst.changeIntent = proputil.GetIntent(props, "changeIntent", nil)
	inst.submitIntent = proputil.GetIntent(props, "submitIntent", nil)
	inst.formID = proputil.GetString(props, "formID", inst.formID)

	// ✨ CRITICAL: When value changes, update cursorPos to prevent out-of-bounds
	// This fixes panic in InsertText when cursorPos > len(value)
	newValue := proputil.GetString(props, "value", inst.value)
	if inst.inputType == TypeNumber {
		newValue = sanitizeEditableNumberValue(newValue, inst.allowNegative, inst.allowDecimal)
	}
	if newValue != inst.value {
		inst.value = newValue
		inst.cursorPos = utf8.RuneCountInString(inst.value)
		inst.cursorModel.ResetBlink()
	}
	inst.maxLen = proputil.GetInt(props, "maxLen", inst.maxLen)
	inst.searchVariant = proputil.GetBool(props, propSearchVariant, inst.searchVariant)
	if inst.searchVariant && inst.inputType == TypePassword {
		inst.inputType = TypeText
	}

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
		oldInputType != inst.inputType ||
		oldAllowNegative != inst.allowNegative ||
		oldAllowDecimal != inst.allowDecimal ||
		oldPlaceholder != inst.placeholder ||
		oldPrefix != inst.prefix ||
		oldSuffix != inst.suffix ||
		oldAddonBefore != inst.addonBefore ||
		oldAddonAfter != inst.addonAfter ||
		oldSearchVariant != inst.searchVariant

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:           inst.key,
		propPlaceholder:   inst.placeholder,
		propInputType:     inst.inputType,
		propPrefix:        inst.prefix,
		propSuffix:        inst.suffix,
		propAddonBefore:   inst.addonBefore,
		propAddonAfter:    inst.addonAfter,
		propAllowNegative: inst.allowNegative,
		propAllowDecimal:  inst.allowDecimal,
		propValue:         inst.value,
		propDisabled:      inst.state.Disabled,
		propReadOnly:      inst.state.Active,
		propSearchVariant: inst.searchVariant,
		propCursorConfig:  inst.cursorConfig,
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
		boxWidth = inst.defaultBoxWidth()
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

	addonBeforeWidth := paint.StringWidth(inst.addonBefore)
	addonAfterWidth := paint.StringWidth(inst.addonAfter)
	borderX := x
	if addonBeforeWidth > 0 {
		borderX += addonBeforeWidth + addonGapWidth
	}
	addonY := y
	if useFullBorder {
		addonY = y + 1
	}

	if addonBeforeWidth > 0 {
		cmds = append(cmds, paint.DrawCmd{
			X:     x,
			Y:     addonY,
			Text:  inst.addonBefore,
			Style: inst.resolveAddonStyle(),
		})
	}

	// Draw full border if present
	if useFullBorder {
		cmds = append(cmds, inst.paintBorder(borderX, y, boxWidth, boxHeight)...)
	}

	// Calculate content position (inside border/brackets)
	contentX := borderX + borderWidth
	contentY := y + borderWidth
	contentWidth := boxWidth - borderWidth*2
	if contentWidth < 1 {
		contentWidth = 1
	}

	prefix := inst.effectivePrefix()
	suffix := inst.effectiveSuffix()
	rawEditable := inst.displayEditableText()
	text := inst.buildDisplayText(prefix, rawEditable, suffix, contentWidth)

	// Resolve style
	inputStyle := inst.resolveStyle()

	// Draw with brackets if BorderNone
	if useBrackets {
		// Draw [text] format
		bracketStyle := inst.resolveBorderColor()
		cmds = append(cmds,
			paint.DrawCmd{X: borderX, Y: contentY, Text: "[", Style: style.Style{FG: bracketStyle}},
			paint.DrawCmd{X: borderX + 1, Y: contentY, Text: text, Style: inputStyle},
			paint.DrawCmd{X: borderX + 1 + contentWidth, Y: contentY, Text: "]", Style: style.Style{FG: bracketStyle}},
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
		cursorCol, cursorChar := inst.computeCursorOverlay(rawEditable, contentWidth, paint.StringWidth(prefix), paint.StringWidth(suffix))
		if cursorCol >= 0 && cursorCol < contentWidth {
			if cmd, ok := inst.cursorModel.DrawCmd(contentX+cursorCol, contentY, cursorChar, inputStyle); ok {
				cmds = append(cmds, cmd)
			}
		}
	}

	if addonAfterWidth > 0 {
		addonAfterX := borderX + boxWidth
		if addonBeforeWidth > 0 || addonAfterWidth > 0 {
			addonAfterX += addonGapWidth
		}
		cmds = append(cmds, paint.DrawCmd{
			X:     addonAfterX,
			Y:     addonY,
			Text:  inst.addonAfter,
			Style: inst.resolveAddonStyle(),
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

func (inst *Instance) defaultBoxWidth() int {
	return inst.innerContentWidth() + 2
}

func (inst *Instance) innerContentWidth() int {
	prefixWidth := paint.StringWidth(inst.effectivePrefix())
	suffixWidth := paint.StringWidth(inst.effectiveSuffix())
	editableWidth := paint.StringWidth(inst.displayEditableText())

	width := prefixWidth + editableWidth + suffixWidth
	if inst.width > 0 {
		width = inst.width
	}

	if inst.maxLen > 0 {
		maxEditableWidth := inst.maxLen + prefixWidth + suffixWidth
		if width > maxEditableWidth {
			width = maxEditableWidth
		}
	}

	if width < 10 {
		width = 10
	}

	return width
}

func (inst *Instance) measureTotalWidth() int {
	totalWidth := inst.defaultBoxWidth()
	if inst.addonBefore != "" {
		totalWidth += paint.StringWidth(inst.addonBefore) + addonGapWidth
	}
	if inst.addonAfter != "" {
		totalWidth += addonGapWidth + paint.StringWidth(inst.addonAfter)
	}
	return totalWidth
}

func (inst *Instance) displayEditableText() string {
	switch inst.inputType {
	case TypePassword:
		if inst.value == "" {
			return inst.placeholder
		}
		return strings.Repeat("*", utf8.RuneCountInString(inst.value))
	default:
		if inst.value == "" {
			return inst.placeholder
		}
		return inst.value
	}
}

func (inst *Instance) effectivePrefix() string {
	if inst.prefix != "" {
		return inst.prefix
	}
	if inst.searchVariant {
		return searchVariantPrefix
	}
	return ""
}

func (inst *Instance) effectiveSuffix() string {
	return inst.suffix
}

func (inst *Instance) buildDisplayText(prefix, editable, suffix string, contentWidth int) string {
	if contentWidth <= 0 {
		return ""
	}

	prefixWidth := paint.StringWidth(prefix)
	suffixWidth := paint.StringWidth(suffix)
	editableWidth := contentWidth - prefixWidth - suffixWidth
	if editableWidth < 0 {
		editableWidth = 0
	}

	editableText := editable
	if editableWidth > 0 {
		editableText = inst.padText(editable, editableWidth)
	} else {
		editableText = ""
	}

	text := prefix + editableText + suffix
	textWidth := paint.StringWidth(text)
	if textWidth > contentWidth {
		return truncateByDisplayWidth(text, contentWidth)
	}
	if textWidth < contentWidth {
		return text + strings.Repeat(" ", contentWidth-textWidth)
	}
	return text
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
	s := inst.resolveBaseStyle()

	// Placeholder styling
	if inst.value == "" && inst.placeholder != "" {
		s = s.Foreground(theme.Placeholder())
	}

	return s
}

func (inst *Instance) resolveBaseStyle() style.Style {
	s := inst.inputStyle

	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}

	if inst.state.Disabled {
		s = s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	} else if inst.state.Focused {
		s = s.Foreground(theme.Focus()).Underline(true).Bold(true)
	} else if inst.state.Hovered {
		s = s.Underline(true)
	}

	return s
}

func (inst *Instance) resolveAddonStyle() style.Style {
	s := inst.resolveBaseStyle()
	if !inst.state.Disabled && s.FG == theme.Focus() {
		s = s.Foreground(theme.Muted()).Underline(false).Bold(false)
	}
	if s.FG == "" {
		s = s.Foreground(theme.Muted())
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
		wasFocused := inst.state.Focused
		inst.state.Focused = focused
		if wasFocused && !focused {
			if inst.normalizeValueOnBlur() {
				inst.emitFieldValueChanged()
			}
			inst.emitFieldBlur()
		}
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

	if text == "" {
		return false
	}

	runes := []rune(inst.value)

	// ✨ CRITICAL: Clamp cursorPos to valid range to prevent panic
	// This handles edge cases where cursorPos was set incorrectly
	maxPos := len(runes)
	if inst.cursorPos < 0 {
		inst.cursorPos = 0
	} else if inst.cursorPos > maxPos {
		inst.cursorPos = maxPos
	}

	textRunes := []rune(text)
	if inst.inputType == TypeNumber {
		textRunes = filterNumberInsert(runes, inst.cursorPos, textRunes, inst.allowNegative, inst.allowDecimal)
		if len(textRunes) == 0 {
			return false
		}
	}

	// Check max length
	if inst.maxLen > 0 {
		currentLen := len(runes)
		textLen := len(textRunes)
		if currentLen+textLen > inst.maxLen {
			return false
		}
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
	innerWidth := 0
	if boxWidth > 0 {
		innerWidth = boxWidth - 2 // left+right border/bracket
	} else if inst.width > 0 {
		innerWidth = inst.width
	} else {
		return 0
	}

	editableWidth := innerWidth - paint.StringWidth(inst.effectivePrefix()) - paint.StringWidth(inst.effectiveSuffix())
	if editableWidth < 0 {
		return 0
	}
	return editableWidth
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
	if inst.inputType == TypeNumber {
		value = sanitizeEditableNumberValue(value, inst.allowNegative, inst.allowDecimal)
	}
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
func (inst *Instance) computeCursorOverlay(rawText string, contentWidth, prefixWidth, suffixWidth int) (int, string) {
	if contentWidth <= 0 {
		return -1, ""
	}

	editableWidth := contentWidth - prefixWidth - suffixWidth
	if editableWidth <= 0 {
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

	cursorCol := prefixWidth + displayWidthByRuneIndex(runes, cursor)
	maxCursorCol := prefixWidth + editableWidth - 1
	if cursorCol > maxCursorCol {
		cursorCol = maxCursorCol
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
	case propDisabled:
		return inst.state.Disabled, true
	case propValue:
		return inst.value, true
	case propPlaceholder:
		return inst.placeholder, true
	case propInputType:
		return inst.inputType, true
	case propAllowNegative:
		return inst.allowNegative, true
	case propAllowDecimal:
		return inst.allowDecimal, true
	case propPrefix:
		return inst.prefix, true
	case propSuffix:
		return inst.suffix, true
	case propAddonBefore:
		return inst.addonBefore, true
	case propAddonAfter:
		return inst.addonAfter, true
	case propChangeIntent:
		return inst.changeIntent, true
	case propSubmitIntent:
		return inst.submitIntent, true
	case propSearchVariant:
		return inst.searchVariant, true
	default:
		return nil, false
	}
}

// SetProp sets a prop value.
func (inst *Instance) SetProp(key string, value interface{}) {
	switch key {
	case propDisabled:
		if v, ok := value.(bool); ok {
			inst.state.Disabled = v
			inst.syncCursorVisibility()
			inst.dirty = true
		}
	case propValue:
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

	totalWidth := inst.measureTotalWidth()
	totalHeight := 1
	if inst.borderStyle != layout.BorderNone {
		totalHeight = 3
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

func (inst *Instance) normalizeValueOnBlur() bool {
	if inst.inputType != TypeNumber {
		return false
	}

	normalized := normalizeBlurNumberValue(inst.value, inst.allowNegative, inst.allowDecimal)
	if normalized == inst.value {
		return false
	}

	inst.value = normalized
	inst.cursorPos = utf8.RuneCountInString(inst.value)
	inst.cursorModel.ResetBlink()
	inst.dirty = true
	return true
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
	if v, ok := props[propInputType]; ok {
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
