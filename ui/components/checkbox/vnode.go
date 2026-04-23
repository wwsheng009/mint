package checkbox

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/optiongroup"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propChecked       = "checked"
	propDisabled      = "disabled"
	propFormID        = "formID"
	propIndeterminate = "indeterminate"
	propKey           = "key"
	propLabel         = "label"
	propStyle         = "style"
	propToggleIntent  = "toggleIntent"
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the checkbox description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	label string
	style style.Style

	// === Intent Props (no closures!) ===
	toggleIntent intent.Intent // Structured intent instead of func(bool)
	formID       string        // Form ID for Form integration (Phase 6)

	// === State Props (declarative, actual state managed by Instance) ===
	disabled      bool
	checked       bool
	indeterminate bool

	// === Box Model (via interface) ===
	rtui.BoxModelMixin
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new Checkbox VNode.
func New(label string) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("checkbox"),
		label:        label,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (c *VNode) Key() string {
	return c.key
}

// SetKey sets the component key - returns VNode for chaining.
func (c *VNode) SetKey(key string) rtui.VNode {
	c.key = key
	return c
}

// Tag returns the tag name.
func (c *VNode) Tag() string {
	return "checkbox"
}

// Style returns the visual style.
func (c *VNode) Style() style.Style {
	return c.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (c *VNode) SetStyle(s style.Style) rtui.VNode {
	c.style = s
	return c
}

// Children returns child nodes (checkbox has no children).
func (c *VNode) Children() []rtui.VNode {
	return nil
}

// SetChildren is a no-op for checkbox - returns VNode for chaining.
func (c *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	return c
}

// GetLayer returns the rendering layer.
func (c *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (c *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return c
}

// Props returns the node properties.
func (c *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:           c.key,
		propLabel:         c.label,
		propStyle:         c.style,
		propToggleIntent:  c.toggleIntent,
		propFormID:        c.formID,
		propDisabled:      c.disabled,
		propChecked:       c.checked,
		propIndeterminate: c.indeterminate,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (c *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p[propKey].(string); ok {
		c.key = v
	}
	if v, ok := p[propLabel].(string); ok {
		c.label = v
	}
	if v, ok := p[propStyle].(style.Style); ok {
		c.style = v
	}
	if v, ok := p[propToggleIntent].(intent.Intent); ok {
		c.toggleIntent = v
	}
	if v, ok := p[propFormID].(string); ok {
		c.formID = v
	}
	if v, ok := p[propDisabled].(bool); ok {
		c.disabled = v
	}
	if v, ok := p[propChecked].(bool); ok {
		c.checked = v
	}
	if v, ok := p[propIndeterminate].(bool); ok {
		c.indeterminate = v
	}
	return c
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new CheckboxInstance from this VNode description.
func (c *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		propKey:           c.key,
		propLabel:         c.label,
		propStyle:         c.style,
		propToggleIntent:  c.toggleIntent,
		propFormID:        c.formID,
		propDisabled:      c.disabled,
		propChecked:       c.checked,
		propIndeterminate: c.indeterminate,
	}
	return NewInstance(props)
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetLabel sets the checkbox label.
func (c *VNode) SetLabel(label string) *VNode {
	c.label = label
	return c
}

// SetDisabled sets the disabled state.
func (c *VNode) SetDisabled(disabled bool) *VNode {
	c.disabled = disabled
	return c
}

// SetChecked sets the checked state (declarative).
func (c *VNode) SetChecked(checked bool) *VNode {
	c.checked = checked
	return c
}

// SetIndeterminate sets the indeterminate state (declarative).
func (c *VNode) SetIndeterminate(indeterminate bool) *VNode {
	c.indeterminate = indeterminate
	return c
}

// SetIntent sets the toggle intent (replaces OnChange closure).
func (c *VNode) SetIntent(toggleIntent intent.Intent) *VNode {
	c.toggleIntent = toggleIntent
	return c
}

// SetFormID sets the form ID for Form integration.
// When set, the component will emit FormFieldChangeIntent/FormFieldBlurIntent.
func (c *VNode) SetFormID(formID string) *VNode {
	c.formID = formID
	return c
}

// SetStyleProps sets the visual style.
func (c *VNode) SetStyleProps(s style.Style) *VNode {
	c.style = s
	return c
}

// =============================================================================
// Intent Methods (replacing closures)
// =============================================================================

// OnToggle sets the intent to emit when toggled.
// This replaces the old OnChange(func(bool)) closure pattern.
func (c *VNode) OnToggle(toggleIntent intent.Intent) *VNode {
	return c.SetIntent(toggleIntent)
}

// =============================================================================
// Props Accessors (for Instance creation)
// =============================================================================

// Label returns the checkbox label.
func (c *VNode) Label() string {
	return c.label
}

// Disabled returns the disabled state.
func (c *VNode) Disabled() bool {
	return c.disabled
}

// Checked returns the checked state.
func (c *VNode) Checked() bool {
	return c.checked
}

// Indeterminate returns the indeterminate state.
func (c *VNode) Indeterminate() bool {
	return c.indeterminate
}

// ToggleIntent returns the toggle intent.
func (c *VNode) ToggleIntent() intent.Intent {
	return c.toggleIntent
}

// Option re-exports the option type used by CheckboxGroup.
type Option = optiongroup.Option

// Orientation re-exports the layout orientation used by CheckboxGroup.
type Orientation = optiongroup.Orientation

const (
	OrientationVertical   = optiongroup.OrientationVertical
	OrientationHorizontal = optiongroup.OrientationHorizontal
)

// GroupVNode is the CheckboxGroup description.
type GroupVNode struct {
	*optiongroup.VNode
}

var (
	_ rtui.VNode           = (*GroupVNode)(nil)
	_ rtui.InstanceFactory = (*GroupVNode)(nil)
)

// NewGroup creates a new CheckboxGroup VNode.
func NewGroup(options []Option) *GroupVNode {
	return &GroupVNode{
		VNode: optiongroup.New(options).Multiple(),
	}
}

// Tag returns the tag name.
func (g *GroupVNode) Tag() string {
	return "checkboxgroup"
}

// CreateInstance creates a new CheckboxGroup instance.
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

// SetSelecteds sets the selected values.
func (g *GroupVNode) SetSelecteds(selecteds []string) *GroupVNode {
	g.VNode.SetSelecteds(selecteds)
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

// OnSelect sets the intent to emit when values change.
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

// Selecteds returns the selected values.
func (g *GroupVNode) Selecteds() []string {
	return g.VNode.Selecteds()
}

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the Checkbox VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: Checkbox uses BoxModelMixin for padding/margin, and has no border.
func (c *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   c.BoxModelMixin.Padding()[3],
			Right:  c.BoxModelMixin.Padding()[1],
			Top:    c.BoxModelMixin.Padding()[0],
			Bottom: c.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   c.BoxModelMixin.Margin()[3],
			Right:  c.BoxModelMixin.Margin()[1],
			Top:    c.BoxModelMixin.Margin()[0],
			Bottom: c.BoxModelMixin.Margin()[2],
		},
		// Checkbox typically doesn't have a border
		Border: layout.Border{Style: layout.BorderNone},
	}
}
