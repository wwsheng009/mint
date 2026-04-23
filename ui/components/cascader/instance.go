package cascader

import (
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

type popupColumn struct {
	Level       int
	Options     []Option
	Highlighted int
	Selected    int
	Active      bool
}

// Instance is the runtime entity for Cascader components.
type Instance struct {
	key               string
	componentID       string
	parent            rtui.ComponentInstance
	childInstances    []rtui.ComponentInstance
	options           []Option
	cascaderStyle     style.Style
	width             int
	placeholder       string
	separator         string
	changeOnSelect    bool
	valueControlled   bool
	disabled          bool
	changeIntent      runtimeintent.Intent
	changeIntentField runtimeintent.FieldIntent
	formID            string

	selectedPath []string
	open         bool
	focused      bool
	activeLevel  int
	highlighted  []int
	bounds       [4]int
	dirty        bool

	intentEmitter func(runtimeintent.Intent)
}

var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ rtui.TreeNode              = (*Instance)(nil)
	_ rtui.TreeContainer         = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new Cascader instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:               proputil.GetString(props, propKey, ""),
		componentID:       proputil.GetString(props, propComponentID, ""),
		options:           getOptionsProp(props),
		cascaderStyle:     proputil.GetStyle(props, propStyle, style.Style{}),
		width:             proputil.GetInt(props, propWidth, defaultTriggerWidth),
		placeholder:       proputil.GetString(props, propPlaceholder, defaultPlaceholder),
		separator:         proputil.GetString(props, propSeparator, defaultSeparator),
		changeOnSelect:    proputil.GetBool(props, propChangeOnSelect, false),
		valueControlled:   proputil.GetBool(props, propValueControlled, false),
		disabled:          proputil.GetBool(props, propDisabled, false),
		changeIntent:      proputil.GetIntent(props, propChangeIntent, nil),
		changeIntentField: getChangeIntentFieldProp(props, propChangeIntent),
		formID:            proputil.GetString(props, propFormID, ""),
		dirty:             true,
	}

	if inst.valueControlled {
		inst.selectedPath = getStringSliceProp(props, propValue, nil)
	} else {
		inst.selectedPath = getStringSliceProp(props, propDefaultValue, nil)
		if len(inst.selectedPath) == 0 {
			inst.selectedPath = getStringSliceProp(props, propValue, nil)
		}
	}

	inst.normalizeConfiguration()
	inst.resetNavigationFromSelection()
	return inst
}

func (inst *Instance) Key() string { return inst.key }

func (inst *Instance) SetKey(key string) {
	inst.key = key
}

func (inst *Instance) Parent() interface{} {
	return inst.parent
}

func (inst *Instance) SetParent(parent rtui.ComponentInstance) {
	inst.parent = parent
}

func (inst *Instance) Children() []rtui.ComponentInstance {
	return append([]rtui.ComponentInstance(nil), inst.childInstances...)
}

func (inst *Instance) AddChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	for index, existing := range inst.childInstances {
		if existing == child || existing.Key() == child.Key() {
			inst.childInstances[index] = child
			if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
				setter.SetParent(inst)
			}
			return
		}
	}
	inst.childInstances = append(inst.childInstances, child)
	if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
		setter.SetParent(inst)
	}
}

func (inst *Instance) RemoveChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	for index, existing := range inst.childInstances {
		if existing != child {
			continue
		}
		inst.childInstances = append(inst.childInstances[:index], inst.childInstances[index+1:]...)
		if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
			setter.SetParent(nil)
		}
		return
	}
}

func (inst *Instance) ClearChildren() {
	for _, child := range inst.childInstances {
		if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
			setter.SetParent(nil)
		}
	}
	inst.childInstances = inst.childInstances[:0]
}

func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}

func (inst *Instance) Destroy() {}

func (inst *Instance) OnMount() {}

