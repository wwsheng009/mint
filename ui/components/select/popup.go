package selectcomp

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type popupVNode struct {
	*rtui.ElementVNode
}

func (v *popupVNode) SetProps(p rtui.Props) rtui.VNode {
	v.ElementVNode.SetProps(v.ElementVNode.Props().Merge(p))
	return v
}

func (v *popupVNode) GetLayer() rtui.Layer {
	return rtui.LayerOverlay
}

func (v *popupVNode) SetLayer(l rtui.Layer) rtui.VNode {
	v.SetProp("_layer", l)
	return v
}

func (v *popupVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props["key"] = v.Key()
	return newPopupInstance(props)
}

type popupInstance struct {
	key               string
	parent            rtui.ComponentInstance
	selectID          string
	componentID       string
	options           []Option
	popupStyle        style.Style
	selectionMode     SelectionMode
	selectedIndex     int
	selectedIndices   []int
	highlightedIndex  int
	scrollOffset      int
	filterOption      bool
	filterPlaceholder string
	filterQuery       string
	maxVisibleRows    int
	minWidth          int
	disabled          bool
	closeOnOutside    bool
	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent
	formID            string
	focused           bool
	bounds            [4]int
	dirty             bool
	intentEmitter     func(intent.Intent)
	overlayCallbacks  *overlayCallbacks
}

var (
	_ rtui.ComponentInstance     = (*popupInstance)(nil)
	_ rtui.PaintableInstance     = (*popupInstance)(nil)
	_ rtui.ActionHandlerInstance = (*popupInstance)(nil)
	_ rtui.FocusableInstance     = (*popupInstance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*popupInstance)(nil)
	_ interface {
		Parent() interface{}
	} = (*popupInstance)(nil)
	_ selectIntentSource = (*popupInstance)(nil)
)

func newPopupInstance(props rtui.Props) *popupInstance {
	inst := &popupInstance{dirty: true}
	inst.SetProps(props)
	return inst
}

func (inst *popupInstance) Key() string                        { return inst.key }
func (inst *popupInstance) SetKey(key string)                  { inst.key = key }
func (inst *popupInstance) Init(props rtui.Props)              { inst.SetProps(props) }
func (inst *popupInstance) Destroy()                           { selectPopupRegistry.unregister(inst) }
func (inst *popupInstance) OnMount()                           { selectPopupRegistry.register(inst); inst.dirty = true }
func (inst *popupInstance) OnUnmount()                         { selectPopupRegistry.unregister(inst) }
func (inst *popupInstance) MarkDirty()                         { inst.dirty = true }
func (inst *popupInstance) IsDirty() bool                      { return inst.dirty }
func (inst *popupInstance) GetContext() *rtui.ComponentContext { return nil }
func (inst *popupInstance) Parent() interface{}                { return inst.parent }
func (inst *popupInstance) SetParent(parent rtui.ComponentInstance) {
	inst.parent = parent
}

func (inst *popupInstance) rows() popupRows {
	return buildPopupRows(
		inst.options,
		inst.selectionMode,
		inst.filterOption,
		inst.filterPlaceholder,
		inst.filterQuery,
	)
}

