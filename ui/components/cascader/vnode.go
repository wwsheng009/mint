package cascader

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propChangeIntent    = "changeIntent"
	propChangeOnSelect  = "changeOnSelect"
	propComponentID     = "componentID"
	propDefaultValue    = "defaultValue"
	propDisabled        = "disabled"
	propFormID          = "formID"
	propKey             = "key"
	propOptions         = "options"
	propPlaceholder     = "placeholder"
	propSeparator       = "separator"
	propStyle           = "style"
	propValue           = "value"
	propValueControlled = "valueControlled"
	propWidth           = "width"
)

const (
	defaultPlaceholder    = "Please select"
	defaultSeparator      = " / "
	defaultFieldSeparator = "/"
	defaultTriggerWidth   = 18
)

// Option represents a single cascader option node.
type Option struct {
	Value    string
	Label    string
	Disabled bool
	Children []Option
}

// Node creates a cascader node with optional child options.
func Node(value, label string, children ...Option) Option {
	return Option{
		Value:    value,
		Label:    label,
		Children: append([]Option(nil), children...),
	}
}

// Leaf creates a leaf cascader option.
func Leaf(value, label string) Option {
	return Node(value, label)
}

// VNode is the declarative description of a Cascader component.
type VNode struct {
	*rtui.ElementVNode

	key             string
	componentID     string
	cascaderStyle   style.Style
	width           int
	placeholder     string
	separator       string
	options         []Option
	value           []string
	defaultValue    []string
	valueControlled bool
	disabled        bool
	changeOnSelect  bool
	changeIntent    intent.Intent
	formID          string
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Cascader VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("cascader"),
		width:        defaultTriggerWidth,
		placeholder:  defaultPlaceholder,
		separator:    defaultSeparator,
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to component ID/key.
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

func (v *VNode) Tag() string { return "cascader" }

func (v *VNode) Style() style.Style { return v.cascaderStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.cascaderStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propChangeIntent:    v.changeIntent,
		propChangeOnSelect:  v.changeOnSelect,
		propComponentID:     v.componentID,
		propDefaultValue:    append([]string(nil), v.defaultValue...),
		propDisabled:        v.disabled,
		propFormID:          v.formID,
		propKey:             v.key,
		propOptions:         append([]Option(nil), v.options...),
		propPlaceholder:     v.placeholder,
		propSeparator:       v.separator,
		propStyle:           v.cascaderStyle,
		propValue:           append([]string(nil), v.value...),
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
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if placeholder, ok := props[propPlaceholder].(string); ok {
		v.placeholder = placeholder
	}
	if separator, ok := props[propSeparator].(string); ok {
		v.separator = separator
	}
	if options, ok := props[propOptions].([]Option); ok {
		v.options = append([]Option(nil), options...)
	}
	if value, ok := props[propValue].([]string); ok {
		v.value = append([]string(nil), value...)
	}
	if defaultValue, ok := props[propDefaultValue].([]string); ok {
		v.defaultValue = append([]string(nil), defaultValue...)
	}
	if valueControlled, ok := props[propValueControlled].(bool); ok {
		v.valueControlled = valueControlled
	}
	if disabled, ok := props[propDisabled].(bool); ok {
		v.disabled = disabled
	}
	if changeOnSelect, ok := props[propChangeOnSelect].(bool); ok {
		v.changeOnSelect = changeOnSelect
	}
	if changeIntent, ok := props[propChangeIntent].(intent.Intent); ok {
		v.changeIntent = changeIntent
	}
	if formID, ok := props[propFormID].(string); ok {
		v.formID = formID
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.cascaderStyle = s
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

func (v *VNode) SetOptions(options []Option) *VNode {
	v.options = append([]Option(nil), options...)
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

func (v *VNode) SetValue(value []string) *VNode {
	v.value = append([]string(nil), value...)
	v.valueControlled = true
	return v
}

func (v *VNode) SetDefaultValue(value []string) *VNode {
	v.defaultValue = append([]string(nil), value...)
	return v
}

func (v *VNode) SetDisabled(disabled bool) *VNode {
	v.disabled = disabled
	return v
}

func (v *VNode) SetChangeOnSelect(enabled bool) *VNode {
	v.changeOnSelect = enabled
	return v
}

func (v *VNode) SetSeparator(separator string) *VNode {
	v.separator = separator
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