func (inst *Instance) OnUnmount() {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldOptions := append([]Option(nil), inst.options...)
	oldSelected := append([]string(nil), inst.selectedPath...)
	oldOpen := inst.open
	oldFocused := inst.focused
	oldActiveLevel := inst.activeLevel
	oldHighlighted := append([]int(nil), inst.highlighted...)
	oldWidth := inst.width
	oldPlaceholder := inst.placeholder
	oldSeparator := inst.separator
	oldDisabled := inst.disabled
	oldChangeOnSelect := inst.changeOnSelect
	oldControlled := inst.valueControlled

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.componentID = proputil.GetString(props, propComponentID, inst.componentID)
	inst.options = getOptionsPropOr(props, inst.options)
	inst.cascaderStyle = proputil.GetStyle(props, propStyle, inst.cascaderStyle)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.placeholder = proputil.GetString(props, propPlaceholder, inst.placeholder)
	inst.separator = proputil.GetString(props, propSeparator, inst.separator)
	inst.changeOnSelect = proputil.GetBool(props, propChangeOnSelect, inst.changeOnSelect)
	inst.valueControlled = proputil.GetBool(props, propValueControlled, inst.valueControlled)
	inst.disabled = proputil.GetBool(props, propDisabled, inst.disabled)
	inst.changeIntent = proputil.GetIntent(props, propChangeIntent, inst.changeIntent)
	inst.changeIntentField = getChangeIntentFieldProp(props, propChangeIntent)
	inst.formID = proputil.GetString(props, propFormID, inst.formID)

	if value, ok := props[propValue].([]string); ok && inst.valueControlled {
		inst.selectedPath = append([]string(nil), value...)
	}
	if defaultValue, ok := props[propDefaultValue].([]string); ok && !inst.valueControlled && len(oldSelected) == 0 {
		inst.selectedPath = append([]string(nil), defaultValue...)
	}

	inst.normalizeConfiguration()
	if !reflect.DeepEqual(oldOptions, inst.options) || !equalStringSlices(oldSelected, inst.selectedPath) || oldControlled != inst.valueControlled {
		inst.resetNavigationFromSelection()
	}
	if inst.disabled {
		inst.open = false
	}

	changed := !reflect.DeepEqual(oldOptions, inst.options) ||
		!equalStringSlices(oldSelected, inst.selectedPath) ||
		oldOpen != inst.open ||
		oldFocused != inst.focused ||
		oldActiveLevel != inst.activeLevel ||
		!equalIntSlices(oldHighlighted, inst.highlighted) ||
		oldWidth != inst.width ||
		oldPlaceholder != inst.placeholder ||
		oldSeparator != inst.separator ||
		oldDisabled != inst.disabled ||
		oldChangeOnSelect != inst.changeOnSelect ||
		oldControlled != inst.valueControlled

	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propChangeIntent:    inst.changeIntent,
		propChangeOnSelect:  inst.changeOnSelect,
		propComponentID:     inst.componentID,
		propDisabled:        inst.disabled,
		propFormID:          inst.formID,
		propKey:             inst.key,
		propOptions:         append([]Option(nil), inst.options...),
		propPlaceholder:     inst.placeholder,
		propSeparator:       inst.separator,
		propStyle:           inst.cascaderStyle,
		propValue:           append([]string(nil), inst.selectedPath...),
		propValueControlled: inst.valueControlled,
		propWidth:           inst.width,
	}
}

func (inst *Instance) MarkDirty() {
	inst.dirty = true
}

func (inst *Instance) IsDirty() bool {
	return inst.dirty
}

func (inst *Instance) ClearDirty() {
	inst.dirty = false
}

func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

func (inst *Instance) SetIntentEmitter(fn func(runtimeintent.Intent)) {
	inst.intentEmitter = fn
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	triggerWidth := inst.triggerPaintWidth()
	cmds := []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  inst.triggerText(triggerWidth),
		Style: inst.resolveStyle(),
	}}
	if !inst.open {
		return cmds
	}
	return append(cmds, inst.paintPopupAt(x, y+1)...)
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := inst.triggerWidth()
	if inst.open {
		width = maxInt(width, inst.popupWidth())
	}
	height := 1
	if inst.open {
		height += inst.popupHeight()
	}
	return layout.Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(height),
	}
}

