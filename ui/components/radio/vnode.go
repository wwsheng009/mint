package radio

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/optiongroup"
)

const (
	propChecked      = "checked"
	propDisabled     = "disabled"
	propFormID       = "formID"
	propKey          = "key"
	propLabel        = "label"
	propSelectIntent = "selectIntent"
	propStyle        = "style"
)

// VNode is the radio description.
// It contains ONLY declarative information.
type VNode struct {
	*rtui.ElementVNode

	key string

	label string
	style style.Style

	selectIntent intent.Intent
	formID       string

	disabled bool
	checked  bool

	rtui.BoxModelMixin
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// New creates a new Radio VNode.
func New(label string) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("radio"),
		label:        label,
	}
}

// Key returns the component key.
func (r *VNode) Key() string {
	return r.key
}

// SetKey sets the component key.
func (r *VNode) SetKey(key string) rtui.VNode {
	r.key = key
	return r
}

// Tag returns the tag name.
func (r *VNode) Tag() string {
	return "radio"
}

// Style returns the visual style.
func (r *VNode) Style() style.Style {
	return r.style
}

// SetStyle sets the visual style.
func (r *VNode) SetStyle(s style.Style) rtui.VNode {
	r.style = s
	return r
}

// Children returns child nodes.
func (r *VNode) Children() []rtui.VNode {
	return nil
}

// SetChildren is a no-op for radio.
func (r *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	return r
}

// GetLayer returns the rendering layer.
func (r *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer.
func (r *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return r
}

// Props returns the node properties.
func (r *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:          r.key,
		propLabel:        r.label,
		propStyle:        r.style,
		propSelectIntent: r.selectIntent,
		propFormID:       r.formID,
		propDisabled:     r.disabled,
		propChecked:      r.checked,
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
	if v, ok := p[propStyle].(style.Style); ok {
		r.style = v
	}
	if v, ok := p[propSelectIntent].(intent.Intent); ok {
		r.selectIntent = v
	}
	if v, ok := p[propFormID].(string); ok {
		r.formID = v
	}
	if v, ok := p[propDisabled].(bool); ok {
		r.disabled = v
	}
	if v, ok := p[propChecked].(bool); ok {
		r.checked = v
	}
	return r
}

// CreateInstance creates a new Radio instance.
func (r *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(r.Props())
}

// SetLabel sets the radio label.
func (r *VNode) SetLabel(label string) *VNode {
	r.label = label
	return r
}

// SetDisabled sets the disabled state.
func (r *VNode) SetDisabled(disabled bool) *VNode {
	r.disabled = disabled
	return r
}

// SetChecked sets the checked state.
func (r *VNode) SetChecked(checked bool) *VNode {
	r.checked = checked
	return r
}

// SetIntent sets the select intent.
func (r *VNode) SetIntent(selectIntent intent.Intent) *VNode {
	r.selectIntent = selectIntent
	return r
}

// SetFormID sets the form ID for Form integration.
func (r *VNode) SetFormID(formID string) *VNode {
	r.formID = formID
	return r
}

// SetStyleProps sets the visual style.
func (r *VNode) SetStyleProps(s style.Style) *VNode {
	r.style = s
	return r
}

// OnSelect sets the intent to emit when selected.
func (r *VNode) OnSelect(selectIntent intent.Intent) *VNode {
	return r.SetIntent(selectIntent)
}

// Label returns the radio label.
func (r *VNode) Label() string {
	return r.label
}

// Disabled returns the disabled state.
func (r *VNode) Disabled() bool {
	return r.disabled
}

// Checked returns the checked state.
func (r *VNode) Checked() bool {
	return r.checked
}

// SelectIntent returns the select intent.
func (r *VNode) SelectIntent() intent.Intent {
	return r.selectIntent
}

// GetBoxModel returns the box model for the Radio VNode.
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

// Option re-exports the option type used by RadioGroup.
type Option = optiongroup.Option

// Orientation re-exports the layout orientation used by RadioGroup.
type Orientation = optiongroup.Orientation

const (
	OrientationVertical   = optiongroup.OrientationVertical
	OrientationHorizontal = optiongroup.OrientationHorizontal
)

// GroupVNode is the RadioGroup description.
type GroupVNode struct {
	*optiongroup.VNode
}

var (
	_ rtui.VNode           = (*GroupVNode)(nil)
	_ rtui.InstanceFactory = (*GroupVNode)(nil)
)

// NewGroup creates a new RadioGroup VNode.
func NewGroup(options []Option) *GroupVNode {
	return &GroupVNode{
		VNode: optiongroup.New(options).Single(),
	}
}

// Tag returns the tag name.
func (g *GroupVNode) Tag() string {
	return "radiogroup"
}

// CreateInstance creates a new RadioGroup instance.
func (g *GroupVNode) CreateInstance() rtui.ComponentInstance {
	inst := g.VNode.CreateInstance()
	groupInst, _ := inst.(*optiongroup.Instance)
	return &GroupInstance{Instance: groupInst}
}

// SetLabel sets the group label.
func (g *GroupVNode) SetLabel(label string) *GroupVNode {
	g.VNode.SetLabel(label)
	return g
}

// SetDisabled sets the disabled state.
func (g *GroupVNode) SetDisabled(disabled bool) *GroupVNode {
	g.VNode.SetDisabled(disabled)
	return g
}

// SetSelected sets the selected value.
func (g *GroupVNode) SetSelected(selected string) *GroupVNode {
	g.VNode.SetSelected(selected)
	return g
}

// SetIntent sets the select intent.
func (g *GroupVNode) SetIntent(selectIntent intent.Intent) *GroupVNode {
	g.VNode.SetIntent(selectIntent)
	return g
}

// SetStyleProps sets the visual style.
func (g *GroupVNode) SetStyleProps(s style.Style) *GroupVNode {
	g.VNode.SetStyle(s)
	return g
}

// SetOrientation sets the layout orientation.
func (g *GroupVNode) SetOrientation(orientation Orientation) *GroupVNode {
	g.VNode.SetOrientation(orientation)
	return g
}

// SetSpacing sets the gap between options.
func (g *GroupVNode) SetSpacing(spacing int) *GroupVNode {
	g.VNode.SetSpacing(spacing)
	return g
}

// SetOptions replaces the option list.
func (g *GroupVNode) SetOptions(options []Option) *GroupVNode {
	g.VNode.SetProps(rtui.Props{"options": options})
	return g
}

// OnSelect sets the intent to emit when a value is selected.
func (g *GroupVNode) OnSelect(selectIntent intent.Intent) *GroupVNode {
	return g.SetIntent(selectIntent)
}

// Vertical sets orientation to vertical.
func (g *GroupVNode) Vertical() *GroupVNode {
	return g.SetOrientation(OrientationVertical)
}

// Horizontal sets orientation to horizontal.
func (g *GroupVNode) Horizontal() *GroupVNode {
	return g.SetOrientation(OrientationHorizontal)
}

// Options returns the option list.
func (g *GroupVNode) Options() []Option {
	if options, ok := g.VNode.Props()["options"].([]optiongroup.Option); ok {
		return options
	}
	return nil
}
