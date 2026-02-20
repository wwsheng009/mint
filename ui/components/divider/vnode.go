// Package divider provides a Fiber-first Divider component.
// Divider displays a horizontal or vertical separator line.
package divider

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types
// =============================================================================

// Style defines the visual style of a divider.
type Style int

const (
	StyleSolid Style = iota // ───────────
	StyleDashed             // - - - - - -
	StyleDotted             // ·· ·· ·· ··
	StyleDouble             // ═══════════
)

// Orientation defines the direction of the divider.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the divider component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	label        string      // optional centered label
	dividerStyle Style       // line style
	orientation  Orientation // horizontal or vertical
	thickness    int         // line thickness (default 1)
	style        style.Style // visual style

	// === Layout Props ===
	fillWidth bool // whether to fill available width
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new Divider VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("divider"),
		dividerStyle: StyleSolid,
		orientation:  Horizontal,
		thickness:    1,
		fillWidth:    true,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (d *VNode) Key() string {
	return d.key
}

// SetKey sets the component key - returns VNode for chaining.
func (d *VNode) SetKey(key string) rtui.VNode {
	d.key = key
	return d
}

// Tag returns the tag name.
func (d *VNode) Tag() string {
	return "divider"
}

// Style returns the visual style.
func (d *VNode) Style() style.Style {
	return d.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (d *VNode) SetStyle(s style.Style) rtui.VNode {
	d.style = s
	return d
}

// Children returns child nodes (divider has no children).
func (d *VNode) Children() []rtui.VNode {
	return nil
}

// SetChildren is a no-op for divider - returns VNode for chaining.
func (d *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// Divider has no children
	return d
}

// GetLayer returns the rendering layer.
func (d *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (d *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return d
}

// Props returns the node properties.
func (d *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":          d.key,
		"label":        d.label,
		"dividerStyle": d.dividerStyle,
		"orientation":  d.orientation,
		"thickness":    d.thickness,
		"style":        d.style,
		"fillWidth":    d.fillWidth,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (d *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		d.key = v
	}
	if v, ok := p["label"].(string); ok {
		d.label = v
	}
	if v, ok := p["dividerStyle"].(Style); ok {
		d.dividerStyle = v
	}
	if v, ok := p["orientation"].(Orientation); ok {
		d.orientation = v
	}
	if v, ok := p["thickness"].(int); ok {
		d.thickness = v
	}
	if v, ok := p["style"].(style.Style); ok {
		d.style = v
	}
	if v, ok := p["fillWidth"].(bool); ok {
		d.fillWidth = v
	}
	return d
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new DividerInstance from this VNode description.
func (d *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		"key":          d.key,
		"label":        d.label,
		"dividerStyle": d.dividerStyle,
		"orientation":  d.orientation,
		"thickness":    d.thickness,
		"style":        d.style,
		"fillWidth":    d.fillWidth,
	}
	return NewInstance(props)
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetLabel sets the centered label text.
func (d *VNode) SetLabel(label string) *VNode {
	d.label = label
	return d
}

// SetDividerStyle sets the line style.
func (d *VNode) SetDividerStyle(s Style) *VNode {
	d.dividerStyle = s
	return d
}

// SetOrientation sets the divider direction.
func (d *VNode) SetOrientation(o Orientation) *VNode {
	d.orientation = o
	return d
}

// SetThickness sets the line thickness.
func (d *VNode) SetThickness(thickness int) *VNode {
	d.thickness = thickness
	return d
}

// SetStyleProps sets the visual style.
func (d *VNode) SetStyleProps(s style.Style) *VNode {
	d.style = s
	return d
}

// SetFillWidth sets whether to fill available width.
func (d *VNode) SetFillWidth(fill bool) *VNode {
	d.fillWidth = fill
	return d
}

// =============================================================================
// Convenience Builder Methods
// =============================================================================

// Horizontal sets orientation to horizontal.
func (d *VNode) Horizontal() *VNode {
	return d.SetOrientation(Horizontal)
}

// Vertical sets orientation to vertical.
func (d *VNode) Vertical() *VNode {
	return d.SetOrientation(Vertical)
}

// Solid sets style to solid line.
func (d *VNode) Solid() *VNode {
	return d.SetDividerStyle(StyleSolid)
}

// Dashed sets style to dashed line.
func (d *VNode) Dashed() *VNode {
	return d.SetDividerStyle(StyleDashed)
}

// Dotted sets style to dotted line.
func (d *VNode) Dotted() *VNode {
	return d.SetDividerStyle(StyleDotted)
}

// Double sets style to double line.
func (d *VNode) Double() *VNode {
	return d.SetDividerStyle(StyleDouble)
}

// =============================================================================
// Props Accessors (for Instance creation)
// =============================================================================

// Label returns the label text.
func (d *VNode) Label() string {
	return d.label
}

// DividerStyle returns the line style.
func (d *VNode) DividerStyle() Style {
	return d.dividerStyle
}

// Orientation returns the divider direction.
func (d *VNode) Orientation() Orientation {
	return d.orientation
}

// Thickness returns the line thickness.
func (d *VNode) Thickness() int {
	return d.thickness
}

// FillWidth returns whether to fill available width.
func (d *VNode) FillWidth() bool {
	return d.fillWidth
}
