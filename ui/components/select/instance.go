package selectcomp

import (
	"fmt"
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
	"github.com/wwsheng009/mint/ui/components/control"
	"github.com/wwsheng009/mint/ui/components/form"
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

const defaultMaxVisibleRows = 6

const popupPortalRootProp = "popupPortalRoot"

// Instance is the runtime entity for Select components.
type Instance struct {
	key               string
	componentID       string
	parent            rtui.ComponentInstance
	childInstances    []rtui.ComponentInstance
	baseOptions       []Option
	createdOptions    []Option
	options           []Option
	selectStyle       style.Style
	width             int
	placeholder       string
	filterOption      bool
	filterPlaceholder string
	maxVisibleRows    int
	overlayPopup      bool
	portalRoot        string
	ownerID           string
	selectID          string
	closeOnOutside    bool
	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent
	formID            string
	selectionMode     SelectionMode

	state            control.InteractionState
	selectedIndex    int
	selectedIndices  []int
	open             bool
	highlightedIndex int
	scrollOffset     int
	filterQuery      string
	bounds           [4]int
	dirty            bool

	intentEmitter    func(intent.Intent)
	behaviors        *control.BehaviorList
	overlayCallbacks *overlayCallbacks
}

var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ rtui.TreeNode              = (*Instance)(nil)
	_ rtui.TreeContainer         = (*Instance)(nil)
	_ control.Instance           = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new SelectInstance from props.
func NewInstance(props rtui.Props) *Instance {
	baseOptions := getOptionsProp(props)
	inst := &Instance{
		key:               proputil.GetString(props, "key", ""),
		componentID:       proputil.GetString(props, "componentID", ""),
		baseOptions:       baseOptions,
		options:           append([]Option(nil), baseOptions...),
		selectStyle:       proputil.GetStyle(props, "style", style.Style{}),
		width:             proputil.GetInt(props, "width", 0),
		placeholder:       proputil.GetString(props, "placeholder", "..."),
		filterOption:      proputil.GetBool(props, propFilterOption, false),
		filterPlaceholder: proputil.GetString(props, propFilterPlaceholder, "type to filter"),
		filterQuery:       proputil.GetString(props, propFilterQuery, ""),
		maxVisibleRows:    proputil.GetInt(props, "maxVisibleRows", defaultMaxVisibleRows),
		overlayPopup:      proputil.GetBool(props, "overlayPopup", false),
		portalRoot:        getPortalRootProp(props, rtui.DefaultOverlayPortalRootID),
		ownerID:           proputil.GetString(props, "ownerID", ""),
		selectID:          proputil.GetString(props, "selectID", ""),
		closeOnOutside:    proputil.GetBool(props, "closeOnOutside", true),
		changeIntent:      proputil.GetIntent(props, "changeIntent", nil),
		changeIntentField: getChangeIntentFieldProp(props, "changeIntent"),
		formID:            proputil.GetString(props, "formID", ""),
		selectionMode:     getSelectionModeProp(props, SelectionSingle),
		selectedIndex:     proputil.GetInt(props, "selectedIndex", -1),
		selectedIndices:   getIntsProp(props, "selectedIndices", nil),
		highlightedIndex:  -1,
		dirty:             true,
		overlayCallbacks:  getOverlayCallbacksProp(props),
	}

	inst.state = control.InteractionState{
		Disabled: proputil.GetBool(props, "disabled", false),
	}
	inst.normalizeSelectionState()
	inst.initBehaviors()
	return inst
}

func (inst *Instance) initBehaviors() {
	inst.behaviors = control.NewBehaviorList(
		&control.FocusableBehavior{},
		&control.HoverableBehavior{},
		&control.DisableableBehavior{},
	)
}

func (inst *Instance) selectIdentity() string {
	return firstNonEmpty(inst.selectID, inst.ownerID, inst.componentID, inst.key)
}

func (inst *Instance) syncOptions() {
	inst.options = mergeOptions(inst.baseOptions, inst.createdOptions)
}

func (inst *Instance) popupRows() popupRows {
	return buildPopupRows(
		inst.options,
		inst.selectionMode,
		inst.filterOption,
		inst.filterPlaceholder,
		inst.filterQuery,
	)
}

func (inst *Instance) canOpenPopup() bool {
	rows := inst.popupRows()
	return rows.showFilter || len(rows.scrollable) > 0
}

func (inst *Instance) applyOverlayControllerState(state overlayControllerState) bool {
	oldOpen := inst.open
	oldHighlight := inst.highlightedIndex
	oldScroll := inst.scrollOffset
	oldSelectedIndex := inst.selectedIndex
	oldSelectedIndices := append([]int(nil), inst.selectedIndices...)
	oldOptions := append([]Option(nil), inst.options...)
	oldFilterQuery := inst.filterQuery

	inst.open = state.open
	inst.highlightedIndex = state.highlightedIndex
	inst.scrollOffset = state.scrollOffset
	inst.selectedIndex = state.selectedIndex
	inst.selectedIndices = append([]int(nil), state.selectedIndices...)
	if state.options != nil {
		inst.options = append([]Option(nil), state.options...)
	}
	inst.filterQuery = state.filterQuery
	inst.normalizeSelectionState()

	changed := oldOpen != inst.open ||
		oldHighlight != inst.highlightedIndex ||
		oldScroll != inst.scrollOffset ||
		oldSelectedIndex != inst.selectedIndex ||
		oldFilterQuery != inst.filterQuery ||
		!equalIntSlices(oldSelectedIndices, inst.selectedIndices) ||
		!equalOptions(oldOptions, inst.options)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) requestOverlayOpen(open bool) bool {
	if inst.overlayCallbacks == nil {
		if open {
			return inst.openDropdown()
		}
		return inst.closeDropdown()
	}
	state := inst.overlayCallbacks.setOpen(open)
	return inst.applyOverlayControllerState(state)
}

func (inst *Instance) requestOverlayHighlight(index int) bool {
	if inst.overlayCallbacks == nil {
		return inst.moveHighlightTo(index)
	}
	state := inst.overlayCallbacks.setHighlight(index)
	return inst.applyOverlayControllerState(state)
}

func (inst *Instance) requestOverlayCommit(index int) bool {
	if inst.overlayCallbacks == nil {
		return inst.activateIndex(index)
	}
	oldSelectedIndex := inst.selectedIndex
	oldSelectedIndices := append([]int(nil), inst.selectedIndices...)
	oldOpen := inst.open
	state := inst.overlayCallbacks.commit(index)
	inst.applyOverlayControllerState(state)
	selectionChanged := oldSelectedIndex != inst.selectedIndex ||
		!equalIntSlices(oldSelectedIndices, inst.selectedIndices)
	closed := oldOpen && !inst.open
	if selectionChanged {
		inst.emitChange()
		inst.emitSelectChange()
	}
	return selectionChanged || closed
}

func (inst *Instance) requestOverlayFilterQuery(query string) bool {
	if inst.overlayCallbacks == nil {
		return inst.setFilterQuery(query)
	}
	state := inst.overlayCallbacks.setFilterQuery(query)
	return inst.applyOverlayControllerState(state)
}

func (inst *Instance) requestOverlayCreateTag() bool {
	if inst.overlayCallbacks == nil {
		return inst.createTagFromQuery()
	}
	oldSelectedIndex := inst.selectedIndex
	oldSelectedIndices := append([]int(nil), inst.selectedIndices...)
	oldOptions := append([]Option(nil), inst.options...)
	state := inst.overlayCallbacks.createTag()
	inst.applyOverlayControllerState(state)
	selectionChanged := oldSelectedIndex != inst.selectedIndex ||
		!equalIntSlices(oldSelectedIndices, inst.selectedIndices)
	optionsChanged := !equalOptions(oldOptions, inst.options)
	if selectionChanged {
		inst.emitChange()
		inst.emitSelectChange()
	}
	return selectionChanged || optionsChanged
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

func (inst *Instance) Destroy() {
	inst.behaviors.OnUnmount(inst)
}

func (inst *Instance) OnMount() {
	inst.behaviors.OnMount(inst)
}

func (inst *Instance) OnUnmount() {
	inst.behaviors.OnUnmount(inst)
}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldOptions := append([]Option(nil), inst.options...)
	oldSelectedIndex := inst.selectedIndex
	oldSelectedIndices := append([]int(nil), inst.selectedIndices...)
	oldDisabled := inst.state.Disabled
	oldWidth := inst.width
	oldPlaceholder := inst.placeholder
	oldFilterOption := inst.filterOption
	oldFilterPlaceholder := inst.filterPlaceholder
	oldMaxVisibleRows := inst.maxVisibleRows
	oldSelectionMode := inst.selectionMode
	oldOverlayPopup := inst.overlayPopup
	oldPortalRoot := inst.portalRoot
	oldOwnerID := inst.ownerID
	oldSelectID := inst.selectID
	oldCloseOnOutside := inst.closeOnOutside

	inst.key = proputil.GetString(props, "key", inst.key)
	inst.componentID = proputil.GetString(props, "componentID", inst.componentID)
	inst.baseOptions = getOptionsProp(props)
	inst.selectStyle = proputil.GetStyle(props, "style", style.Style{})
	inst.width = proputil.GetInt(props, "width", inst.width)
	inst.placeholder = proputil.GetString(props, "placeholder", inst.placeholder)
	inst.filterOption = proputil.GetBool(props, propFilterOption, inst.filterOption)
	inst.filterPlaceholder = proputil.GetString(props, propFilterPlaceholder, inst.filterPlaceholder)
	if value, ok := props[propFilterQuery].(string); ok {
		inst.filterQuery = value
	}
	inst.maxVisibleRows = proputil.GetInt(props, "maxVisibleRows", inst.maxVisibleRows)
	inst.overlayPopup = proputil.GetBool(props, "overlayPopup", inst.overlayPopup)
	inst.portalRoot = getPortalRootProp(props, inst.portalRoot)
	inst.ownerID = proputil.GetString(props, "ownerID", inst.ownerID)
	inst.selectID = proputil.GetString(props, "selectID", inst.selectIdentity())
	inst.closeOnOutside = proputil.GetBool(props, "closeOnOutside", inst.closeOnOutside)
	inst.changeIntent = proputil.GetIntent(props, "changeIntent", nil)
	inst.changeIntentField = getChangeIntentFieldProp(props, "changeIntent")
	inst.formID = proputil.GetString(props, "formID", inst.formID)
	inst.selectionMode = getSelectionModeProp(props, inst.selectionMode)
	if isTagsSelectionMode(inst.selectionMode) {
		inst.filterOption = true
	}
	inst.syncOptions()
	inst.selectedIndex = proputil.GetInt(props, "selectedIndex", inst.selectedIndex)
	inst.selectedIndices = getIntsProp(props, "selectedIndices", inst.selectedIndices)
	inst.overlayCallbacks = getOverlayCallbacksProp(props)

	if v, ok := props[propOpen].(bool); ok {
		inst.open = v
	}
	if v, ok := props[propHighlightedIndex].(int); ok {
		inst.highlightedIndex = v
	}
	if v, ok := props[propScrollOffset].(int); ok {
		inst.scrollOffset = v
	}

	newDisabled := proputil.GetBool(props, "disabled", inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}
	if inst.state.Disabled {
		inst.open = false
	}

	if oldSelectedIndex != inst.selectedIndex ||
		!equalIntSlices(oldSelectedIndices, inst.selectedIndices) ||
		oldSelectionMode != inst.selectionMode {
		if _, controlledHighlight := props[propHighlightedIndex]; !controlledHighlight {
			if isMultiSelectionMode(inst.selectionMode) && len(inst.selectedIndices) > 0 {
				inst.highlightedIndex = inst.selectedIndices[len(inst.selectedIndices)-1]
			} else {
				inst.highlightedIndex = inst.selectedIndex
			}
		}
	}

	inst.normalizeSelectionState()

	changed := !equalOptions(oldOptions, inst.options) ||
		oldSelectedIndex != inst.selectedIndex ||
		!equalIntSlices(oldSelectedIndices, inst.selectedIndices) ||
		oldDisabled != inst.state.Disabled ||
		oldWidth != inst.width ||
		oldPlaceholder != inst.placeholder ||
		oldFilterOption != inst.filterOption ||
		oldFilterPlaceholder != inst.filterPlaceholder ||
		oldMaxVisibleRows != inst.maxVisibleRows ||
		oldSelectionMode != inst.selectionMode ||
		oldOverlayPopup != inst.overlayPopup ||
		oldPortalRoot != inst.portalRoot ||
		oldOwnerID != inst.ownerID ||
		oldSelectID != inst.selectID ||
		oldCloseOnOutside != inst.closeOnOutside

	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:               inst.key,
		propOptions:           append([]Option(nil), inst.options...),
		propSelectedIndex:     inst.selectedIndex,
		propSelectedIndices:   append([]int(nil), inst.selectedIndices...),
		propSelectionMode:     inst.selectionMode,
		propDisabled:          inst.state.Disabled,
		propOpen:              inst.open,
		propFilterOption:      inst.filterOption,
		propFilterPlaceholder: inst.filterPlaceholder,
		propFilterQuery:       inst.filterQuery,
		propOverlayPopup:      inst.overlayPopup,
		propPortalRoot:        inst.portalRoot,
		propOwnerID:           inst.ownerID,
		propSelectID:          inst.selectIdentity(),
		propCloseOnOutside:    inst.closeOnOutside,
		propHighlightedIndex:  inst.highlightedIndex,
		propScrollOffset:      inst.scrollOffset,
		overlayCallbacksProp:  inst.overlayCallbacks,
	}
}

