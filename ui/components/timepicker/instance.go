package timepicker

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

const (
	defaultPlaceholder  = "HH:mm"
	defaultTriggerWidth = 10
	defaultPopupWidth   = 18
	popupCallbacksProp  = "_timepickerPopupCallbacks"
)

type timeSegment int

const (
	segmentHour timeSegment = iota
	segmentMinute
)

var nowFunc = time.Now

// Instance is the runtime entity for TimePicker components.
type Instance struct {
	key               string
	pickerID          string
	componentID       string
	pickerStyle       style.Style
	width             int
	placeholder       string
	valueControlled   bool
	disabled          bool
	portalRoot        string
	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent
	formID            string

	selectedMinutes    int
	highlightedMinutes int
	hasValue           bool
	draft              string
	open               bool
	hovered            bool
	focused            bool
	activeSegment      timeSegment
	bounds             [4]int
	dirty              bool

	parent         rtui.ComponentInstance
	childInstances []rtui.ComponentInstance
	intentEmitter  func(intent.Intent)
}

type popupCallbacks struct {
	setOpen            func(bool) bool
	switchSegment      func(int) bool
	stepSegment        func(int) bool
	setSegmentBoundary func(bool) bool
	commitHighlighted  func() bool
	commitMinutes      func(int) bool
	pickHour           func(int) bool
	pickMinute         func(int) bool
	props              func() rtui.Props
}

type popupVNode struct {
	*rtui.ElementVNode
}

type popupInstance struct {
	key                string
	popupID            string
	popupStyle         style.Style
	selectedMinutes    int
	highlightedMinutes int
	hasValue           bool
	activeSegment      timeSegment
	localStateDirty    bool
	minWidth           int
	disabled           bool
	focused            bool
	bounds             [4]int
	dirty              bool
	callbacks          *popupCallbacks
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.PaintableInstance       = (*Instance)(nil)
	_ rtui.FocusableInstance       = (*Instance)(nil)
	_ rtui.ActionHandlerInstance   = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
	_ rtui.TreeNode                = (*Instance)(nil)
	_ rtui.TreeContainer           = (*Instance)(nil)

	_ rtui.ComponentInstance     = (*popupInstance)(nil)
	_ rtui.PaintableInstance     = (*popupInstance)(nil)
	_ rtui.FocusableInstance     = (*popupInstance)(nil)
	_ rtui.ActionHandlerInstance = (*popupInstance)(nil)
)

// NewInstance creates a new TimePicker instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:               proputil.GetString(props, propKey, ""),
		pickerID:          proputil.GetString(props, propPickerID, ""),
		componentID:       proputil.GetString(props, propComponentID, ""),
		pickerStyle:       proputil.GetStyle(props, propPickerStyle, style.Style{}),
		width:             proputil.GetInt(props, propWidth, defaultTriggerWidth),
		placeholder:       proputil.GetString(props, propPlaceholder, defaultPlaceholder),
		valueControlled:   proputil.GetBool(props, propValueControlled, false),
		disabled:          proputil.GetBool(props, propDisabled, false),
		portalRoot:        proputil.GetString(props, propPortalRoot, rtui.DefaultOverlayPortalRootID),
		changeIntent:      proputil.GetIntent(props, propChangeIntent, nil),
		changeIntentField: getChangeIntentFieldProp(props, propChangeIntent),
		formID:            proputil.GetString(props, propFormID, ""),
		activeSegment:     segmentHour,
		dirty:             true,
	}
	inst.syncInitialValue(props)
	inst.normalizeState()
	return inst
}

func (inst *Instance) syncInitialValue(props rtui.Props) {
	if inst.valueControlled {
		inst.syncFromExternalValue(proputil.GetString(props, propValue, ""))
		return
	}
	if defaultValue := proputil.GetString(props, propDefaultValue, ""); defaultValue != "" {
		inst.syncFromExternalValue(defaultValue)
	}
}

func (inst *Instance) Key() string { return inst.key }

func (inst *Instance) SetKey(key string) { inst.key = key }

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) Destroy() {}

func (inst *Instance) OnMount() {}

func (inst *Instance) OnUnmount() {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.pickerID = proputil.GetString(props, propPickerID, inst.pickerID)
	inst.componentID = proputil.GetString(props, propComponentID, inst.componentID)
	inst.pickerStyle = proputil.GetStyle(props, propPickerStyle, inst.pickerStyle)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.placeholder = proputil.GetString(props, propPlaceholder, inst.placeholder)
	inst.disabled = proputil.GetBool(props, propDisabled, inst.disabled)
	inst.portalRoot = proputil.GetString(props, propPortalRoot, inst.portalRoot)
	inst.changeIntent = proputil.GetIntent(props, propChangeIntent, inst.changeIntent)
	inst.changeIntentField = getChangeIntentFieldProp(props, propChangeIntent)
	inst.formID = proputil.GetString(props, propFormID, inst.formID)

	inst.valueControlled = proputil.GetBool(props, propValueControlled, inst.valueControlled)
	if inst.valueControlled {
		inst.syncFromExternalValue(proputil.GetString(props, propValue, ""))
	}

	inst.normalizeState()
	if inst.open {
		inst.syncPopupState()
	}
	inst.dirty = true
	return true
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propChangeIntent:    inst.changeIntent,
		propComponentID:     inst.componentID,
		propDisabled:        inst.disabled,
		propFormID:          inst.formID,
		propKey:             inst.key,
		propPickerID:        inst.pickerID,
		propPickerStyle:     inst.pickerStyle,
		propPlaceholder:     inst.placeholder,
		propPortalRoot:      inst.portalRoot,
		propValue:           inst.selectedValue(),
		propValueControlled: inst.valueControlled,
		propWidth:           inst.width,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }

