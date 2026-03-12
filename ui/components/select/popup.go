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
	key              string
	parent           rtui.ComponentInstance
	selectID         string
	componentID      string
	options          []Option
	popupStyle       style.Style
	selectionMode    SelectionMode
	selectedIndex    int
	selectedIndices  []int
	highlightedIndex int
	scrollOffset     int
	maxVisibleRows   int
	minWidth         int
	disabled         bool
	closeOnOutside   bool
	changeIntent     intent.Intent
	changeIntentField intent.FieldIntent
	formID           string
	focused          bool
	bounds           [4]int
	dirty            bool
	intentEmitter    func(intent.Intent)
	overlayCallbacks *overlayCallbacks
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

func (inst *popupInstance) SetProps(props rtui.Props) bool {
	oldSelectID := inst.selectID
	oldHighlight := inst.highlightedIndex
	oldScroll := inst.scrollOffset
	oldSelected := inst.selectedIndex
	oldSelectedIndices := append([]int(nil), inst.selectedIndices...)

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
	if len(inst.options) == 0 {
		inst.highlightedIndex = -1
		inst.scrollOffset = 0
	} else {
		if inst.highlightedIndex < 0 || inst.highlightedIndex >= len(inst.options) {
			inst.highlightedIndex = defaultOverlayHighlight(
				inst.selectedIndex,
				inst.selectedIndices,
				inst.selectionMode,
				len(inst.options),
			)
		}
		inst.highlightedIndex = clampInt(inst.highlightedIndex, 0, len(inst.options)-1)
		inst.scrollOffset = clampInt(inst.scrollOffset, 0, inst.maxScrollOffset())
	}

	if oldSelectID != "" && oldSelectID != inst.selectID {
		selectPopupRegistry.unregister(inst)
	}
	if inst.selectID != "" {
		selectPopupRegistry.register(inst)
	}

	changed := oldSelectID != inst.selectID ||
		oldHighlight != inst.highlightedIndex ||
		oldScroll != inst.scrollOffset ||
		oldSelected != inst.selectedIndex ||
		!equalIntSlices(oldSelectedIndices, inst.selectedIndices)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *popupInstance) GetProps() rtui.Props {
	return rtui.Props{
		"key":               inst.key,
		"selectID":          inst.selectID,
		"componentID":       inst.componentID,
		"options":           append([]Option(nil), inst.options...),
		"style":             inst.popupStyle,
		"selectionMode":     inst.selectionMode,
		"selectedIndex":     inst.selectedIndex,
		"selectedIndices":   append([]int(nil), inst.selectedIndices...),
		"highlightedIndex":  inst.highlightedIndex,
		"scrollOffset":      inst.scrollOffset,
		"maxVisibleRows":    inst.maxVisibleRows,
		"minWidth":          inst.minWidth,
		"disabled":          inst.disabled,
		"closeOnOutside":    inst.closeOnOutside,
		"changeIntent":      inst.changeIntent,
		"formID":            inst.formID,
		overlayCallbacksProp: inst.overlayCallbacks,
	}
}

func (inst *popupInstance) Measure(constraints layout.Constraints) layout.Size {
	if len(inst.options) == 0 {
		return layout.Size{}
	}
	width := constraints.ConstrainWidth(inst.popupWidth())
	height := constraints.ConstrainHeight(inst.popupHeight())
	return layout.Size{Width: width, Height: height}
}

func (inst *popupInstance) Paint(x, y int) []paint.DrawCmd {
	if len(inst.options) == 0 {
		return nil
	}
	return inst.paintPopupAt(x, y)
}