func (inst *Instance) MarkDirty() {
	inst.dirty = true
}

func (inst *Instance) IsDirty() bool {
	return inst.dirty
}

func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	return nil
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	triggerWidth := inst.triggerPaintWidth()
	triggerStyle := inst.resolveStyle()
	triggerText := inst.triggerText(triggerWidth)
	cmds := []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  triggerText,
		Style: triggerStyle,
	}}

	if inst.overlayPopup {
		return cmds
	}

	if !inst.open || inst.popupHeight() == 0 {
		return cmds
	}
	return append(cmds, inst.paintPopupAt(x, y+1)...)
}

func (inst *Instance) resolveStyle() style.Style {
	s := inst.selectStyle
	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}

	if inst.state.Disabled {
		return s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	if inst.open || inst.state.Focused {
		return s.Foreground(theme.Focus()).Bold(true)
	}
	if inst.state.Hovered {
		return s.Underline(true)
	}
	return s
}

func (inst *Instance) popupFillStyle() style.Style {
	s := inst.selectStyle
	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}
	if inst.state.Disabled {
		return s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	return s.Background(theme.Surface())
}

func (inst *Instance) popupBorderStyle() style.Style {
	s := style.Style{}.Foreground(theme.Focus()).Background(theme.Surface())
	if inst.state.Disabled {
		return s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	return s
}

func (inst *Instance) popupRowStyle(row popupRow, highlighted bool) style.Style {
	s := inst.popupFillStyle()
	if inst.state.Disabled {
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

func (inst *Instance) paintPopupAt(x, y int) []paint.DrawCmd {
	rows := inst.popupRows()
	if !inst.open || (!rows.showFilter && len(rows.scrollable) == 0) {
		return nil
	}

	popupWidth := inst.popupWidth()
	popupHeight := inst.popupHeight()
	if popupWidth < 4 || popupHeight < 3 {
		return nil
	}

	borderStyle := inst.popupBorderStyle()
	fillStyle := inst.popupFillStyle()
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

	visibleRows := visibleScrollableRowCount(rows, inst.maxVisibleRows)
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
		rowStyle := inst.popupRowStyle(row, isHighlightedTarget(row, inst.highlightedIndex))
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

func (inst *Instance) SetFocus(focused bool) {
	if inst.state.Focused == focused {
		return
	}
	oldState := inst.state
	inst.state.Focused = focused
	if !focused {
		if !inst.overlayPopup || !inst.open {
			inst.closeDropdown()
			inst.emitFieldBlur()
		}
	}
	inst.dirty = true
	inst.behaviors.OnStateChange(inst, oldState, inst.state)
}

func (inst *Instance) HasFocus() bool {
	return inst.state.Focused
}

func (inst *Instance) IsDisabled() bool {
	return inst.state.Disabled
}

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}
	if inst.behaviors.OnAction(inst, act) {
		return true
	}
	if inst.state.Disabled {
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
	case action.ActionHover:
		return inst.handleHover(act)
	case action.ActionClick:
		if mouse, ok := mousePayload(act.Payload); ok && log.RenderLogger.Enabled() {
			log.RenderLogger.Debug("[SelectTrigger] ActionClick select=%s screen=(%d,%d) local=(%d,%d) open=%v",
				inst.selectIdentity(), mouse.X, mouse.Y, mouse.LocalX, mouse.LocalY, inst.open)
		}
		return inst.handleClick(act)
	case action.ActionScroll:
		return inst.handleScroll(act)
	case action.ActionNavigateDown:
		if inst.open {
			return inst.moveHighlight(1)
		}
		if isMultiSelectionMode(inst.selectionMode) || filterEnabledFor(inst.selectionMode, inst.filterOption) {
			return inst.requestOverlayOpen(true)
		}
		next := inst.selectedIndex + 1
		if next < 0 {
			next = 0
		}
		if next >= len(inst.options) {
			next = 0
		}
		return inst.requestOverlayCommit(next)
	case action.ActionNavigateUp:
		if inst.open {
			return inst.moveHighlight(-1)
		}
		if isMultiSelectionMode(inst.selectionMode) || filterEnabledFor(inst.selectionMode, inst.filterOption) {
			return inst.requestOverlayOpen(true)
		}
		next := inst.selectedIndex - 1
		if next < 0 {
			next = len(inst.options) - 1
		}
		return inst.requestOverlayCommit(next)
	case action.ActionNavigateHome:
		if inst.open {
			return inst.moveHighlightTo(inst.firstHighlightTarget())
		}
		return inst.requestOverlayCommit(0)
	case action.ActionNavigateEnd:
		if inst.open {
			return inst.moveHighlightTo(inst.lastHighlightTarget())
		}
		if len(inst.options) == 0 {
			return false
		}
		return inst.requestOverlayCommit(len(inst.options) - 1)
	case action.ActionNavigatePageUp:
		if !inst.open {
			return inst.requestOverlayOpen(true)
		}
		return inst.pageHighlight(-1)
	case action.ActionNavigatePageDown:
		if !inst.open {
			return inst.requestOverlayOpen(true)
		}
		return inst.pageHighlight(1)
	case action.ActionSelect:
		if _, ok := mousePayload(act.Payload); ok {
			// Mouse release is delivered as ActionSelect by App.processMsg.
			// For the trigger we treat release as a no-op; real click behavior is
			// handled on press (ActionClick), while popup rows handle release
			// themselves as first-class components.
			return true
		}
		if inst.open {
			return inst.requestOverlayCommit(inst.highlightedIndex)
		}
		if isMultiSelectionMode(inst.selectionMode) || filterEnabledFor(inst.selectionMode, inst.filterOption) {
			return inst.requestOverlayOpen(true)
		}
		next := inst.selectedIndex + 1
		if next < 0 {
			next = 0
		}
		if next >= len(inst.options) {
			next = 0
		}
		return inst.requestOverlayCommit(next)
	case action.ActionEnter, action.ActionSubmit:
		if !inst.open {
			return inst.requestOverlayOpen(true)
		}
		return inst.requestOverlayCommit(inst.highlightedIndex)
	case action.ActionMouseRelease:
		// Real runtime maps mouse release to ActionMouseRelease. The trigger
		// should not treat release as a second select/submit step.
		if mouse, ok := mousePayload(act.Payload); ok && log.RenderLogger.Enabled() {
			log.RenderLogger.Debug("[SelectTrigger] ActionMouseRelease select=%s screen=(%d,%d) local=(%d,%d) open=%v",
				inst.selectIdentity(), mouse.X, mouse.Y, mouse.LocalX, mouse.LocalY, inst.open)
		}
		return true
	case action.ActionCancel:
		return inst.requestOverlayOpen(false)
	}
	return false
}

func (inst *Instance) handleClick(act *action.Action) bool {
	mouse, ok := mousePayload(act.Payload)
	if !ok {
		if inst.open {
			return inst.requestOverlayCommit(inst.highlightedIndex)
		}
		return inst.requestOverlayOpen(true)
	}

	if mouse.LocalY <= 0 || inst.overlayPopup {
		if inst.open {
			return inst.requestOverlayOpen(false)
		}
		return inst.requestOverlayOpen(true)
	}

	if !inst.open {
		return inst.requestOverlayOpen(true)
	}

	index, hit := inst.optionIndexAt(mouse.LocalX, mouse.LocalY)
	if !hit {
		return true
	}

	inst.requestOverlayHighlight(index)
	return inst.requestOverlayCommit(index)
}

func (inst *Instance) handleHover(act *action.Action) bool {
	if !inst.open {
		return false
	}
	mouse, ok := mousePayload(act.Payload)
	if !ok {
		return false
	}
	index, hit := inst.optionIndexAt(mouse.LocalX, mouse.LocalY)
	if !hit || index == inst.highlightedIndex {
		return false
	}
	return inst.requestOverlayHighlight(index)
}

func (inst *Instance) handleScroll(act *action.Action) bool {
	if !inst.open {
		return false
	}
	delta, ok := scrollutil.DeltaFromAction(act)
	if !ok || delta == 0 {
		return false
	}
	if delta > 0 {
		return inst.moveHighlight(1)
	}
	return inst.moveHighlight(-1)
}

func (inst *Instance) firstHighlightTarget() int {
	return firstSelectableTarget(inst.popupRows())
}

func (inst *Instance) lastHighlightTarget() int {
	return lastSelectableTarget(inst.popupRows())
}

func (inst *Instance) setFilterQuery(query string) bool {
	if !filterEnabledFor(inst.selectionMode, inst.filterOption) {
		return false
	}

	query = sanitizeFilterText(query)
	if inst.overlayCallbacks != nil {
		return inst.requestOverlayFilterQuery(query)
	}
	if inst.filterQuery == query {
		return false
	}

	inst.filterQuery = query
	inst.normalizeSelectionState()
	inst.dirty = true
	return true
}

func (inst *Instance) clearFilterQueryValue() bool {
	return inst.setFilterQuery("")
}

func (inst *Instance) appendFilterText(text string) bool {
	if !filterEnabledFor(inst.selectionMode, inst.filterOption) {
		return false
	}
	text = sanitizeFilterText(text)
	if strings.TrimSpace(text) == "" {
		return false
	}
	if !inst.open {
		inst.requestOverlayOpen(true)
	}
	return inst.setFilterQuery(inst.filterQuery + text)
}

func (inst *Instance) backspaceFilterText() bool {
	if !filterEnabledFor(inst.selectionMode, inst.filterOption) || inst.filterQuery == "" {
		return false
	}
	runes := []rune(inst.filterQuery)
	if len(runes) == 0 {
		return false
	}
	return inst.setFilterQuery(string(runes[:len(runes)-1]))
}

func (inst *Instance) clearFilterText() bool {
	if !filterEnabledFor(inst.selectionMode, inst.filterOption) || inst.filterQuery == "" {
		return false
	}
	return inst.clearFilterQueryValue()
}

// SelectNext selects the next option in single mode.
func (inst *Instance) SelectNext() {
	if !inst.canOpenPopup() && len(inst.options) == 0 {
		return
	}
	if isMultiSelectionMode(inst.selectionMode) || filterEnabledFor(inst.selectionMode, inst.filterOption) {
		if !inst.open {
			inst.openDropdown()
			return
		}
		inst.moveHighlight(1)
		return
	}

	next := inst.selectedIndex + 1
	if next < 0 {
		next = 0
	}
	if next >= len(inst.options) {
		next = 0
	}
	inst.applySingleSelection(next, false)
}

// SelectPrev selects the previous option in single mode.
func (inst *Instance) SelectPrev() {
	if !inst.canOpenPopup() && len(inst.options) == 0 {
		return
	}
	if isMultiSelectionMode(inst.selectionMode) || filterEnabledFor(inst.selectionMode, inst.filterOption) {
		if !inst.open {
			inst.openDropdown()
			return
		}
		inst.moveHighlight(-1)
		return
	}

	next := inst.selectedIndex - 1
	if next < 0 {
		next = len(inst.options) - 1
	}
	inst.applySingleSelection(next, false)
}

// SetSelectedIndex sets the selected index.
func (inst *Instance) SetSelectedIndex(idx int) {
	if idx < -1 || idx >= len(inst.options) {
		return
	}
	if isMultiSelectionMode(inst.selectionMode) {
		inst.SetSelectedIndices([]int{idx})
		return
	}
	if inst.selectedIndex == idx && len(inst.selectedIndices) == boolToIntSliceLen(idx >= 0) {
		return
	}
	inst.selectedIndex = idx
	if idx >= 0 {
		inst.selectedIndices = []int{idx}
		inst.highlightedIndex = idx
	} else {
		inst.selectedIndices = nil
		inst.highlightedIndex = -1
	}
	inst.ensureHighlightVisible()
	inst.dirty = true
	inst.markOverlayDirty()
	inst.emitSelectChange()
}

// SetSelectedIndices sets the selected indices.
func (inst *Instance) SetSelectedIndices(indices []int) {
	normalized := normalizeIndices(indices, len(inst.options))
	if equalIntSlices(inst.selectedIndices, normalized) {
		return
	}
	inst.selectedIndices = normalized
	if len(normalized) == 0 {
		inst.selectedIndex = -1
	} else if isMultiSelectionMode(inst.selectionMode) {
		inst.selectedIndex = normalized[len(normalized)-1]
	} else {
		inst.selectedIndex = normalized[0]
		inst.selectedIndices = []int{inst.selectedIndex}
	}
	inst.highlightedIndex = inst.selectedIndex
	inst.ensureHighlightVisible()
	inst.dirty = true
	inst.markOverlayDirty()
	inst.emitSelectChange()
}

// SelectedIndex returns the selected index.
func (inst *Instance) SelectedIndex() int {
	return inst.selectedIndex
}

// SelectedIndices returns the selected indices.
func (inst *Instance) SelectedIndices() []int {
	return append([]int(nil), inst.selectedIndices...)
}

// SelectedValue returns the primary selected value.
func (inst *Instance) SelectedValue() string {
	if inst.selectedIndex >= 0 && inst.selectedIndex < len(inst.options) {
		return inst.options[inst.selectedIndex].Value
	}
	return ""
}

// SelectedValues returns all selected values.
func (inst *Instance) SelectedValues() []string {
	values := make([]string, 0, len(inst.selectedIndices))
	for _, idx := range inst.selectedIndices {
		if idx >= 0 && idx < len(inst.options) {
			values = append(values, inst.options[idx].Value)
		}
	}
	return values
}

// SelectedLabel returns the primary selected label.
func (inst *Instance) SelectedLabel() string {
	if inst.selectedIndex >= 0 && inst.selectedIndex < len(inst.options) {
		return inst.options[inst.selectedIndex].Label
	}
	return ""
}

// SelectedLabels returns all selected labels.
func (inst *Instance) SelectedLabels() []string {
	labels := make([]string, 0, len(inst.selectedIndices))
	for _, idx := range inst.selectedIndices {
		if idx >= 0 && idx < len(inst.options) {
			labels = append(labels, inst.options[idx].Label)
		}
	}
	return labels
}

func (inst *Instance) emitChange() {
	inst.emitFieldValueChanged()
}

func (inst *Instance) GetState() *control.InteractionState {
	return &inst.state
}

func (inst *Instance) SetState(state control.InteractionState) {
	oldState := inst.state
	inst.state = state
	inst.behaviors.OnStateChange(inst, oldState, inst.state)
}

func (inst *Instance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil {
		inst.intentEmitter(i)
	}
}

func (inst *Instance) HandleIntent(i intent.Intent) bool {
	if !intent.ShouldHandleIntentWithID(inst.componentID, i) {
		return false
	}

	switch v := i.(type) {
	case SelectNextIntent:
		if isMultiSelectionMode(inst.selectionMode) && inst.open {
			return inst.moveHighlight(1)
		}
		if isMultiSelectionMode(inst.selectionMode) || filterEnabledFor(inst.selectionMode, inst.filterOption) {
			return inst.requestOverlayOpen(true)
		}
		next := inst.selectedIndex + 1
		if next < 0 {
			next = 0
		}
		if next >= len(inst.options) {
			next = 0
		}
		return inst.requestOverlayCommit(next)
	case SelectPrevIntent:
		if isMultiSelectionMode(inst.selectionMode) && inst.open {
			return inst.moveHighlight(-1)
		}
		if isMultiSelectionMode(inst.selectionMode) || filterEnabledFor(inst.selectionMode, inst.filterOption) {
			return inst.requestOverlayOpen(true)
		}
		next := inst.selectedIndex - 1
		if next < 0 {
			next = len(inst.options) - 1
		}
		return inst.requestOverlayCommit(next)
	case SelectByIndexIntent:
		if v.Index < highlightCreateTag || v.Index >= len(inst.options) {
			return false
		}
		if isMultiSelectionMode(inst.selectionMode) {
			if v.Index < 0 {
				inst.SetSelectedIndices(nil)
				return true
			}
			return inst.toggleIndex(v.Index)
		}
		inst.SetSelectedIndex(v.Index)
		return true
	case SelectByValueIntent:
		for idx, opt := range inst.options {
			if opt.Value != v.Value {
				continue
			}
			if isMultiSelectionMode(inst.selectionMode) {
				return inst.toggleIndex(idx)
			}
			inst.SetSelectedIndex(idx)
			return true
		}
	case SelectSetOpenIntent:
		return inst.requestOverlayOpen(v.Open)
	case SelectHighlightIntent:
		return inst.requestOverlayHighlight(v.Index)
	case SelectCommitIndexIntent:
		return inst.requestOverlayCommit(v.Index)
	}

	return false
}

func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *Instance) GetStyle() style.Style {
	return inst.selectStyle
}

func (inst *Instance) SetStyle(s style.Style) {
	inst.selectStyle = s
}

func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case propDisabled:
		return inst.state.Disabled, true
	case propSelectedIndex:
		return inst.selectedIndex, true
	case propSelectedIndices:
		return append([]int(nil), inst.selectedIndices...), true
	case propSelectionMode:
		return inst.selectionMode, true
	case propOpen:
		return inst.open, true
	case propOptions:
		return inst.options, true
	case propFilterOption:
		return inst.filterOption, true
	default:
		return nil, false
	}
}