func (inst *Instance) IsDirty() bool { return inst.dirty }

func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Parent() interface{} { return inst.parent }

func (inst *Instance) SetParent(parent rtui.ComponentInstance) { inst.parent = parent }

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

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	if !inst.open {
		return nil
	}

	popup := newPopupVNode(inst.popupProps())
	popup.SetKey(inst.key + "-popup")
	popup.SetID(inst.popupID())

	portal := rtui.NewElement("box")
	portal.SetKey(inst.key + "-popup-portal")
	portal.SetID(inst.popupID() + "-portal")
	portal.SetLayer(rtui.LayerOverlay)
	portal.SetProps(rtui.Props{
		"position": "absolute",
		"left":     0,
		"top":      0,
		"width":    inst.popupWidth(),
		"height":   inst.popupHeight(),
	})
	portal.SetPortalRoot(inst.portalRoot)
	portal.SetAnchorTo(inst.anchorID(), rttypes.AnchorBottomLeft)
	portal.SetPortalPosition(rttypes.PositionAbsolute)
	portal.SetChildren([]rtui.VNode{popup})
	return []rtui.VNode{portal}
}

func (inst *Instance) popupCallbacks() *popupCallbacks {
	return &popupCallbacks{
		setOpen:            inst.setOpen,
		switchSegment:      inst.switchSegment,
		stepSegment:        inst.stepSegment,
		setSegmentBoundary: inst.setSegmentBoundary,
		commitHighlighted:  inst.commitHighlighted,
		commitMinutes:      inst.commitMinutesFromPopup,
		pickHour:           inst.pickHour,
		pickMinute:         inst.pickMinute,
		props:              inst.popupProps,
	}
}

func (inst *Instance) popupProps() rtui.Props {
	return rtui.Props{
		propKey:             inst.key + "-popup",
		"popupID":           inst.popupID(),
		propPickerStyle:     inst.pickerStyle,
		"selectedMinutes":   inst.selectedMinutes,
		"highlightedMinute": inst.highlightedMinutes,
		"hasValue":          inst.hasValue,
		"activeSegment":     inst.activeSegment,
		"minWidth":          inst.popupWidth(),
		propDisabled:        inst.disabled,
		popupCallbacksProp:  inst.popupCallbacks(),
	}
}

func (inst *Instance) syncPopupState() {
	if popup := inst.findPopupInstance(); popup != nil {
		popup.SetProps(inst.popupProps())
	}
}

func (inst *Instance) findPopupInstance() *popupInstance {
	for _, child := range inst.childInstances {
		if popup := findTimePickerPopupInstance(child); popup != nil {
			return popup
		}
	}
	return nil
}

func findTimePickerPopupInstance(node rtui.ComponentInstance) *popupInstance {
	if node == nil {
		return nil
	}
	if popup, ok := node.(*popupInstance); ok {
		return popup
	}
	tree, ok := node.(rtui.TreeNode)
	if !ok {
		return nil
	}
	for _, child := range tree.Children() {
		if popup := findTimePickerPopupInstance(child); popup != nil {
			return popup
		}
	}
	return nil
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  inst.triggerText(inst.triggerPaintWidth()),
		Style: inst.resolveStyle(),
	}}
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := constraints.ConstrainWidth(inst.triggerWidth())
	height := constraints.ConstrainHeight(1)
	return layout.Size{Width: width, Height: height}
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *Instance) SetFocus(focused bool) {
	if inst.focused == focused {
		return
	}
	inst.focused = focused
	if !focused && !inst.open {
		inst.applyDraftOnBlur()
		inst.emitFieldBlur()
	}
	inst.dirty = true
}

func (inst *Instance) HasFocus() bool { return inst.focused }

