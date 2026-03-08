package textarea

import (
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

// Instance is the runtime entity for Textarea components.
type Instance struct {
	key               string
	placeholder       string
	textareaStyle     style.Style
	rows, cols        int
	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent // For FieldChangeIntent extraction
	submitIntent      intent.Intent
	formID            string // Form ID for Form integration (Phase 6)
	maxLen            int
	cursorConfig      cursor.Config

	state       control.InteractionState
	value       string
	cursorPos   int // Cursor position in rune index
	cursorModel *cursor.Model
	bounds      [4]int
	dirty       bool

	intentEmitter func(intent.Intent)
	behaviors     *control.BehaviorList
}

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

// NewInstance creates a new TextareaInstance from props.
func NewInstance(props rtui.Props) *Instance {
	cursorCfg := getCursorConfigProp(props, "cursorConfig", cursor.DefaultConfig())
	inst := &Instance{
		key:               getStringProp(props, "key", ""),
		placeholder:       getStringProp(props, "placeholder", ""),
		textareaStyle:     getStyleProp(props),
		rows:              getIntProp(props, "rows", 3),
		cols:              getIntProp(props, "cols", 40),
		changeIntent:      getIntentProp(props, "changeIntent"),
		changeIntentField: getChangeIntentFieldProp(props, "changeIntent"),
		formID:            getStringProp(props, "formID", ""),
		submitIntent:      getIntentProp(props, "submitIntent"),
		value:             getStringProp(props, "value", ""),
		maxLen:            getIntProp(props, "maxLen", 0),
		cursorConfig:      cursorCfg,
		cursorModel:       cursor.NewModel(cursorCfg),
		dirty:             true,
	}
	inst.cursorPos = utf8.RuneCountInString(inst.value)

	inst.state = control.InteractionState{
		Disabled: getBoolProp(props, "disabled", false),
	}

	inst.behaviors = control.NewBehaviorList(
		&control.FocusableBehavior{},
		&control.HoverableBehavior{},
		&control.DisableableBehavior{},
	)
	inst.syncCursorVisibility()

	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }

// Parent implements TreeComponent interface (intent bubble).
// Returns nil as Textarea is a leaf component without parent tracking.
func (inst *Instance) Parent() interface{} {
	return nil
}

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              { inst.behaviors.OnUnmount(inst) }
func (inst *Instance) OnMount()              { inst.behaviors.OnMount(inst) }
func (inst *Instance) OnUnmount()            { inst.behaviors.OnUnmount(inst) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldValue := inst.value
	oldDisabled := inst.state.Disabled
	oldCursorConfig := inst.cursorConfig

	inst.placeholder = getStringProp(props, "placeholder", inst.placeholder)
	inst.textareaStyle = getStyleProp(props)
	inst.rows = getIntProp(props, "rows", inst.rows)
	inst.cols = getIntProp(props, "cols", inst.cols)
	inst.changeIntent = getIntentProp(props, "changeIntent")
	inst.changeIntentField = getChangeIntentFieldProp(props, "changeIntent")
	inst.formID = getStringProp(props, "formID", inst.formID)
	inst.submitIntent = getIntentProp(props, "submitIntent")
	newValue := getStringProp(props, "value", inst.value)
	if newValue != inst.value {
		inst.value = newValue
		inst.cursorPos = utf8.RuneCountInString(inst.value)
		inst.cursorModel.ResetBlink()
	}
	inst.maxLen = getIntProp(props, "maxLen", inst.maxLen)

	newDisabled := getBoolProp(props, "disabled", inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}

	newCursorConfig := getCursorConfigProp(props, "cursorConfig", inst.cursorConfig)
	if newCursorConfig != inst.cursorConfig {
		inst.cursorConfig = newCursorConfig
		inst.cursorModel.SetConfig(newCursorConfig)
	}
	inst.syncCursorVisibility()

	changed := oldValue != inst.value ||
		oldDisabled != inst.state.Disabled ||
		oldCursorConfig != inst.cursorConfig
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":          inst.key,
		"value":        inst.value,
		"disabled":     inst.state.Disabled,
		"cursorConfig": inst.cursorConfig,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

// =============================================================================
// PaintableInstance Interface
// =============================================================================

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	content := inst.value
	if content == "" {
		content = inst.placeholder
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	// Calculate dimensions
	width := inst.cols + 4
	height := inst.rows + 2
	if len(lines) > inst.rows {
		height = len(lines) + 2
	}

	// Resolve style
	borderStyle := inst.resolveStyle()

	// Build commands
	var cmds []paint.DrawCmd

	// Top border
	topBorder := "+" + strings.Repeat("-", width-2) + "+"
	cmds = append(cmds, paint.DrawCmd{X: x, Y: y, Text: topBorder, Style: borderStyle})

	contentWidth := width - 2
	contentHeight := height - 2

	// Content lines
	for i, line := range lines {
		if i >= contentHeight {
			break
		}
		lineText := "|" + fitLineToWidth(line, contentWidth) + "|"
		cmds = append(cmds, paint.DrawCmd{X: x, Y: y + 1 + i, Text: lineText, Style: borderStyle})
	}

	// Fill empty rows
	for i := len(lines); i < contentHeight; i++ {
		emptyLine := "|" + strings.Repeat(" ", contentWidth) + "|"
		cmds = append(cmds, paint.DrawCmd{X: x, Y: y + 1 + i, Text: emptyLine, Style: borderStyle})
	}

	// Bottom border
	cmds = append(cmds, paint.DrawCmd{X: x, Y: y + height - 1, Text: topBorder, Style: borderStyle})

	// Draw caret overlay when focused.
	if inst.shouldDrawCursor() && contentWidth > 0 && contentHeight > 0 {
		cursorLine, cursorCol := inst.cursorLineCol()
		if inst.value == "" {
			cursorLine, cursorCol = 0, 0
		}

		if cursorLine < 0 {
			cursorLine = 0
		}
		if cursorLine >= contentHeight {
			cursorLine = contentHeight - 1
		}
		if cursorCol < 0 {
			cursorCol = 0
		}
		if cursorCol >= contentWidth {
			cursorCol = contentWidth - 1
		}

		cursorChar := " "
		if inst.value != "" {
			valueLines := strings.Split(inst.value, "\n")
			if cursorLine >= 0 && cursorLine < len(valueLines) {
				cursorChar = glyphAtDisplayColumn(valueLines[cursorLine], cursorCol)
			}
		}

		if cmd, ok := inst.cursorModel.DrawCmd(x+1+cursorCol, y+1+cursorLine, cursorChar, borderStyle); ok {
			cmds = append(cmds, cmd)
		}
	}

	return cmds
}

func (inst *Instance) resolveStyle() style.Style {
	s := inst.textareaStyle

	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}

	if inst.value == "" && inst.placeholder != "" {
		s = s.Foreground(theme.Placeholder())
	}

	if inst.state.Disabled {
		s = s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	} else if inst.state.Focused {
		s = s.Foreground(theme.Focus()).Bold(true)
	} else if inst.state.Hovered {
		s = s.Underline(true)
	}

	return s
}

// =============================================================================
// FocusableInstance Interface
// =============================================================================

func (inst *Instance) SetFocus(focused bool) {
	if inst.state.Focused != focused {
		oldState := inst.state
		inst.state.Focused = focused
		inst.syncCursorVisibility()
		inst.dirty = true
		inst.behaviors.OnStateChange(inst, oldState, inst.state)
	}
}

func (inst *Instance) HasFocus() bool   { return inst.state.Focused }
func (inst *Instance) IsDisabled() bool { return inst.state.Disabled }

// =============================================================================
// TickableInstance Interface
// =============================================================================

// WantsTick reports whether the textarea caret needs periodic blink updates.
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

func (inst *Instance) HandleAction(act *action.Action) bool {
	if inst.state.Disabled {
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
		return inst.MoveCursorToLineHome()
	case action.ActionCursorEnd:
		return inst.MoveCursorToLineEnd()
	case action.ActionEnter:
		return inst.InsertText("\n")
	case action.ActionSubmit:
		if inst.submitIntent != nil && inst.intentEmitter != nil {
			inst.intentEmitter(inst.submitIntent)
		}
		return true
	}
	return false
}

// =============================================================================
// Textarea-specific Methods
// =============================================================================

func (inst *Instance) InsertText(text string) bool {
	if inst.state.Disabled {
		return false
	}

	if inst.maxLen > 0 {
		currentLen := utf8.RuneCountInString(inst.value)
		textLen := utf8.RuneCountInString(text)
		if currentLen+textLen > inst.maxLen {
			return false
		}
	}

	runes := []rune(inst.value)
	textRunes := []rune(text)
	inst.clampCursorPos(len(runes))

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

// DeleteText deletes text relative to cursor.
func (inst *Instance) DeleteText(direction int) bool {
	if inst.state.Disabled {
		return false
	}

	runes := []rune(inst.value)
	if len(runes) == 0 {
		return false
	}
	inst.clampCursorPos(len(runes))

	if direction < 0 {
		// Backspace
		if inst.cursorPos == 0 {
			return false
		}
		newRunes := make([]rune, 0, len(runes)-1)
		newRunes = append(newRunes, runes[:inst.cursorPos-1]...)
		newRunes = append(newRunes, runes[inst.cursorPos:]...)
		inst.value = string(newRunes)
		inst.cursorPos--
	} else {
		// Delete
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
	inst.emitFieldValueChanged()
	return true
}

func (inst *Instance) SetValue(value string) {
	if inst.value != value {
		inst.value = value
		inst.cursorPos = utf8.RuneCountInString(value)
		inst.cursorModel.ResetBlink()
		inst.dirty = true
	}
}

func (inst *Instance) GetValue() string { return inst.value }

// CursorPos returns the current cursor position.
func (inst *Instance) CursorPos() int { return inst.cursorPos }

// SetCursorPos sets cursor position in rune index.
func (inst *Instance) SetCursorPos(pos int) {
	max := utf8.RuneCountInString(inst.value)
	if pos < 0 {
		pos = 0
	}
	if pos > max {
		pos = max
	}
	if inst.cursorPos != pos {
		inst.cursorPos = pos
		inst.cursorModel.ResetBlink()
		inst.dirty = true
	}
}

// MoveCursor moves cursor by delta runes.
func (inst *Instance) MoveCursor(delta int) bool {
	oldPos := inst.cursorPos
	inst.SetCursorPos(inst.cursorPos + delta)
	return inst.cursorPos != oldPos
}

// MoveCursorToLineHome moves cursor to line start.
func (inst *Instance) MoveCursorToLineHome() bool {
	oldPos := inst.cursorPos
	start, _ := inst.currentLineBounds()
	inst.SetCursorPos(start)
	return inst.cursorPos != oldPos
}

// MoveCursorToLineEnd moves cursor to line end.
func (inst *Instance) MoveCursorToLineEnd() bool {
	oldPos := inst.cursorPos
	_, end := inst.currentLineBounds()
	inst.SetCursorPos(end)
	return inst.cursorPos != oldPos
}

func (inst *Instance) shouldDrawCursor() bool {
	return inst.cursorModel != nil && inst.cursorModel.ShouldPaint()
}

func (inst *Instance) wantsCursorVisible() bool {
	return inst.state.Focused && !inst.state.Disabled
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

func (inst *Instance) clampCursorPos(max int) {
	if inst.cursorPos < 0 {
		inst.cursorPos = 0
	} else if inst.cursorPos > max {
		inst.cursorPos = max
	}
}

func (inst *Instance) currentLineBounds() (start, end int) {
	runes := []rune(inst.value)
	cursor := inst.cursorPos
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}

	start = 0
	for i := 0; i < cursor; i++ {
		if runes[i] == '\n' {
			start = i + 1
		}
	}

	end = len(runes)
	for i := cursor; i < len(runes); i++ {
		if runes[i] == '\n' {
			end = i
			break
		}
	}
	return start, end
}

func (inst *Instance) cursorLineCol() (line int, col int) {
	runes := []rune(inst.value)
	cursor := inst.cursorPos
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}

	line, col = 0, 0
	for i := 0; i < cursor; i++ {
		r := runes[i]
		if r == '\n' {
			line++
			col = 0
			continue
		}
		col += paint.RuneWidth(r)
	}
	return line, col
}

func fitLineToWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}
	clamped := truncateByDisplayWidth(text, width)
	lineWidth := paint.StringWidth(clamped)
	if lineWidth < width {
		clamped += strings.Repeat(" ", width-lineWidth)
	}
	return clamped
}

func truncateByDisplayWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(text)
	var result []rune
	width := 0
	for _, r := range runes {
		rw := paint.RuneWidth(r)
		if width+rw > maxWidth {
			break
		}
		result = append(result, r)
		width += rw
	}
	return string(result)
}

func glyphAtDisplayColumn(text string, col int) string {
	if col < 0 {
		return " "
	}
	width := 0
	for _, r := range text {
		rw := paint.RuneWidth(r)
		if col >= width && col < width+rw {
			return string(r)
		}
		width += rw
	}
	return " "
}

// =============================================================================
// control.Instance Interface
// =============================================================================

func (inst *Instance) GetState() *control.InteractionState { return &inst.state }
func (inst *Instance) SetState(state control.InteractionState) {
	oldState := inst.state
	inst.state = state
	inst.syncCursorVisibility()
	if oldState != inst.state {
		inst.dirty = true
	}
	inst.behaviors.OnStateChange(inst, oldState, inst.state)
}
func (inst *Instance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil {
		inst.intentEmitter(i)
	}
}
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}
func (inst *Instance) SetBounds(x, y, w, h int) { inst.bounds = [4]int{x, y, w, h} }
func (inst *Instance) GetStyle() style.Style    { return inst.textareaStyle }
func (inst *Instance) SetStyle(s style.Style)   { inst.textareaStyle = s }
func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case "disabled":
		return inst.state.Disabled, true
	case "value":
		return inst.value, true
	default:
		return nil, false
	}
}
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
func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) { inst.intentEmitter = fn }
func (inst *Instance) ClearDirty()                             { inst.dirty = false }

