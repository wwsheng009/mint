package switchcomp

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propChecked        = "checked"
	propCheckedLabel   = "checkedLabel"
	propDisabled       = "disabled"
	propFormID         = "formID"
	propKey            = "key"
	propLabel          = "label"
	propStyle          = "style"
	propToggleIntent   = "toggleIntent"
	propUncheckedLabel = "uncheckedLabel"
)

const (
	defaultCheckedLabel   = "ON"
	defaultUncheckedLabel = "OFF"
)

// VNode is the declarative description of a Switch component.
type VNode struct {
	*rtui.ElementVNode

	key string

	label string
	style style.Style

	toggleIntent intent.Intent
	formID       string

	checked        bool
	disabled       bool
	checkedLabel   string
	uncheckedLabel string

	rtui.BoxModelMixin
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// New creates a new Switch VNode.
func New(label string) *VNode {
	return &VNode{
		ElementVNode:   rtui.NewElement("switch"),
		label:          label,
		checkedLabel:   defaultCheckedLabel,
		uncheckedLabel: defaultUncheckedLabel,
	}
}

// Key returns the component key.
func (s *VNode) Key() string {
	return s.key
}

// SetKey sets the component key.
func (s *VNode) SetKey(key string) rtui.VNode {
	s.key = key
	return s
}

// Tag returns the tag name.
func (s *VNode) Tag() string {
	return "switch"
}

// Style returns the visual style.
func (s *VNode) Style() style.Style {
	return s.style
}

// SetStyle sets the visual style.
func (s *VNode) SetStyle(st style.Style) rtui.VNode {
	s.style = st
	return s
}

// Children returns child nodes.
func (s *VNode) Children() []rtui.VNode {
	return nil
}

// SetChildren is a no-op for switch.
func (s *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	return s
}

// GetLayer returns the rendering layer.
func (s *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer.
func (s *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return s
}

// Props returns the node properties.
func (s *VNode) Props() rtui.Props {
	return rtui.Props{
		propChecked:        s.checked,
		propCheckedLabel:   s.checkedLabel,
		propDisabled:       s.disabled,
		propFormID:         s.formID,
		propKey:            s.key,
		propLabel:          s.label,
		propStyle:          s.style,
		propToggleIntent:   s.toggleIntent,
		propUncheckedLabel: s.uncheckedLabel,
	}
}

// SetProps sets the node properties.
func (s *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p[propKey].(string); ok {
		s.key = v
	}
	if v, ok := p[propLabel].(string); ok {
		s.label = v
	}
	if v, ok := p[propStyle].(style.Style); ok {
		s.style = v
	}
	if v, ok := p[propToggleIntent].(intent.Intent); ok {
		s.toggleIntent = v
	}
	if v, ok := p[propFormID].(string); ok {
		s.formID = v
	}
	if v, ok := p[propDisabled].(bool); ok {
		s.disabled = v
	}
	if v, ok := p[propChecked].(bool); ok {
		s.checked = v
	}
	if v, ok := p[propCheckedLabel].(string); ok {
		s.checkedLabel = v
	}
	if v, ok := p[propUncheckedLabel].(string); ok {
		s.uncheckedLabel = v
	}
	return s
}

// CreateInstance creates a new Switch instance from this VNode.
func (s *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(s.Props())
}

// SetLabel sets the switch label.
func (s *VNode) SetLabel(label string) *VNode {
	s.label = label
	return s
}

// SetDisabled sets the disabled state.
func (s *VNode) SetDisabled(disabled bool) *VNode {
	s.disabled = disabled
	return s
}

// SetChecked sets the checked state.
func (s *VNode) SetChecked(checked bool) *VNode {
	s.checked = checked
	return s
}

// SetIntent sets the toggle intent.
func (s *VNode) SetIntent(toggleIntent intent.Intent) *VNode {
	s.toggleIntent = toggleIntent
	return s
}

// SetFormID sets the form ID for Form integration.
func (s *VNode) SetFormID(formID string) *VNode {
	s.formID = formID
	return s
}

// SetCheckedLabel sets the label displayed for the checked state.
func (s *VNode) SetCheckedLabel(label string) *VNode {
	s.checkedLabel = label
	return s
}

// SetUncheckedLabel sets the label displayed for the unchecked state.
func (s *VNode) SetUncheckedLabel(label string) *VNode {
	s.uncheckedLabel = label
	return s
}

// SetLabels sets both checked and unchecked labels.
func (s *VNode) SetLabels(checkedLabel, uncheckedLabel string) *VNode {
	s.checkedLabel = checkedLabel
	s.uncheckedLabel = uncheckedLabel
	return s
}

// SetStyleProps sets the visual style.
func (s *VNode) SetStyleProps(st style.Style) *VNode {
	s.style = st
	return s
}

// OnToggle sets the intent to emit when toggled.
func (s *VNode) OnToggle(toggleIntent intent.Intent) *VNode {
	return s.SetIntent(toggleIntent)
}

// Label returns the descriptive label.
func (s *VNode) Label() string {
	return s.label
}

// Disabled returns the disabled state.
func (s *VNode) Disabled() bool {
	return s.disabled
}

// Checked returns the checked state.
func (s *VNode) Checked() bool {
	return s.checked
}

// ToggleIntent returns the toggle intent.
func (s *VNode) ToggleIntent() intent.Intent {
	return s.toggleIntent
}

// CheckedLabel returns the checked-state track label.
func (s *VNode) CheckedLabel() string {
	return s.checkedLabel
}

// UncheckedLabel returns the unchecked-state track label.
func (s *VNode) UncheckedLabel() string {
	return s.uncheckedLabel
}

// GetBoxModel returns the box model for the Switch VNode.
func (s *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   s.BoxModelMixin.Padding()[3],
			Right:  s.BoxModelMixin.Padding()[1],
			Top:    s.BoxModelMixin.Padding()[0],
			Bottom: s.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   s.BoxModelMixin.Margin()[3],
			Right:  s.BoxModelMixin.Margin()[1],
			Top:    s.BoxModelMixin.Margin()[0],
			Bottom: s.BoxModelMixin.Margin()[2],
		},
		Border: layout.Border{Style: layout.BorderNone},
	}
}