func (inst *Instance) SetProp(key string, value interface{}) {
	switch key {
	case propDisabled:
		if v, ok := value.(bool); ok {
			inst.state.Disabled = v
			if v {
				inst.open = false
			}
			inst.dirty = true
			inst.markOverlayDirty()
		}
	case propSelectedIndex:
		if v, ok := value.(int); ok {
			inst.SetSelectedIndex(v)
		}
	case propSelectedIndices:
		if v, ok := value.([]int); ok {
			inst.SetSelectedIndices(v)
		}
	case propSelectionMode:
		if v, ok := value.(SelectionMode); ok {
			inst.selectionMode = v
			if isTagsSelectionMode(v) {
				inst.filterOption = true
			}
			inst.normalizeSelectionState()
			inst.dirty = true
			inst.markOverlayDirty()
		}
	case propFilterOption:
		if v, ok := value.(bool); ok {
			inst.filterOption = v
			inst.normalizeSelectionState()
			inst.dirty = true
		}
	}
}

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

func (inst *Instance) ClearDirty() {
	inst.dirty = false
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	width := inst.triggerWidth()
	if inst.open && !inst.overlayPopup {
		width = maxInt(width, inst.popupWidth())
	}
	height := 1
	if inst.open && !inst.overlayPopup && inst.popupHeight() > 0 {
		height += inst.popupHeight()
	}

	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)
	return layout.Size{Width: width, Height: height}
}