func (inst *Instance) IsDisabled() bool { return inst.disabled }

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil || inst.disabled {
		return false
	}

	switch act.Type {
	case action.ActionMouseEnter, action.ActionHover:
		if !inst.hovered {
			inst.hovered = true
			inst.dirty = true
		}
		return false
	case action.ActionMouseLeave:
		if inst.hovered {
			inst.hovered = false
			inst.dirty = true
		}
		return false
	case action.ActionClick:
		return inst.handleClick(act)
	case action.ActionMouseRelease:
		return true
	case action.ActionSelect:
		if _, ok := mousePayload(act.Payload); ok {
			return true
		}
		if inst.open {
			return inst.commitHighlighted()
		}
		return inst.openPicker()
	case action.ActionEnter, action.ActionSubmit:
		if inst.open {
			return inst.commitHighlighted()
		}
		return inst.openPicker()
	case action.ActionCancel:
		if inst.open {
			return inst.setOpen(false)
		}
		return inst.resetDraftToSelection()
	case action.ActionNavigateLeft:
		if !inst.open {
			return inst.openPicker()
		}
		return inst.switchSegment(-1)
	case action.ActionNavigateRight:
		if !inst.open {
			return inst.openPicker()
		}
		return inst.switchSegment(1)
	case action.ActionNavigateUp:
		if !inst.open {
			inst.openPicker()
		}
		return inst.stepSegment(-1)
	case action.ActionNavigateDown:
		if !inst.open {
			inst.openPicker()
		}
		return inst.stepSegment(1)
	case action.ActionNavigateHome:
		if !inst.open {
			inst.openPicker()
		}
		return inst.setSegmentBoundary(false)
	case action.ActionNavigateEnd:
		if !inst.open {
			inst.openPicker()
		}
		return inst.setSegmentBoundary(true)
	case action.ActionNavigatePageUp:
		if !inst.open {
			inst.openPicker()
		}
		return inst.adjustHour(-1)
	case action.ActionNavigatePageDown:
		if !inst.open {
			inst.openPicker()
		}
		return inst.adjustHour(1)
	case action.ActionInputChar:
		if inst.open {
			return false
		}
		if value, ok := act.GetPayloadRune(); ok {
			return inst.appendDraft(string(value))
		}
		return false
	case action.ActionInputText:
		if inst.open {
			return false
		}
		if value, ok := act.GetPayloadString(); ok {
			return inst.appendDraft(value)
		}
		return false
	case action.ActionBackspace:
		if inst.open {
			return false
		}
		return inst.backspaceDraft()
	case action.ActionDeleteChar, action.ActionClear:
		return inst.clearSelection(true)
	default:
		return false
	}
}

func (inst *Instance) handleClick(act *action.Action) bool {
	if _, ok := mousePayload(act.Payload); ok {
		return inst.setOpen(!inst.open)
	}
	if inst.open {
		return inst.commitHighlighted()
	}
	return inst.openPicker()
}

func (inst *Instance) setOpen(open bool) bool {
	if inst.open == open {
		return false
	}
	inst.open = open
	if open {
		inst.prepareOpenState()
	}
	inst.dirty = true
	return true
}

func (inst *Instance) openPicker() bool {
	return inst.setOpen(true)
}

func (inst *Instance) prepareOpenState() {
	inst.activeSegment = segmentHour
	switch {
	case inst.draft != "":
		if parsed, ok := parseTimeValue(inst.draft); ok {
			inst.highlightedMinutes = parsed
			return
		}
	case inst.hasValue:
		inst.highlightedMinutes = inst.selectedMinutes
		return
	}
	inst.highlightedMinutes = currentTimeMinutes()
}

func (inst *Instance) switchSegment(delta int) bool {
	next := segmentHour
	if inst.activeSegment == segmentHour {
		next = segmentMinute
	}
	if delta < 0 && inst.activeSegment == segmentMinute {
		next = segmentHour
	}
	if delta > 0 && inst.activeSegment == segmentHour {
		next = segmentMinute
	}
	if delta == 0 || next == inst.activeSegment {
		return false
	}
	inst.activeSegment = next
	inst.dirty = true
	inst.syncPopupState()
	return true
}

func (inst *Instance) stepSegment(delta int) bool {
	base := inst.highlightBaseMinutes()
	hour, minute := splitMinutes(base)
	switch inst.activeSegment {
	case segmentHour:
		hour = wrapMod(hour+delta, 24)
	case segmentMinute:
		minute = wrapMod(minute+delta, 60)
	}
	return inst.setHighlightedMinutes(hour*60 + minute)
}

func (inst *Instance) setSegmentBoundary(toEnd bool) bool {
	base := inst.highlightBaseMinutes()
	hour, minute := splitMinutes(base)
	switch inst.activeSegment {
	case segmentHour:
		if toEnd {
			hour = 23
		} else {
			hour = 0
		}
	case segmentMinute:
		if toEnd {
			minute = 59
		} else {
			minute = 0
		}
	}
	return inst.setHighlightedMinutes(hour*60 + minute)
}

func (inst *Instance) adjustHour(delta int) bool {
	base := inst.highlightBaseMinutes()
	hour, minute := splitMinutes(base)
	hour = wrapMod(hour+delta, 24)
	return inst.setHighlightedMinutes(hour*60 + minute)
}

