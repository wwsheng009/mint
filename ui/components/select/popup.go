package selectcomp

import (
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type popupVNode struct {
	*rtui.ElementVNode
}

const popupAnchorOffsetY = 0

func newPopupRuntimeVNode(owner *Instance) rtui.VNode {
	if owner == nil || owner.ownerID == "" {
		return nil
	}

	componentID := owner.componentID
	if componentID == "" {
		componentID = owner.ownerID
	}

	surface := &popupVNode{ElementVNode: rtui.NewElement("select-popup")}
	surface.SetKey(owner.ownerID + "-popup")
	surface.SetID(owner.ownerID + "-popup")
	surface.SetLayer(rtui.LayerOverlay)
	surface.SetProps(rtui.Props{
		"ownerID":          owner.ownerID,
		"componentID":      componentID,
		"options":          append([]Option(nil), owner.options...),
		"style":            owner.selectStyle,
		"selectionMode":    owner.selectionMode,
		"selectedIndex":    owner.selectedIndex,
		"selectedIndices":  append([]int(nil), owner.selectedIndices...),
		"highlightedIndex": owner.highlightedIndex,
		"scrollOffset":     owner.scrollOffset,
		"maxVisibleRows":   owner.maxVisibleRows,
		"minWidth":         owner.triggerWidth(),
		"disabled":         owner.state.Disabled,
		"closeOnOutside":   owner.closeOnOutside,
	})

	portal := rtui.NewElement("box")
	portal.SetKey(owner.ownerID + "-popup-portal")
	portal.SetID(owner.ownerID + "-popup-portal")
	portal.SetProps(rtui.Props{
		"position": "absolute",
		"left":     0,
		"top":      0,
		"width":    1,
		"height":   1,
	})
	portal.SetLayer(rtui.LayerOverlay)
	if ownerID := owner.ownerID; ownerID != "" {
		portal.SetAnchorTo(ownerID, rttypes.AnchorBottomLeft)
	}
	if owner.portalRoot != "" {
		portal.SetPortalRoot(owner.portalRoot)
	}
	portal.SetPortalPosition(rttypes.PositionAbsolute)
	portal.SetProp("left", 0)
	if popupAnchorOffsetY != 0 {
		portal.SetProp("top", popupAnchorOffsetY)
	}
	portal.SetChildren([]rtui.VNode{surface})
	return portal
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
	ownerID          string
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
	focused          bool
	bounds           [4]int
	dirty            bool
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
)

func newPopupInstance(props rtui.Props) *popupInstance {
	inst := &popupInstance{dirty: true}
	inst.SetProps(props)
	return inst
}

func (inst *popupInstance) Key() string                        { return inst.key }
func (inst *popupInstance) SetKey(key string)                  { inst.key = key }
func (inst *popupInstance) Init(props rtui.Props)              { inst.SetProps(props) }
func (inst *popupInstance) Destroy()                           { inst.unregister() }
func (inst *popupInstance) OnMount()                           { selectOverlayRegistry.registerPopup(inst.ownerID, inst) }
func (inst *popupInstance) OnUnmount()                         { inst.unregister() }
func (inst *popupInstance) MarkDirty()                         { inst.dirty = true }
func (inst *popupInstance) IsDirty() bool                      { return inst.dirty }
func (inst *popupInstance) GetContext() *rtui.ComponentContext { return nil }
func (inst *popupInstance) Parent() interface{} {
	if inst.parent != nil {
		return inst.parent
	}
	return selectOverlayRegistry.trigger(inst.ownerID)
}
func (inst *popupInstance) SetParent(parent rtui.ComponentInstance) {
	inst.parent = parent
}

func (inst *popupInstance) unregister() {
	selectOverlayRegistry.unregisterPopup(inst.ownerID, inst)
}

func (inst *popupInstance) SetProps(props rtui.Props) bool {
	oldOwnerID := inst.ownerID
	oldHighlight := inst.highlightedIndex
	oldScroll := inst.scrollOffset
	oldSelected := inst.selectedIndex
	oldSelectedIndices := append([]int(nil), inst.selectedIndices...)

	inst.key = getStringProp(props, "key", inst.key)
	inst.ownerID = getStringProp(props, "ownerID", inst.ownerID)
	inst.componentID = getStringProp(props, "componentID", inst.componentID)
	inst.options = getOptionsProp(props)
	inst.popupStyle = getStyleProp(props)
	inst.selectionMode = getSelectionModeProp(props, inst.selectionMode)
	inst.selectedIndex = getIntProp(props, "selectedIndex", inst.selectedIndex)
	inst.selectedIndices = getIntsProp(props, "selectedIndices", inst.selectedIndices)
	inst.highlightedIndex = getIntProp(props, "highlightedIndex", inst.highlightedIndex)
	inst.scrollOffset = getIntProp(props, "scrollOffset", inst.scrollOffset)
	inst.maxVisibleRows = getIntProp(props, "maxVisibleRows", inst.maxVisibleRows)
	inst.minWidth = getIntProp(props, "minWidth", inst.minWidth)
	inst.disabled = getBoolProp(props, "disabled", inst.disabled)
	inst.closeOnOutside = getBoolProp(props, "closeOnOutside", inst.closeOnOutside)

	inst.selectedIndices = normalizeIndices(inst.selectedIndices, len(inst.options))
	if inst.maxVisibleRows <= 0 {
		inst.maxVisibleRows = defaultMaxVisibleRows
	}
	if inst.highlightedIndex < 0 && len(inst.options) > 0 {
		if inst.selectedIndex >= 0 {
			inst.highlightedIndex = inst.selectedIndex
		} else {
			inst.highlightedIndex = 0
		}
	}
	if len(inst.options) > 0 {
		inst.highlightedIndex = clampInt(inst.highlightedIndex, 0, len(inst.options)-1)
		inst.scrollOffset = clampInt(inst.scrollOffset, 0, inst.maxScrollOffset())
	}

	if oldOwnerID != "" && oldOwnerID != inst.ownerID {
		selectOverlayRegistry.unregisterPopup(oldOwnerID, inst)
	}
	if inst.ownerID != "" {
		selectOverlayRegistry.registerPopup(inst.ownerID, inst)
	}

	changed := oldOwnerID != inst.ownerID ||
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
		"key":              inst.key,
		"ownerID":          inst.ownerID,
		"componentID":      inst.componentID,
		"options":          append([]Option(nil), inst.options...),
		"style":            inst.popupStyle,
		"selectionMode":    inst.selectionMode,
		"selectedIndex":    inst.selectedIndex,
		"selectedIndices":  append([]int(nil), inst.selectedIndices...),
		"highlightedIndex": inst.highlightedIndex,
		"scrollOffset":     inst.scrollOffset,
		"maxVisibleRows":   inst.maxVisibleRows,
		"minWidth":         inst.minWidth,
		"disabled":         inst.disabled,
		"closeOnOutside":   inst.closeOnOutside,
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
		return inst.moveHighlight(1)
	case action.ActionNavigateUp:
		return inst.moveHighlight(-1)
	case action.ActionNavigateHome:
		return inst.moveHighlightTo(0)
	case action.ActionNavigateEnd:
		return inst.moveHighlightTo(len(inst.options) - 1)
	case action.ActionNavigatePageUp:
		return inst.pageHighlight(-1)
	case action.ActionNavigatePageDown:
		return inst.pageHighlight(1)
	case action.ActionHover:
		mouse, ok := popupMousePayload(act.Payload)
		if !ok {
			return false
		}
		return inst.highlightIndexAt(mouse.LocalX, mouse.LocalY)
	case action.ActionClick:
		mouse, ok := popupMousePayload(act.Payload)
		if ok {
			index, hit := inst.optionIndexAt(mouse.LocalX, mouse.LocalY)
			if !hit {
				return true
			}
			inst.applyOwnerState()
			inst.setOwnerHighlight(index)
			inst.commitOwnerIndex(index)
			return true
		}
		inst.applyOwnerState()
		inst.commitOwnerIndex(inst.highlightedIndex)
		return true
	case action.ActionScroll:
		delta, ok := scrollDeltaFromAction(act)
		if !ok || delta == 0 {
			return false
		}
		next := clampInt(inst.highlightedIndex+scrollDirection(delta), 0, len(inst.options)-1)
		if next == inst.highlightedIndex {
			return false
		}
		return inst.setOwnerHighlight(next)
	case action.ActionSelect, action.ActionEnter, action.ActionSubmit:
		inst.applyOwnerState()
		return inst.commitOwnerIndex(inst.highlightedIndex)
	case action.ActionCancel:
		inst.applyOwnerState()
		return inst.setOwnerOpen(false)
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

func (inst *popupInstance) moveHighlight(delta int) bool {
	inst.applyOwnerState()
	return inst.moveHighlightTo(clampInt(inst.highlightedIndex+delta, 0, len(inst.options)-1))
}

func (inst *popupInstance) moveHighlightTo(index int) bool {
	next := clampInt(index, 0, len(inst.options)-1)
	if next == inst.highlightedIndex {
		return false
	}
	return inst.setOwnerHighlight(next)
}

func (inst *popupInstance) pageHighlight(direction int) bool {
	pageSize := maxInt(1, inst.visibleRowCount())
	next := inst.highlightedIndex + direction*pageSize
	return inst.moveHighlightTo(next)
}

func (inst *popupInstance) highlightIndexAt(localX, localY int) bool {
	index, hit := inst.optionIndexAt(localX, localY)
	if !hit || index == inst.highlightedIndex {
		return false
	}
	return inst.setOwnerHighlight(index)
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

func (inst *popupInstance) owner() *Instance {
	if inst == nil {
		return nil
	}
	if parent := inst.Parent(); parent != nil {
		if owner, ok := parent.(*Instance); ok {
			return owner
		}
	}
	return selectOverlayRegistry.trigger(inst.ownerID)
}

func (inst *popupInstance) applyOwnerState() {
	owner := inst.owner()
	if owner == nil {
		return
	}
	inst.selectedIndex = owner.selectedIndex
	inst.selectedIndices = append([]int(nil), owner.selectedIndices...)
	inst.highlightedIndex = owner.highlightedIndex
	inst.scrollOffset = owner.scrollOffset
	inst.disabled = owner.state.Disabled
}

func (inst *popupInstance) setOwnerHighlight(index int) bool {
	owner := inst.owner()
	if owner == nil {
		intent.Emit(inst, SelectHighlightIntent{Index: index, ComponentID: inst.componentID})
		return true
	}
	handled := owner.HandleIntent(SelectHighlightIntent{Index: index, ComponentID: inst.componentID})
	inst.applyOwnerState()
	if handled {
		inst.dirty = true
	}
	return handled
}

func (inst *popupInstance) commitOwnerIndex(index int) bool {
	owner := inst.owner()
	if owner == nil {
		intent.Emit(inst, SelectCommitIndexIntent{Index: index, ComponentID: inst.componentID})
		return true
	}
	handled := owner.HandleIntent(SelectCommitIndexIntent{Index: index, ComponentID: inst.componentID})
	inst.applyOwnerState()
	inst.dirty = true
	return handled
}

func (inst *popupInstance) setOwnerOpen(open bool) bool {
	owner := inst.owner()
	if owner == nil {
		intent.Emit(inst, SelectSetOpenIntent{Open: open, ComponentID: inst.componentID})
		return true
	}
	handled := owner.HandleIntent(SelectSetOpenIntent{Open: open, ComponentID: inst.componentID})
	inst.applyOwnerState()
	inst.dirty = true
	return handled
}