func (inst *Instance) emitSelectChange() {
	if inst.intentEmitter == nil {
		return
	}

	inst.intentEmitter(SelectChangeIntent{
		SelectedIndex:   inst.selectedIndex,
		SelectedIndices: inst.SelectedIndices(),
		SelectedValue:   inst.SelectedValue(),
		SelectedValues:  inst.SelectedValues(),
		SelectedLabel:   inst.SelectedLabel(),
		SelectedLabels:  inst.SelectedLabels(),
		Mode:            inst.selectionMode,
		ComponentID:     inst.componentID,
	})
}

func (inst *Instance) emitFieldValueChanged() {
	if inst.intentEmitter == nil {
		return
	}

	value := inst.fieldValue()
	if inst.formID != "" {
		if inst.changeIntentField != nil {
			formIntent := form.FieldChange(inst.formID, inst.changeIntentField.GetField(), value, true)
			intent.Emit(inst, formIntent)
		}
		return
	}

	if inst.changeIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: value,
		})
	} else if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
}

func (inst *Instance) emitFieldBlur() {
	if inst.intentEmitter == nil || inst.formID == "" || inst.changeIntentField == nil {
		return
	}
	intent.Emit(inst, form.FieldBlur(inst.formID, inst.changeIntentField.GetField(), inst.fieldValue()))
}