func (inst *Instance) SetFocus(focused bool) {
	if inst.focused == focused {
		return
	}
	inst.focused = focused
	if !focused {
		inst.closePopup()
		inst.emitFieldBlur()
	}
	inst.dirty = true
}

func (inst *Instance) HasFocus() bool {
	return inst.focused
}

func (inst *Instance) IsDisabled() bool {
	return inst.disabled
}

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil || inst.disabled {
		return false
	}

	switch act.Type {
	case action.ActionClick:
		return inst.handleClick(act)
	case action.ActionMouseRelease:
		return inst.handleMouseRelease(act)
	case action.ActionSelect:
		if _, ok := mousePayload(act.Payload); ok {
			return inst.handleMouseRelease(act)
		}
		if !inst.open {
			return inst.openPopup()
		}
		return inst.activateCurrent()
	case action.ActionHover:
		return inst.handleHover(act)
	case action.ActionEnter, action.ActionSubmit:
		if !inst.open {
			return inst.openPopup()
		}
		return inst.activateCurrent()
	case action.ActionNavigateDown:
		if !inst.open {
			return inst.openPopup()
		}
		return inst.moveWithinLevel(inst.activeLevel, 1)
	case action.ActionNavigateUp:
		if !inst.open {
			return inst.openPopup()
		}
		return inst.moveWithinLevel(inst.activeLevel, -1)
	case action.ActionNavigateLeft:
		if !inst.open {
			return false
		}
		return inst.moveLeft()
	case action.ActionNavigateRight:
		if !inst.open {
			return inst.openPopup()
		}
		return inst.moveRight()
	case action.ActionNavigateHome:
		if !inst.open {
			return inst.openPopup()
		}
		return inst.moveToBoundary(inst.activeLevel, false)
	case action.ActionNavigateEnd:
		if !inst.open {
			return inst.openPopup()
		}
		return inst.moveToBoundary(inst.activeLevel, true)
	case action.ActionCancel:
		return inst.closePopup()
	case action.ActionBlur:
		inst.SetFocus(false)
		return true
	}
	return false
}

func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// SelectedValue returns the committed path value joined by "/".
func (inst *Instance) SelectedValue() string {
	return strings.Join(inst.selectedPath, defaultFieldSeparator)
}

// SelectedValues returns the committed path values.
func (inst *Instance) SelectedValues() []string {
	return append([]string(nil), inst.selectedPath...)
}

// SelectedLabels returns the committed path labels.
func (inst *Instance) SelectedLabels() []string {
	_, labels, _ := resolvePath(inst.options, inst.selectedPath)
	return labels
}

func (inst *Instance) handleClick(act *action.Action) bool {
	mouse, ok := mousePayload(act.Payload)
	if !ok {
		if inst.open {
			return inst.activateCurrent()
		}
		return inst.openPopup()
	}

	if mouse.LocalY <= 0 {
		if inst.open {
			return inst.closePopup()
		}
		return inst.openPopup()
	}

	if !inst.open {
		return inst.openPopup()
	}

	level, index, hit := inst.popupHit(mouse.LocalX, mouse.LocalY)
	if !hit {
		return true
	}
	return inst.handlePopupSelection(level, index)
}

func (inst *Instance) handleMouseRelease(act *action.Action) bool {
	mouse, ok := mousePayload(act.Payload)
	if !ok {
		return true
	}
	if mouse.LocalY <= 0 {
		return true
	}
	if !inst.open {
		return true
	}
	level, index, hit := inst.popupHit(mouse.LocalX, mouse.LocalY)
	if !hit {
		return true
	}
	return inst.handlePopupSelection(level, index)
}