func (inst *popupInstance) SetProps(props rtui.Props) bool {
	oldSelectID := inst.selectID
	oldOptions := append([]Option(nil), inst.options...)
	oldFilterOption := inst.filterOption
	oldFilterPlaceholder := inst.filterPlaceholder
	oldHighlight := inst.highlightedIndex
	oldScroll := inst.scrollOffset
	oldSelected := inst.selectedIndex
	oldSelectedIndices := append([]int(nil), inst.selectedIndices...)
	oldFilterQuery := inst.filterQuery

	inst.key = proputil.GetString(props, "key", inst.key)
	inst.selectID = proputil.GetString(props, "selectID", inst.selectID)
	inst.componentID = proputil.GetString(props, "componentID", inst.componentID)
	inst.options = getOptionsProp(props)
	inst.popupStyle = proputil.GetStyle(props, "style", style.Style{})
	inst.selectionMode = getSelectionModeProp(props, inst.selectionMode)
	inst.selectedIndex = proputil.GetInt(props, "selectedIndex", inst.selectedIndex)
	inst.selectedIndices = getIntsProp(props, "selectedIndices", inst.selectedIndices)
	inst.highlightedIndex = proputil.GetInt(props, "highlightedIndex", inst.highlightedIndex)
	inst.scrollOffset = proputil.GetInt(props, "scrollOffset", inst.scrollOffset)
	inst.filterOption = proputil.GetBool(props, propFilterOption, inst.filterOption)
	inst.filterPlaceholder = proputil.GetString(props, propFilterPlaceholder, inst.filterPlaceholder)
	inst.filterQuery = sanitizeFilterText(proputil.GetString(props, "filterQuery", inst.filterQuery))
	inst.maxVisibleRows = proputil.GetInt(props, "maxVisibleRows", inst.maxVisibleRows)
	inst.minWidth = proputil.GetInt(props, "minWidth", inst.minWidth)
	inst.disabled = proputil.GetBool(props, "disabled", inst.disabled)
	inst.closeOnOutside = proputil.GetBool(props, "closeOnOutside", inst.closeOnOutside)
	inst.changeIntent = proputil.GetIntent(props, "changeIntent", nil)
	inst.changeIntentField = getChangeIntentFieldProp(props, "changeIntent")
	inst.formID = proputil.GetString(props, "formID", inst.formID)
	inst.overlayCallbacks = getOverlayCallbacksProp(props)

	inst.selectedIndex, inst.selectedIndices = normalizeOverlaySelection(
		inst.selectionMode,
		inst.selectedIndex,
		inst.selectedIndices,
		len(inst.options),
	)
	if inst.maxVisibleRows <= 0 {
		inst.maxVisibleRows = defaultMaxVisibleRows
	}
	if isTagsSelectionMode(inst.selectionMode) {
		inst.filterOption = true
	}
	inst.highlightedIndex, inst.scrollOffset = normalizePopupHighlight(
		inst.rows(),
		inst.highlightedIndex,
		inst.scrollOffset,
		inst.maxVisibleRows,
		inst.selectionMode,
		inst.selectedIndex,
		inst.selectedIndices,
	)

	if oldSelectID != "" && oldSelectID != inst.selectID {
		selectPopupRegistry.unregister(inst)
	}
	if inst.selectID != "" {
		selectPopupRegistry.register(inst)
	}

	changed := oldSelectID != inst.selectID ||
		!equalOptions(oldOptions, inst.options) ||
		oldFilterOption != inst.filterOption ||
		oldFilterPlaceholder != inst.filterPlaceholder ||
		oldHighlight != inst.highlightedIndex ||
		oldScroll != inst.scrollOffset ||
		oldSelected != inst.selectedIndex ||
		oldFilterQuery != inst.filterQuery ||
		!equalIntSlices(oldSelectedIndices, inst.selectedIndices)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *popupInstance) GetProps() rtui.Props {
	return rtui.Props{
		"key":                 inst.key,
		"selectID":            inst.selectID,
		"componentID":         inst.componentID,
		"options":             append([]Option(nil), inst.options...),
		"style":               inst.popupStyle,
		"selectionMode":       inst.selectionMode,
		"selectedIndex":       inst.selectedIndex,
		"selectedIndices":     append([]int(nil), inst.selectedIndices...),
		"highlightedIndex":    inst.highlightedIndex,
		"scrollOffset":        inst.scrollOffset,
		propFilterOption:      inst.filterOption,
		propFilterPlaceholder: inst.filterPlaceholder,
		"filterQuery":         inst.filterQuery,
		"maxVisibleRows":      inst.maxVisibleRows,
		"minWidth":            inst.minWidth,
		"disabled":            inst.disabled,
		"closeOnOutside":      inst.closeOnOutside,
		"changeIntent":        inst.changeIntent,
		"formID":              inst.formID,
		overlayCallbacksProp:  inst.overlayCallbacks,
	}
}

