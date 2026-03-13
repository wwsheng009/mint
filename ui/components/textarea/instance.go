package textarea

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
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Textarea components.
type Instance struct {
	key                    string
	placeholder            string
	textareaStyle          style.Style
	rows, cols             int
	scrollOffset           int
	scrollOffsetControlled bool
	showScrollbar          bool
	scrollbarStyle         style.Style
	changeIntent           intent.Intent
	changeIntentField      intent.FieldIntent // For FieldChangeIntent extraction
	submitIntent           intent.Intent
	formID                 string // Form ID for Form integration (Phase 6)
	maxLen                 int
	cursorConfig           cursor.Config

	state         control.InteractionState
	value         string
	cursorPos     int // Cursor position in rune index
	cursorGoal    int // Preferred cursor display column during vertical moves.
	hasCursorGoal bool
	cursorModel   *cursor.Model
	bounds        [4]int
	dirty         bool

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
		key:                    proputil.GetString(props, "key", ""),
		placeholder:            proputil.GetString(props, "placeholder", ""),
		textareaStyle:          proputil.GetStyle(props, "style", style.Style{}),
		rows:                   proputil.GetInt(props, "rows", 3),
		cols:                   proputil.GetInt(props, "cols", 40),
		scrollOffset:           proputil.GetInt(props, "scrollOffset", 0),
		scrollOffsetControlled: proputil.GetBool(props, "scrollOffsetControlled", false),
		showScrollbar:          proputil.GetBool(props, "showScrollbar", true),
		scrollbarStyle:         proputil.GetStyle(props, "scrollbarStyle", style.Style{}),
		changeIntent:           proputil.GetIntent(props, "changeIntent", nil),
		changeIntentField:      getChangeIntentFieldProp(props, "changeIntent"),
		formID:                 proputil.GetString(props, "formID", ""),
		submitIntent:           proputil.GetIntent(props, "submitIntent", nil),
		value:                  proputil.GetString(props, "value", ""),
		maxLen:                 proputil.GetInt(props, "maxLen", 0),
		cursorConfig:           cursorCfg,
		cursorModel:            cursor.NewModel(cursorCfg),
		dirty:                  true,
	}
	inst.cursorPos = utf8.RuneCountInString(inst.value)

	inst.state = control.InteractionState{
		Disabled: proputil.GetBool(props, "disabled", false),
	}

	inst.behaviors = control.NewBehaviorList(
		&control.FocusableBehavior{},
		&control.HoverableBehavior{},
		&control.DisableableBehavior{},
	)
	inst.syncCursorVisibility()
	inst.ensureCursorVisible()

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
	oldScrollOffset := inst.scrollOffset
	oldScrollOffsetControlled := inst.scrollOffsetControlled
	oldShowScrollbar := inst.showScrollbar
	oldScrollbarStyle := inst.scrollbarStyle

	inst.placeholder = proputil.GetString(props, "placeholder", inst.placeholder)
	inst.textareaStyle = proputil.GetStyle(props, "style", style.Style{})
	inst.rows = proputil.GetInt(props, "rows", inst.rows)
	inst.cols = proputil.GetInt(props, "cols", inst.cols)
	if controlled, ok := props[propScrollOffsetControlled].(bool); ok {
		inst.scrollOffsetControlled = controlled
	}
	if inst.scrollOffsetControlled {
		inst.scrollOffset = proputil.GetInt(props, "scrollOffset", inst.scrollOffset)
	} else if offset, ok := props[propScrollOffset].(int); ok {
		inst.scrollOffset = offset
	}
	inst.showScrollbar = proputil.GetBool(props, "showScrollbar", inst.showScrollbar)
	inst.scrollbarStyle = proputil.GetStyle(props, "scrollbarStyle", style.Style{})
	inst.changeIntent = proputil.GetIntent(props, "changeIntent", nil)
	inst.changeIntentField = getChangeIntentFieldProp(props, "changeIntent")
	inst.formID = proputil.GetString(props, "formID", inst.formID)
	inst.submitIntent = proputil.GetIntent(props, "submitIntent", nil)
	newValue := proputil.GetString(props, "value", inst.value)
	if newValue != inst.value {
		inst.value = newValue
		inst.cursorPos = utf8.RuneCountInString(inst.value)
		inst.clearCursorGoal()
		inst.cursorModel.ResetBlink()
		inst.ensureCursorVisible()
	}
	inst.maxLen = proputil.GetInt(props, "maxLen", inst.maxLen)

	newDisabled := proputil.GetBool(props, "disabled", inst.state.Disabled)
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
		oldCursorConfig != inst.cursorConfig ||
		oldScrollOffset != inst.scrollOffset ||
		oldScrollOffsetControlled != inst.scrollOffsetControlled ||
		oldShowScrollbar != inst.showScrollbar ||
		oldScrollbarStyle != inst.scrollbarStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:                    inst.key,
		propValue:                  inst.value,
		propDisabled:               inst.state.Disabled,
		propScrollOffsetControlled: inst.scrollOffsetControlled,
		propScrollOffset:           inst.scrollOffset,
		propShowScrollbar:          inst.showScrollbar,
		propScrollbarStyle:         inst.scrollbarStyle,
		propCursorConfig:           inst.cursorConfig,
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

	// Calculate dimensions
	width := inst.cols + 4
	height := inst.rows + 2
	contentWidth := width - 2
	if contentWidth < 1 {
		contentWidth = 1
	}

	contentHeight := inst.rows
	if contentHeight < 1 {
		contentHeight = 1
	}
	height = contentHeight + 2

	logicalLines := strings.Split(content, "\n")
	if len(logicalLines) == 0 {
		logicalLines = []string{""}
	}
	wrappedLines := wrapLinesByDisplayWidth(logicalLines, contentWidth)
	if len(wrappedLines) == 0 {
		wrappedLines = []string{""}
	}

	// Resolve style
	borderStyle := inst.resolveStyle()

	// Build commands
	var cmds []paint.DrawCmd

	// Top border
	topBorder := "+" + strings.Repeat("-", width-2) + "+"
	cmds = append(cmds, paint.DrawCmd{X: x, Y: y, Text: topBorder, Style: borderStyle})

	cursorLine, cursorCol := inst.cursorLineColWrapped(contentWidth)
	if inst.value == "" {
		cursorLine, cursorCol = 0, 0
	}

	viewport := scrollutil.NewVerticalViewport(len(wrappedLines), contentHeight, inst.scrollOffset)
	inst.scrollOffset = viewport.Offset

	// Content lines
	visibleStart, visibleEnd := viewport.VisibleRange()
	visibleLines := wrappedLines[visibleStart:visibleEnd]
	for i, line := range visibleLines {
		if i >= contentHeight {
			break
		}
		lineText := "|" + fitLineToWidth(line, contentWidth) + "|"
		cmds = append(cmds, paint.DrawCmd{X: x, Y: y + 1 + i, Text: lineText, Style: borderStyle})
	}

	// Fill empty rows
	for i := len(visibleLines); i < contentHeight; i++ {
		emptyLine := "|" + strings.Repeat(" ", contentWidth) + "|"
		cmds = append(cmds, paint.DrawCmd{X: x, Y: y + 1 + i, Text: emptyLine, Style: borderStyle})
	}

	// Bottom border
	cmds = append(cmds, paint.DrawCmd{X: x, Y: y + height - 1, Text: topBorder, Style: borderStyle})

	if inst.showScrollbar {
		scrollbarStyle := inst.scrollbarStyle
		if scrollbarStyle.FG == "" {
			scrollbarStyle = scrollbarStyle.Foreground(borderStyle.FG)
		}
		cmds = append(cmds, scrollutil.DrawVerticalScrollbar(
			x+width-1,
			y+1,
			contentHeight,
			viewport,
			scrollbarStyle,
			scrollutil.DefaultVerticalScrollbarConfig(),
		)...)
	}

	// Draw caret overlay when focused.
	if inst.shouldDrawCursor() && contentWidth > 0 && contentHeight > 0 {
		cursorLine = cursorLine - visibleStart
		if cursorCol < 0 {
			cursorCol = 0
		}
		if cursorCol >= contentWidth {
			cursorCol = contentWidth - 1
		}

		if cursorLine < 0 || cursorLine >= contentHeight {
			return cmds
		}

		cursorChar := inst.cursorGlyphAtCursor()
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
		wasFocused := inst.state.Focused
		inst.state.Focused = focused
		if wasFocused && !focused {
			inst.emitFieldBlur()
		}
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
	case action.ActionScroll:
		if delta, ok := scrollutil.DeltaFromAction(act); ok {
			return inst.ScrollBy(delta)
		}
		return false
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
	case action.ActionCursorUp:
		return inst.MoveCursorUp()
	case action.ActionCursorDown:
		return inst.MoveCursorDown()
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

// ScrollBy scrolls viewport by delta rows.
func (inst *Instance) ScrollBy(delta int) bool {
	viewport := scrollutil.NewVerticalViewport(inst.wrappedLineCountForContent(), inst.viewportRows(), inst.scrollOffset)
	if !viewport.ScrollBy(delta) {
		return false
	}
	inst.scrollOffset = viewport.Offset
	inst.dirty = true
	return true
}

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
	inst.clearCursorGoal()
	inst.cursorModel.ResetBlink()
	inst.ensureCursorVisible()
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

	inst.clearCursorGoal()
	inst.cursorModel.ResetBlink()
	inst.ensureCursorVisible()
	inst.dirty = true
	inst.emitFieldValueChanged()
	return true
}