func (inst *Instance) handleHover(act *action.Action) bool {
	if !inst.open {
		return false
	}
	mouse, ok := mousePayload(act.Payload)
	if !ok {
		return false
	}
	level, index, hit := inst.popupHit(mouse.LocalX, mouse.LocalY)
	if !hit {
		return false
	}
	changed := inst.setHighlight(level, index)
	if inst.activeLevel != level {
		inst.activeLevel = level
		changed = true
	}
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) handlePopupSelection(level, index int) bool {
	changed := inst.setHighlight(level, index)
	if inst.activeLevel != level {
		inst.activeLevel = level
		changed = true
	}
	option, ok := inst.optionAtLevel(level, index)
	if !ok {
		if changed {
			inst.dirty = true
		}
		return changed
	}
	if len(option.Children) > 0 && !inst.changeOnSelect {
		if inst.moveRight() {
			return true
		}
		if changed {
			inst.dirty = true
		}
		return changed
	}
	if inst.commitActivePath() {
		return true
	}
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) activateCurrent() bool {
	values, _, current, ok := inst.activePath()
	if !ok {
		return false
	}
	if len(current.Children) > 0 && !inst.changeOnSelect {
		return inst.moveRight()
	}
	return inst.commitSelection(values)
}

func (inst *Instance) openPopup() bool {
	if inst.open || len(inst.options) == 0 {
		return false
	}
	inst.resetNavigationFromSelection()
	inst.open = true
	inst.dirty = true
	return true
}

func (inst *Instance) closePopup() bool {
	if !inst.open {
		return false
	}
	inst.open = false
	inst.resetNavigationFromSelection()
	inst.dirty = true
	return true
}

func (inst *Instance) commitActivePath() bool {
	values, _, _, ok := inst.activePath()
	if !ok {
		return false
	}
	return inst.commitSelection(values)
}

func (inst *Instance) commitSelection(values []string) bool {
	values = normalizePathValues(inst.options, values)
	selectionChanged := !equalStringSlices(inst.selectedPath, values)
	wasOpen := inst.open

	inst.selectedPath = append([]string(nil), values...)
	inst.open = false
	inst.resetNavigationFromSelection()

	if selectionChanged || wasOpen {
		inst.dirty = true
	}
	if selectionChanged {
		inst.emitChange()
	}
	return selectionChanged || wasOpen
}

func (inst *Instance) moveWithinLevel(level, delta int) bool {
	options := inst.optionsAtLevel(level)
	if len(options) == 0 {
		return false
	}
	current := inst.highlightAt(level)
	next := nextEnabledIndex(options, current, delta)
	if next < 0 {
		return false
	}
	changed := inst.setHighlight(level, next)
	if inst.activeLevel != level {
		inst.activeLevel = level
		changed = true
	}
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) moveToBoundary(level int, last bool) bool {
	options := inst.optionsAtLevel(level)
	if len(options) == 0 {
		return false
	}
	index := firstEnabledIndex(options)
	if last {
		index = lastEnabledIndex(options)
	}
	if index < 0 {
		return false
	}
	changed := inst.setHighlight(level, index)
	if inst.activeLevel != level {
		inst.activeLevel = level
		changed = true
	}
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) moveLeft() bool {
	if inst.activeLevel <= 0 {
		return false
	}
	inst.activeLevel--
	inst.dirty = true
	return true
}

func (inst *Instance) moveRight() bool {
	values, _, current, ok := inst.activePath()
	if !ok || len(current.Children) == 0 {
		return false
	}
	nextLevel := len(values)
	nextIndex := inst.highlightAt(nextLevel)
	if nextIndex < 0 || nextIndex >= len(current.Children) {
		nextIndex = firstEnabledIndex(current.Children)
	}
	if nextIndex < 0 {
		return false
	}
	changed := inst.setHighlight(nextLevel, nextIndex)
	if inst.activeLevel != nextLevel {
		inst.activeLevel = nextLevel
		changed = true
	}
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) setHighlight(level, index int) bool {
	if level < 0 {
		return false
	}
	next := make([]int, level+1)
	for i := range next {
		next[i] = -1
	}
	copy(next, inst.highlighted)
	if next[level] == index && len(inst.highlighted) == len(next) {
		return false
	}
	next[level] = index
	inst.highlighted = next
	return true
}

