// Package optiongroup provides a Store + Reducer compatible option group component.
// OptionGroup supports single-select (radio) and multi-select (checkbox) modes.
package optiongroup

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types
// =============================================================================

// SelectMode defines the selection behavior.
type SelectMode int

const (
	ModeSingle   SelectMode = iota // Radio button behavior (select one)
	ModeMultiple                   // Checkbox behavior (select multiple)
)

// Orientation defines the layout direction.
type Orientation int

const (
	OrientationVertical   Orientation = iota // Options stacked vertically
	OrientationHorizontal                     // Options arranged horizontally
)

// Option represents a single option in the group.
type Option struct {
	Value string
	Label string
}

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the optiongroup component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	label  string
	style  style.Style

	// === Intent Props (no closures!) ===
	selectIntent intent.Intent // Structured intent instead of func(string, bool)

	// === State Props (declarative, actual state managed by Instance) ===
	disabled  bool
	mode      SelectMode
	options   []Option
	selected  string   // For ModeSingle: selected value
	selecteds []string // For ModeMultiple: selected values (comma-separated in FieldMap)

	// === Layout Props ===
	orientation Orientation
	spacing     int // gap between options

	// === Parent Callback (set by Instance for children) ===
	optionSelectFunc SelectOptionFunc // Passed to child OptionVNodes

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

// New creates a new OptionGroup VNode.
func New(options []Option) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("optiongroup"),
		options:      options,
		mode:         ModeSingle,
		orientation:  OrientationVertical,
		spacing:      1,
		selected:     "",
		selecteds:    []string{},
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (o *VNode) Key() string {
	return o.key
}

// SetKey sets the component key - returns VNode for chaining.
func (o *VNode) SetKey(key string) rtui.VNode {
	o.key = key
	return o
}

// Tag returns the tag name.
func (o *VNode) Tag() string {
	return "optiongroup"
}