func (inst *Instance) SetValue(value string) {
	if inst.value != value {
		inst.value = value
		inst.cursorPos = utf8.RuneCountInString(value)
		inst.clearCursorGoal()
		inst.cursorModel.ResetBlink()
		inst.ensureCursorVisible()
		inst.dirty = true
	}
}

func (inst *Instance) GetValue() string { return inst.value }

// CursorPos returns the current cursor position.
func (inst *Instance) CursorPos() int { return inst.cursorPos }

// SetCursorPos sets cursor position in rune index.
func (inst *Instance) SetCursorPos(pos int) {
	inst.clearCursorGoal()
	inst.setCursorPos(pos)
}

func (inst *Instance) setCursorPos(pos int) {
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
		inst.ensureCursorVisible()
		inst.dirty = true
	}
}

// MoveCursor moves cursor by delta runes.
func (inst *Instance) MoveCursor(delta int) bool {
	oldPos := inst.cursorPos
	inst.SetCursorPos(inst.cursorPos + delta)
	return inst.cursorPos != oldPos
}

// MoveCursorUp moves cursor to previous line while preserving column when possible.
func (inst *Instance) MoveCursorUp() bool {
	maxWidth := inst.wrapContentWidth()
	runes := []rune(inst.value)
	oldPos := inst.cursorPos
	inst.clampCursorPos(len(runes))

	lineAt, colAt := buildWrappedCursorMap(runes, maxWidth)
	currentLine := lineAt[inst.cursorPos]
	if currentLine == 0 {
		return false
	}

	col := inst.cursorGoal
	if !inst.hasCursorGoal {
		col = colAt[inst.cursorPos]
		inst.cursorGoal = col
		inst.hasCursorGoal = true
	}

	newPos, ok := findCursorPosForWrappedLineCol(lineAt, colAt, currentLine-1, col)
	if !ok {
		return false
	}
	inst.setCursorPos(newPos)
	return inst.cursorPos != oldPos
}