func (inst *Instance) pickHour(hour int) bool {
	base := inst.highlightBaseMinutes()
	_, minute := splitMinutes(base)
	changed := inst.setHighlightedMinutes(wrapMod(hour, 24)*60 + minute)
	if inst.activeSegment != segmentMinute {
		inst.activeSegment = segmentMinute
		inst.dirty = true
		inst.syncPopupState()
		changed = true
	}
	return changed
}

func (inst *Instance) pickMinute(minute int) bool {
	base := inst.highlightBaseMinutes()
	hour, _ := splitMinutes(base)
	inst.setHighlightedMinutes(hour*60 + wrapMod(minute, 60))
	return inst.commitHighlighted()
}

func (inst *Instance) setHighlightedMinutes(total int) bool {
	total = wrapMinutes(total)
	if total == inst.highlightedMinutes {
		return false
	}
	inst.highlightedMinutes = total
	inst.dirty = true
	inst.syncPopupState()
	return true
}

func (inst *Instance) commitHighlighted() bool {
	return inst.commitMinutes(inst.highlightBaseMinutes(), true, true)
}

func (inst *Instance) commitMinutesFromPopup(total int) bool {
	return inst.commitMinutes(total, true, true)
}

func (inst *Instance) commitMinutes(total int, closePopup bool, emit bool) bool {
	total = wrapMinutes(total)
	formatted := formatTimeValue(total)
	changed := !inst.hasValue || inst.selectedMinutes != total || inst.draft != formatted || (closePopup && inst.open)
	inst.selectedMinutes = total
	inst.hasValue = true
	inst.draft = formatted
	inst.highlightedMinutes = total
	if closePopup {
		inst.open = false
	}
	inst.dirty = true
	if inst.open {
		inst.syncPopupState()
	}
	if emit {
		inst.emitChange()
	}
	return changed
}

func (inst *Instance) clearSelection(emit bool) bool {
	if !inst.hasValue && inst.draft == "" && !inst.open {
		return false
	}
	inst.hasValue = false
	inst.selectedMinutes = 0
	inst.draft = ""
	if inst.open {
		inst.highlightedMinutes = currentTimeMinutes()
		inst.activeSegment = segmentHour
		inst.syncPopupState()
	}
	inst.dirty = true
	if emit {
		inst.emitChange()
	}
	return true
}

func (inst *Instance) appendDraft(text string) bool {
	sanitized := sanitizeTimeInput(text)
	if sanitized == "" {
		return false
	}
	next := normalizeEditableTimeDraft(inst.draft + sanitized)
	if next == inst.draft {
		return false
	}
	inst.draft = next
	inst.dirty = true
	if isCompleteTimeValue(next) {
		if parsed, ok := parseTimeValue(next); ok {
			inst.commitMinutes(parsed, false, true)
		}
	}
	return true
}

func (inst *Instance) backspaceDraft() bool {
	if inst.draft == "" {
		return false
	}
	runes := []rune(inst.draft)
	next := string(runes[:len(runes)-1])
	inst.draft = next
	inst.dirty = true
	if next == "" {
		return inst.clearSelection(true)
	}
	if isCompleteTimeValue(next) {
		if parsed, ok := parseTimeValue(next); ok {
			inst.commitMinutes(parsed, false, true)
		}
	}
	return true
}

func (inst *Instance) resetDraftToSelection() bool {
	want := inst.selectedValue()
	if inst.draft == want {
		return false
	}
	inst.draft = want
	inst.dirty = true
	return true
}

func (inst *Instance) applyDraftOnBlur() {
	if inst.draft == "" {
		return
	}
	if normalized, ok := normalizeBlurTimeValue(inst.draft); ok {
		if parsed, parsedOK := parseTimeValue(normalized); parsedOK {
			inst.commitMinutes(parsed, false, true)
			return
		}
	}
	inst.resetDraftToSelection()
}

func (inst *Instance) emitChange() {
	inst.emitFieldValueChanged()
	inst.emitTimeChange()
}

func (inst *Instance) emitFieldValueChanged() {
	if inst.intentEmitter == nil {
		return
	}
	value := inst.selectedValue()
	if inst.formID != "" && inst.changeIntentField != nil {
		intent.Emit(inst, form.FieldChange(inst.formID, inst.changeIntentField.GetField(), value, true))
		return
	}
	if inst.changeIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
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
	intent.Emit(inst, form.FieldBlur(inst.formID, inst.changeIntentField.GetField(), inst.selectedValue()))
}

func (inst *Instance) emitTimeChange() {
	if inst.intentEmitter == nil {
		return
	}
	if !inst.hasValue {
		inst.intentEmitter(TimeChangeIntent{
			Value:       "",
			ComponentID: inst.componentID,
		})
		return
	}
	hour, minute := splitMinutes(inst.selectedMinutes)
	inst.intentEmitter(TimeChangeWithID(
		inst.componentID,
		inst.selectedValue(),
		hour,
		minute,
	))
}

func (inst *Instance) resolveStyle() style.Style {
	s := inst.pickerStyle
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
	if inst.hovered {
		return s.Underline(true)
	}
	return s
}

