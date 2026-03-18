package popconfirm

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

type confirmClickIntent struct{ ComponentID string }
type cancelClickIntent struct{ ComponentID string }

func (confirmClickIntent) IntentType() string              { return "popconfirm.confirmClickIntent" }
func (confirmClickIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (confirmClickIntent) IsTransition() bool              { return false }
func (confirmClickIntent) IsGlobal() bool                  { return false }
func (confirmClickIntent) StayPressed() bool               { return true }
func (i confirmClickIntent) GetComponentID() string        { return i.ComponentID }
func (cancelClickIntent) IntentType() string               { return "popconfirm.cancelClickIntent" }
func (cancelClickIntent) Priority() intent.ActionPriority  { return intent.PriorityNormal }
func (cancelClickIntent) IsTransition() bool               { return false }
func (cancelClickIntent) IsGlobal() bool                   { return false }
func (cancelClickIntent) StayPressed() bool                { return true }
func (i cancelClickIntent) GetComponentID() string         { return i.ComponentID }

// Instance is the runtime entity for Popconfirm components.
type Instance struct {
	key               string
	componentID       string
	anchorID          string
	title             string
	description       string
	placement         Placement
	trigger           TriggerMode
	open              bool
	openControlled    bool
	disabled          bool
	showArrow         bool
	showCancel        bool
	gapRows           int
	maxWidth          int
	okText            string
	cancelText        string
	rootStyle         style.Style
	overlayStyle      style.Style
	titleStyle        style.Style
	textStyle         style.Style
	okButtonStyle     style.Style
	confirmIntent     intent.Intent
	cancelIntent      intent.Intent
	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent
	parent            rtui.ComponentInstance
	childInstances    []rtui.ComponentInstance
	intentEmitter     func(intent.Intent)
	bounds            [4]int
	dirty             bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
	_ rtui.TreeNode                = (*Instance)(nil)
	_ rtui.TreeContainer           = (*Instance)(nil)
	_ rtui.ActionHandlerInstance   = (*Instance)(nil)
	_ intent.IntentHandler         = (*Instance)(nil)
	_ intent.TreeComponent         = (*Instance)(nil)
)

func NewInstance(props rtui.Props) *Instance {
	openControlled := proputil.GetBool(props, propOpenControlled, false)
	open := proputil.GetBool(props, propInitialOpen, false)
	if openControlled {
		open = proputil.GetBool(props, propOpen, false)
	}
	inst := &Instance{
		key:               proputil.GetString(props, propKey, ""),
		componentID:       proputil.GetString(props, propComponentID, ""),
		anchorID:          proputil.GetString(props, propAnchorID, ""),
		title:             proputil.GetString(props, propTitle, ""),
		description:       proputil.GetString(props, propDescription, ""),
		placement:         getPlacementProp(props, PlacementTop),
		trigger:           getTriggerProp(props, TriggerClick),
		open:              open,
		openControlled:    openControlled,
		disabled:          proputil.GetBool(props, propDisabled, false),
		showArrow:         proputil.GetBool(props, propShowArrow, true),
		showCancel:        proputil.GetBool(props, propShowCancel, true),
		gapRows:           proputil.GetInt(props, propGapRows, 1),
		maxWidth:          proputil.GetInt(props, propMaxWidth, 36),
		okText:            proputil.GetString(props, propOkText, "OK"),
		cancelText:        proputil.GetString(props, propCancelText, "Cancel"),
		rootStyle:         proputil.GetStyle(props, propRootStyle, style.Style{}),
		overlayStyle:      proputil.GetStyle(props, propOverlayStyle, style.Style{}),
		titleStyle:        proputil.GetStyle(props, propTitleStyle, style.Style{}),
		textStyle:         proputil.GetStyle(props, propTextStyle, style.Style{}),
		okButtonStyle:     proputil.GetStyle(props, propOkButtonStyle, style.Style{}),
		confirmIntent:     proputil.GetIntent(props, propConfirmIntent, nil),
		cancelIntent:      proputil.GetIntent(props, propCancelIntent, nil),
		changeIntent:      proputil.GetIntent(props, propChangeIntent, nil),
		changeIntentField: getFieldIntentProp(props, propChangeIntentField),
		dirty:             true,
	}
	inst.normalize()
	return inst
}

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              {}
func (inst *Instance) OnMount()              {}
func (inst *Instance) OnUnmount()            {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := *inst
	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.componentID = proputil.GetString(props, propComponentID, inst.componentID)
	inst.anchorID = proputil.GetString(props, propAnchorID, inst.anchorID)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.description = proputil.GetString(props, propDescription, inst.description)
	inst.placement = getPlacementProp(props, inst.placement)
	inst.trigger = getTriggerProp(props, inst.trigger)
	nextControlled := proputil.GetBool(props, propOpenControlled, inst.openControlled)
	if nextControlled {
		inst.open = proputil.GetBool(props, propOpen, inst.open)
	} else if old.openControlled && !nextControlled {
		inst.open = proputil.GetBool(props, propInitialOpen, inst.open)
	}
	inst.openControlled = nextControlled
	inst.disabled = proputil.GetBool(props, propDisabled, inst.disabled)
	inst.showArrow = proputil.GetBool(props, propShowArrow, inst.showArrow)
	inst.showCancel = proputil.GetBool(props, propShowCancel, inst.showCancel)
	inst.gapRows = proputil.GetInt(props, propGapRows, inst.gapRows)
	inst.maxWidth = proputil.GetInt(props, propMaxWidth, inst.maxWidth)
	inst.okText = proputil.GetString(props, propOkText, inst.okText)
	inst.cancelText = proputil.GetString(props, propCancelText, inst.cancelText)
	inst.rootStyle = proputil.GetStyle(props, propRootStyle, inst.rootStyle)
	inst.overlayStyle = proputil.GetStyle(props, propOverlayStyle, inst.overlayStyle)
	inst.titleStyle = proputil.GetStyle(props, propTitleStyle, inst.titleStyle)
	inst.textStyle = proputil.GetStyle(props, propTextStyle, inst.textStyle)
	inst.okButtonStyle = proputil.GetStyle(props, propOkButtonStyle, inst.okButtonStyle)
	inst.confirmIntent = proputil.GetIntent(props, propConfirmIntent, inst.confirmIntent)
	inst.cancelIntent = proputil.GetIntent(props, propCancelIntent, inst.cancelIntent)
	inst.changeIntent = proputil.GetIntent(props, propChangeIntent, inst.changeIntent)
	inst.changeIntentField = getFieldIntentProp(props, propChangeIntentField)
	inst.normalize()

	changed := old.key != inst.key ||
		old.componentID != inst.componentID ||
		old.anchorID != inst.anchorID ||
		old.title != inst.title ||
		old.description != inst.description ||
		old.placement != inst.placement ||
		old.trigger != inst.trigger ||
		old.open != inst.open ||
		old.openControlled != inst.openControlled ||
		old.disabled != inst.disabled ||
		old.showArrow != inst.showArrow ||
		old.showCancel != inst.showCancel ||
		old.gapRows != inst.gapRows ||
		old.maxWidth != inst.maxWidth ||
		old.okText != inst.okText ||
		old.cancelText != inst.cancelText ||
		old.rootStyle != inst.rootStyle ||
		old.overlayStyle != inst.overlayStyle ||
		old.titleStyle != inst.titleStyle ||
		old.textStyle != inst.textStyle ||
		old.okButtonStyle != inst.okButtonStyle ||
		old.confirmIntent != inst.confirmIntent ||
		old.cancelIntent != inst.cancelIntent ||
		old.changeIntent != inst.changeIntent ||
		old.changeIntentField != inst.changeIntentField
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propAnchorID:          inst.anchorID,
		propCancelIntent:      inst.cancelIntent,
		propCancelText:        inst.cancelText,
		propChangeIntent:      inst.changeIntent,
		propChangeIntentField: inst.changeIntentField,
		propComponentID:       inst.componentID,
		propConfirmIntent:     inst.confirmIntent,
		propDescription:       inst.description,
		propDisabled:          inst.disabled,
		propGapRows:           inst.gapRows,
		propKey:               inst.key,
		propMaxWidth:          inst.maxWidth,
		propOkButtonStyle:     inst.okButtonStyle,
		propOkText:            inst.okText,
		propOpen:              inst.open,
		propOpenControlled:    inst.openControlled,
		propOverlayStyle:      inst.overlayStyle,
		propPlacement:         inst.placement,
		propRootStyle:         inst.rootStyle,
		propShowArrow:         inst.showArrow,
		propShowCancel:        inst.showCancel,
		propTextStyle:         inst.textStyle,
		propTitle:             inst.title,
		propTitleStyle:        inst.titleStyle,
		propTrigger:           inst.trigger,
	}
}