func (inst *Instance) highlightAt(level int) int {
	if level < 0 || level >= len(inst.highlighted) {
		return -1
	}
	return inst.highlighted[level]
}

func (inst *Instance) activePath() ([]string, []string, Option, bool) {
	if len(inst.options) == 0 || len(inst.highlighted) == 0 {
		return nil, nil, Option{}, false
	}
	currentOptions := inst.options
	limit := minInt(inst.activeLevel+1, len(inst.highlighted))
	if limit <= 0 {
		return nil, nil, Option{}, false
	}
	values := make([]string, 0, limit)
	labels := make([]string, 0, limit)
	var current Option
	for level := 0; level < limit; level++ {
		index := inst.highlighted[level]
		if index < 0 || index >= len(currentOptions) {
			return nil, nil, Option{}, false
		}
		current = currentOptions[index]
		values = append(values, current.Value)
		labels = append(labels, current.Label)
		currentOptions = current.Children
	}
	return values, labels, current, true
}

func (inst *Instance) optionAtLevel(level, index int) (Option, bool) {
	options := inst.optionsAtLevel(level)
	if index < 0 || index >= len(options) {
		return Option{}, false
	}
	return options[index], true
}

func (inst *Instance) optionsAtLevel(level int) []Option {
	if level < 0 {
		return nil
	}
	current := inst.options
	if level == 0 {
		return current
	}
	for depth := 0; depth < level; depth++ {
		index := inst.highlightAt(depth)
		if index < 0 || index >= len(current) {
			return nil
		}
		current = current[index].Children
	}
	return current
}

func (inst *Instance) popupColumns() []popupColumn {
	if len(inst.options) == 0 {
		return nil
	}
	selectedIndices, _, _ := resolvePath(inst.options, inst.selectedPath)
	columns := make([]popupColumn, 0, maxInt(1, len(inst.highlighted)))
	current := inst.options
	for level := 0; ; level++ {
		highlighted := -1
		if level < len(inst.highlighted) {
			highlighted = inst.highlighted[level]
		}
		selected := -1
		if level < len(selectedIndices) {
			selected = selectedIndices[level]
		}
		columns = append(columns, popupColumn{
			Level:       level,
			Options:     current,
			Highlighted: highlighted,
			Selected:    selected,
			Active:      level == inst.activeLevel,
		})
		if highlighted < 0 || highlighted >= len(current) || len(current[highlighted].Children) == 0 {
			break
		}
		current = current[highlighted].Children
	}
	return columns
}

func (inst *Instance) popupColumnWidths() []int {
	columns := inst.popupColumns()
	widths := make([]int, len(columns))
	for i, column := range columns {
		widths[i] = popupColumnWidth(column.Options)
	}
	return widths
}

func (inst *Instance) popupWidth() int {
	widths := inst.popupColumnWidths()
	if len(widths) == 0 {
		return 0
	}
	total := 1
	for _, width := range widths {
		total += width + 3
	}
	return total
}

func (inst *Instance) popupHeight() int {
	columns := inst.popupColumns()
	if len(columns) == 0 {
		return 0
	}
	rows := 0
	for _, column := range columns {
		rows = maxInt(rows, len(column.Options))
	}
	if rows == 0 {
		return 0
	}
	return rows + 2
}

func (inst *Instance) popupHit(localX, localY int) (level, index int, ok bool) {
	if !inst.open {
		return 0, 0, false
	}
	row := localY - 2
	if row < 0 {
		return 0, 0, false
	}
	columns := inst.popupColumns()
	widths := inst.popupColumnWidths()
	cursor := 0
	for colIndex, column := range columns {
		cellWidth := widths[colIndex] + 2
		contentStart := cursor + 1
		contentEnd := contentStart + cellWidth
		if localX >= contentStart && localX < contentEnd {
			if row < len(column.Options) {
				return column.Level, row, true
			}
			return 0, 0, false
		}
		cursor += cellWidth + 1
	}
	return 0, 0, false
}

