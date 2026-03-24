package datepicker

import (
	"fmt"
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
	dateLayout          = "2006-01-02"
	defaultPlaceholder  = "YYYY-MM-DD"
	defaultTriggerWidth = 16
	defaultPopupWidth   = 23
	popupCallbacksProp  = "_datepickerPopupCallbacks"
)

var nowFunc = time.Now

// Instance is the runtime entity for DatePicker components.
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

	selectedDate    time.Time
	hasValue        bool
	displayMonth    time.Time
	highlightedDate time.Time
	draft           string
	open            bool
	hovered         bool
	focused         bool
	bounds          [4]int
	dirty           bool

	parent         rtui.ComponentInstance
	childInstances []rtui.ComponentInstance
	intentEmitter  func(intent.Intent)
}

type popupCallbacks struct {
	setOpen           func(bool) bool
	navigateDay       func(int) bool
	navigateWeek      func(int) bool
	navigateMonth     func(int) bool
	jumpWeekBoundary  func(bool) bool
	commitHighlighted func() bool
	commitDate        func(time.Time) bool
	props             func() rtui.Props
}

type popupVNode struct {
	*rtui.ElementVNode
}

type popupInstance struct {
	key             string
	popupID         string
	popupStyle      style.Style
	displayMonth    time.Time
	selectedDate    time.Time
	hasValue        bool
	highlightedDate time.Time
	localStateDirty bool
	minWidth        int
	disabled        bool
	focused         bool
	bounds          [4]int
	dirty           bool
	callbacks       *popupCallbacks
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

// NewInstance creates a new DatePicker instance.
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
		setOpen:           inst.setOpen,
		navigateDay:       inst.navigateDay,
		navigateWeek:      inst.navigateWeek,
		navigateMonth:     inst.navigateMonth,
		jumpWeekBoundary:  inst.jumpWeekBoundary,
		commitHighlighted: inst.commitHighlighted,
		commitDate:        inst.commitDateFromPopup,
		props:             inst.popupProps,
	}
}

func (inst *Instance) popupProps() rtui.Props {
	return rtui.Props{
		propKey:            inst.key + "-popup",
		"popupID":          inst.popupID(),
		propPickerStyle:    inst.pickerStyle,
		"displayMonth":     inst.displayMonth,
		"selectedDate":     inst.selectedDate,
		"hasValue":         inst.hasValue,
		"highlighted":      inst.highlightedDate,
		"minWidth":         inst.popupWidth(),
		propDisabled:       inst.disabled,
		popupCallbacksProp: inst.popupCallbacks(),
	}
}

func (inst *Instance) syncPopupState() {
	if popup := inst.findPopupInstance(); popup != nil {
		popup.SetProps(inst.popupProps())
	}
}

func (inst *Instance) findPopupInstance() *popupInstance {
	for _, child := range inst.childInstances {
		if popup := findDatePickerPopupInstance(child); popup != nil {
			return popup
		}
	}
	return nil
}

func findDatePickerPopupInstance(node rtui.ComponentInstance) *popupInstance {
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
		if popup := findDatePickerPopupInstance(child); popup != nil {
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
			inst.openPicker()
		}
		return inst.navigateDay(-1)
	case action.ActionNavigateRight:
		if !inst.open {
			inst.openPicker()
		}
		return inst.navigateDay(1)
	case action.ActionNavigateUp:
		if !inst.open {
			inst.openPicker()
		}
		return inst.navigateWeek(-1)
	case action.ActionNavigateDown:
		if !inst.open {
			inst.openPicker()
		}
		return inst.navigateWeek(1)
	case action.ActionNavigateHome:
		if !inst.open {
			inst.openPicker()
		}
		return inst.jumpWeekBoundary(false)
	case action.ActionNavigateEnd:
		if !inst.open {
			inst.openPicker()
		}
		return inst.jumpWeekBoundary(true)
	case action.ActionNavigatePageUp:
		if !inst.open {
			inst.openPicker()
		}
		return inst.navigateMonth(-1)
	case action.ActionNavigatePageDown:
		if !inst.open {
			inst.openPicker()
		}
		return inst.navigateMonth(1)
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
	switch {
	case inst.draft != "":
		if parsed, ok := parseDateValue(inst.draft); ok {
			inst.highlightedDate = parsed
			inst.displayMonth = monthStart(parsed)
			return
		}
	case inst.hasValue:
		inst.highlightedDate = inst.selectedDate
		inst.displayMonth = monthStart(inst.selectedDate)
		return
	}
	today := currentDate()
	inst.highlightedDate = today
	inst.displayMonth = monthStart(today)
}