func (inst *Instance) triggerWidth() int {
	width := maxInt(defaultTriggerWidth, paint.StringWidth(inst.placeholder)+4)
	width = maxInt(width, paint.StringWidth(inst.displayLabel())+4)
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

func (inst *Instance) popupWidth() int {
	return maxInt(defaultPopupWidth, inst.triggerWidth())
}

func (inst *Instance) popupHeight() int {
	return 9
}

func (inst *Instance) triggerText(width int) string {
	innerWidth := maxInt(0, width-4)
	label := truncateWithEllipsis(inst.displayLabel(), innerWidth)
	return "< " + padDisplayWidth(label, innerWidth) + " >"
}

func (inst *Instance) displayLabel() string {
	if strings.TrimSpace(inst.draft) != "" {
		return inst.draft
	}
	if value := inst.selectedValue(); value != "" {
		return value
	}
	return inst.placeholder
}

func (inst *Instance) selectedValue() string {
	if !inst.hasValue {
		return ""
	}
	return formatTimeValue(inst.selectedMinutes)
}

func (inst *Instance) highlightBaseMinutes() int {
	return wrapMinutes(inst.highlightedMinutes)
}

func (inst *Instance) normalizeState() {
	if strings.TrimSpace(inst.placeholder) == "" {
		inst.placeholder = defaultPlaceholder
	}
	if inst.width <= 0 {
		inst.width = defaultTriggerWidth
	}
	if inst.portalRoot == "" {
		inst.portalRoot = rtui.DefaultOverlayPortalRootID
	}
	if inst.activeSegment != segmentMinute {
		inst.activeSegment = segmentHour
	}
	if inst.hasValue {
		inst.selectedMinutes = wrapMinutes(inst.selectedMinutes)
		inst.highlightedMinutes = inst.selectedMinutes
		inst.draft = inst.selectedValue()
		return
	}
	if inst.highlightedMinutes == 0 && strings.TrimSpace(inst.draft) == "" {
		inst.highlightedMinutes = currentTimeMinutes()
	}
}

func (inst *Instance) syncFromExternalValue(raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		inst.hasValue = false
		inst.selectedMinutes = 0
		inst.draft = ""
		inst.highlightedMinutes = currentTimeMinutes()
		return
	}
	if parsed, ok := parseTimeValue(raw); ok {
		inst.selectedMinutes = parsed
		inst.hasValue = true
		inst.draft = formatTimeValue(parsed)
		inst.highlightedMinutes = parsed
		return
	}
	inst.hasValue = false
	inst.selectedMinutes = 0
	inst.draft = raw
	inst.highlightedMinutes = currentTimeMinutes()
}

func (inst *Instance) popupID() string { return inst.anchorID() + "-popup" }

func (inst *Instance) anchorID() string {
	return firstNonEmpty(inst.pickerID, inst.componentID, inst.key)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return "timepicker"
}

func newPopupVNode(props rtui.Props) *popupVNode {
	node := &popupVNode{ElementVNode: rtui.NewElement("timepicker-popup")}
	node.SetProps(props)
	return node
}

func (v *popupVNode) SetProps(p rtui.Props) rtui.VNode {
	existing := v.ElementVNode.Props()
	if existing == nil {
		existing = make(rtui.Props)
	}
	v.ElementVNode.SetProps(existing.Merge(p))
	return v
}

func (v *popupVNode) GetLayer() rtui.Layer { return rtui.LayerOverlay }

func (v *popupVNode) SetLayer(l rtui.Layer) rtui.VNode {
	v.SetProp("_layer", l)
	return v
}

func (v *popupVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props[propKey] = v.Key()
	return newPopupInstance(props)
}

func newPopupInstance(props rtui.Props) *popupInstance {
	inst := &popupInstance{dirty: true}
	inst.SetProps(props)
	return inst
}

func (inst *popupInstance) Key() string { return inst.key }

func (inst *popupInstance) SetKey(key string) { inst.key = key }

func (inst *popupInstance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *popupInstance) Destroy() {}

func (inst *popupInstance) OnMount() {}

func (inst *popupInstance) OnUnmount() {}

func (inst *popupInstance) SetProps(props rtui.Props) bool {
	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.popupID = proputil.GetString(props, "popupID", inst.popupID)
	inst.popupStyle = proputil.GetStyle(props, propPickerStyle, inst.popupStyle)
	inst.selectedMinutes = wrapMinutes(proputil.GetInt(props, "selectedMinutes", inst.selectedMinutes))
	if !inst.localStateDirty {
		inst.highlightedMinutes = wrapMinutes(proputil.GetInt(props, "highlightedMinute", inst.highlightedMinutes))
		if segment, ok := props["activeSegment"].(timeSegment); ok {
			inst.activeSegment = segment
		}
	}
	inst.hasValue = proputil.GetBool(props, "hasValue", inst.hasValue)
	inst.minWidth = proputil.GetInt(props, "minWidth", inst.minWidth)
	inst.disabled = proputil.GetBool(props, propDisabled, inst.disabled)
	if callbacks, ok := props[popupCallbacksProp].(*popupCallbacks); ok {
		inst.callbacks = callbacks
	}
	if inst.minWidth < defaultPopupWidth {
		inst.minWidth = defaultPopupWidth
	}
	if inst.activeSegment != segmentMinute {
		inst.activeSegment = segmentHour
	}
	inst.dirty = true
	return true
}