func (inst *popupInstance) HandleAction(act *action.Action) bool {
	if act == nil || inst.disabled || len(inst.options) == 0 {
		return false
	}

	switch act.Type {
	case action.ActionNavigateDown:
		return inst.setHighlight(inst.highlightedIndex + 1)
	case action.ActionNavigateUp:
		return inst.setHighlight(inst.highlightedIndex - 1)
	case action.ActionNavigateHome:
		return inst.setHighlight(0)
	case action.ActionNavigateEnd:
		return inst.setHighlight(len(inst.options) - 1)
	case action.ActionNavigatePageUp:
		return inst.setHighlight(inst.highlightedIndex - maxInt(1, inst.visibleRowCount()))
	case action.ActionNavigatePageDown:
		return inst.setHighlight(inst.highlightedIndex + maxInt(1, inst.visibleRowCount()))
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
		return inst.setHighlight(inst.highlightedIndex + scrollDirection(delta))
	case action.ActionSelect, action.ActionEnter, action.ActionSubmit:
		return inst.commit(inst.highlightedIndex)
	case action.ActionCancel:
		return inst.requestClose()
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
	return inst.disabled || len(inst.options) == 0
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
	markerWidth := inst.optionMarkerWidth()
	labelWidth := 0
	for _, opt := range inst.options {
		labelWidth = maxInt(labelWidth, paint.StringWidth(opt.Label))
	}
	contentWidth := labelWidth
	if markerWidth > 0 {
		contentWidth += markerWidth + 1
	}
	return maxInt(inst.minWidth, maxInt(contentWidth+2, 6))
}

func (inst *popupInstance) popupHeight() int {
	if len(inst.options) == 0 {
		return 0
	}
	return inst.visibleRowCount() + 2
}

func (inst *popupInstance) visibleRowCount() int {
	if len(inst.options) == 0 {
		return 0
	}
	rows := inst.maxVisibleRows
	if rows <= 0 {
		rows = defaultMaxVisibleRows
	}
	return minInt(len(inst.options), rows)
}

func (inst *popupInstance) maxScrollOffset() int {
	return maxInt(0, len(inst.options)-inst.visibleRowCount())
}

func (inst *popupInstance) setHighlight(index int) bool {
	if len(inst.options) == 0 {
		return false
	}
	index = clampInt(index, 0, len(inst.options)-1)
	if inst.overlayCallbacks != nil {
		state := inst.overlayCallbacks.setHighlight(index)
		oldHighlight := inst.highlightedIndex
		oldScroll := inst.scrollOffset
		inst.selectedIndex = state.selectedIndex
		inst.selectedIndices = append([]int(nil), state.selectedIndices...)
		inst.highlightedIndex = state.highlightedIndex
		inst.scrollOffset = state.scrollOffset
		inst.dirty = oldHighlight != inst.highlightedIndex || oldScroll != inst.scrollOffset
		return oldHighlight != inst.highlightedIndex || oldScroll != inst.scrollOffset
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
		return false
	}
	oldSelectedIndex := inst.selectedIndex
	oldSelectedIndices := append([]int(nil), inst.selectedIndices...)
	closed := false
	if inst.overlayCallbacks != nil {
		state := inst.overlayCallbacks.commit(index)
		inst.selectedIndex = state.selectedIndex
		inst.selectedIndices = append([]int(nil), state.selectedIndices...)
		inst.highlightedIndex = state.highlightedIndex
		inst.scrollOffset = state.scrollOffset
		inst.dirty = true
		closed = !state.open
	} else {
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
	if selectionChanged {
		emitFieldValueChangedFrom(
			inst,
			inst.changeIntent,
			inst.changeIntentField,
			inst.formID,
			inst.selectionMode,
			inst.selectedIndex,
			inst.selectedIndices,
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
	return selectionChanged || closed
}

func (inst *popupInstance) requestClose() bool {
	if inst.overlayCallbacks != nil {
		wasOpen := true
		state := inst.overlayCallbacks.setOpen(false)
		inst.selectedIndex = state.selectedIndex
		inst.selectedIndices = append([]int(nil), state.selectedIndices...)
		inst.highlightedIndex = state.highlightedIndex
		inst.scrollOffset = state.scrollOffset
		inst.dirty = true
		return wasOpen && !state.open
	}
	inst.dirty = true
	return true
}

func (inst *popupInstance) ensureHighlightVisible() {
	if len(inst.options) == 0 || inst.highlightedIndex < 0 {
		inst.scrollOffset = 0
		return
	}
	maxOffset := inst.maxScrollOffset()
	if inst.scrollOffset > maxOffset {
		inst.scrollOffset = maxOffset
	}
	if inst.highlightedIndex < inst.scrollOffset {
		inst.scrollOffset = inst.highlightedIndex
	}
	if visible := inst.visibleRowCount(); visible > 0 && inst.highlightedIndex >= inst.scrollOffset+visible {
		inst.scrollOffset = inst.highlightedIndex - visible + 1
	}
	inst.scrollOffset = clampInt(inst.scrollOffset, 0, maxOffset)
}

func (inst *popupInstance) optionIndexAt(localX, localY int) (int, bool) {
	if localX < 0 || localY < 1 {
		return 0, false
	}
	row := localY - 1
	if row < 0 || row >= inst.visibleRowCount() {
		return 0, false
	}
	index := inst.scrollOffset + row
	if index < 0 || index >= len(inst.options) {
		return 0, false
	}
	return index, true
}

func (inst *popupInstance) paintPopupAt(x, y int) []paint.DrawCmd {
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

	visibleRows := inst.visibleRowCount()
	for row := 0; row < visibleRows; row++ {
		rowY := y + 1 + row
		cmds = append(cmds,
			paint.DrawCmd{X: x, Y: rowY, Text: "│", Style: borderStyle},
			paint.DrawCmd{X: x + 1, Y: rowY, Text: strings.Repeat(" ", contentWidth), Style: fillStyle},
			paint.DrawCmd{X: x + popupWidth - 1, Y: rowY, Text: "│", Style: borderStyle},
		)
		optionIndex := inst.scrollOffset + row
		if optionIndex >= len(inst.options) {
			continue
		}
		rowStyle := popupOptionRowStyleFor(inst.popupStyle, inst.disabled, optionIndex == inst.highlightedIndex)
		rowText := inst.optionRowText(optionIndex, contentWidth)
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
	selected := containsInt(inst.selectedIndices, index)
	if inst.selectionMode == SelectionMultiple {
		if selected {
			return "[x]"
		}
		return "[ ]"
	}
	if selected {
		return "●"
	}
	return "○"
}

func (inst *popupInstance) optionMarkerWidth() int {
	if inst.selectionMode == SelectionMultiple {
		return 3
	}
	return 1
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

func popupOptionRowStyleFor(base style.Style, disabled, highlighted bool) style.Style {
	s := popupFillStyleFor(base, disabled)
	if disabled {
		return s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
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