func (inst *Instance) paintPopupAt(x, y int) []paint.DrawCmd {
	columns := inst.popupColumns()
	widths := inst.popupColumnWidths()
	if len(columns) == 0 || len(widths) == 0 {
		return nil
	}

	maxRows := 0
	for _, column := range columns {
		maxRows = maxInt(maxRows, len(column.Options))
	}
	if maxRows == 0 {
		return nil
	}

	borderStyle := inst.popupBorderStyle()
	cmds := []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  popupBorderLine(widths),
		Style: borderStyle,
	}}

	for row := 0; row < maxRows; row++ {
		rowY := y + 1 + row
		cursor := x
		for colIndex, column := range columns {
			contentWidth := widths[colIndex]
			cmds = append(cmds, paint.DrawCmd{
				X:     cursor,
				Y:     rowY,
				Text:  "|",
				Style: borderStyle,
			})

			cellText := " " + strings.Repeat(" ", contentWidth) + " "
			cellStyle := inst.popupFillStyle()
			if row < len(column.Options) {
				option := column.Options[row]
				selected := row == column.Selected
				highlighted := row == column.Highlighted
				cellText = " " + popupRowText(option, contentWidth, selected) + " "
				cellStyle = inst.popupRowStyle(option, highlighted, selected, column.Active)
			}
			cmds = append(cmds, paint.DrawCmd{
				X:     cursor + 1,
				Y:     rowY,
				Text:  cellText,
				Style: cellStyle,
			})
			cursor += contentWidth + 3
		}
		cmds = append(cmds, paint.DrawCmd{
			X:     cursor,
			Y:     rowY,
			Text:  "|",
			Style: borderStyle,
		})
	}

	cmds = append(cmds, paint.DrawCmd{
		X:     x,
		Y:     y + 1 + maxRows,
		Text:  popupBorderLine(widths),
		Style: borderStyle,
	})
	return cmds
}

func (inst *Instance) resolveStyle() style.Style {
	s := inst.cascaderStyle
	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}
	if inst.disabled {
		return s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	if inst.open || inst.focused {
		return s.Foreground(theme.Focus()).Bold(true)
	}
	return s
}