func (inst *popupInstance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:             inst.key,
		"popupID":           inst.popupID,
		propPickerStyle:     inst.popupStyle,
		"selectedMinutes":   inst.selectedMinutes,
		"highlightedMinute": inst.highlightedMinutes,
		"hasValue":          inst.hasValue,
		"activeSegment":     inst.activeSegment,
		"minWidth":          inst.minWidth,
		propDisabled:        inst.disabled,
		popupCallbacksProp:  inst.callbacks,
	}
}

func (inst *popupInstance) MarkDirty() { inst.dirty = true }

func (inst *popupInstance) IsDirty() bool { return inst.dirty }

func (inst *popupInstance) GetContext() *rtui.ComponentContext { return nil }

func (inst *popupInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *popupInstance) Measure(constraints layout.Constraints) layout.Size {
	return layout.Size{
		Width:  constraints.ConstrainWidth(inst.popupWidth()),
		Height: constraints.ConstrainHeight(inst.popupHeight()),
	}
}

func (inst *popupInstance) Paint(x, y int) []paint.DrawCmd {
	width := inst.popupWidth()
	height := inst.popupHeight()
	contentWidth := width - 2
	fillStyle := inst.fillStyle()
	borderStyle := inst.borderStyle()
	titleStyle := inst.titleStyle()
	headerStyle := inst.headerStyle()
	cmds := []paint.DrawCmd{
		{X: x, Y: y, Text: "┌" + strings.Repeat("─", width-2) + "┐", Style: borderStyle},
		{X: x, Y: y + height - 1, Text: "└" + strings.Repeat("─", width-2) + "┘", Style: borderStyle},
	}

	for row := 1; row < height-1; row++ {
		cmds = append(cmds,
			paint.DrawCmd{X: x, Y: y + row, Text: "│", Style: borderStyle},
			paint.DrawCmd{X: x + 1, Y: y + row, Text: strings.Repeat(" ", contentWidth), Style: fillStyle},
			paint.DrawCmd{X: x + width - 1, Y: y + row, Text: "│", Style: borderStyle},
		)
	}

	cmds = append(cmds, paint.DrawCmd{
		X:     x + 1,
		Y:     y + 1,
		Text:  padDisplayWidth(centerText("Time", contentWidth), contentWidth),
		Style: titleStyle,
	})

	leftOffset, sepOffset, rightOffset := inst.columnOffsets()
	leftCol, sepX, rightCol := x+leftOffset, x+sepOffset, x+rightOffset
	cmds = append(cmds,
		paint.DrawCmd{X: leftCol, Y: y + 2, Text: "HH", Style: headerStyleForSegment(headerStyle, inst.activeSegment == segmentHour)},
		paint.DrawCmd{X: sepX, Y: y + 2, Text: ":", Style: headerStyle},
		paint.DrawCmd{X: rightCol, Y: y + 2, Text: "MM", Style: headerStyleForSegment(headerStyle, inst.activeSegment == segmentMinute)},
	)

	baseHour, baseMinute := splitMinutes(inst.highlightedMinutes)
	selectedHour, selectedMinute := splitMinutes(inst.selectedMinutes)
	for row := 0; row < 5; row++ {
		offset := row - 2
		rowY := y + 3 + row
		hourValue := wrapMod(baseHour+offset, 24)
		minuteValue := wrapMod(baseMinute+offset, 60)
		cmds = append(cmds,
			paint.DrawCmd{
				X:     leftCol,
				Y:     rowY,
				Text:  fmt.Sprintf("%02d", hourValue),
				Style: inst.segmentValueStyle(segmentHour, hourValue, baseHour, selectedHour),
			},
			paint.DrawCmd{
				X:     sepX,
				Y:     rowY,
				Text:  ":",
				Style: fillStyle.Foreground(theme.Muted()),
			},
			paint.DrawCmd{
				X:     rightCol,
				Y:     rowY,
				Text:  fmt.Sprintf("%02d", minuteValue),
				Style: inst.segmentValueStyle(segmentMinute, minuteValue, baseMinute, selectedMinute),
			},
		)
	}

	return cmds
}

func (inst *popupInstance) SetFocus(focused bool) {
	inst.focused = focused
	inst.dirty = true
}

func (inst *popupInstance) HasFocus() bool { return inst.focused }

func (inst *popupInstance) IsDisabled() bool { return inst.disabled }