func (inst *Instance) fieldValue() string {
	if isTagsSelectionMode(inst.selectionMode) {
		return joinSelectedValues(inst.options, inst.selectedIndices)
	}
	if isMultiSelectionMode(inst.selectionMode) {
		return joinIndices(inst.selectedIndices)
	}
	return fmt.Sprintf("%d", inst.selectedIndex)
}

func (inst *Instance) normalizeSelectionState() {
	if inst.maxVisibleRows <= 0 {
		inst.maxVisibleRows = defaultMaxVisibleRows
	}
	if strings.TrimSpace(inst.placeholder) == "" {
		inst.placeholder = "..."
	}
	if strings.TrimSpace(inst.filterPlaceholder) == "" {
		inst.filterPlaceholder = "type to filter"
	}
	if isTagsSelectionMode(inst.selectionMode) {
		inst.filterOption = true
	}
	inst.filterQuery = sanitizeFilterText(inst.filterQuery)

	count := len(inst.options)
	inst.selectedIndices = normalizeIndices(inst.selectedIndices, count)

	switch inst.selectionMode {
	case SelectionMultiple, SelectionTags:
		if len(inst.selectedIndices) == 0 && inst.selectedIndex >= 0 && inst.selectedIndex < count {
			inst.selectedIndices = []int{inst.selectedIndex}
		}
		if len(inst.selectedIndices) > 0 {
			if !containsInt(inst.selectedIndices, inst.selectedIndex) {
				inst.selectedIndex = inst.selectedIndices[len(inst.selectedIndices)-1]
			}
		} else {
			inst.selectedIndex = -1
		}
	default:
		if count == 0 {
			inst.selectedIndex = -1
			inst.selectedIndices = nil
		} else {
			if inst.selectedIndex >= count {
				inst.selectedIndex = count - 1
			}
			if inst.selectedIndex < -1 {
				inst.selectedIndex = -1
			}
			if inst.selectedIndex >= 0 {
				inst.selectedIndices = []int{inst.selectedIndex}
			} else {
				inst.selectedIndices = nil
			}
		}
	}

	if !inst.canOpenPopup() {
		inst.open = false
		inst.highlightedIndex, inst.scrollOffset = -1, 0
		return
	}

	rows := inst.popupRows()
	inst.highlightedIndex, inst.scrollOffset = normalizePopupHighlight(
		rows,
		inst.highlightedIndex,
		inst.scrollOffset,
		inst.maxVisibleRows,
		inst.selectionMode,
		inst.selectedIndex,
		inst.selectedIndices,
	)
}