func (inst *popupInstance) Measure(constraints layout.Constraints) layout.Size {
	rows := inst.rows()
	if !rows.showFilter && len(rows.scrollable) == 0 {
		return layout.Size{}
	}
	width := constraints.ConstrainWidth(inst.popupWidth())
	height := constraints.ConstrainHeight(inst.popupHeight())
	return layout.Size{Width: width, Height: height}
}

func (inst *popupInstance) Paint(x, y int) []paint.DrawCmd {
	rows := inst.rows()
	if !rows.showFilter && len(rows.scrollable) == 0 {
		return nil
	}
	return inst.paintPopupAt(x, y)
}

func (inst *popupInstance) HandleAction(act *action.Action) bool {
	if act == nil || inst.disabled {
		return false
	}

	switch act.Type {
	case action.ActionInputChar:
		if value, ok := act.GetPayloadRune(); ok {
			return inst.appendFilterText(string(value))
		}
		return false
	case action.ActionInputText:
		if value, ok := act.GetPayloadString(); ok {
			return inst.appendFilterText(value)
		}
		return false
	case action.ActionBackspace:
		return inst.backspaceFilterText()
	case action.ActionDeleteChar, action.ActionClear:
		return inst.clearFilterText()
	case action.ActionCancel:
		return inst.requestClose()
	}

	rows := inst.rows()
	if len(selectableTargets(rows)) == 0 {
		return false
	}

	switch act.Type {
	case action.ActionNavigateDown:
		return inst.setHighlight(nextHighlightTarget(rows, inst.highlightedIndex, 1))
	case action.ActionNavigateUp:
		return inst.setHighlight(nextHighlightTarget(rows, inst.highlightedIndex, -1))
	case action.ActionNavigateHome:
		return inst.setHighlight(firstSelectableTarget(rows))
	case action.ActionNavigateEnd:
		return inst.setHighlight(lastSelectableTarget(rows))
	case action.ActionNavigatePageUp:
		return inst.setHighlight(pageHighlightTarget(rows, inst.highlightedIndex, -1, maxInt(1, inst.visibleRowCount())))
	case action.ActionNavigatePageDown:
		return inst.setHighlight(pageHighlightTarget(rows, inst.highlightedIndex, 1, maxInt(1, inst.visibleRowCount())))
	case action.ActionHover:
		mouse, ok := popupMousePayload(act.Payload)
		if !ok {
			return false
		}
		index, hit := inst.optionIndexAt(mouse.LocalX, mouse.LocalY)
		if !hit {
			return false
		}
		return inst.setHighlight(index)
	case action.ActionClick:
		mouse, ok := popupMousePayload(act.Payload)
		if ok {
			if log.RenderLogger.Enabled() {
				log.RenderLogger.Debug("[SelectPopup] ActionClick select=%s screen=(%d,%d) local=(%d,%d) highlight=%d",
					inst.selectID, mouse.X, mouse.Y, mouse.LocalX, mouse.LocalY, inst.highlightedIndex)
			}
			index, hit := inst.optionIndexAt(mouse.LocalX, mouse.LocalY)
			if !hit {
				return true
			}
			inst.setHighlight(index)
			return inst.commit(index)
		}
		return inst.commit(inst.highlightedIndex)
	case action.ActionMouseRelease:
		mouse, ok := popupMousePayload(act.Payload)
		if ok {
			if log.RenderLogger.Enabled() {
				log.RenderLogger.Debug("[SelectPopup] ActionMouseRelease select=%s screen=(%d,%d) local=(%d,%d) highlight=%d",
					inst.selectID, mouse.X, mouse.Y, mouse.LocalX, mouse.LocalY, inst.highlightedIndex)
			}
			index, hit := inst.optionIndexAt(mouse.LocalX, mouse.LocalY)
			if !hit {
				return true
			}
			inst.setHighlight(index)
			return inst.commit(index)
		}
		return inst.commit(inst.highlightedIndex)
	case action.ActionScroll:
		delta, ok := scrollDeltaFromAction(act)
		if !ok || delta == 0 {
			return false
		}
		return inst.setHighlight(nextHighlightTarget(rows, inst.highlightedIndex, scrollDirection(delta)))
	case action.ActionSelect, action.ActionEnter, action.ActionSubmit:
		return inst.commit(inst.highlightedIndex)
	}
	return false
}