func (inst *Instance) MarkDirty()                              { inst.dirty = true }
func (inst *Instance) IsDirty() bool                           { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext      { return nil }
func (inst *Instance) Parent() interface{}                     { return inst.parent }
func (inst *Instance) SetParent(parent rtui.ComponentInstance) { inst.parent = parent }
func (inst *Instance) Children() []rtui.ComponentInstance {
	return append([]rtui.ComponentInstance(nil), inst.childInstances...)
}

func (inst *Instance) AddChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	for i, existing := range inst.childInstances {
		if existing == child || existing.Key() == child.Key() {
			inst.childInstances[i] = child
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
	for i, existing := range inst.childInstances {
		if existing != child {
			continue
		}
		inst.childInstances = append(inst.childInstances[:i], inst.childInstances[i+1:]...)
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

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) { inst.intentEmitter = fn }

func (inst *Instance) SetBounds(x, y, w, h int) { inst.bounds = [4]int{x, y, w, h} }

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil || inst.disabled {
		return false
	}
	switch inst.trigger {
	case TriggerHover:
		switch act.Type {
		case action.ActionMouseEnter:
			return inst.setOpen(true, TriggerHover)
		case action.ActionMouseLeave, action.ActionCancel:
			return inst.setOpen(false, TriggerHover)
		}
	case TriggerManual:
		if act.Type == action.ActionCancel {
			return inst.setOpen(false, TriggerManual)
		}
	default:
		switch act.Type {
		case action.ActionClick, action.ActionEnter, action.ActionSelect:
			return inst.setOpen(!inst.open, TriggerClick)
		case action.ActionCancel:
			return inst.setOpen(false, TriggerClick)
		}
	}
	return false
}

func (inst *Instance) HandleIntent(i intent.Intent) bool {
	if !intent.ShouldHandleIntentWithID(inst.componentID, i) {
		return false
	}
	switch i.(type) {
	case PopconfirmToggleIntent:
		return inst.setOpen(!inst.open, inst.trigger)
	case PopconfirmOpenIntent:
		return inst.setOpen(true, inst.trigger)
	case PopconfirmCloseIntent:
		return inst.setOpen(false, inst.trigger)
	case confirmClickIntent:
		inst.setOpen(false, TriggerClick)
		inst.emitResult(PopconfirmConfirmIntent{ComponentID: inst.componentID}, inst.confirmIntent)
		return true
	case cancelClickIntent:
		inst.setOpen(false, TriggerClick)
		inst.emitResult(PopconfirmCancelIntent{ComponentID: inst.componentID}, inst.cancelIntent)
		return true
	default:
		return false
	}
}

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	if !inst.open || strings.TrimSpace(inst.title) == "" {
		return nil
	}
	portal := rtui.NewElement("box")
	portal.SetKey(inst.key + "-portal")
	portal.SetProps(rtui.Props{
		"position": "absolute",
		"left":     inst.portalOffsetX(),
		"top":      inst.portalOffsetY(),
		"width":    1,
		"height":   1,
	})
	portal.SetLayer(rtui.LayerOverlay)
	portal.SetPortalRoot(rtui.DefaultOverlayPortalRootID)
	portal.SetAnchorTo(inst.anchorID, inst.anchor())
	portal.SetPortalPosition(rttypes.PositionAbsolute)
	portal.SetChildren([]rtui.VNode{inst.buildOverlaySurface()})
	return []rtui.VNode{portal}
}

func (inst *Instance) buildOverlaySurface() rtui.VNode {
	children := make([]rtui.VNode, 0, 4)
	title := textcomp.New(inst.title).Bold(true).SetStyleProps(style.NewStyle().Foreground(theme.Text()).Bold(true).Merge(inst.titleStyle))
	children = append(children, title)
	if desc := strings.TrimSpace(inst.description); desc != "" {
		body := textcomp.New(desc).
			SetWrap(true).
			SetMaxWidth(inst.maxWidth).
			SetStyleProps(style.NewStyle().Foreground(theme.Text()).Merge(inst.textStyle))
		children = append(children, body)
	}
	children = append(children, inst.buildActionRow())

	surface := rtui.VStackBuilder(children...).Gap(1).AlignCross(rtui.AlignStart)
	surface.Width(inst.overlayWidth())
	surface.SingleBorder()
	surface.SetBorderColor(theme.Primary())
	surface.SetStyleProps(style.NewStyle().Background(theme.Surface()).Foreground(theme.Text()).Merge(inst.overlayStyle))
	node := surface.Build()
	node.SetKey(inst.key + "-surface")
	return node
}

func (inst *Instance) buildActionRow() rtui.VNode {
	children := make([]rtui.VNode, 0, 3)
	if inst.showCancel {
		cancelBtn := button.NewBuilder(inst.cancelText).
			Key(inst.key + "-cancel").
			Secondary().
			Small().
			OnPress(cancelClickIntent{ComponentID: inst.componentID}).
			Build()
		children = append(children, cancelBtn)
	}
	okBuilder := button.NewBuilder(inst.okText).
		Key(inst.key + "-ok").
		Primary().
		Small().
		OnPress(confirmClickIntent{ComponentID: inst.componentID})
	if !inst.okButtonStyle.IsEmpty() {
		okBuilder.Style(inst.okButtonStyle)
	}
	children = append(children, okBuilder.Build())

	row := rtui.HStackBuilder(children...).Gap(1).AlignCross(rtui.AlignCenter)
	node := row.Build()
	node.SetKey(inst.key + "-actions")
	return node
}

func (inst *Instance) setOpen(next bool, trigger TriggerMode) bool {
	if inst.open == next && !inst.openControlled {
		return false
	}
	if !inst.openControlled {
		inst.open = next
		inst.dirty = true
	}
	inst.emitChange(next, trigger)
	return true
}

func (inst *Instance) emitChange(open bool, trigger TriggerMode) {
	if inst.intentEmitter == nil {
		return
	}
	inst.intentEmitter(PopconfirmChangeIntent{
		ComponentID: inst.componentID,
		Open:        open,
		Trigger:     trigger,
	})
	if inst.changeIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: fmt.Sprintf("%t", open),
		})
	}
	if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
}