func (inst *Instance) openDropdown() bool {
	if inst.overlayCallbacks != nil {
		return inst.requestOverlayOpen(true)
	}
	if inst.open || !inst.canOpenPopup() {
		return false
	}
	inst.open = true
	inst.ensureHighlightVisible()
	inst.dirty = true
	return true
}

func (inst *Instance) closeDropdown() bool {
	if inst.overlayCallbacks != nil {
		return inst.requestOverlayOpen(false)
	}
	if !inst.open {
		return false
	}
	inst.open = false
	if inst.filterQuery != "" {
		inst.filterQuery = ""
	}
	inst.normalizeSelectionState()
	inst.dirty = true
	if !inst.state.Focused {
		inst.emitFieldBlur()
	}
	return true
}

func (inst *Instance) moveHighlight(delta int) bool {
	rows := inst.popupRows()
	if inst.overlayCallbacks != nil {
		if len(selectableTargets(rows)) == 0 {
			return false
		}
		if !inst.open {
			return inst.requestOverlayOpen(true)
		}
		next := nextHighlightTarget(rows, inst.highlightedIndex, delta)
		if next == inst.highlightedIndex {
			return false
		}
		return inst.requestOverlayHighlight(next)
	}
	if len(selectableTargets(rows)) == 0 {
		return false
	}
	if !inst.open {
		return inst.openDropdown()
	}
	next := nextHighlightTarget(rows, inst.highlightedIndex, delta)
	if next == inst.highlightedIndex {
		return false
	}
	inst.highlightedIndex = next
	inst.ensureHighlightVisible()
	inst.dirty = true
	return true
}