func (inst *Instance) popupBorderStyle() style.Style {
	if inst.disabled {
		return style.Style{}.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	return style.Style{}.Foreground(theme.Focus()).Background(theme.Surface())
}

func (inst *Instance) popupFillStyle() style.Style {
	if inst.disabled {
		return style.Style{}.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	return style.Style{}.Foreground(theme.Text()).Background(theme.Surface())
}

func (inst *Instance) popupRowStyle(option Option, highlighted, selected, active bool) style.Style {
	if option.Disabled {
		return style.Style{}.Foreground(theme.DisabledFG()).Background(theme.Surface())
	}
	if highlighted {
		return style.Style{}.Foreground(theme.BG()).Background(theme.Select()).Bold(true)
	}
	if selected {
		s := style.Style{}.Foreground(theme.Focus()).Background(theme.Surface()).Bold(true)
		if active {
			return s.Underline(true)
		}
		return s
	}
	return inst.popupFillStyle()
}

func (inst *Instance) triggerWidth() int {
	label := inst.triggerLabel()
	width := maxInt(defaultTriggerWidth, paint.StringWidth(label)+4)
	if inst.width > 0 {
		width = inst.width
	}
	return maxInt(width, 8)
}

func (inst *Instance) triggerPaintWidth() int {
	if inst.bounds[2] > 0 {
		return maxInt(inst.bounds[2], 8)
	}
	return inst.triggerWidth()
}

func (inst *Instance) triggerText(width int) string {
	innerWidth := maxInt(0, width-4)
	label := truncateWithEllipsis(inst.triggerLabel(), innerWidth)
	return "< " + padDisplayWidth(label, innerWidth) + " >"
}

func (inst *Instance) triggerLabel() string {
	labels := inst.SelectedLabels()
	if len(labels) == 0 {
		return inst.placeholder
	}
	return strings.Join(labels, inst.separator)
}

func (inst *Instance) normalizeConfiguration() {
	if inst.width <= 0 {
		inst.width = defaultTriggerWidth
	}
	if strings.TrimSpace(inst.placeholder) == "" {
		inst.placeholder = defaultPlaceholder
	}
	if inst.separator == "" {
		inst.separator = defaultSeparator
	}
	inst.selectedPath = normalizePathValues(inst.options, inst.selectedPath)
	if len(inst.options) == 0 {
		inst.open = false
		inst.activeLevel = 0
		inst.highlighted = nil
		return
	}
	if inst.activeLevel < 0 {
		inst.activeLevel = 0
	}
}

func (inst *Instance) resetNavigationFromSelection() {
	if len(inst.options) == 0 {
		inst.highlighted = nil
		inst.activeLevel = 0
		return
	}
	indices, _, _ := resolvePath(inst.options, inst.selectedPath)
	if len(indices) > 0 {
		inst.highlighted = append([]int(nil), indices...)
		inst.activeLevel = len(indices) - 1
		return
	}
	first := firstEnabledIndex(inst.options)
	if first < 0 {
		inst.highlighted = nil
		inst.activeLevel = 0
		return
	}
	inst.highlighted = []int{first}
	inst.activeLevel = 0
}

func (inst *Instance) emitChange() {
	inst.emitFieldValueChanged()
	inst.emitCascaderChange()
}

func (inst *Instance) emitFieldValueChanged() {
	if inst.intentEmitter == nil {
		return
	}
	value := inst.SelectedValue()
	if inst.formID != "" && inst.changeIntentField != nil {
		runtimeintent.Emit(inst, form.FieldChange(inst.formID, inst.changeIntentField.GetField(), value, true))
		return
	}
	if inst.changeIntentField != nil {
		inst.intentEmitter(runtimeintent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: value,
		})
		return
	}
	if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
}

func (inst *Instance) emitFieldBlur() {
	if inst.intentEmitter == nil || inst.formID == "" || inst.changeIntentField == nil {
		return
	}
	runtimeintent.Emit(inst, form.FieldBlur(inst.formID, inst.changeIntentField.GetField(), inst.SelectedValue()))
}

func (inst *Instance) emitCascaderChange() {
	if inst.intentEmitter == nil {
		return
	}
	values := inst.SelectedValues()
	labels := inst.SelectedLabels()
	inst.intentEmitter(ChangeWithID(
		inst.componentID,
		strings.Join(values, defaultFieldSeparator),
		strings.Join(labels, inst.separator),
		values,
		labels,
	))
}

func popupBorderLine(widths []int) string {
	var builder strings.Builder
	builder.WriteString("+")
	for _, width := range widths {
		builder.WriteString(strings.Repeat("-", width+2))
		builder.WriteString("+")
	}
	return builder.String()
}

func popupColumnWidth(options []Option) int {
	width := 8
	for _, option := range options {
		width = maxInt(width, paint.StringWidth(option.Label)+4)
	}
	return width
}

func popupRowText(option Option, width int, selected bool) string {
	marker := "  "
	if selected {
		marker = "* "
	}
	suffix := "  "
	if len(option.Children) > 0 {
		suffix = " >"
	}
	labelWidth := maxInt(1, width-paint.StringWidth(marker)-paint.StringWidth(suffix))
	label := truncateWithEllipsis(option.Label, labelWidth)
	return marker + padDisplayWidth(label, labelWidth) + suffix
}

func resolvePath(options []Option, values []string) ([]int, []string, bool) {
	if len(values) == 0 {
		return nil, nil, false
	}
	current := options
	indices := make([]int, 0, len(values))
	labels := make([]string, 0, len(values))
	for _, value := range values {
		index := findOptionIndex(current, value)
		if index < 0 {
			break
		}
		option := current[index]
		indices = append(indices, index)
		labels = append(labels, option.Label)
		current = option.Children
	}
	if len(indices) == 0 {
		return nil, nil, false
	}
	return indices, labels, true
}

func normalizePathValues(options []Option, values []string) []string {
	if len(values) == 0 {
		return nil
	}
	current := options
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		index := findOptionIndex(current, value)
		if index < 0 {
			break
		}
		normalized = append(normalized, current[index].Value)
		current = current[index].Children
	}
	return normalized
}

