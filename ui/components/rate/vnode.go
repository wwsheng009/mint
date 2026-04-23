package rate

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey          = "key"
	propLabel        = "label"
	propValue        = "value"
	propCount        = "count"
	propAllowClear   = "allowClear"
	propCharacter    = "character"
	propDisabled     = "disabled"
	propShowValue    = "showValue"
	propStyle        = "style"
	propChangeIntent = "changeIntent"
	propFormID       = "formID"
)

const (
	defaultCount     = 5
	defaultCharacter = "★"
)

// VNode is the declarative description of a Rate component.
type VNode struct {
	*rtui.ElementVNode

	key   string
	label string
	style style.Style

	changeIntent intent.Intent
	formID       string

	value      int
	count      int
	allowClear bool
	character  string
	disabled   bool
	showValue  bool

	rtui.BoxModelMixin
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// New creates a new Rate VNode with defaults.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("rate"),
		count:        defaultCount,
		allowClear:   true,
		character:    defaultCharacter,
	}
}

// Key returns the component key.
func (r *VNode) Key() string { return r.key }

// SetKey sets the component key.
func (r *VNode) SetKey(key string) rtui.VNode { r.key = key; return r }

// Tag returns the tag name.
func (r *VNode) Tag() string { return "rate" }

// Style returns the visual style.
func (r *VNode) Style() style.Style { return r.style }

// SetStyle sets the visual style.
func (r *VNode) SetStyle(st style.Style) rtui.VNode { r.style = st; return r }

// Children returns nil (no children).
func (r *VNode) Children() []rtui.VNode { return nil }

// SetChildren is a no-op for rate.
func (r *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return r }

// GetLayer returns the rendering layer.
func (r *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

// SetLayer sets the rendering layer.
func (r *VNode) SetLayer(l rtui.Layer) rtui.VNode { return r }

// Props returns the node properties.
func (r *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:          r.key,
		propLabel:        r.label,
		propValue:        r.value,
		propCount:        r.count,
		propAllowClear:   r.allowClear,
		propCharacter:    r.character,
		propDisabled:     r.disabled,
		propShowValue:    r.showValue,
		propStyle:        r.style,
		propChangeIntent: r.changeIntent,
		propFormID:       r.formID,
	}
}

// SetProps sets the node properties.
func (r *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p[propKey].(string); ok {
		r.key = v
	}
	if v, ok := p[propLabel].(string); ok {
		r.label = v
	}
	if v, ok := p[propValue].(int); ok {
		r.value = v
	}
	if v, ok := p[propCount].(int); ok {
		r.count = v
	}
	if v, ok := p[propAllowClear].(bool); ok {
		r.allowClear = v
	}
	if v, ok := p[propCharacter].(string); ok {
		r.character = v
	}
	if v, ok := p[propDisabled].(bool); ok {
		r.disabled = v
	}
	if v, ok := p[propShowValue].(bool); ok {
		r.showValue = v
	}
	if v, ok := p[propStyle].(style.Style); ok {
		r.style = v
	}
	if v, ok := p[propChangeIntent].(intent.Intent); ok {
		r.changeIntent = v
	}
	if v, ok := p[propFormID].(string); ok {
		r.formID = v
	}
	return r
}

// CreateInstance creates a new Rate instance from this VNode.
func (r *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(r.Props())
}

// SetLabel sets the descriptive label.
func (r *VNode) SetLabel(label string) *VNode { r.label = label; return r }

// SetValue sets the current value.
func (r *VNode) SetValue(value int) *VNode { r.value = value; return r }

// SetCount sets the total number of stars.
func (r *VNode) SetCount(count int) *VNode { r.count = count; return r }

// SetAllowClear controls whether clicking/toggling can clear rating.
func (r *VNode) SetAllowClear(allow bool) *VNode { r.allowClear = allow; return r }

// SetCharacter sets the filled star character.
func (r *VNode) SetCharacter(ch string) *VNode { r.character = ch; return r }

// SetDisabled sets the disabled state.
func (r *VNode) SetDisabled(disabled bool) *VNode { r.disabled = disabled; return r }

// SetShowValue controls whether current rating text is shown.
func (r *VNode) SetShowValue(show bool) *VNode { r.showValue = show; return r }

// SetStyleProps sets the visual style.
func (r *VNode) SetStyleProps(st style.Style) *VNode { r.style = st; return r }

// SetIntent sets the intent to emit on value change.
func (r *VNode) SetIntent(i intent.Intent) *VNode { r.changeIntent = i; return r }

// SetFormID sets the form ID for Form integration.
func (r *VNode) SetFormID(formID string) *VNode { r.formID = formID; return r }

// GetBoxModel returns the box model for the Rate VNode.
func (r *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   r.BoxModelMixin.Padding()[3],
			Right:  r.BoxModelMixin.Padding()[1],
			Top:    r.BoxModelMixin.Padding()[0],
			Bottom: r.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   r.BoxModelMixin.Margin()[3],
			Right:  r.BoxModelMixin.Margin()[1],
			Top:    r.BoxModelMixin.Margin()[0],
			Bottom: r.BoxModelMixin.Margin()[2],
		},
		Border: layout.Border{Style: layout.BorderNone},
	}
}