func (inst *Instance) navigateDay(delta int) bool {
	base := inst.highlightBaseDate()
	next := normalizeDate(base.AddDate(0, 0, delta))
	if sameDate(next, inst.highlightedDate) {
		return false
	}
	inst.highlightedDate = next
	inst.displayMonth = monthStart(next)
	inst.dirty = true
	inst.syncPopupState()
	return true
}

func (inst *Instance) navigateWeek(delta int) bool {
	return inst.navigateDay(delta * 7)
}

func (inst *Instance) navigateMonth(delta int) bool {
	base := inst.highlightBaseDate()
	next := addMonthsClamped(base, delta)
	if sameDate(next, inst.highlightedDate) {
		return false
	}
	inst.highlightedDate = next
	inst.displayMonth = monthStart(next)
	inst.dirty = true
	inst.syncPopupState()
	return true
}

func (inst *Instance) jumpWeekBoundary(toEnd bool) bool {
	base := inst.highlightBaseDate()
	offset := weekdayIndexMonday(base.Weekday())
	if toEnd {
		offset = 6 - offset
	} else {
		offset = -offset
	}
	return inst.navigateDay(offset)
}

func (inst *Instance) commitHighlighted() bool {
	return inst.commitDate(inst.highlightBaseDate(), true, true)
}

func (inst *Instance) commitDateFromPopup(date time.Time) bool {
	return inst.commitDate(date, true, true)
}

func (inst *Instance) commitDate(date time.Time, closePopup bool, emit bool) bool {
	date = normalizeDate(date)
	changed := !inst.hasValue || !sameDate(inst.selectedDate, date) || inst.draft != formatDateValue(date) || (closePopup && inst.open)
	inst.selectedDate = date
	inst.hasValue = true
	inst.draft = formatDateValue(date)
	inst.highlightedDate = date
	inst.displayMonth = monthStart(date)
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
	inst.selectedDate = time.Time{}
	inst.draft = ""
	if inst.open {
		today := currentDate()
		inst.highlightedDate = today
		inst.displayMonth = monthStart(today)
		inst.syncPopupState()
	}
	inst.dirty = true
	if emit {
		inst.emitChange()
	}
	return true
}