func (inst *popupInstance) HandleAction(act *action.Action) bool {
	if act == nil || inst.disabled || inst.callbacks == nil {
		return false
	}
	switch act.Type {
	case action.ActionClick:
		return inst.handleClick(act)
	case action.ActionMouseRelease:
		return true
	case action.ActionSelect:
		if _, ok := mousePayload(act.Payload); ok {
			return true
		}
		return inst.commitHighlighted()
	case action.ActionEnter, action.ActionSubmit:
		return inst.commitHighlighted()
	case action.ActionCancel:
		return inst.callbacks.setOpen(false)
	case action.ActionNavigateLeft:
		return inst.switchSegment(-1)
	case action.ActionNavigateRight:
		return inst.switchSegment(1)
	case action.ActionNavigateUp:
		return inst.stepSegment(-1)
	case action.ActionNavigateDown:
		return inst.stepSegment(1)
	case action.ActionNavigateHome:
		return inst.setSegmentBoundary(false)
	case action.ActionNavigateEnd:
		return inst.setSegmentBoundary(true)
	case action.ActionNavigatePageUp:
		return inst.adjustHour(-1)
	case action.ActionNavigatePageDown:
		return inst.adjustHour(1)
	default:
		return false
	}
}

func (inst *popupInstance) handleClick(act *action.Action) bool {
	mouse, ok := mousePayload(act.Payload)
	if !ok {
		return false
	}
	localX := mouse.LocalX
	localY := mouse.LocalY
	if localY < 3 || localY > 7 {
		return true
	}
	row := localY - 3
	offset := row - 2
	baseHour, baseMinute := splitMinutes(inst.highlightedMinutes)
	leftOffset, _, rightOffset := inst.columnOffsets()
	if localX >= leftOffset && localX <= leftOffset+1 {
		return inst.pickHour(wrapMod(baseHour+offset, 24))
	}
	if localX >= rightOffset && localX <= rightOffset+1 {
		return inst.pickMinute(wrapMod(baseMinute+offset, 60))
	}
	return true
}

func (inst *popupInstance) switchSegment(delta int) bool {
	next := inst.activeSegment
	if delta < 0 && inst.activeSegment == segmentMinute {
		next = segmentHour
	}
	if delta > 0 && inst.activeSegment == segmentHour {
		next = segmentMinute
	}
	if delta == 0 || next == inst.activeSegment {
		return false
	}
	inst.activeSegment = next
	inst.localStateDirty = true
	inst.dirty = true
	return true
}

func (inst *popupInstance) stepSegment(delta int) bool {
	baseHour, baseMinute := splitMinutes(inst.highlightedMinutes)
	switch inst.activeSegment {
	case segmentHour:
		baseHour = wrapMod(baseHour+delta, 24)
	case segmentMinute:
		baseMinute = wrapMod(baseMinute+delta, 60)
	}
	return inst.setHighlightedMinutes(baseHour*60 + baseMinute)
}

func (inst *popupInstance) setSegmentBoundary(toEnd bool) bool {
	baseHour, baseMinute := splitMinutes(inst.highlightedMinutes)
	switch inst.activeSegment {
	case segmentHour:
		if toEnd {
			baseHour = 23
		} else {
			baseHour = 0
		}
	case segmentMinute:
		if toEnd {
			baseMinute = 59
		} else {
			baseMinute = 0
		}
	}
	return inst.setHighlightedMinutes(baseHour*60 + baseMinute)
}

func (inst *popupInstance) adjustHour(delta int) bool {
	baseHour, baseMinute := splitMinutes(inst.highlightedMinutes)
	baseHour = wrapMod(baseHour+delta, 24)
	return inst.setHighlightedMinutes(baseHour*60 + baseMinute)
}

func (inst *popupInstance) pickHour(hour int) bool {
	_, minute := splitMinutes(inst.highlightedMinutes)
	changed := inst.setHighlightedMinutes(wrapMod(hour, 24)*60 + minute)
	if inst.activeSegment != segmentMinute {
		inst.activeSegment = segmentMinute
		inst.dirty = true
		changed = true
	}
	return changed
}

func (inst *popupInstance) pickMinute(minute int) bool {
	hour, _ := splitMinutes(inst.highlightedMinutes)
	inst.setHighlightedMinutes(hour*60 + wrapMod(minute, 60))
	return inst.commitHighlighted()
}

func (inst *popupInstance) setHighlightedMinutes(total int) bool {
	total = wrapMinutes(total)
	if total == inst.highlightedMinutes {
		return false
	}
	inst.highlightedMinutes = total
	inst.localStateDirty = true
	inst.dirty = true
	return true
}

func (inst *popupInstance) commitHighlighted() bool {
	if inst.callbacks == nil || inst.callbacks.commitMinutes == nil {
		return false
	}
	return inst.callbacks.commitMinutes(inst.highlightedMinutes)
}

func (inst *popupInstance) popupWidth() int { return maxInt(inst.minWidth, defaultPopupWidth) }

func (inst *popupInstance) popupHeight() int { return 9 }

func (inst *popupInstance) fillStyle() style.Style {
	s := inst.popupStyle
	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}
	if inst.disabled {
		return s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	return s.Background(theme.Surface())
}