// MoveCursorDown moves cursor to next line while preserving column when possible.
func (inst *Instance) MoveCursorDown() bool {
	maxWidth := inst.wrapContentWidth()
	runes := []rune(inst.value)
	oldPos := inst.cursorPos
	inst.clampCursorPos(len(runes))

	lineAt, colAt := buildWrappedCursorMap(runes, maxWidth)
	currentLine := lineAt[inst.cursorPos]
	maxLine := lineAt[len(lineAt)-1]
	if currentLine >= maxLine {
		return false
	}

	col := inst.cursorGoal
	if !inst.hasCursorGoal {
		col = colAt[inst.cursorPos]
		inst.cursorGoal = col
		inst.hasCursorGoal = true
	}
	newPos, ok := findCursorPosForWrappedLineCol(lineAt, colAt, currentLine+1, col)
	if !ok {
		return false
	}
	inst.setCursorPos(newPos)
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

func (inst *Instance) clearCursorGoal() {
	inst.cursorGoal = 0
	inst.hasCursorGoal = false
}

func (inst *Instance) viewportRows() int {
	if inst.rows < 1 {
		return 1
	}
	return inst.rows
}

func (inst *Instance) wrappedLineCountForContent() int {
	width := inst.wrapContentWidth()
	if width < 1 {
		width = 1
	}
	content := inst.value
	if content == "" {
		content = inst.placeholder
	}
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	wrapped := wrapLinesByDisplayWidth(lines, width)
	if len(wrapped) == 0 {
		return 1
	}
	return len(wrapped)
}

func (inst *Instance) ensureCursorVisible() {
	if inst.scrollOffsetControlled {
		return
	}
	width := inst.wrapContentWidth()
	if width < 1 {
		width = 1
	}
	cursorLine, _ := inst.cursorLineColWrapped(width)
	viewport := scrollutil.NewVerticalViewport(inst.wrappedLineCountForContent(), inst.viewportRows(), inst.scrollOffset)
	viewport.EnsureVisible(cursorLine)
	inst.scrollOffset = viewport.Offset
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

func (inst *Instance) cursorLineColWrapped(maxWidth int) (line int, col int) {
	if maxWidth <= 0 {
		return 0, 0
	}
	runes := []rune(inst.value)
	cursor := inst.cursorPos
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	lineAt, colAt := buildWrappedCursorMap(runes, maxWidth)
	return lineAt[cursor], colAt[cursor]
}

func (inst *Instance) cursorGlyphAtCursor() string {
	if inst.value == "" {
		return " "
	}
	runes := []rune(inst.value)
	cursor := inst.cursorPos
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(runes) {
		return " "
	}
	if runes[cursor] == '\n' {
		return " "
	}
	return string(runes[cursor])
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

func wrapLinesByDisplayWidth(lines []string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{""}
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, wrapLineByDisplayWidth(line, maxWidth)...)
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}

func wrapLineByDisplayWidth(line string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{""}
	}
	if line == "" {
		return []string{""}
	}

	segments := make([]string, 0, 1)
	var current []rune
	currentWidth := 0

	for _, r := range line {
		rw := paint.RuneWidth(r)
		if rw > maxWidth {
			if len(current) > 0 {
				segments = append(segments, string(current))
				current = current[:0]
				currentWidth = 0
			}
			segments = append(segments, string(r))
			continue
		}

		if currentWidth+rw > maxWidth {
			segments = append(segments, string(current))
			current = current[:0]
			currentWidth = 0
		}

		current = append(current, r)
		currentWidth += rw

		if currentWidth == maxWidth {
			segments = append(segments, string(current))
			current = current[:0]
			currentWidth = 0
		}
	}

	if len(current) > 0 {
		segments = append(segments, string(current))
	}
	if len(segments) == 0 {
		segments = append(segments, "")
	}
	return segments
}

func (inst *Instance) wrapContentWidth() int {
	width := inst.cols + 2
	if width < 1 {
		return 1
	}
	return width
}

func buildWrappedCursorMap(runes []rune, maxWidth int) ([]int, []int) {
	if maxWidth <= 0 {
		maxWidth = 1
	}
	lineAt := make([]int, len(runes)+1)
	colAt := make([]int, len(runes)+1)
	line, col := 0, 0
	lineAt[0], colAt[0] = line, col

	for index, r := range runes {
		if r == '\n' {
			line++
			col = 0
			lineAt[index+1], colAt[index+1] = line, col
			continue
		}

		rw := paint.RuneWidth(r)
		if rw <= 0 {
			rw = 1
		}
		if col+rw > maxWidth {
			line++
			col = 0
		}

		col += rw
		if col >= maxWidth {
			line++
			col = 0
		}

		lineAt[index+1], colAt[index+1] = line, col
	}

	return lineAt, colAt
}

func findCursorPosForWrappedLineCol(lineAt, colAt []int, targetLine, targetCol int) (int, bool) {
	if len(lineAt) == 0 || len(lineAt) != len(colAt) {
		return 0, false
	}
	if targetLine < 0 {
		return 0, false
	}
	if targetCol < 0 {
		targetCol = 0
	}

	maxLine := lineAt[len(lineAt)-1]
	if targetLine > maxLine {
		return 0, false
	}

	bestPos := -1
	bestCol := -1
	for position := 0; position < len(lineAt); position++ {
		line := lineAt[position]
		col := colAt[position]

		if line < targetLine {
			continue
		}
		if line > targetLine {
			break
		}

		if col == targetCol {
			return position, true
		}
		if col < targetCol {
			if col > bestCol {
				bestCol = col
				bestPos = position
			}
			continue
		}
		if bestPos != -1 {
			return bestPos, true
		}
		return position, true
	}

	if bestPos != -1 {
		return bestPos, true
	}
	return 0, false
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
	case propDisabled:
		return inst.state.Disabled, true
	case propValue:
		return inst.value, true
	default:
		return nil, false
	}
}
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