func (inst *Instance) emitResult(primary intent.Intent, extra intent.Intent) {
	if inst.intentEmitter == nil {
		return
	}
	if primary != nil {
		inst.intentEmitter(primary)
	}
	if extra != nil {
		inst.intentEmitter(extra)
	}
}

func (inst *Instance) normalize() {
	if inst.anchorID == "" {
		inst.anchorID = inst.componentID + "-anchor"
	}
	if inst.anchorID == "-anchor" || inst.anchorID == "" {
		inst.anchorID = "popconfirm-anchor"
	}
	if inst.gapRows < 0 {
		inst.gapRows = 0
	}
	if inst.maxWidth <= 0 {
		inst.maxWidth = 36
	}
	if strings.TrimSpace(inst.okText) == "" {
		inst.okText = "OK"
	}
	if strings.TrimSpace(inst.cancelText) == "" {
		inst.cancelText = "Cancel"
	}
}

func (inst *Instance) overlayWidth() int {
	width := paint.StringWidth(inst.title)
	for _, line := range strings.Split(strings.ReplaceAll(inst.description, "\r\n", "\n"), "\n") {
		lineWidth := paint.StringWidth(line)
		if lineWidth > width {
			width = lineWidth
		}
	}
	actionWidth := paint.StringWidth("[" + inst.okText + "]")
	if inst.showCancel {
		actionWidth += 1 + paint.StringWidth("["+inst.cancelText+"]")
	}
	if actionWidth > width {
		width = actionWidth
	}
	if inst.maxWidth > 0 && width > inst.maxWidth {
		width = inst.maxWidth
	}
	if width < 12 {
		width = 12
	}
	return width + 4
}