func (inst *popupInstance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *popupInstance) SetFocus(focused bool) {
	if inst.focused == focused {
		return
	}
	inst.focused = focused
	inst.dirty = true
}

func (inst *popupInstance) HasFocus() bool {
	return inst.focused
}

func (inst *popupInstance) IsDisabled() bool {
	return inst.disabled || len(selectableTargets(inst.rows())) == 0
}

func (inst *popupInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *popupInstance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

func (inst *popupInstance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil {
		inst.intentEmitter(i)
	}
}

func (inst *popupInstance) containsPoint(screenX, screenY int) bool {
	x, y, width, height := inst.GetBounds()
	return screenX >= x && screenX < x+width && screenY >= y && screenY < y+height
}

func (inst *popupInstance) popupWidth() int {
	contentWidth := popupContentWidth(inst.rows(), inst.selectionMode)
	return maxInt(inst.minWidth, maxInt(contentWidth+2, 6))
}

func (inst *popupInstance) popupHeight() int {
	rows := inst.rows()
	if !rows.showFilter && len(rows.scrollable) == 0 {
		return 0
	}
	height := inst.visibleRowCount() + 2
	if rows.showFilter {
		height++
	}
	return height
}

func (inst *popupInstance) visibleRowCount() int {
	return visibleScrollableRowCount(inst.rows(), inst.maxVisibleRows)
}

func (inst *popupInstance) maxScrollOffset() int {
	return maxScrollOffsetForRows(inst.rows(), inst.maxVisibleRows)
}

func (inst *popupInstance) setFilterQuery(query string) bool {
	if !filterEnabledFor(inst.selectionMode, inst.filterOption) {
		return false
	}

	query = sanitizeFilterText(query)
	if inst.overlayCallbacks != nil {
		state := inst.overlayCallbacks.setFilterQuery(query)
		oldHighlight := inst.highlightedIndex
		oldScroll := inst.scrollOffset
		oldFilterQuery := inst.filterQuery
		oldOptions := append([]Option(nil), inst.options...)
		inst.selectedIndex = state.selectedIndex
		inst.selectedIndices = append([]int(nil), state.selectedIndices...)
		inst.highlightedIndex = state.highlightedIndex
		inst.scrollOffset = state.scrollOffset
		if state.options != nil {
			inst.options = append([]Option(nil), state.options...)
		}
		inst.filterQuery = state.filterQuery
		inst.dirty = oldHighlight != inst.highlightedIndex ||
			oldScroll != inst.scrollOffset ||
			oldFilterQuery != inst.filterQuery ||
			!equalOptions(oldOptions, inst.options)
		return inst.dirty
	}

	if inst.filterQuery == query {
		return false
	}
	inst.filterQuery = query
	inst.ensureHighlightVisible()
	inst.dirty = true
	return true
}

func (inst *popupInstance) appendFilterText(text string) bool {
	if !filterEnabledFor(inst.selectionMode, inst.filterOption) {
		return false
	}
	text = sanitizeFilterText(text)
	if strings.TrimSpace(text) == "" {
		return false
	}
	return inst.setFilterQuery(inst.filterQuery + text)
}

func (inst *popupInstance) backspaceFilterText() bool {
	if !filterEnabledFor(inst.selectionMode, inst.filterOption) || inst.filterQuery == "" {
		return false
	}
	runes := []rune(inst.filterQuery)
	if len(runes) == 0 {
		return false
	}
	return inst.setFilterQuery(string(runes[:len(runes)-1]))
}

func (inst *popupInstance) clearFilterText() bool {
	if !filterEnabledFor(inst.selectionMode, inst.filterOption) || inst.filterQuery == "" {
		return false
	}
	return inst.setFilterQuery("")
}

func (inst *popupInstance) setHighlight(index int) bool {
	rows := inst.rows()
	if len(selectableTargets(rows)) == 0 {
		return false
	}
	if rowPositionForHighlight(rows, index) < 0 {
		index = defaultHighlightTarget(rows, inst.selectionMode, inst.selectedIndex, inst.selectedIndices)
	}
	if inst.overlayCallbacks != nil {
		state := inst.overlayCallbacks.setHighlight(index)
		oldHighlight := inst.highlightedIndex
		oldScroll := inst.scrollOffset
		oldFilterQuery := inst.filterQuery
		oldOptions := append([]Option(nil), inst.options...)
		inst.selectedIndex = state.selectedIndex
		inst.selectedIndices = append([]int(nil), state.selectedIndices...)
		inst.highlightedIndex = state.highlightedIndex
		inst.scrollOffset = state.scrollOffset
		if state.options != nil {
			inst.options = append([]Option(nil), state.options...)
		}
		inst.filterQuery = state.filterQuery
		changed := oldHighlight != inst.highlightedIndex ||
			oldScroll != inst.scrollOffset ||
			oldFilterQuery != inst.filterQuery ||
			!equalOptions(oldOptions, inst.options)
		inst.dirty = changed
		return changed
	}
	if index == inst.highlightedIndex {
		return false
	}
	inst.highlightedIndex = index
	inst.ensureHighlightVisible()
	inst.dirty = true
	return true
}

func (inst *popupInstance) commit(index int) bool {
	if index < 0 || index >= len(inst.options) {
		if index != highlightCreateTag {
			return false
		}
	}
	oldSelectedIndex := inst.selectedIndex
	oldSelectedIndices := append([]int(nil), inst.selectedIndices...)
	oldOptions := append([]Option(nil), inst.options...)
	closed := false
	if inst.overlayCallbacks != nil {
		var state overlayControllerState
		if index == highlightCreateTag {
			state = inst.overlayCallbacks.createTag()
		} else {
			state = inst.overlayCallbacks.commit(index)
		}
		inst.selectedIndex = state.selectedIndex
		inst.selectedIndices = append([]int(nil), state.selectedIndices...)
		inst.highlightedIndex = state.highlightedIndex
		inst.scrollOffset = state.scrollOffset
		if state.options != nil {
			inst.options = append([]Option(nil), state.options...)
		}
		inst.filterQuery = state.filterQuery
		inst.dirty = true
		closed = !state.open
	} else {
		if index == highlightCreateTag {
			return false
		}
		nextIndex, nextIndices, _, shouldClose := applyOverlayCommit(
			inst.selectionMode,
			len(inst.options),
			inst.selectedIndex,
			inst.selectedIndices,
			index,
		)
		inst.selectedIndex = nextIndex
		inst.selectedIndices = nextIndices
		inst.highlightedIndex = clampIndexForOptions(index, len(inst.options))
		inst.ensureHighlightVisible()
		inst.dirty = true
		closed = shouldClose
	}

	selectionChanged := oldSelectedIndex != inst.selectedIndex ||
		!equalIntSlices(oldSelectedIndices, inst.selectedIndices)
	optionsChanged := !equalOptions(oldOptions, inst.options)
	if selectionChanged {
		emitFieldValueChangedFrom(
			inst,
			inst.changeIntent,
			inst.changeIntentField,
			inst.formID,
			inst.selectionMode,
			inst.selectedIndex,
			inst.selectedIndices,
			inst.options,
		)
		emitSelectChangeFrom(
			inst,
			inst.componentID,
			inst.selectionMode,
			inst.options,
			inst.selectedIndex,
			inst.selectedIndices,
		)
	}
	return selectionChanged || optionsChanged || closed
}

func (inst *popupInstance) requestClose() bool {
	if inst.overlayCallbacks != nil {
		wasOpen := true
		state := inst.overlayCallbacks.setOpen(false)
		inst.selectedIndex = state.selectedIndex
		inst.selectedIndices = append([]int(nil), state.selectedIndices...)
		inst.highlightedIndex = state.highlightedIndex
		inst.scrollOffset = state.scrollOffset
		if state.options != nil {
			inst.options = append([]Option(nil), state.options...)
		}
		inst.filterQuery = state.filterQuery
		inst.dirty = true
		return wasOpen && !state.open
	}
	inst.dirty = true
	return true
}

func (inst *popupInstance) ensureHighlightVisible() {
	inst.highlightedIndex, inst.scrollOffset = normalizePopupHighlight(
		inst.rows(),
		inst.highlightedIndex,
		inst.scrollOffset,
		inst.maxVisibleRows,
		inst.selectionMode,
		inst.selectedIndex,
		inst.selectedIndices,
	)
}

func (inst *popupInstance) optionIndexAt(localX, localY int) (int, bool) {
	if localX < 0 || localY < 1 {
		return 0, false
	}
	return popupHitTarget(inst.rows(), inst.scrollOffset, inst.maxVisibleRows, localY)
}

func (inst *popupInstance) paintPopupAt(x, y int) []paint.DrawCmd {
	rows := inst.rows()
	popupWidth := inst.popupWidth()
	popupHeight := inst.popupHeight()
	if popupWidth < 4 || popupHeight < 3 {
		return nil
	}

	borderStyle := popupBorderStyleFor(inst.popupStyle, inst.disabled)
	fillStyle := popupFillStyleFor(inst.popupStyle, inst.disabled)
	contentWidth := popupWidth - 2
	top := "┌" + strings.Repeat("─", maxInt(0, popupWidth-2)) + "┐"
	bottom := "└" + strings.Repeat("─", maxInt(0, popupWidth-2)) + "┘"
	cmds := []paint.DrawCmd{{X: x, Y: y, Text: top, Style: borderStyle}}

	contentY := y + 1
	if rows.showFilter {
		cmds = append(cmds,
			paint.DrawCmd{X: x, Y: contentY, Text: "│", Style: borderStyle},
			paint.DrawCmd{X: x + 1, Y: contentY, Text: strings.Repeat(" ", contentWidth), Style: fillStyle},
			paint.DrawCmd{X: x + popupWidth - 1, Y: contentY, Text: "│", Style: borderStyle},
			paint.DrawCmd{
				X:     x + 1,
				Y:     contentY,
				Text:  padDisplayWidth(truncateWithEllipsis(popupFilterText(rows.filterQuery, rows.filterPlaceholder), contentWidth), contentWidth),
				Style: fillStyle,
			},
		)
		contentY++
	}

	visibleRows := inst.visibleRowCount()
	for rowOffset := 0; rowOffset < visibleRows; rowOffset++ {
		rowY := contentY + rowOffset
		cmds = append(cmds,
			paint.DrawCmd{X: x, Y: rowY, Text: "│", Style: borderStyle},
			paint.DrawCmd{X: x + 1, Y: rowY, Text: strings.Repeat(" ", contentWidth), Style: fillStyle},
			paint.DrawCmd{X: x + popupWidth - 1, Y: rowY, Text: "│", Style: borderStyle},
		)
		rowIndex := inst.scrollOffset + rowOffset
		if rowIndex >= len(rows.scrollable) {
			continue
		}
		row := rows.scrollable[rowIndex]
		rowStyle := popupOptionRowStyleFor(inst.popupStyle, inst.disabled, row, isHighlightedTarget(row, inst.highlightedIndex))
		rowText := popupRowText(row, contentWidth, inst.selectionMode, inst.selectedIndices)
		cmds = append(cmds, paint.DrawCmd{
			X:     x + 1,
			Y:     rowY,
			Text:  rowText,
			Style: rowStyle,
		})
	}

	cmds = append(cmds, paint.DrawCmd{X: x, Y: y + popupHeight - 1, Text: bottom, Style: borderStyle})
	if inst.scrollOffset > 0 {
		cmds = append(cmds, paint.DrawCmd{X: x + popupWidth - 2, Y: y, Text: "↑", Style: borderStyle})
	}
	if inst.scrollOffset < inst.maxScrollOffset() {
		cmds = append(cmds, paint.DrawCmd{X: x + popupWidth - 2, Y: y + popupHeight - 1, Text: "↓", Style: borderStyle})
	}
	return cmds
}

func (inst *popupInstance) optionRowText(index, width int) string {
	if index < 0 || index >= len(inst.options) {
		return strings.Repeat(" ", width)
	}
	marker := inst.optionMarker(index)
	labelWidth := width
	rowText := ""
	if marker != "" {
		rowText = padDisplayWidth(marker, inst.optionMarkerWidth())
		if width > inst.optionMarkerWidth() {
			rowText += " "
			labelWidth = width - inst.optionMarkerWidth() - 1
		} else {
			labelWidth = 0
		}
	}
	rowText += padDisplayWidth(truncateWithEllipsis(inst.options[index].Label, labelWidth), labelWidth)
	return padDisplayWidth(rowText, width)
}

func (inst *popupInstance) optionMarker(index int) string {
	return optionMarkerForMode(inst.selectionMode, inst.selectedIndices, index)
}

func (inst *popupInstance) optionMarkerWidth() int {
	return markerWidthForMode(inst.selectionMode)
}

func popupFillStyleFor(base style.Style, disabled bool) style.Style {
	s := base
	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}
	if disabled {
		return s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	return s.Background(theme.Surface())
}

func popupBorderStyleFor(base style.Style, disabled bool) style.Style {
	s := style.Style{}.Foreground(theme.Focus()).Background(theme.Surface())
	if disabled {
		return s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	return s
}

func popupOptionRowStyleFor(base style.Style, disabled bool, row popupRow, highlighted bool) style.Style {
	s := popupFillStyleFor(base, disabled)
	if disabled {
		return s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	switch row.kind {
	case popupRowGroup:
		return s.Foreground(theme.Focus()).Bold(true)
	case popupRowCreateTag:
		if highlighted {
			return s.Foreground(theme.BG()).Background(theme.Select()).Bold(true)
		}
		return s.Foreground(theme.Focus()).Bold(true)
	case popupRowEmpty:
		return s.Foreground(theme.DisabledFG())
	}
	if highlighted {
		return s.Foreground(theme.BG()).Background(theme.Select()).Bold(true)
	}
	return s
}

func popupMousePayload(payload any) (*runtimemsg.MouseMsg, bool) {
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

func scrollDeltaFromAction(act *action.Action) (int, bool) {
	if act == nil {
		return 0, false
	}
	switch value := act.Payload.(type) {
	case *runtimemsg.MouseMsg:
		if value != nil && value.Delta != 0 {
			return value.Delta, true
		}
	case runtimemsg.MouseMsg:
		if value.Delta != 0 {
			return value.Delta, true
		}
	}
	return 0, false
}

func scrollDirection(delta int) int {
	if delta > 0 {
		return 1
	}
	if delta < 0 {
		return -1
	}
	return 0
}
