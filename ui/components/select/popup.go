package selectcomp

import (
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
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
	return rtui.LayerBase
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
	key            string
	ownerID        string
	closeOnOutside bool
	bounds         [4]int
	dirty          bool
}

var (
	_ rtui.ComponentInstance     = (*popupInstance)(nil)
	_ rtui.PaintableInstance     = (*popupInstance)(nil)
	_ rtui.ActionHandlerInstance = (*popupInstance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
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
func (inst *popupInstance) MarkDirty()                         { inst.dirty = true }
func (inst *popupInstance) IsDirty() bool                      { return inst.dirty }
func (inst *popupInstance) GetContext() *rtui.ComponentContext { return nil }

func (inst *popupInstance) Destroy() {
	inst.unregister()
}

func (inst *popupInstance) OnMount() {
	selectOverlayRegistry.registerPopup(inst.ownerID, inst)
}

func (inst *popupInstance) OnUnmount() {
	inst.unregister()
}

func (inst *popupInstance) unregister() {
	selectOverlayRegistry.unregisterPopup(inst.ownerID, inst)
}

func (inst *popupInstance) SetProps(props rtui.Props) bool {
	oldOwnerID := inst.ownerID
	oldCloseOnOutside := inst.closeOnOutside

	inst.key = getStringProp(props, "key", inst.key)
	inst.ownerID = getStringProp(props, "ownerID", inst.ownerID)
	inst.closeOnOutside = getBoolProp(props, "closeOnOutside", inst.closeOnOutside)
	if oldOwnerID != "" && oldOwnerID != inst.ownerID {
		selectOverlayRegistry.unregisterPopup(oldOwnerID, inst)
	}
	if inst.ownerID != "" {
		selectOverlayRegistry.registerPopup(inst.ownerID, inst)
	}

	changed := oldOwnerID != inst.ownerID || oldCloseOnOutside != inst.closeOnOutside
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *popupInstance) GetProps() rtui.Props {
	return rtui.Props{
		"key":            inst.key,
		"ownerID":        inst.ownerID,
		"closeOnOutside": inst.closeOnOutside,
	}
}

func (inst *popupInstance) Measure(constraints layout.Constraints) layout.Size {
	owner := selectOverlayRegistry.trigger(inst.ownerID)
	if owner == nil || !owner.open {
		return layout.Size{}
	}
	width := constraints.ConstrainWidth(owner.popupWidth())
	height := constraints.ConstrainHeight(owner.popupHeight())
	return layout.Size{Width: width, Height: height}
}

func (inst *popupInstance) Paint(x, y int) []paint.DrawCmd {
	owner := selectOverlayRegistry.trigger(inst.ownerID)
	if owner == nil || !owner.open {
		return nil
	}
	return owner.paintPopupAt(x, y)
}

func (inst *popupInstance) HandleAction(act *action.Action) bool {
	owner := selectOverlayRegistry.trigger(inst.ownerID)
	if act == nil || owner == nil || !owner.open || owner.state.Disabled {
		return false
	}

	switch act.Type {
	case action.ActionHover:
		mouse, ok := popupMousePayload(act.Payload)
		if !ok {
			return false
		}
		index, hit := owner.popupOptionIndexAt(mouse.LocalX, mouse.LocalY, 1)
		if !hit || index == owner.highlightedIndex {
			return false
		}
		owner.highlightedIndex = index
		owner.ensureHighlightVisible()
		owner.MarkDirty()
		return true
	case action.ActionClick:
		mouse, ok := popupMousePayload(act.Payload)
		if !ok {
			return owner.activateIndex(owner.highlightedIndex)
		}
		index, hit := owner.popupOptionIndexAt(mouse.LocalX, mouse.LocalY, 1)
		if !hit {
			return true
		}
		owner.highlightedIndex = index
		owner.ensureHighlightVisible()
		return owner.activateIndex(index)
	case action.ActionScroll:
		return owner.handleScroll(act)
	case action.ActionSelect, action.ActionEnter, action.ActionSubmit:
		return owner.activateIndex(owner.highlightedIndex)
	case action.ActionCancel:
		return owner.closeDropdown()
	}
	return false
}

func (inst *popupInstance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *popupInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *popupInstance) containsPoint(screenX, screenY int) bool {
	x, y, width, height := inst.GetBounds()
	return screenX >= x && screenX < x+width && screenY >= y && screenY < y+height
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