func (inst *Instance) anchor() rttypes.Anchor {
	switch inst.placement {
	case PlacementBottom, PlacementBottomLeft:
		return rttypes.AnchorTopLeft
	case PlacementBottomRight:
		return rttypes.AnchorTopRight
	case PlacementTop:
		return rttypes.AnchorBottom
	case PlacementTopRight:
		return rttypes.AnchorBottomRight
	default:
		return rttypes.AnchorBottomLeft
	}
}

func (inst *Instance) portalOffsetX() int {
	switch inst.placement {
	case PlacementTop, PlacementBottom:
		return inst.bounds[2] / 2
	default:
		return 0
	}
}

func (inst *Instance) portalOffsetY() int {
	switch inst.placement {
	case PlacementBottom, PlacementBottomLeft, PlacementBottomRight:
		return inst.bounds[3] + inst.gapRows
	default:
		return -inst.gapRows
	}
}

func getPlacementProp(props rtui.Props, def Placement) Placement {
	if value, ok := props[propPlacement]; ok {
		if placement, ok := value.(Placement); ok {
			return placement
		}
	}
	return def
}

func getTriggerProp(props rtui.Props, def TriggerMode) TriggerMode {
	if value, ok := props[propTrigger]; ok {
		if trigger, ok := value.(TriggerMode); ok {
			return trigger
		}
	}
	return def
}

func getFieldIntentProp(props rtui.Props, key string) intent.FieldIntent {
	if value, ok := props[key]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	return nil
}