func (inst *Instance) appendDraft(text string) bool {
	sanitized := sanitizeDateInput(text)
	if sanitized == "" {
		return false
	}
	next := truncateDateDraft(inst.draft + sanitized)
	if next == inst.draft {
		return false
	}
	inst.draft = next
	inst.dirty = true
	if parsed, ok := parseDateValue(next); ok {
		inst.commitDate(parsed, false, true)
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
	if parsed, ok := parseDateValue(next); ok {
		inst.commitDate(parsed, false, true)
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
	if parsed, ok := parseDateValue(inst.draft); ok {
		inst.commitDate(parsed, false, true)
		return
	}
	inst.resetDraftToSelection()
}

func (inst *Instance) emitChange() {
	inst.emitFieldValueChanged()
	inst.emitDateChange()
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

func (inst *Instance) emitDateChange() {
	if inst.intentEmitter == nil {
		return
	}
	if !inst.hasValue {
		inst.intentEmitter(DateChangeIntent{
			Value:       "",
			ComponentID: inst.componentID,
		})
		return
	}
	inst.intentEmitter(DateChangeWithID(
		inst.componentID,
		inst.selectedValue(),
		inst.selectedDate.Year(),
		int(inst.selectedDate.Month()),
		inst.selectedDate.Day(),
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
	return 10
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
	return formatDateValue(inst.selectedDate)
}

func (inst *Instance) highlightBaseDate() time.Time {
	if !inst.highlightedDate.IsZero() {
		return inst.highlightedDate
	}
	if inst.hasValue {
		return inst.selectedDate
	}
	return currentDate()
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
	if inst.hasValue {
		inst.selectedDate = normalizeDate(inst.selectedDate)
		inst.highlightedDate = inst.selectedDate
		inst.displayMonth = monthStart(inst.selectedDate)
		inst.draft = inst.selectedValue()
		return
	}
	if inst.displayMonth.IsZero() {
		inst.displayMonth = monthStart(currentDate())
	}
	if inst.highlightedDate.IsZero() {
		inst.highlightedDate = currentDate()
	}
}

func (inst *Instance) syncFromExternalValue(raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		inst.hasValue = false
		inst.selectedDate = time.Time{}
		inst.draft = ""
		inst.displayMonth = monthStart(currentDate())
		inst.highlightedDate = currentDate()
		return
	}
	if parsed, ok := parseDateValue(raw); ok {
		inst.selectedDate = parsed
		inst.hasValue = true
		inst.draft = formatDateValue(parsed)
		inst.displayMonth = monthStart(parsed)
		inst.highlightedDate = parsed
		return
	}
	inst.hasValue = false
	inst.selectedDate = time.Time{}
	inst.draft = raw
	inst.displayMonth = monthStart(currentDate())
	inst.highlightedDate = currentDate()
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
	return "datepicker"
}

func newPopupVNode(props rtui.Props) *popupVNode {
	node := &popupVNode{ElementVNode: rtui.NewElement("datepicker-popup")}
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
	if !inst.localStateDirty {
		if month, ok := props["displayMonth"].(time.Time); ok {
			inst.displayMonth = monthStart(month)
		}
		if highlighted, ok := props["highlighted"].(time.Time); ok {
			inst.highlightedDate = normalizeDate(highlighted)
		}
	}
	if selected, ok := props["selectedDate"].(time.Time); ok {
		inst.selectedDate = normalizeDate(selected)
	}
	inst.hasValue = proputil.GetBool(props, "hasValue", inst.hasValue)
	if month, ok := props["displayMonth"].(time.Time); ok && inst.displayMonth.IsZero() {
		inst.displayMonth = monthStart(month)
	}
	if highlighted, ok := props["highlighted"].(time.Time); ok && inst.highlightedDate.IsZero() {
		inst.highlightedDate = normalizeDate(highlighted)
	}
	inst.minWidth = proputil.GetInt(props, "minWidth", inst.minWidth)
	inst.disabled = proputil.GetBool(props, propDisabled, inst.disabled)
	if callbacks, ok := props[popupCallbacksProp].(*popupCallbacks); ok {
		inst.callbacks = callbacks
	}
	if inst.minWidth < defaultPopupWidth {
		inst.minWidth = defaultPopupWidth
	}
	inst.dirty = true
	return true
}

func (inst *popupInstance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:            inst.key,
		"popupID":          inst.popupID,
		propPickerStyle:    inst.popupStyle,
		"displayMonth":     inst.displayMonth,
		"selectedDate":     inst.selectedDate,
		"hasValue":         inst.hasValue,
		"highlighted":      inst.highlightedDate,
		"minWidth":         inst.minWidth,
		propDisabled:       inst.disabled,
		popupCallbacksProp: inst.callbacks,
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

	title := inst.monthTitle()
	cmds = append(cmds,
		paint.DrawCmd{X: x + 1, Y: y + 1, Text: "‹", Style: borderStyle},
		paint.DrawCmd{X: x + width - 2, Y: y + 1, Text: "›", Style: borderStyle},
		paint.DrawCmd{X: x + 2, Y: y + 1, Text: padDisplayWidth(centerText(title, maxInt(0, contentWidth-2)), maxInt(0, contentWidth-2)), Style: inst.titleStyle()},
	)

	cmds = append(cmds, paint.DrawCmd{
		X:     x + 1,
		Y:     y + 2,
		Text:  padDisplayWidth("Mo Tu We Th Fr Sa Su", contentWidth),
		Style: inst.weekdayStyle(),
	})

	grid := calendarGrid(inst.displayMonth)
	for row := 0; row < 6; row++ {
		rowY := y + 3 + row
		for col := 0; col < 7; col++ {
			index := row*7 + col
			cellDate := grid[index]
			cellText := fmt.Sprintf("%2d ", cellDate.Day())
			cmds = append(cmds, paint.DrawCmd{
				X:     x + 1 + col*3,
				Y:     rowY,
				Text:  cellText,
				Style: inst.dayStyle(cellDate),
			})
		}
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
		return inst.navigateDay(-1)
	case action.ActionNavigateRight:
		return inst.navigateDay(1)
	case action.ActionNavigateUp:
		return inst.navigateWeek(-1)
	case action.ActionNavigateDown:
		return inst.navigateWeek(1)
	case action.ActionNavigateHome:
		return inst.jumpWeekBoundary(false)
	case action.ActionNavigateEnd:
		return inst.jumpWeekBoundary(true)
	case action.ActionNavigatePageUp:
		return inst.navigateMonth(-1)
	case action.ActionNavigatePageDown:
		return inst.navigateMonth(1)
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
	if localY == 1 {
		if localX <= 2 {
			return inst.navigateMonth(-1)
		}
		if localX >= inst.popupWidth()-2 {
			return inst.navigateMonth(1)
		}
		return true
	}
	if localY < 3 || localY > 8 || localX < 1 || localX > 21 {
		return true
	}
	col := (localX - 1) / 3
	row := localY - 3
	index := row*7 + col
	grid := calendarGrid(inst.displayMonth)
	if index < 0 || index >= len(grid) {
		return true
	}
	return inst.callbacks.commitDate(grid[index])
}

func (inst *popupInstance) navigateDay(delta int) bool {
	next := normalizeDate(inst.highlightedDate.AddDate(0, 0, delta))
	if sameDate(next, inst.highlightedDate) {
		return false
	}
	inst.highlightedDate = next
	inst.displayMonth = monthStart(next)
	inst.localStateDirty = true
	inst.dirty = true
	return true
}

func (inst *popupInstance) navigateWeek(delta int) bool {
	return inst.navigateDay(delta * 7)
}

func (inst *popupInstance) navigateMonth(delta int) bool {
	next := addMonthsClamped(inst.highlightedDate, delta)
	if sameDate(next, inst.highlightedDate) {
		return false
	}
	inst.highlightedDate = next
	inst.displayMonth = monthStart(next)
	inst.localStateDirty = true
	inst.dirty = true
	return true
}

func (inst *popupInstance) jumpWeekBoundary(toEnd bool) bool {
	offset := weekdayIndexMonday(inst.highlightedDate.Weekday())
	if toEnd {
		offset = 6 - offset
	} else {
		offset = -offset
	}
	return inst.navigateDay(offset)
}

func (inst *popupInstance) commitHighlighted() bool {
	return inst.callbacks.commitDate(inst.highlightedDate)
}

func (inst *popupInstance) popupWidth() int { return maxInt(inst.minWidth, defaultPopupWidth) }

func (inst *popupInstance) popupHeight() int { return 10 }

func (inst *popupInstance) monthTitle() string { return inst.displayMonth.Format("2006-01") }

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

func (inst *popupInstance) weekdayStyle() style.Style {
	return style.NewStyle().Foreground(theme.Muted()).Background(theme.Surface()).Bold(true)
}

func (inst *popupInstance) dayStyle(day time.Time) style.Style {
	if inst.disabled {
		return style.NewStyle().Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}
	switch {
	case inst.hasValue && sameDate(day, inst.selectedDate):
		return style.NewStyle().Foreground(theme.BG()).Background(theme.Primary()).Bold(true)
	case sameDate(day, inst.highlightedDate):
		return style.NewStyle().Foreground(theme.Text()).Background(theme.Select()).Bold(true)
	case sameDate(day, currentDate()):
		return style.NewStyle().Foreground(theme.BG()).Background(theme.Accent()).Bold(true)
	case !sameMonth(day, inst.displayMonth):
		return style.NewStyle().Foreground(theme.DisabledFG()).Background(theme.Surface())
	default:
		return style.NewStyle().Foreground(theme.Text()).Background(theme.Surface())
	}
}

func parseDateValue(raw string) (time.Time, bool) {
	parsed, err := time.ParseInLocation(dateLayout, strings.TrimSpace(raw), time.UTC)
	if err != nil {
		return time.Time{}, false
	}
	return normalizeDate(parsed), true
}

func formatDateValue(date time.Time) string {
	return normalizeDate(date).Format(dateLayout)
}

func normalizeDate(date time.Time) time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func monthStart(date time.Time) time.Time {
	date = normalizeDate(date)
	if date.IsZero() {
		date = currentDate()
	}
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func currentDate() time.Time {
	return normalizeDate(nowFunc().UTC())
}

func sameDate(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return a.IsZero() && b.IsZero()
	}
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func sameMonth(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return false
	}
	return a.Year() == b.Year() && a.Month() == b.Month()
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func addMonthsClamped(date time.Time, delta int) time.Time {
	base := normalizeDate(date)
	targetMonth := monthStart(base).AddDate(0, delta, 0)
	day := minInt(base.Day(), daysInMonth(targetMonth.Year(), targetMonth.Month()))
	return normalizeDate(time.Date(targetMonth.Year(), targetMonth.Month(), day, 0, 0, 0, 0, time.UTC))
}

func weekdayIndexMonday(weekday time.Weekday) int {
	if weekday == time.Sunday {
		return 6
	}
	return int(weekday - time.Monday)
}

func calendarGrid(month time.Time) []time.Time {
	start := monthStart(month)
	firstWeekday := weekdayIndexMonday(start.Weekday())
	gridStart := start.AddDate(0, 0, -firstWeekday)
	grid := make([]time.Time, 42)
	for index := range grid {
		grid[index] = normalizeDate(gridStart.AddDate(0, 0, index))
	}
	return grid
}

func sanitizeDateInput(text string) string {
	if text == "" {
		return ""
	}
	var builder strings.Builder
	for _, r := range text {
		if (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func truncateDateDraft(text string) string {
	runes := []rune(text)
	if len(runes) <= len(dateLayout) {
		return string(runes)
	}
	return string(runes[:len(dateLayout)])
}

func getChangeIntentFieldProp(props rtui.Props, key string) intent.FieldIntent {
	if value, ok := props[key]; ok {
		if fieldIntent, ok := value.(intent.FieldIntent); ok {
			return fieldIntent
		}
	}
	return nil
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