func (inst *popupInstance) borderStyle() style.Style {
	if inst.disabled {
		return style.NewStyle().Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	return style.NewStyle().Foreground(theme.Focus()).Background(theme.Surface()).Bold(true)
}

func (inst *popupInstance) titleStyle() style.Style {
	return style.NewStyle().Foreground(theme.Text()).Background(theme.Surface()).Bold(true)
}

func (inst *popupInstance) headerStyle() style.Style {
	return style.NewStyle().Foreground(theme.Muted()).Background(theme.Surface()).Bold(true)
}

func headerStyleForSegment(base style.Style, active bool) style.Style {
	if !active {
		return base
	}
	return base.Foreground(theme.Focus())
}

func (inst *popupInstance) segmentValueStyle(segment timeSegment, value, highlighted, selected int) style.Style {
	if inst.disabled {
		return style.NewStyle().Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	switch {
	case value == highlighted && inst.activeSegment == segment:
		return style.NewStyle().Foreground(theme.Text()).Background(theme.Select()).Bold(true)
	case value == highlighted:
		return style.NewStyle().Foreground(theme.Text()).Background(theme.Surface()).Underline(true)
	case inst.hasValue && value == selected:
		return style.NewStyle().Foreground(theme.BG()).Background(theme.Primary()).Bold(true)
	default:
		return style.NewStyle().Foreground(theme.Text()).Background(theme.Surface())
	}
}

func (inst *popupInstance) columnOffsets() (leftCol, sepX, rightCol int) {
	contentWidth := inst.popupWidth() - 2
	rowWidth := paint.StringWidth("HH:MM")
	start := 1 + maxInt(0, (contentWidth-rowWidth)/2)
	return start, start + 2, start + 3
}

func parseTimeValue(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return 0, false
	}
	if len(parts[0]) == 0 || len(parts[0]) > 2 || len(parts[1]) == 0 || len(parts[1]) > 2 {
		return 0, false
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, false
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, false
	}
	return hour*60 + minute, true
}

func formatTimeValue(totalMinutes int) string {
	hour, minute := splitMinutes(totalMinutes)
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

func normalizeBlurTimeValue(raw string) (string, bool) {
	if parsed, ok := parseTimeValue(normalizeEditableTimeDraft(raw)); ok {
		return formatTimeValue(parsed), true
	}
	return "", false
}

func isCompleteTimeValue(raw string) bool {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) != 2 || len(parts[1]) != 2 {
		return false
	}
	_, ok := parseTimeValue(raw)
	return ok
}

func sanitizeTimeInput(text string) string {
	if text == "" {
		return ""
	}
	var builder strings.Builder
	for _, r := range text {
		if (r >= '0' && r <= '9') || r == ':' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func normalizeEditableTimeDraft(text string) string {
	var builder strings.Builder
	seenColon := false
	hourDigits := 0
	minuteDigits := 0
	for _, r := range []rune(text) {
		switch {
		case r >= '0' && r <= '9':
			if !seenColon {
				if hourDigits >= 2 {
					continue
				}
				hourDigits++
			} else {
				if minuteDigits >= 2 {
					continue
				}
				minuteDigits++
			}
			builder.WriteRune(r)
		case r == ':':
			if seenColon || hourDigits == 0 {
				continue
			}
			seenColon = true
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func getChangeIntentFieldProp(props rtui.Props, key string) intent.FieldIntent {
	if value, ok := props[key]; ok {
		if fieldIntent, ok := value.(intent.FieldIntent); ok {
			return fieldIntent
		}
	}
	return nil
}

func currentTimeMinutes() int {
	now := nowFunc().UTC()
	return wrapMinutes(now.Hour()*60 + now.Minute())
}

func wrapMinutes(total int) int {
	return wrapMod(total, 24*60)
}

func wrapMod(value, mod int) int {
	if mod <= 0 {
		return 0
	}
	value %= mod
	if value < 0 {
		value += mod
	}
	return value
}

func splitMinutes(total int) (hour, minute int) {
	total = wrapMinutes(total)
	return total / 60, total % 60
}

func hourPart(total int) int {
	hour, _ := splitMinutes(total)
	return hour
}

func centerText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	textWidth := paint.StringWidth(text)
	if textWidth >= width {
		return truncateWithEllipsis(text, width)
	}
	left := (width - textWidth) / 2
	right := width - textWidth - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

func truncateWithEllipsis(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if paint.StringWidth(text) <= width {
		return text
	}
	if width <= 1 {
		return "…"
	}
	runes := []rune(text)
	var builder strings.Builder
	currentWidth := 0
	for _, r := range runes {
		runeWidth := paint.RuneWidth(r)
		if currentWidth+runeWidth >= width {
			break
		}
		builder.WriteRune(r)
		currentWidth += runeWidth
	}
	return strings.TrimRight(builder.String(), " ") + "…"
}

func padDisplayWidth(text string, width int) string {
	currentWidth := paint.StringWidth(text)
	if currentWidth >= width {
		return text
	}
	return text + strings.Repeat(" ", width-currentWidth)
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