// =============================================================================
// Measurable Interface
// =============================================================================

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	width := inst.cols + 4
	height := inst.rows + 2

	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
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
		if inst.changeIntentField != nil {
			formIntent := form.FieldChange(
				inst.formID,
				inst.changeIntentField.GetField(),
				inst.value,
				true, // isDirty
			)
			intent.Emit(inst, formIntent)
		}
		return
	}

	// Original MVP behavior: emit FieldChangeIntent
	if inst.changeIntentField != nil {
		changeIntent := intent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
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

	if inst.changeIntentField != nil {
		blurIntent := form.FieldBlur(
			inst.formID,
			inst.changeIntentField.GetField(),
			inst.value,
		)
		intent.Emit(inst, blurIntent)
	}
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

func getChangeIntentFieldProp(props rtui.Props, key string) intent.FieldIntent {
	if v, ok := props[key]; ok {
		if fieldIntent, ok := v.(intent.FieldIntent); ok {
			return fieldIntent
		}
	}
	return nil
}

func getCursorConfigProp(props rtui.Props, key string, def cursor.Config) cursor.Config {
	if v, ok := props[key]; ok {
		if cfg, ok := v.(cursor.Config); ok {
			return cursor.NormalizeConfig(cfg)
		}
	}
	return cursor.NormalizeConfig(def)
}