func findOptionIndex(options []Option, value string) int {
	for index, option := range options {
		if option.Value == value {
			return index
		}
	}
	return -1
}

func firstEnabledIndex(options []Option) int {
	for index, option := range options {
		if !option.Disabled {
			return index
		}
	}
	return -1
}

func lastEnabledIndex(options []Option) int {
	for index := len(options) - 1; index >= 0; index-- {
		if !options[index].Disabled {
			return index
		}
	}
	return -1
}

func nextEnabledIndex(options []Option, current, delta int) int {
	if len(options) == 0 {
		return -1
	}
	if current < 0 || current >= len(options) {
		if delta >= 0 {
			return firstEnabledIndex(options)
		}
		return lastEnabledIndex(options)
	}
	for step := 1; step <= len(options); step++ {
		index := current + (step * delta)
		for index < 0 {
			index += len(options)
		}
		index %= len(options)
		if !options[index].Disabled {
			return index
		}
	}
	if !options[current].Disabled {
		return current
	}
	return -1
}

func getOptionsProp(props rtui.Props) []Option {
	if value, ok := props[propOptions]; ok {
		if options, ok := value.([]Option); ok {
			return append([]Option(nil), options...)
		}
	}
	return nil
}

func getOptionsPropOr(props rtui.Props, def []Option) []Option {
	if value, ok := props[propOptions]; ok {
		if options, ok := value.([]Option); ok {
			return append([]Option(nil), options...)
		}
	}
	return append([]Option(nil), def...)
}

func getStringSliceProp(props rtui.Props, key string, def []string) []string {
	if value, ok := props[key]; ok {
		if slice, ok := value.([]string); ok {
			return append([]string(nil), slice...)
		}
	}
	return append([]string(nil), def...)
}

func getChangeIntentFieldProp(props rtui.Props, key string) runtimeintent.FieldIntent {
	if value, ok := props[key]; ok {
		if fieldIntent, ok := value.(runtimeintent.FieldIntent); ok {
			return fieldIntent
		}
	}
	return nil
}

func mousePayload(payload any) (*runtimemsg.MouseMsg, bool) {
	switch value := payload.(type) {
	case *runtimemsg.MouseMsg:
		if value != nil {
			return value, true
		}
	case runtimemsg.MouseMsg:
		copy := value
		return &copy, true
	}
	return nil, false
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func truncateWithEllipsis(content string, width int) string {
	if width <= 0 {
		return ""
	}
	const ellipsis = "..."
	if paint.StringWidth(content) <= width {
		return content
	}
	if width <= len(ellipsis) {
		return ellipsis[:width]
	}
	trimmed := truncateByDisplayWidth(content, width-len(ellipsis))
	return strings.TrimRight(trimmed, " ") + ellipsis
}

func truncateByDisplayWidth(content string, width int) string {
	if width <= 0 {
		return ""
	}
	var builder strings.Builder
	currentWidth := 0
	for _, r := range content {
		runeWidth := paint.RuneWidth(r)
		if currentWidth+runeWidth > width {
			break
		}
		builder.WriteRune(r)
		currentWidth += runeWidth
	}
	return builder.String()
}

func padDisplayWidth(content string, width int) string {
	content = truncateByDisplayWidth(content, width)
	padding := width - paint.StringWidth(content)
	if padding <= 0 {
		return content
	}
	return content + strings.Repeat(" ", padding)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