func (inst *Instance) moveHighlightTo(index int) bool {
	rows := inst.popupRows()
	if inst.overlayCallbacks != nil {
		if len(selectableTargets(rows)) == 0 {
			return false
		}
		if !inst.open {
			inst.requestOverlayOpen(true)
		}
		next := index
		if rowPositionForHighlight(rows, next) < 0 {
			next = defaultHighlightTarget(rows, inst.selectionMode, inst.selectedIndex, inst.selectedIndices)
		}
		if next == inst.highlightedIndex {
			return false
		}
		return inst.requestOverlayHighlight(next)
	}
	if len(selectableTargets(rows)) == 0 {
		return false
	}
	if !inst.open {
		inst.open = true
	}
	next := index
	if rowPositionForHighlight(rows, next) < 0 {
		next = defaultHighlightTarget(rows, inst.selectionMode, inst.selectedIndex, inst.selectedIndices)
	}
	if next == inst.highlightedIndex {
		return false
	}
	inst.highlightedIndex = next
	inst.ensureHighlightVisible()
	inst.dirty = true
	return true
}

func (inst *Instance) pageHighlight(direction int) bool {
	rows := inst.popupRows()
	if len(selectableTargets(rows)) == 0 {
		return false
	}
	pageSize := maxInt(1, visibleScrollableRowCount(rows, inst.maxVisibleRows))
	return inst.moveHighlightTo(pageHighlightTarget(rows, inst.highlightedIndex, direction, pageSize))
}

func (inst *Instance) activateIndex(index int) bool {
	if index == highlightCreateTag {
		return inst.createTagFromQuery()
	}
	if inst.overlayCallbacks != nil {
		return inst.requestOverlayCommit(index)
	}
	if index < 0 || index >= len(inst.options) {
		return false
	}
	if isMultiSelectionMode(inst.selectionMode) {
		return inst.toggleIndex(index)
	}
	return inst.applySingleSelection(index, true)
}

func (inst *Instance) createTagFromQuery() bool {
	if !isTagsSelectionMode(inst.selectionMode) {
		return false
	}
	if inst.overlayCallbacks != nil {
		return inst.requestOverlayCreateTag()
	}

	query := strings.TrimSpace(inst.filterQuery)
	if query == "" {
		return false
	}
	if existing := findExactOptionIndex(inst.options, query); existing >= 0 {
		inst.filterQuery = ""
		if !inst.open {
			inst.open = true
		}
		return inst.toggleIndex(existing)
	}

	inst.createdOptions = append(inst.createdOptions, createTagOption(query))
	inst.syncOptions()
	newIndex := len(inst.options) - 1
	inst.filterQuery = ""
	if !inst.open {
		inst.open = true
	}
	return inst.toggleIndex(newIndex)
}

func (inst *Instance) applySingleSelection(index int, close bool) bool {
	if index < 0 || index >= len(inst.options) {
		return false
	}

	selectionChanged := inst.selectedIndex != index ||
		len(inst.selectedIndices) != 1 ||
		inst.selectedIndices[0] != index
	closed := close && inst.open

	inst.selectedIndex = index
	inst.selectedIndices = []int{index}
	inst.highlightedIndex = index
	if close {
		inst.open = false
		inst.filterQuery = ""
	}
	inst.ensureHighlightVisible()

	if selectionChanged || closed {
		inst.dirty = true
	}
	if selectionChanged {
		inst.emitChange()
		inst.emitSelectChange()
	}
	return selectionChanged || closed
}

func (inst *Instance) toggleIndex(index int) bool {
	if index < 0 || index >= len(inst.options) {
		return false
	}
	if !isMultiSelectionMode(inst.selectionMode) {
		return inst.applySingleSelection(index, true)
	}

	next := append([]int(nil), inst.selectedIndices...)
	pos := indexOfInt(next, index)
	if pos >= 0 {
		next = append(next[:pos], next[pos+1:]...)
	} else {
		next = append(next, index)
	}
	next = normalizeIndices(next, len(inst.options))
	if equalIntSlices(inst.selectedIndices, next) {
		return false
	}

	inst.selectedIndices = next
	if pos >= 0 {
		if inst.selectedIndex == index {
			if len(inst.selectedIndices) > 0 {
				inst.selectedIndex = inst.selectedIndices[len(inst.selectedIndices)-1]
			} else {
				inst.selectedIndex = -1
			}
		}
	} else {
		inst.selectedIndex = index
	}
	inst.highlightedIndex = index
	if isTagsSelectionMode(inst.selectionMode) {
		inst.filterQuery = ""
	}
	inst.ensureHighlightVisible()
	inst.dirty = true
	inst.emitChange()
	inst.emitSelectChange()
	return true
}

func (inst *Instance) triggerWidth() int {
	longest := paint.StringWidth(inst.placeholder)
	for _, opt := range inst.options {
		longest = maxInt(longest, paint.StringWidth(opt.Label))
	}
	longest = maxInt(longest, paint.StringWidth(inst.triggerDisplayLabel()))
	longest = maxInt(longest, 10)

	width := longest + 4
	if inst.width > 0 {
		width = inst.width
	}
	return maxInt(width, 6)
}

func (inst *Instance) popupWidth() int {
	contentWidth := popupContentWidth(inst.popupRows(), inst.selectionMode)
	width := contentWidth + 2
	width = maxInt(width, inst.triggerWidth())
	return maxInt(width, 6)
}

func (inst *Instance) popupHeight() int {
	rows := inst.popupRows()
	if !rows.showFilter && len(rows.scrollable) == 0 {
		return 0
	}
	height := visibleScrollableRowCount(rows, inst.maxVisibleRows) + 2
	if rows.showFilter {
		height++
	}
	return height
}

