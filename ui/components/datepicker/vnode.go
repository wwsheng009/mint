package datepicker

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propChangeIntent    = "changeIntent"
	propComponentID     = "componentID"
	propDefaultValue    = "defaultValue"
	propDisabled        = "disabled"
	propFormID          = "formID"
	propKey             = "key"
	propPickerStyle     = "style"
	propPickerID        = "pickerID"
	propPlaceholder     = "placeholder"
	propPortalRoot      = "portalRoot"
	propValue           = "value"
	propValueControlled = "valueControlled"
	propWidth           = "width"
)

// VNode is the declarative description of a DatePicker component.
type VNode struct {
	*rtui.ElementVNode

	key             string
	componentID     string
	pickerStyle     style.Style
	width           int
	placeholder     string
	value           string
	defaultValue    string
	valueControlled bool
	disabled        bool
	portalRoot      string
	changeIntent    intent.Intent
	formID          string
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new DatePicker VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("datepicker"),
		placeholder:  defaultPlaceholder,
		portalRoot:   rtui.DefaultOverlayPortalRootID,
		width:        16,
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to component ID/key for popup anchoring.
func (v *VNode) ID() string {
	if id := v.ElementVNode.ID(); id != "" {
		return id
	}
	if v.componentID != "" {
		return v.componentID
	}
	return v.key
}

func (v *VNode) SetID(id string) rtui.VNode {
	v.ElementVNode.SetID(id)
	return v
}

func (v *VNode) Tag() string { return "datepicker" }

func (v *VNode) Style() style.Style { return v.pickerStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.pickerStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propChangeIntent:    v.changeIntent,
		propComponentID:     v.componentID,
		propDefaultValue:    v.defaultValue,
		propDisabled:        v.disabled,
		propFormID:          v.formID,
		propKey:             v.key,
		propPickerID:        v.ID(),
		propPickerStyle:     v.pickerStyle,
		propPlaceholder:     v.placeholder,
		propPortalRoot:      v.portalRoot,
		propValue:           v.value,
		propValueControlled: v.valueControlled,
		propWidth:           v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if componentID, ok := props[propComponentID].(string); ok {
		v.componentID = componentID
	}
	if pickerID, ok := props[propPickerID].(string); ok && pickerID != "" {
		v.ElementVNode.SetID(pickerID)
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if placeholder, ok := props[propPlaceholder].(string); ok {
		v.placeholder = placeholder
	}
	if value, ok := props[propValue].(string); ok {
		v.value = value
	}
	if defaultValue, ok := props[propDefaultValue].(string); ok {
		v.defaultValue = defaultValue
	}
	if valueControlled, ok := props[propValueControlled].(bool); ok {
		v.valueControlled = valueControlled
	}
	if disabled, ok := props[propDisabled].(bool); ok {
		v.disabled = disabled
	}
	if portalRoot, ok := props[propPortalRoot].(string); ok {
		v.portalRoot = portalRoot
	}
	if changeIntent, ok := props[propChangeIntent].(intent.Intent); ok {
		v.changeIntent = changeIntent
	}
	if formID, ok := props[propFormID].(string); ok {
		v.formID = formID
	}
	if s, ok := props[propPickerStyle].(style.Style); ok {
		v.pickerStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetComponentID(id string) *VNode {
	v.componentID = id
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

func (v *VNode) SetPlaceholder(placeholder string) *VNode {
	v.placeholder = placeholder
	return v
}

func (v *VNode) SetValue(value string) *VNode {
	v.value = value
	v.valueControlled = true
	return v
}

func (v *VNode) SetDefaultValue(value string) *VNode {
	v.defaultValue = value
	return v
}

func (v *VNode) SetDisabled(disabled bool) *VNode {
	v.disabled = disabled
	return v
}

func (v *VNode) SetPopupPortalRoot(root string) *VNode {
	v.portalRoot = root
	return v
}

func (v *VNode) SetChangeIntent(changeIntent intent.Intent) *VNode {
	v.changeIntent = changeIntent
	return v
}

func (v *VNode) SetFormID(formID string) *VNode {
	v.formID = formID
	return v
}