// Style returns the visual style.
func (o *VNode) Style() style.Style {
	return o.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (o *VNode) SetStyle(s style.Style) rtui.VNode {
	o.style = s
	return o
}

// Children returns child nodes - the options as individual OptionVNodes.
func (o *VNode) Children() []rtui.VNode {
	if o.options == nil {
		return nil
	}
	children := make([]rtui.VNode, len(o.options))
	for i, opt := range o.options {
		child := NewOptionVNodeDeferred(opt.Value, opt.Label, i, o.mode)
		// Apply parent disabled state to child
		if o.disabled {
			child.SetDisabled(true)
		}
		// Apply the selectFunc if it's been set by the parent Instance
		if o.optionSelectFunc != nil {
			child.SetSelectFunc(o.optionSelectFunc)
		}
		// Pass parentCallback as a prop for later use
		child.SetProps(rtui.Props{"parentCallback": o.optionSelectFunc})
		children[i] = child
	}
	return children
}

// SetChildren is not supported for OptionGroup - options are managed internally.
// Returns VNode for chaining (no effect).
func (o *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// OptionGroup generates its own children from options list
	// Manual child setting is not supported
	return o
}

// SetOptionCallbacks sets the selection callback for all child OptionVNodes.
// This should be called by the parent OptionGroupInstance after it's created.
func (o *VNode) SetOptionCallbacks(selectFunc SelectOptionFunc) {
	o.optionSelectFunc = selectFunc

	// Update existing child VNodes (if they've been created)
	if o.options != nil {
		for i := range o.options {
			// We need to update the VNode, but we don't have direct access to child VNodes here
			// The Children() method creates new VNodes each time, so we rely on future renders
		}
	}
}

// GetChildOptionVNodes returns the child OptionVNodes for direct update.
// This is used by the parent Instance to set callbacks on existing child nodes.
func (o *VNode) GetChildOptionVNodes() []*OptionVNode {
	if o.options == nil {
		return nil
	}
	children := make([]*OptionVNode, len(o.options))
	for i, opt := range o.options {
		child := NewOptionVNodeDeferred(opt.Value, opt.Label, i, o.mode)
		if o.optionSelectFunc != nil {
			child.SetSelectFunc(o.optionSelectFunc)
		}
		if o.disabled {
			child.SetDisabled(true)
		}
		children[i] = child
	}
	return children
}

// GetLayer returns the rendering layer.
func (o *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (o *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return o
}

// Props returns the node properties.
func (o *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":           o.key,
		"label":         o.label,
		"style":         o.style,
		"selectIntent":  o.selectIntent,
		"disabled":      o.disabled,
		"mode":          o.mode,
		"options":       o.options,
		"selected":      o.selected,
		"selecteds":     o.selecteds,
		"orientation":   o.orientation,
		"spacing":       o.spacing,
		"parentCallback": o.optionSelectFunc, // Passed to children for selection
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (o *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		o.key = v
	}
	if v, ok := p["label"].(string); ok {
		o.label = v
	}
	if v, ok := p["style"].(style.Style); ok {
		o.style = v
	}
	if v, ok := p["selectIntent"].(intent.Intent); ok {
		o.selectIntent = v
	}
	if v, ok := p["disabled"].(bool); ok {
		o.disabled = v
	}
	if v, ok := p["mode"].(SelectMode); ok {
		o.mode = v
	}
	if v, ok := p["options"].([]Option); ok {
		o.options = v
	}
	if v, ok := p["selected"].(string); ok {
		o.selected = v
	}
	if v, ok := p["selecteds"].([]string); ok {
		o.selecteds = v
	}
	if v, ok := p["orientation"].(Orientation); ok {
		o.orientation = v
	}
	if v, ok := p["spacing"].(int); ok {
		o.spacing = v
	}
	return o
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new OptionGroupInstance from this VNode description.
func (o *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		"key":          o.key,
		"label":        o.label,
		"style":        o.style,
		"selectIntent": o.selectIntent,
		"disabled":     o.disabled,
		"mode":         o.mode,
		"options":      o.options,
		"selected":     o.selected,
		"selecteds":    o.selecteds,
		"orientation":  o.orientation,
		"spacing":      o.spacing,
	}
	inst := NewInstance(props)

	// Set the callback so child OptionVNodes can select options
	o.optionSelectFunc = inst.SelectOption

	return inst
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetLabel sets the optiongroup label.
func (o *VNode) SetLabel(label string) *VNode {
	o.label = label
	return o
}

// SetDisabled sets the disabled state.
func (o *VNode) SetDisabled(disabled bool) *VNode {
	o.disabled = disabled
	return o
}

// SetMode sets the selection mode (single or multiple).
func (o *VNode) SetMode(mode SelectMode) *VNode {
	o.mode = mode
	return o
}

// SetSelected sets the selected value (for ModeSingle).
func (o *VNode) SetSelected(selected string) *VNode {
	o.selected = selected
	return o
}

// SetSelecteds sets the selected values (for ModeMultiple).
func (o *VNode) SetSelecteds(selecteds []string) *VNode {
	o.selecteds = selecteds
	return o
}

// SetIntent sets the select intent (replaces closure).
func (o *VNode) SetIntent(selectIntent intent.Intent) *VNode {
	o.selectIntent = selectIntent
	return o
}

// SetStyleProps sets the visual style.
func (o *VNode) SetStyleProps(s style.Style) *VNode {
	o.style = s
	return o
}

// SetOrientation sets the layout orientation.
func (o *VNode) SetOrientation(orientation Orientation) *VNode {
	o.orientation = orientation
	return o
}

// SetSpacing sets the gap between options.
func (o *VNode) SetSpacing(spacing int) *VNode {
	o.spacing = spacing
	return o
}

// =============================================================================
// Convenience Builder Methods
// =============================================================================

// Single sets mode to single-select (radio behavior).
func (o *VNode) Single() *VNode {
	return o.SetMode(ModeSingle)
}

// Multiple sets mode to multi-select (checkbox behavior).
func (o *VNode) Multiple() *VNode {
	return o.SetMode(ModeMultiple)
}

// Vertical sets orientation to vertical.
func (o *VNode) Vertical() *VNode {
	return o.SetOrientation(OrientationVertical)
}

// Horizontal sets orientation to horizontal.
func (o *VNode) Horizontal() *VNode {
	return o.SetOrientation(OrientationHorizontal)
}

// =============================================================================
// Intent Methods (replacing closures)
// =============================================================================

// OnSelect sets the intent to emit when an option is selected.
func (o *VNode) OnSelect(selectIntent intent.Intent) *VNode {
	return o.SetIntent(selectIntent)
}

// =============================================================================
// Props Accessors (for Instance creation)
// =============================================================================

// Label returns the component label.
func (o *VNode) Label() string {
	return o.label
}

// Disabled returns the disabled state.
func (o *VNode) Disabled() bool {
	return o.disabled
}

// Mode returns the selection mode.
func (o *VNode) Mode() SelectMode {
	return o.mode
}

// Options returns the option list.
func (o *VNode) Options() []Option {
	return o.options
}

// Selected returns the selected value (for ModeSingle).
func (o *VNode) Selected() string {
	return o.selected
}

// Selecteds returns the selected values (for ModeMultiple).
func (o *VNode) Selecteds() []string {
	return o.selecteds
}

// SelectIntent returns the select intent.
func (o *VNode) SelectIntent() intent.Intent {
	return o.selectIntent
}

// Orientation returns the layout orientation.
func (o *VNode) Orientation() Orientation {
	return o.orientation
}

// Spacing returns the gap between options.
func (o *VNode) Spacing() int {
	return o.spacing
}

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the OptionGroup VNode.
func (o *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   o.BoxModelMixin.Padding()[3],
			Right:  o.BoxModelMixin.Padding()[1],
			Top:    o.BoxModelMixin.Padding()[0],
			Bottom: o.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   o.BoxModelMixin.Margin()[3],
			Right:  o.BoxModelMixin.Margin()[1],
			Top:    o.BoxModelMixin.Margin()[0],
			Bottom: o.BoxModelMixin.Margin()[2],
		},
		Border: layout.Border{Style: layout.BorderNone},
	}
}