func (inst *Instance) visibleRowCount() int {
	return visibleScrollableRowCount(inst.popupRows(), inst.maxVisibleRows)
}

func (inst *Instance) maxScrollOffset() int {
	return maxScrollOffsetForRows(inst.popupRows(), inst.maxVisibleRows)
}

func (inst *Instance) ensureHighlightVisible() {
	inst.highlightedIndex, inst.scrollOffset = normalizePopupHighlight(
		inst.popupRows(),
		inst.highlightedIndex,
		inst.scrollOffset,
		inst.maxVisibleRows,
		inst.selectionMode,
		inst.selectedIndex,
		inst.selectedIndices,
	)
}

func (inst *Instance) optionIndexAt(localX, localY int) (int, bool) {
	return inst.popupOptionIndexAt(localX, localY, 2)
}

func (inst *Instance) popupOptionIndexAt(localX, localY, rowStart int) (int, bool) {
	if !inst.open || localX < 0 || localY < rowStart {
		return 0, false
	}
	return popupHitTarget(inst.popupRows(), inst.scrollOffset, inst.maxVisibleRows, localY-rowStart+1)
}

func (inst *Instance) optionRowText(index, width int) string {
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

func (inst *Instance) optionMarker(index int) string {
	return optionMarkerForMode(inst.selectionMode, inst.selectedIndices, index)
}

func (inst *Instance) optionMarkerWidth() int {
	return markerWidthForMode(inst.selectionMode)
}

func (inst *Instance) triggerText(width int) string {
	innerWidth := maxInt(0, width-4)
	label := truncateWithEllipsis(inst.triggerDisplayLabel(), innerWidth)
	return "< " + padDisplayWidth(label, innerWidth) + " >"
}

func (inst *Instance) triggerDisplayLabel() string {
	if isTagsSelectionMode(inst.selectionMode) {
		switch len(inst.selectedIndices) {
		case 0:
			return inst.placeholder
		default:
			return strings.Join(inst.SelectedLabels(), ", ")
		}
	}
	if inst.selectionMode == SelectionMultiple {
		switch len(inst.selectedIndices) {
		case 0:
			return inst.placeholder
		case 1:
			return inst.labelAt(inst.selectedIndices[0])
		default:
			return fmt.Sprintf("%d selected", len(inst.selectedIndices))
		}
	}
	if label := inst.SelectedLabel(); label != "" {
		return label
	}
	if len(inst.options) > 0 {
		return inst.options[0].Label
	}
	return inst.placeholder
}

func (inst *Instance) labelAt(index int) string {
	if index >= 0 && index < len(inst.options) {
		return inst.options[index].Label
	}
	return ""
}

func (inst *Instance) triggerPaintWidth() int {
	if inst.bounds[2] > 0 {
		return maxInt(inst.bounds[2], 6)
	}
	if inst.width > 0 {
		return maxInt(inst.width, 6)
	}
	return maxInt(paint.StringWidth(inst.triggerDisplayLabel())+4, 6)
}

func (inst *Instance) containsPoint(screenX, screenY int) bool {
	x, y, width, height := inst.GetBounds()
	return screenX >= x && screenX < x+width && screenY >= y && screenY < y+height
}

func (inst *Instance) syncOverlayRegistration() {
	// Overlay popup is now rendered declaratively by the wrapper component.
}

func (inst *Instance) unregisterOverlay() {
	// Overlay popup is now rendered declaratively by the wrapper component.
}

func (inst *Instance) markOverlayDirty() {
	// Overlay popup is now rendered declaratively by the wrapper component.
}

func getPortalRootProp(props rtui.Props, def string) string {
	if v, ok := props[popupPortalRootProp]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return proputil.GetString(props, "portalRoot", def)
}

func getIntsProp(props rtui.Props, key string, def []int) []int {
	if v, ok := props[key]; ok {
		if values, ok := v.([]int); ok {
			return append([]int(nil), values...)
		}
	}
	return append([]int(nil), def...)
}

func getChangeIntentFieldProp(props rtui.Props, key string) intent.FieldIntent {
	if v, ok := props[key]; ok {
		if fieldIntent, ok := v.(intent.FieldIntent); ok {
			return fieldIntent
		}
	}
	return nil
}

func getSelectionModeProp(props rtui.Props, def SelectionMode) SelectionMode {
	if v, ok := props[propSelectionMode]; ok {
		if mode, ok := v.(SelectionMode); ok {
			return mode
		}
	}
	return def
}

func getOptionsProp(props rtui.Props) []Option {
	if v, ok := props[propOptions]; ok {
		if opts, ok := v.([]Option); ok {
			return append([]Option(nil), opts...)
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

func equalOptions(a, b []Option) bool {
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

func normalizeIndices(indices []int, optionCount int) []int {
	if optionCount <= 0 || len(indices) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(indices))
	result := make([]int, 0, len(indices))
	for _, index := range indices {
		if index < 0 || index >= optionCount {
			continue
		}
		if _, exists := seen[index]; exists {
			continue
		}
		seen[index] = struct{}{}
		result = append(result, index)
	}
	return result
}

func containsInt(values []int, target int) bool {
	return indexOfInt(values, target) >= 0
}

func indexOfInt(values []int, target int) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

func joinIndices(indices []int) string {
	if len(indices) == 0 {
		return ""
	}
	parts := make([]string, len(indices))
	for i, index := range indices {
		parts[i] = fmt.Sprintf("%d", index)
	}
	return strings.Join(parts, ",")
}

func truncateWithEllipsis(content string, width int) string {
	if width <= 0 {
		return ""
	}
	const ellipsis = "…"
	if paint.StringWidth(content) <= width {
		return content
	}
	ellipsisWidth := paint.StringWidth(ellipsis)
	if width <= ellipsisWidth {
		return ellipsis
	}
	trimmed := strings.TrimRight(truncateByDisplayWidth(content, width-ellipsisWidth), " ")
	if trimmed == "" {
		return truncateByDisplayWidth(content, width)
	}
	return trimmed + ellipsis
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

func clampInt(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
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

func boolToIntSliceLen(ok bool) int {
	if ok {
		return 1
	}
	return 0
}
